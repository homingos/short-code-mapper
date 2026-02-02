package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/google/martian/v3/log"

	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/campaign-svc/models"
	"github.com/homingos/campaign-svc/types/consts"
	"github.com/homingos/campaign-svc/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type CampaignDaoImpl struct {
	lgr *zap.SugaredLogger
	db  *mongo.Database
}

func createCampaignIndexes(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := db.Collection(consts.CampaignCollection)
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "short_code", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "client_id", Value: 1},
				{Key: "name", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.D{
					{Key: "is_active", Value: true},
				}),
		},
		{
			Keys: bson.D{
				{Key: "client_id", Value: 1},
				{Key: "is_active", Value: 1},
				{Key: "name", Value: 1},
			},
			Options: options.Index(),
		}, {
			Keys: bson.D{
				{Key: "client_id", Value: 1},
				{Key: "is_active", Value: 1},
				{Key: "short_code", Value: 1},
			},
			Options: options.Index(),
		}, {
			Keys: bson.D{
				{Key: "client_id", Value: 1},
				{Key: "is_active", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index(),
		},
		{
			Keys: bson.D{
				{Key: "is_active", Value: 1},
				{Key: "name", Value: 1},
			},
			Options: options.Index(),
		},
		{
			Keys: bson.D{
				{Key: "is_active", Value: 1},
				{Key: "short_code", Value: 1},
			},
			Options: options.Index(),
		}}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := coll.Indexes().CreateMany(ctx, indexes, opts)
	if err != nil {
		fmt.Println(err)
	}
}

func createCampaignGroupIndexes(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := db.Collection("campaign_groups")
	indexModel := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "client_id", Value: 1},
				{Key: "name", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.D{
					{Key: "is_active", Value: true},
				}),
		},
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	coll.Indexes().CreateMany(ctx, indexModel, opts)
}

func NewCampaignDao(lgr *zap.SugaredLogger, db *mongo.Database) *CampaignDaoImpl {
	createCampaignIndexes(db)
	createCampaignGroupIndexes(db)
	return &CampaignDaoImpl{lgr, db}
}

func (impl *CampaignDaoImpl) GetClientCampaignDao(ClientId string, filterDto *dtos.CampaignFilterDto) (interface{}, error) {
	client_id, err := primitive.ObjectIDFromHex(ClientId)
	if err != nil {
		return nil, err
	}
	nameFilter := filterDto.Name
	if filterDto.Offset == 0 {
		filterDto.Offset = filterDto.Page * filterDto.PageSize
	}
	regexPattern := bson.M{"$regex": primitive.Regex{Pattern: nameFilter, Options: "i"}}
	match := bson.M{
		"$match": bson.M{
			"is_active": true,
			"client_id": client_id,
			"$or": []bson.M{
				{"name": regexPattern},
				{"short_code": regexPattern},
			},
		},
	}
	if filterDto.Publish != nil {
		match["$match"].(bson.M)["publish"] = filterDto.Publish
	}

	sort := bson.M{
		"$sort": bson.M{
			"updated_at": -1,
		},
	}
	skip := bson.M{
		"$skip": filterDto.Offset,
	}
	limit := bson.M{
		"$limit": filterDto.PageSize * 1,
	}
	lookup := bson.M{
		"$lookup": bson.M{
			"from":         "experiences",
			"localField":   "_id",
			"foreignField": "campaign_id",
			"as":           "experiences",
			"pipeline": []bson.M{
				{
					"$match": bson.M{"is_active": true},
				},
				{
					"$sort": bson.M{"created_at": -1},
				},
				{
					"$project": bson.M{
						"_id":              1,
						"status":           1,
						"images":           bson.M{"$arrayToObject": "$images"},
						"template_details": 1,
					},
				},
				{
					"$project": bson.M{
						"_id":                     1,
						"status":                  1,
						"images.original":         1,
						"images.color_compressed": 1,
						"images.spawn":            1,
						"template_details":        1,
					},
				},
			},
		},
	}
	setStatus := bson.M{
		"$set": bson.M{
			"status": bson.M{
				"$cond": bson.M{
					"if": bson.M{
						"$gt": bson.A{
							bson.M{"$size": bson.M{
								"$filter": bson.M{
									"input": "$experiences",
									"as":    "exp",
									"cond":  bson.M{"$eq": bson.A{"$$exp.status", "FAILED"}},
								},
							}},
							0,
						},
					},
					"then": "FAILED",
					"else": bson.M{
						"$cond": bson.M{
							"if": bson.M{
								"$gt": bson.A{
									bson.M{"$size": bson.M{
										"$filter": bson.M{
											"input": "$experiences",
											"as":    "exp",
											"cond":  bson.M{"$eq": bson.A{"$$exp.status", "DRAFT"}},
										},
									}},
									0,
								},
							},
							"then": "DRAFT",
							"else": bson.M{
								"$cond": bson.M{
									"if":   bson.M{"$eq": bson.A{bson.M{"$size": "$experiences"}, 0}},
									"then": "CREATED",
									"else": bson.M{
										"$cond": bson.M{
											"if": bson.M{
												"$eq": bson.A{
													bson.M{"$size": "$experiences"},
													bson.M{"$size": bson.M{"$filter": bson.M{
														"input": "$experiences",
														"as":    "exp",
														"cond":  bson.M{"$eq": bson.A{"$$exp.status", "PROCESSED"}},
													}}},
												},
											},
											"then": "PROCESSED",
											"else": "PROCESSING",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	project := bson.M{
		"$project": bson.M{
			"id":            "$_id",
			"_id":           0,
			"name":          1,
			"client_id":     1,
			"short_code":    1,
			"is_active":     1,
			"publish":       1,
			"status":        1,
			"air_tracking":  1,
			"track_type":    1,
			"created_at":    1,
			"updated_at":    1,
			"created_by":    1,
			"feature_flags": 1,
			"copyright":     1,
			"experiences":   1,
		},
	}
	facet := bson.M{
		"$facet": bson.M{
			"count": []bson.M{
				{"$count": "totalDocs"},
			},
			"data": []bson.M{
				skip,
				limit,
				lookup,
				setStatus,
				project,
			},
		},
	}

	projectFinal := bson.M{
		"$project": bson.M{
			"data": "$data",
			"total_document": bson.M{"$ifNull": []interface{}{
				bson.M{"$arrayElemAt": bson.A{"$count.totalDocs", 0}},
				0,
			}},
		},
	}
	pipelines := []bson.M{match, sort, facet, projectFinal}
	data := []map[string]interface{}{}
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := coll.Aggregate(ctx, pipelines)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &data); err != nil {
		return nil, err
	}
	return data[0], nil
}

func (impl *CampaignDaoImpl) GetClientCampaignV2Dao(clientID string, filterDto *dtos.CampaignFilterDto) (interface{}, error) {
	client_id, err := primitive.ObjectIDFromHex(clientID)
	if err != nil {
		return nil, err
	}

	match := bson.M{
		"$match": bson.M{
			"client_id": client_id,
			"is_active": true,
		},
	}

	if filterDto.Draft != nil && *filterDto.Draft {
		match["$match"].(bson.M)["publish"] = false
		match["$match"].(bson.M)["$or"] = []bson.M{
			{
				"expires_at": bson.M{
					"$exists": false,
				},
			},
			{
				"expires_at": bson.M{
					"$exists": true,
					"$gt":     time.Now().UnixMilli(),
				},
			},
		}
	}

	if filterDto.Publish != nil {
		match["$match"].(bson.M)["publish"] = filterDto.Publish
		match["$match"].(bson.M)["expires_at"] = bson.M{"$gt": time.Now().UnixMilli()}
	}

	if filterDto.Publish != nil && *filterDto.Publish && filterDto.Draft != nil && *filterDto.Draft {
		delete(match["$match"].(bson.M), "publish")
		delete(match["$match"].(bson.M), "$or")
		// expiry filter remains
	}

	nameFilter := bson.M{"$match": bson.M{}}
	if filterDto.Name != "" {
		regexPattern := bson.M{"$regex": primitive.Regex{Pattern: filterDto.Name, Options: "i"}}
		nameFilter["$match"] = bson.M{
			"$or": []bson.M{
				{"name": regexPattern},
				{"short_code": regexPattern},
				{"group_name": regexPattern},
			},
		}
	}

	lookup := bson.M{
		"$lookup": bson.M{
			"from":         "experiences",
			"localField":   "_id",
			"foreignField": "campaign_id",
			"as":           "experiences",
			"pipeline": []bson.M{
				{"$match": bson.M{"is_active": true}},
				{"$sort": bson.M{"created_at": -1}},
				{"$project": bson.M{
					"_id":               1,
					"status":            1,
					"images":            bson.M{"$arrayToObject": "$images"},
					"template_details":  1,
					"template_category": 1,
				}},
				{"$project": bson.M{
					"_id":                     1,
					"status":                  1,
					"images.original":         1,
					"images.color_compressed": 1,
					"images.std_compressed":   1,
					"images.spawn":            1,
					"template_details":        1,
					"template_category":       1,
				}},
			},
		},
	}

	unwindExp := bson.M{
		"$unwind": bson.M{
			"path":                       "$experiences",
			"preserveNullAndEmptyArrays": true,
		},
	}

	setStatus := bson.M{
		"$addFields": bson.M{
			"status": bson.M{
				"$ifNull": bson.A{"$experiences.status", consts.Created},
			},
		},
	}
	grouping := bson.M{
		"$group": bson.M{
			"_id":       "$group_name",
			"group_id":  bson.M{"$first": "$group_id"},
			"campaigns": bson.M{"$push": "$$ROOT"},
		},
	}

	project := bson.M{
		"$project": bson.M{
			"_id":        0,
			"group_id":   1,
			"campaigns":  1,
			"group_name": "$_id",
		},
	}

	sortCampaignsInGroup := bson.M{
		"$set": bson.M{
			"campaigns": bson.M{
				"$function": bson.M{
					"body": `function(campaigns) {
						return campaigns.sort((a, b) => b.updated_at - a.updated_at);
					}`,
					"args": bson.A{"$campaigns"},
					"lang": "js",
				},
			},
		},
	}

	setMaxUpdatedAt := bson.M{
		"$set": bson.M{
			"max_updated_at": bson.M{
				"$max": bson.M{
					"$map": bson.M{
						"input": "$campaigns",
						"as":    "c",
						"in":    "$$c.updated_at",
					},
				},
			},
		},
	}

	sortGroups := bson.M{
		"$sort": bson.M{
			"max_updated_at": -1,
		},
	}

	finalProject := bson.M{
		"$project": bson.M{
			"max_updated_at": 0,
		},
	}

	pipelines := []bson.M{
		match,
		nameFilter,
		lookup,
		unwindExp,
		setStatus,
		grouping,
		project,
		sortCampaignsInGroup,
		setMaxUpdatedAt,
		sortGroups,
		finalProject,
	}

	data := []map[string]interface{}{}
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := coll.Aggregate(ctx, pipelines)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (impl *CampaignDaoImpl) GetAllCampaignDao(GetCampaignDto *dtos.GetCampaignDto) (interface{}, error) {
	shortCode := GetCampaignDto.ShortCode
	regexPattern := bson.M{"$regex": primitive.Regex{Pattern: shortCode, Options: "i"}}
	match := bson.M{
		"$match": bson.M{
			"is_active": true,
			"$or": []bson.M{
				{"name": regexPattern},
				{"short_code": regexPattern},
			},
		},
	}
	sort := bson.M{
		"$sort": bson.M{
			"created_at": -1,
		},
	}
	skip := bson.M{
		"$skip": GetCampaignDto.Page * GetCampaignDto.PageSize,
	}
	limit := bson.M{
		"$limit": GetCampaignDto.PageSize * 1,
	}
	lookup := bson.M{
		"$lookup": bson.M{
			"from":         "experiences",
			"localField":   "_id",
			"foreignField": "campaign_id",
			"as":           "experiences",
			"pipeline": []bson.M{
				{
					"$match": bson.M{"is_active": true},
				},
				{
					"$project": bson.M{
						"_id":    1,
						"status": 1,
					},
				},
				{
					"$project": bson.M{
						"_id":    1,
						"status": 1,
					},
				},
			},
		},
	}
	setStatus := bson.M{
		"$set": bson.M{
			"status": bson.M{
				"$cond": bson.M{
					"if": bson.M{
						"$gt": bson.A{
							bson.M{"$size": bson.M{
								"$filter": bson.M{
									"input": "$experiences",
									"as":    "exp",
									"cond":  bson.M{"$eq": bson.A{"$$exp.status", "FAILED"}},
								},
							}},
							0,
						},
					},
					"then": "FAILED",
					"else": bson.M{
						"$cond": bson.M{
							"if": bson.M{
								"$gt": bson.A{
									bson.M{"$size": bson.M{
										"$filter": bson.M{
											"input": "$experiences",
											"as":    "exp",
											"cond":  bson.M{"$eq": bson.A{"$$exp.status", "DRAFT"}},
										},
									}},
									0,
								},
							},
							"then": "DRAFT",
							"else": bson.M{
								"$cond": bson.M{
									"if":   bson.M{"$eq": bson.A{bson.M{"$size": "$experiences"}, 0}},
									"then": "CREATED",
									"else": bson.M{
										"$cond": bson.M{
											"if": bson.M{
												"$eq": bson.A{
													bson.M{"$size": "$experiences"},
													bson.M{"$size": bson.M{"$filter": bson.M{
														"input": "$experiences",
														"as":    "exp",
														"cond":  bson.M{"$eq": bson.A{"$$exp.status", "PROCESSED"}},
													}}},
												},
											},
											"then": "PROCESSED",
											"else": "PROCESSING",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	project := bson.M{
		"$project": bson.M{
			"id":           "$_id",
			"_id":          0,
			"name":         1,
			"client_id":    1,
			"short_code":   1,
			"publish":      1,
			"status":       1,
			"track_type":   1,
			"air_tracking": 1,
			"created_at":   1,
			"updated_at":   1,
			"created_by":   1,
			"edited_by":    1,
			"total_experiences": bson.M{
				"$size": "$experiences",
			},
		},
	}
	facet := bson.M{
		"$facet": bson.M{
			"count": []bson.M{
				{"$count": "totalDocs"},
			},
			"data": []bson.M{
				skip,
				limit,
				lookup,
				setStatus,
				project,
			},
		},
	}

	projectFinal := bson.M{
		"$project": bson.M{
			"campaigns": "$data",
			"total_document": bson.M{"$ifNull": []interface{}{
				bson.M{"$arrayElemAt": bson.A{"$count.totalDocs", 0}},
				0,
			}},
		},
	}
	pipelines := []bson.M{match, sort, facet, projectFinal}
	data := []map[string]interface{}{}
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := coll.Aggregate(ctx, pipelines)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &data); err != nil {
		return nil, err
	}
	return data[0], nil
}

func (impl *CampaignDaoImpl) GetCampaignGroupsDao(clientID string) ([]*models.CampaignGroup, error) {
	clientObjID, err := primitive.ObjectIDFromHex(clientID)
	if err != nil {
		return nil, fmt.Errorf("invalid client ID: %v", err)
	}

	filter := bson.M{"client_id": clientObjID, "is_active": true, "name": bson.M{"$ne": consts.CampaignGroupDemo}}
	var campaignGroups []*models.CampaignGroup

	coll := impl.db.Collection(consts.CampaignGroupCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "updated_at", Value: -1}})

	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &campaignGroups); err != nil {
		return nil, err
	}

	return campaignGroups, nil
}

func (impl *CampaignDaoImpl) CreateCampaignDao(ctx context.Context, sessionCtx *mongo.SessionContext, campaignReqDto *dtos.CampaignRequestDto) (*models.Campaign, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var FeatureFlags models.FeatureFlags
	if campaignReqDto.FeatureFlags == nil {
		FeatureFlags = models.FeatureFlags{
			EnableNextar:            true,
			EnableGeoVideos:         false,
			EnableVideoFullscreen:   false,
			EnableQRButton:          false,
			EnableRecording:         false,
			EnableIosStreaming:      true,
			EnableAndroidStreaming:  true,
			EnableAdaptiveStreaming: true,
			EnableScreenCapture:     true,
		}
	} else {
		FeatureFlags = *campaignReqDto.FeatureFlags
	}
	clientId, err := primitive.ObjectIDFromHex(campaignReqDto.ClientId)
	if err != nil {
		return nil, err
	}
	var campaign models.Campaign
	campaign.ID = primitive.NewObjectID()
	campaign.ClientId = clientId
	campaign.Name = campaignReqDto.Name
	campaign.ShortCode = campaignReqDto.ShortCode
	campaign.TrackType = campaignReqDto.TrackType
	campaign.Status = consts.Created
	if campaignReqDto.TrackType == "GROUND" {
		campaign.Scan.ScanText = consts.DefaultGroundScanText
	} else {
		campaign.Scan.ScanText = consts.DefaultCardScanText
	}
	campaign.IsActive = true
	campaign.FeatureFlags = FeatureFlags
	if campaignReqDto.AirTracking != nil {
		campaign.AirTracking = campaignReqDto.AirTracking
		if *campaign.AirTracking == true {
			campaign.FeatureFlags.EnableAirBoard = true
			campaign.FeatureFlags.EnableAutoPlay = true
			campaign.FeatureFlags.EnableVideoFullscreen = true
		}
	}
	campaign.Publish = campaignReqDto.Publish
	campaign.CopyRight.Show = true
	campaign.CopyRight.Content = consts.CopyRightContent
	campaign.CreatedBy = campaignReqDto.CreatedBy
	campaign.CreatedAt = time.Duration(time.Now().UnixMilli())
	campaign.UpdatedAt = time.Duration(time.Now().UnixMilli())
	campaign.Share = campaignReqDto.Share
	coll := impl.db.Collection(consts.CampaignCollection)

	if sessionCtx != nil {
		ctx = *sessionCtx
	}

	_, err = coll.InsertOne(ctx, campaign)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, fmt.Errorf("shoot. That one’s already taken")
		}
		return nil, err
	}
	return &campaign, nil
}

func (impl *CampaignDaoImpl) CreateCampaignV2Dao(ctx context.Context, sessionCtx *mongo.SessionContext, campaignReqDto *dtos.CampaignRequestV2Dto) (*models.Campaign, error) {
	var FeatureFlags models.FeatureFlags
	if campaignReqDto.FeatureFlags == nil {
		FeatureFlags = models.FeatureFlags{
			EnableNextar:            true,
			EnableGeoVideos:         false,
			EnableVideoFullscreen:   false,
			EnableQRButton:          false,
			EnableRecording:         false,
			EnableIosStreaming:      true,
			EnableAndroidStreaming:  true,
			EnableAdaptiveStreaming: true,
			EnableScreenCapture:     true,
			EnableAirBoard:          false,
			EnableAutoPlay:          false,
		}
	} else {
		FeatureFlags = *campaignReqDto.FeatureFlags
	}
	clientId, err := primitive.ObjectIDFromHex(campaignReqDto.ClientId)
	if err != nil {
		return nil, err
	}
	var campaign models.Campaign
	campaign.ID = primitive.NewObjectID()
	campaign.ClientId = clientId
	campaign.Name = campaignReqDto.Name
	campaign.ShortCode = campaignReqDto.ShortCode
	campaign.TrackType = campaignReqDto.TrackType
	campaign.Status = consts.Created
	if campaignReqDto.TrackType == "GROUND" {
		campaign.Scan.ScanText = consts.DefaultGroundScanText
	} else {
		campaign.Scan.ScanText = consts.DefaultCardScanText
	}
	campaign.IsActive = true
	campaign.FeatureFlags = FeatureFlags
	if campaignReqDto.AirTracking != nil {
		campaign.AirTracking = campaignReqDto.AirTracking
		if *campaign.AirTracking == true {
			campaign.FeatureFlags.EnableAirBoard = true
			campaign.FeatureFlags.EnableAutoPlay = true
			campaign.FeatureFlags.EnableVideoFullscreen = true
		}
	}
	campaign.Publish = campaignReqDto.Publish
	campaign.CopyRight.Show = true
	campaign.CopyRight.Content = consts.CopyRightContent
	campaign.CreatedBy = campaignReqDto.CreatedBy
	campaign.CreatedAt = time.Duration(time.Now().UnixMilli())
	campaign.UpdatedAt = time.Duration(time.Now().UnixMilli())
	campaign.Share = campaignReqDto.Share
	campaign.GroupID = campaignReqDto.CampaignGroupId
	campaign.GroupName = campaignReqDto.GroupName
	if campaignReqDto.ExpiresAt != 0 {
		campaign.ExpiresAt = time.Duration(campaignReqDto.ExpiresAt)
	}
	if campaignReqDto.GoLiveAt != 0 {
		campaign.GoLiveAt = campaignReqDto.GoLiveAt
	}

	coll := impl.db.Collection(consts.CampaignCollection)
	if sessionCtx != nil {
		ctx = *sessionCtx
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err = coll.InsertOne(ctx, campaign)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Errorf("create_campaign_repo_error: ", err.Error())
			return nil, fmt.Errorf("duplicate_key: That one’s already taken")
		}
		return nil, err
	}
	return &campaign, nil
}

func (impl *CampaignDaoImpl) CreateCampaignGroupDao(ctx context.Context, sessionCtx *mongo.SessionContext, dto *dtos.CampaignGroupCreateDto) (*models.CampaignGroup, error) {
	clientObjID, err := primitive.ObjectIDFromHex(dto.ClientID)
	if err != nil {
		return nil, err
	}

	now := time.Duration(time.Now().UnixMilli())

	// Set filter: only match existing active groups with same name and client ID
	filter := bson.M{
		"client_id": clientObjID,
		"name":      dto.GroupName,
		"is_active": true,
	}

	// Update fields
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":        primitive.NewObjectID(),
			"client_id":  clientObjID,
			"name":       dto.GroupName,
			"is_active":  true,
			"created_at": now,
			"created_by": dto.CreatedBy,
		},
		"$set": bson.M{
			"updated_at": now,
		},
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	coll := impl.db.Collection(consts.CampaignGroupCollection)
	if sessionCtx != nil {
		ctx = *sessionCtx
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var campaignGroup models.CampaignGroup
	err = coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&campaignGroup)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, fmt.Errorf("group name '%s' is already taken", dto.GroupName)
		}
		return nil, err
	}

	return &campaignGroup, nil
}

func (impl *CampaignDaoImpl) CreateBulkCampaignDao(sessionCtx *mongo.SessionContext, Campaigns []*models.Campaign) (string, error) {
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	document := make([]interface{}, len(Campaigns))
	for i, campaign := range Campaigns {
		document[i] = campaign
	}
	if sessionCtx != nil {
		ctx = *sessionCtx
	}
	_, err := coll.InsertMany(ctx, document)
	if err != nil {
		return "", err
	}
	return "Campaign Created Successfully", nil
}

func (impl *CampaignDaoImpl) GetCampaignDao(ID string, clientId string) (*models.Campaign, error) {
	objID, err := primitive.ObjectIDFromHex(ID)
	query := bson.M{"_id": objID}
	if err != nil {
		query = bson.M{"short_code": ID}
	}
	if clientId != "" {
		ClientId, err := primitive.ObjectIDFromHex(clientId)
		if err != nil {
			return nil, err
		}

		query["client_id"] = ClientId
	}
	var campaign models.Campaign
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result := coll.FindOne(ctx, query)
	err = result.Decode(&campaign)
	return &campaign, err
}

func (impl *CampaignDaoImpl) GetCampaignDaoByShortCode(ID string) (*models.Campaign, error) {
	objID, err := primitive.ObjectIDFromHex(ID)
	query := bson.M{"_id": objID}
	if err != nil {
		query = bson.M{"short_code": ID}
	}
	var campaign models.Campaign
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result := coll.FindOne(ctx, query)
	err = result.Decode(&campaign)
	return &campaign, err
}

func (impl *CampaignDaoImpl) GetClientDemoCampaignsCount(clientID string) (int64, error) {
	clientObjID, err := primitive.ObjectIDFromHex(clientID)
	if err != nil {
		return 0, fmt.Errorf("invalid client ID: %v", err)
	}
	query := bson.M{"client_id": clientObjID, "group_name": consts.CampaignGroupDemo}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := impl.db.Collection(consts.CampaignCollection)
	count, err := coll.CountDocuments(ctx, query)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (impl *CampaignDaoImpl) UpdateCampaignDao(objID primitive.ObjectID, campaignUpdateDto *dtos.CampaignUpdateDto) (*models.Campaign, error) {
	filter := bson.M{"_id": objID}
	updateMap := utils.StructToMap(campaignUpdateDto)
	delete(updateMap, "network_info")
	delete(updateMap, "expiry_date")
	updates := bson.M{"$set": updateMap}
	updateMap["updated_at"] = time.Now().UnixMilli()
	if campaignUpdateDto.Scan != nil {
		if campaignUpdateDto.Scan.ImageUrl != "" {
			updateMap["scan.image_url"] = campaignUpdateDto.Scan.ImageUrl
			updateMap["scan.compressed_image_url"] = ""
		}
		if campaignUpdateDto.Scan.ScanText != "" {
			updateMap["scan.scan_text"] = campaignUpdateDto.Scan.ScanText
		}
		delete(updateMap, "scan")
	}
	if campaignUpdateDto.ShowCopyRight != nil {
		updateMap["copyright.show"] = campaignUpdateDto.ShowCopyRight
		updateMap["copyright.content"] = consts.CopyRightContent
		delete(updateMap, "show_copyright")
	}
	if campaignUpdateDto.ExpiresAt != 0 {
		updateMap["expires_at"] = campaignUpdateDto.ExpiresAt
	}

	if campaignUpdateDto.GoliveAt != 0 {
		updateMap["golive_at"] = campaignUpdateDto.GoliveAt
	}
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	var campaign models.Campaign
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result := coll.FindOneAndUpdate(ctx, filter, updates, &opt)
	if err := result.Err(); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, fmt.Errorf("shoot. That one’s already taken")
		}
		return nil, err
	}
	if err := result.Decode(&campaign); err != nil {
		return nil, err
	}
	return &campaign, nil
}

// Hard Delete, should not be accessed ever
func (impl *CampaignDaoImpl) DeleteCampaignDao(objID primitive.ObjectID) (*models.Campaign, error) {
	filter := bson.M{"_id": objID}
	opt := options.FindOneAndDeleteOptions{}
	var campaign models.Campaign
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result := coll.FindOneAndDelete(ctx, filter, &opt)
	if err := result.Err(); err != nil {
		return nil, err
	}
	if err := result.Decode(&campaign); err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (impl *CampaignDaoImpl) GetCampaignExperiencesDao(ctx context.Context, ID string, optionalParams ...string) (interface{}, error) {
	var ClientId string
	if len(optionalParams) > 0 {
		ClientId = optionalParams[0]
	}
	fmt.Print(ClientId, "clientid")
	objID, err := primitive.ObjectIDFromHex(ID)
	filter := bson.M{
		"_id":       objID,
		"is_active": true,
	}
	if err != nil {
		filter = bson.M{
			"short_code": ID,
			"is_active":  true,
		}
	}
	if ClientId != "" {
		clientId, err := primitive.ObjectIDFromHex(ClientId)
		if err != nil {
			return nil, err
		}
		filter["client_id"] = clientId
	}
	match := bson.M{"$match": filter}
	lookup := bson.M{
		"$lookup": bson.M{
			"from":         "experiences",
			"localField":   "_id",
			"foreignField": "campaign_id",
			"as":           "experiences",
			"pipeline": []bson.M{{
				"$match": bson.M{
					"is_active": true,
				}},
			},
		},
	}
	unwind := bson.M{
		"$unwind": bson.M{
			"path":                       "$experiences",
			"preserveNullAndEmptyArrays": true,
		},
	}
	sort := bson.M{
		"$sort": bson.M{
			"experiences.created_at": -1,
		},
	}
	project := bson.M{
		"$project": bson.M{
			"_id":                           1,
			"id":                            "$_id",
			"name":                          1,
			"short_code":                    1,
			"feature_flags":                 1,
			"group_name":                    1,
			"expires_at":                    1,
			"golive_at":                     1,
			"is_active":                     1,
			"status":                        1,
			"publish":                       1,
			"track_type":                    1,
			"air_tracking":                  1,
			"flam_logo":                     consts.LOGO,
			"logo_width":                    bson.M{"$literal": consts.LogoWidth},
			"copyright":                     1,
			"scan":                          1,
			"created_at":                    1,
			"updated_at":                    1,
			"experiences.id":                "$experiences._id",
			"experiences.name":              1,
			"experiences.is_active":         1,
			"experiences.canvas":            1,
			"experiences.created_at":        1,
			"experiences.scan_text":         1,
			"experiences.status":            1,
			"experiences.aspect_ratio":      1,
			"experiences.ui_elements":       1,
			"experiences.variant":           1,
			"experiences.updated_at":        1,
			"experiences.campaign_id":       1,
			"experiences.playback_scale":    1,
			"experiences.engagment_options": 1,
			"experiences.reward_enabled":    1,
			"experiences.share_meta":        1,
			"experiences.overlay":           1,
			"experiences.scene":             1,
			"experiences.workflow_error":    1,
			"experiences.workflow_id":       1,
			"experiences.total_task":        1,
			"experiences.credit_deduct":     1,
			"experiences.mask":              1,
			"experiences.rewards":           1,
			"experiences.template_details":  1,
			"experiences.qr_panel":          1,
			"experiences.images":            bson.M{"$ifNull": bson.A{bson.M{"$arrayToObject": "$experiences.images"}, "$$REMOVE"}},
			"experiences.videos":            bson.M{"$ifNull": bson.A{bson.M{"$arrayToObject": "$experiences.videos"}, "$$REMOVE"}},
			"experiences.audios":            bson.M{"$ifNull": bson.A{bson.M{"$arrayToObject": "$experiences.audios"}, "$$REMOVE"}},
			"experiences.3d_assets":         bson.M{"$ifNull": bson.A{bson.M{"$arrayToObject": "$experiences.3d_assets"}, "$$REMOVE"}},
			"experiences.template_category": 1,
			"experiences.video_generation":  1,
		},
	}

	addStatus := bson.M{
		"$addFields": bson.M{
			"status": bson.M{
				"$ifNull": bson.A{"$experiences.status", consts.Created},
			},
			"experiences": bson.M{
				"$cond": bson.A{
					bson.M{"$eq": bson.A{"$experiences", bson.M{}}},
					nil,
					bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$experiences", nil}},
						nil,
						"$experiences",
					},
					},
				},
			},
		},
	}
	pipelines := []bson.M{match, lookup, unwind, sort, project, addStatus}
	var data []map[string]interface{}
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cursor, err := coll.Aggregate(ctx, pipelines)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &data); err != nil {
		return nil, err
	}
	if len(data) > 0 {
		return data[0], nil
	}
	return nil, nil
}

func (impl *CampaignDaoImpl) GetAppExperiencesDao(ctx context.Context, shortcode string) (interface{}, error) {

	filter := bson.M{
		"short_code": shortcode,
		"is_active":  true,
	}

	match := bson.M{"$match": filter}
	lookup := bson.M{
		"$lookup": bson.M{
			"from":         "experiences",
			"localField":   "_id",
			"foreignField": "campaign_id",
			"as":           "experiences",
			"pipeline": []bson.M{{
				"$match": bson.M{
					"is_active": true,
				}},
			},
		},
	}
	unwind := bson.M{
		"$unwind": bson.M{
			"path":                       "$experiences",
			"preserveNullAndEmptyArrays": true,
		},
	}

	project := bson.M{
		"$project": bson.M{
			"_id":                           1,
			"name":                          1,
			"short_code":                    1,
			"feature_flags":                 1,
			"group_name":                    1,
			"expires_at":                    1,
			"golive_at":                     1,
			"is_active":                     1,
			"status":                        1,
			"publish":                       1,
			"track_type":                    1,
			"flam_logo":                     consts.LOGO,
			"logo_width":                    bson.M{"$literal": consts.LogoWidth},
			"copyright":                     1,
			"scan":                          1,
			"created_at":                    1,
			"updated_at":                    1,
			"experiences.id":                "$experiences._id",
			"experiences.name":              1,
			"experiences.is_active":         1,
			"experiences.canvas":            1,
			"experiences.created_at":        1,
			"experiences.scan_text":         1,
			"experiences.status":            1,
			"experiences.aspect_ratio":      1,
			"experiences.ui_elements":       1,
			"experiences.variant":           1,
			"experiences.updated_at":        1,
			"experiences.campaign_id":       1,
			"experiences.playback_scale":    1,
			"experiences.engagment_options": 1,
			"experiences.reward_enabled":    1,
			"experiences.share_meta":        1,
			"experiences.overlay":           1,
			"experiences.scene":             1,
			"experiences.workflow_error":    1,
			"experiences.workflow_id":       1,
			"experiences.total_task":        1,
			"experiences.credit_deduct":     1,
			"experiences.mask":              1,
			"experiences.rewards":           1,
			"experiences.template_details":  1,
			"experiences.qr_panel":          1,
			"experiences.images":            bson.M{"$ifNull": bson.A{bson.M{"$arrayToObject": "$experiences.images"}, "$$REMOVE"}},
			"experiences.videos":            bson.M{"$ifNull": bson.A{bson.M{"$arrayToObject": "$experiences.videos"}, "$$REMOVE"}},
			"experiences.audios":            bson.M{"$ifNull": bson.A{bson.M{"$arrayToObject": "$experiences.audios"}, "$$REMOVE"}},
			"experiences.3d_assets":         bson.M{"$ifNull": bson.A{bson.M{"$arrayToObject": "$experiences.3d_assets"}, "$$REMOVE"}},
			"experiences.template_category": 1,
		},
	}
	group := bson.M{
		"$group": bson.M{
			"_id":           "$_id",
			"id":            bson.M{"$first": "$_id"},
			"copyright":     bson.M{"$first": "$copyright"},
			"name":          bson.M{"$first": "$name"},
			"short_code":    bson.M{"$first": "$short_code"},
			"feature_flags": bson.M{"$first": "$feature_flags"},
			"track_type":    bson.M{"$first": "$track_type"},
			"scan":          bson.M{"$first": "$scan"},
			"created_at":    bson.M{"$first": "$created_at"},
			"updated_at":    bson.M{"$first": "$updated_at"},
			"group_name":    bson.M{"$first": "$group_name"},
			"expires_at":    bson.M{"$first": "$expires_at"},
			"golive_at":     bson.M{"$first": "$golive_at"},
			"status":        bson.M{"$first": "$status"},
			"publish":       bson.M{"$first": "$publish"},
			"flam_logo":     bson.M{"$first": "$flam_logo"},
			"logo_width":    bson.M{"$first": "$logo_width"},
			"experiences": bson.M{"$push": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$ne": bson.A{"$experiences", bson.M{}}},
					"then": "$experiences",
					"else": bson.M{},
				},
			}},
		},
	}
	addFields := bson.M{
		"$addFields": bson.M{
			"experiences": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$eq": bson.A{"$experiences", bson.A{}}},
					"then": bson.A{},
					"else": bson.M{"$filter": bson.M{
						"input": "$experiences",
						"as":    "experience",
						"cond":  bson.M{"$ne": bson.A{"$$experience", bson.M{}}},
					}},
				},
			},
		},
	}
	addStatus := bson.M{
		"$addFields": bson.M{
			"status": bson.M{
				"$cond": bson.M{
					"if": bson.M{
						"$gt": bson.A{bson.M{"$size": "$experiences"}, 0},
					},
					"then": bson.M{"$arrayElemAt": bson.A{"$experiences.status", 0}},
					"else": "$status",
				},
			},
		},
	}
	pipelines := []bson.M{match, lookup, unwind, project, group, addFields, addStatus}
	var data []map[string]interface{}
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cursor, err := coll.Aggregate(ctx, pipelines)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &data); err != nil {
		return nil, err
	}
	if len(data) > 0 {
		return data[0], nil
	}
	return nil, nil
}

func (impl *CampaignDaoImpl) PostbackCampaignDao(shortCode string, updateMap map[string]interface{}) (*models.Campaign, error) {
	var campaign models.Campaign
	filter := bson.M{"short_code": shortCode}
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if updateMap["scan_compressed_image_url"] != nil {
		after := options.After
		opt := options.FindOneAndUpdateOptions{
			ReturnDocument: &after,
		}
		updateCampaign := map[string]interface{}{}
		setCampaign := bson.M{"$set": updateCampaign}
		updateCampaign["updated_at"] = time.Now().UnixMilli()
		updateCampaign["scan.compressed_image_url"] = updateMap["scan_compressed_image_url"]

		result := coll.FindOneAndUpdate(ctx, filter, setCampaign, &opt)
		if err := result.Err(); err != nil {
			return nil, err
		}
		if err := result.Decode(&campaign); err != nil {
			return nil, err
		}
	}
	return &campaign, nil
}

func (impl *CampaignDaoImpl) CheckProjectExists(ID string) (*models.Project, error) {
	var project *models.Project
	projectID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": projectID}
	coll := impl.db.Collection(consts.ProjectCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result := coll.FindOne(ctx, filter)
	err = result.Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("project not found for this ID")
		}
		return nil, err
	}
	return project, nil
}

func (impl *CampaignDaoImpl) FdbCampaignDao(expID, fdbUrl string) error {
	// Match stage to find active campaign with shortcode
	expObjID, err := primitive.ObjectIDFromHex(expID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": expObjID, "is_active": true}

	addToSetOperations := bson.M{}
	addToSetOperations["images"] = models.Image{K: "fdb", V: fdbUrl}
	update := bson.M{
		"$addToSet": addToSetOperations,
	}

	coll := impl.db.Collection(consts.ExperienceCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update experience: %v", err)
	}

	return nil
}

func (impl *CampaignDaoImpl) GetExperincesCampaignDao(campaignID string) ([]models.Experience, error) {
	campaignObjId, err := primitive.ObjectIDFromHex(campaignID)
	if err != nil {
		return nil, fmt.Errorf("invalid campaign ID: %v", err)
	}
	query := bson.M{"campaign_id": campaignObjId, "is_active": true}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var exps []models.Experience
	coll := impl.db.Collection(consts.ExperienceCollection)
	cursor, err := coll.Find(ctx, query)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &exps); err != nil {
		return nil, err
	}
	return exps, nil
}

func (impl *CampaignDaoImpl) UpdateExperienceCreditDeduct(ExpId primitive.ObjectID, CreditAllowanceID string) error {
	CreditAllowanceObID, err := primitive.ObjectIDFromHex(CreditAllowanceID)
	if err != nil {
		return fmt.Errorf("invalid credit allowance ID: %v", err)
	}
	filter := bson.M{"_id": ExpId}
	coll := impl.db.Collection(consts.ExperienceCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	updateExp := map[string]interface{}{}
	set := bson.M{"$set": updateExp}
	updateExp["updated_at"] = time.Now().UnixMilli()
	updateExp["credit_deduct"] = true
	updateExp["credit_allowance_id"] = CreditAllowanceObID
	result := coll.FindOneAndUpdate(ctx, filter, set)
	if err := result.Err(); err != nil {
		impl.lgr.Infof("error consuming credit: %v", err)
		return err
	}
	return nil
}

func (impl *CampaignDaoImpl) CheckCampaingsByName(Names []string, ClientID primitive.ObjectID) ([]string, error) {
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"name": bson.M{"$in": Names}, "is_active": true, "client_id": ClientID}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var existingNames []string
	for cursor.Next(ctx) {
		var doc struct {
			Name string `bson:"name"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		existingNames = append(existingNames, doc.Name)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return existingNames, nil
}

func (impl *CampaignDaoImpl) GetBulkCampaignDao(clientID, shortCode string, page, pageSize int, publish *bool) (interface{}, error) {
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("shortCode", shortCode)
	clientObjID, err := primitive.ObjectIDFromHex(clientID)
	if err != nil {
		return nil, err
	}
	match := bson.M{
		"$match": bson.M{
			"source":    shortCode,
			"is_active": true,
			"client_id": clientObjID,
		},
	}
	sort := bson.M{
		"$sort": bson.M{
			"created_at": -1,
		},
	}

	if publish != nil {
		match["$match"].(bson.M)["publish"] = *publish
	}

	experienceMatch := bson.M{
		"$match": bson.M{
			"is_active": true,
		},
	}

	lookup := bson.M{
		"$lookup": bson.M{
			"from":         "experiences",
			"localField":   "_id",
			"foreignField": "campaign_id",
			"as":           "experiences",
			"pipeline": []bson.M{
				experienceMatch,
				{
					"$project": bson.M{
						"_id":         1,
						"status":      1,
						"ui_elements": 1,
					},
				},
			},
		},
	}

	filterExperiences := bson.M{
		"$match": bson.M{
			"experiences": bson.M{
				"$ne": bson.A{},
			},
		},
	}

	setStatus := bson.M{
		"$set": bson.M{
			"status": bson.M{
				"$cond": bson.M{
					"if": bson.M{
						"$gt": bson.A{
							bson.M{"$size": bson.M{
								"$filter": bson.M{
									"input": "$experiences",
									"as":    "exp",
									"cond":  bson.M{"$eq": bson.A{"$$exp.status", "FAILED"}},
								},
							}},
							0,
						},
					},
					"then": "FAILED",
					"else": bson.M{
						"$cond": bson.M{
							"if": bson.M{
								"$gt": bson.A{
									bson.M{"$size": bson.M{
										"$filter": bson.M{
											"input": "$experiences",
											"as":    "exp",
											"cond":  bson.M{"$eq": bson.A{"$$exp.status", "DRAFT"}},
										},
									}},
									0,
								},
							},
							"then": "DRAFT",
							"else": bson.M{
								"$cond": bson.M{
									"if":   bson.M{"$eq": bson.A{bson.M{"$size": "$experiences"}, 0}},
									"then": "CREATED",
									"else": bson.M{
										"$cond": bson.M{
											"if": bson.M{
												"$eq": bson.A{
													bson.M{"$size": "$experiences"},
													bson.M{"$size": bson.M{"$filter": bson.M{
														"input": "$experiences",
														"as":    "exp",
														"cond":  bson.M{"$eq": bson.A{"$$exp.status", "PROCESSED"}},
													}}},
												},
											},
											"then": "PROCESSED",
											"else": "PROCESSING",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	project := bson.M{
		"$project": bson.M{
			"_id":        1,
			"name":       1,
			"client_id":  1,
			"short_code": 1,
			"publish":    1,
			"status":     1,
			"track_type": 1,
			"qr_config":  1,
			"source":     1,
			"group_name": 1,
			"ui_elements": bson.M{
				"$cond": bson.M{
					"if": bson.M{"$gt": bson.A{bson.M{"$size": "$experiences"}, 0}},
					"then": bson.M{
						"$arrayElemAt": bson.A{"$experiences.ui_elements", 0},
					},
					"else": nil,
				},
			},
		},
	}

	offset := page * pageSize
	skip := bson.M{
		"$skip": offset,
	}
	limit := bson.M{
		"$limit": pageSize,
	}

	facet := bson.M{
		"$facet": bson.M{
			"count": []bson.M{
				{"$count": "totalDocs"},
			},
			"data": []bson.M{
				skip,
				limit,
				lookup,
				filterExperiences,
				setStatus,
				project,
			},
		},
	}

	projectFinal := bson.M{
		"$project": bson.M{
			"campaigns": "$data",
			"total_document": bson.M{"$ifNull": []interface{}{
				bson.M{"$arrayElemAt": bson.A{"$count.totalDocs", 0}},
				0,
			}},
		},
	}

	pipelines := []bson.M{match, sort, facet, projectFinal}
	data := []map[string]interface{}{}
	cursor, err := coll.Aggregate(ctx, pipelines)
	defer cursor.Close(ctx)

	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (impl *CampaignDaoImpl) GetCampaignBySourceCodeDao(shortcode string) ([]models.Campaign, error) {
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"source": shortcode, "is_active": true}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var campaigns []models.Campaign
	if err := cursor.All(ctx, &campaigns); err != nil {
		return nil, err
	}
	return campaigns, nil
}

func (impl *CampaignDaoImpl) GetCampaignWithExperienceBySourceCodeDao(shortcode string, pending bool) (*dtos.CampaignsExperiences, error) {
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	match := bson.M{
		"$match": bson.M{
			"source":    shortcode,
			"is_active": true,
		},
	}
	experienceMatch := bson.M{
		"$match": bson.M{
			"is_active": true,
		},
	}

	if pending {
		experienceMatch["$match"].(bson.M)["status"] = bson.M{"$ne": consts.Processed}
	}

	lookup := bson.M{
		"$lookup": bson.M{
			"from":         "experiences",
			"localField":   "_id",
			"foreignField": "campaign_id",
			"as":           "experiences",
			"pipeline": []bson.M{
				experienceMatch,
			},
		},
	}

	// Add this after lookup to flatten all experiences across campaigns
	unwind := bson.M{
		"$unwind": bson.M{
			"path":                       "$experiences",
			"preserveNullAndEmptyArrays": true,
		},
	}

	// Group everything together
	groupAll := bson.M{
		"$group": bson.M{
			"_id": nil,
			"campaigns": bson.M{
				"$addToSet": bson.M{
					"_id":         "$_id",
					"name":        "$name",
					"client_id":   "$client_id",
					"short_code":  "$short_code",
					"qr_config":   "$qr_config",
					"experiences": "$experiences",
				},
			},
			"allExperiences": bson.M{"$push": "$experiences"},
		},
	}

	// Calculate overall status
	addStatus := bson.M{
		"$addFields": bson.M{
			"status": bson.M{
				"$cond": bson.M{
					"if": bson.M{
						"$in": bson.A{"PROCESSING", "$allExperiences.status"},
					},
					"then": "PROCESSING",
					"else": bson.M{
						"$cond": bson.M{
							"if": bson.M{
								"$in": bson.A{"FAILED", "$allExperiences.status"},
							},
							"then": "FAILED",
							"else": "PROCESSED",
						},
					},
				},
			},
		},
	}

	project := bson.M{
		"$project": bson.M{
			"_id":       1,
			"status":    1,
			"campaigns": 1,
		},
	}

	pipelines := []bson.M{match, lookup, unwind, groupAll, addStatus, project}
	var data []dtos.CampaignsExperiences
	cursor, err := coll.Aggregate(ctx, pipelines)
	defer cursor.Close(ctx)

	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &data); err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return &data[0], nil
}

func (impl *CampaignDaoImpl) GetClientCampaignsDAO(clientID primitive.ObjectID) ([]dtos.ClientCampaignsInfo, error) {
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{"client_id": clientID, "is_active": true}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var clientCampaigns []dtos.ClientCampaignsInfo
	if err := cursor.All(ctx, &clientCampaigns); err != nil {
		return nil, err
	}

	if len(clientCampaigns) == 0 {
		return []dtos.ClientCampaignsInfo{}, nil
	}

	return clientCampaigns, nil
}

func (impl *CampaignDaoImpl) GetCampaignByShortCodesDao(shortCodes []string, clientId string) ([]models.Campaign, error) {
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientObjId, err := primitive.ObjectIDFromHex(clientId)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"short_code": bson.M{"$in": shortCodes}, "client_id": clientObjId, "is_active": true}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var campaigns []models.Campaign
	if err := cursor.All(ctx, &campaigns); err != nil {
		return nil, err
	}
	return campaigns, nil
}
func (impl *CampaignDaoImpl) GetShortcodesByMilvusRefID(IDs []string) ([]string, error) {
	coll := impl.db.Collection(consts.CampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// clientObjId, err := primitive.ObjectIDFromHex(clientId)
	// if err != nil {
	// 	return nil, err
	// }

	fmt.Println("IDS in repo: ", IDs)
	filter := bson.M{"milvus_ref_id": bson.M{"$in": IDs}, "is_active": true}
	pipelines := []bson.M{
		{
			"$match": filter,
		},
		{
			"$group": bson.M{
				"_id": nil,
				"short_codes": bson.M{
					"$push": "$short_code",
				},
			},
		},
		{
			"$project": bson.M{
				"_id": 0,
			},
		},
	}
	var data []map[string][]string

	cursor, err := coll.Aggregate(ctx, pipelines)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &data); err != nil {
		return nil, err
	}
	if len(data) > 0 {
		return data[0]["short_codes"], nil
	}

	return nil, nil
}
