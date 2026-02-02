package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/campaign-svc/models"
	"github.com/homingos/campaign-svc/types/consts"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type CategoryDaoImpl struct {
	lgr *zap.SugaredLogger
	db  *mongo.Database
}

func createCategoryIndexes(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := db.Collection(consts.CategoryCollection)
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "site_code", Value: 1}},
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
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := coll.Indexes().CreateMany(ctx, indexes, opts)
	if err != nil {
		fmt.Println(err)
	}
}

func NewCategoryDao(lgr *zap.SugaredLogger, db *mongo.Database) *CategoryDaoImpl {
	createCategoryIndexes(db)
	return &CategoryDaoImpl{lgr: lgr, db: db}
}

func (impl *CategoryDaoImpl) buildMongoUpdate(updateCategoryDto *dtos.UpdateCategoryDto) bson.M {
	updateFields := bson.M{}
	if updateCategoryDto.Name != "" {
		updateFields["name"] = updateCategoryDto.Name
	}
	if updateCategoryDto.BrandInfo != nil {
		updateFields["brand_info"] = updateCategoryDto.BrandInfo
	}
	if updateCategoryDto.ShareMeta != nil {
		updateFields["share_meta"] = updateCategoryDto.ShareMeta
	}
	if updateCategoryDto.Categories != nil {
		updateFields["categories"] = updateCategoryDto.Categories
	}
	if len(updateFields) > 0 {
		updateFields["updated_at"] = time.Duration(time.Now().UnixMilli())
	}
	return updateFields
}

func (impl *CategoryDaoImpl) GetCategoriesBySiteCodeDao(ctx context.Context, siteCode string, shortCodes []string, text string) (interface{}, error) {
	collection := impl.db.Collection(consts.CategoryCollection)

	// If text is provided, use a simpler pipeline that returns just shortcode strings
	if text != "" {
		if shortCodes == nil {
			shortCodes = []string{}
		}

		pipeline := []bson.M{
			{"$match": bson.M{"site_code": siteCode, "is_active": true}},
			// Unwind categories
			{"$unwind": "$categories"},
			// Unwind campaigns within each category
			{"$unwind": "$categories.campaigns"},
			// Filter to only shortcodes in the search results
			{"$match": bson.M{"categories.campaigns": bson.M{"$in": shortCodes}}},
			// Lookup campaign by short_code
			{"$lookup": bson.M{
				"from":         consts.CampaignCollection,
				"localField":   "categories.campaigns",
				"foreignField": "short_code",
				"as":           "campaign_doc",
				"pipeline": []bson.M{
					{"$match": bson.M{"is_active": true}},
				},
			}},
			{"$unwind": bson.M{"path": "$campaign_doc", "preserveNullAndEmptyArrays": false}},
			// Lookup experience and check status is processed
			{"$lookup": bson.M{
				"from":         "experiences",
				"localField":   "campaign_doc._id",
				"foreignField": "campaign_id",
				"as":           "experience",
				"pipeline": []bson.M{
					{"$match": bson.M{"is_active": true, "status": consts.Processed}},
					{"$limit": 1},
				},
			}},
			// Filter out campaigns without processed experiences
			{"$match": bson.M{"experience": bson.M{"$ne": bson.A{}}}},
			// Group back by category
			{"$group": bson.M{
				"_id": bson.M{
					"doc_id":        "$_id",
					"category_name": "$categories.name",
				},
				"client_id":  bson.M{"$first": "$client_id"},
				"site_code":  bson.M{"$first": "$site_code"},
				"brand_info": bson.M{"$first": "$brand_info"},
				"share_meta": bson.M{"$first": "$share_meta"},
				"campaigns":  bson.M{"$push": "$categories.campaigns"},
			}},
			// Group by document to collect categories
			{"$group": bson.M{
				"_id":        "$_id.doc_id",
				"client_id":  bson.M{"$first": "$client_id"},
				"site_code":  bson.M{"$first": "$site_code"},
				"brand_info": bson.M{"$first": "$brand_info"},
				"share_meta": bson.M{"$first": "$share_meta"},
				"categories": bson.M{
					"$push": bson.M{
						"name":      "$_id.category_name",
						"campaigns": "$campaigns",
					},
				},
			}},
		}

		cursor, err := collection.Aggregate(ctx, pipeline)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		var categories []dtos.CategorySearchResponseDto
		if err := cursor.All(ctx, &categories); err != nil {
			return nil, err
		}
		if len(categories) == 0 {
			return nil, nil
		}
		categories[0].OrderButtonText = "Order"
		return &categories[0], nil
	}

	var category dtos.CategoryAppResponseDto

	match := bson.M{
		"$match": bson.M{
			"site_code": siteCode,
			"is_active": true,
			// "categories.campaigns": bson.M{"$in": shortCodes},
		},
	}

	// Add index to categories array to preserve order
	addCategoryIndex := bson.M{
		"$addFields": bson.M{
			"categories": bson.M{
				"$map": bson.M{
					"input": bson.M{"$range": bson.A{0, bson.M{"$size": "$categories"}}},
					"as":    "idx",
					"in": bson.M{
						"$mergeObjects": bson.A{
							bson.M{"$arrayElemAt": bson.A{"$categories", "$$idx"}},
							bson.M{"_category_index": "$$idx"},
						},
					},
				},
			},
		},
	}

	// Unwind categories array
	unwindCategories := bson.M{
		"$unwind": "$categories",
	}

	// Add index to campaigns array within each category to preserve order
	// Convert campaigns array from strings to objects with index
	addCampaignIndex := bson.M{
		"$addFields": bson.M{
			"categories.campaigns": bson.M{
				"$map": bson.M{
					"input": bson.M{"$range": bson.A{0, bson.M{"$size": "$categories.campaigns"}}},
					"as":    "idx",
					"in": bson.M{
						"short_code":      bson.M{"$arrayElemAt": bson.A{"$categories.campaigns", "$$idx"}},
						"_campaign_index": "$$idx",
					},
				},
			},
		},
	}

	// Unwind campaigns array within each category
	unwindCampaigns := bson.M{
		"$unwind": "$categories.campaigns",
	}

	// Lookup campaign documents by short_code
	lookupCampaigns := bson.M{
		"$lookup": bson.M{
			"from":         consts.CampaignCollection,
			"localField":   "categories.campaigns.short_code",
			"foreignField": "short_code",
			"as":           "campaign_doc",
			"pipeline": []bson.M{
				{"$match": bson.M{"is_active": true}},
			},
		},
	}

	// Unwind campaign documents
	unwindCampaignDoc := bson.M{
		"$unwind": bson.M{
			"path":                       "$campaign_doc",
			"preserveNullAndEmptyArrays": true,
		},
	}

	// Filter out documents where campaign_doc is null or empty (inactive campaigns)
	filterInactiveCampaigns := bson.M{
		"$match": bson.M{
			"campaign_doc._id": bson.M{"$exists": true, "$ne": nil},
		},
	}

	// Lookup experiences for each campaign
	lookupExperiences := bson.M{
		"$lookup": bson.M{
			"from":         "experiences",
			"localField":   "campaign_doc._id",
			"foreignField": "campaign_id",
			"as":           "experiences_array",
			"pipeline": []bson.M{
				{"$match": bson.M{"is_active": true, "status": consts.Processed}},
				{"$limit": 1},
				{
					"$project": bson.M{
						"_id":            1,
						"aspect_ratio":   1,
						"canvas":         1,
						"images":         bson.M{"$arrayToObject": "$images"},
						"playback_scale": 1,
						"videos":         bson.M{"$arrayToObject": "$videos"},
						"variant":        1,
						"catalogue_details": bson.M{
							"name":     "$catalogue_details.name",
							"currency": "$catalogue_details.currency",
							"price":    "$catalogue_details.price",
						},
					},
				},
			},
		},
	}

	// Project campaign with experiences as single object
	projectCampaign := bson.M{
		"$project": bson.M{
			"_id":            1,
			"client_id":      1,
			"site_code":      1,
			"brand_info":     1,
			"share_meta":     1,
			"category_name":  "$categories.name",
			"category_index": "$categories._category_index",
			"campaign_index": "$categories.campaigns._campaign_index",
			"campaign": bson.M{
				"_id":         bson.M{"$toString": "$campaign_doc._id"},
				"icon_url":    "$campaign_doc.icon_url",
				"name":        "$campaign_doc.name",
				"short_code":  "$campaign_doc.short_code",
				"experiences": bson.M{"$arrayElemAt": bson.A{"$experiences_array", 0}},
			},
		},
	}

	// Filter out campaigns without active experiences
	filterCampaignsWithoutExperiences := bson.M{
		"$match": bson.M{
			"$and": []bson.M{
				{"campaign.experiences": bson.M{"$exists": true}},
				{"campaign.experiences": bson.M{"$ne": nil}},
				{"campaign.experiences._id": bson.M{"$exists": true}},
				{"campaign.experiences._id": bson.M{"$ne": nil}},
				{"campaign.experiences._id": bson.M{"$ne": ""}},
			},
		},
	}

	// Group by category name to collect campaigns, preserving order
	groupByCategory := bson.M{
		"$group": bson.M{
			"_id": bson.M{
				"doc_id":         "$_id",
				"category_name":  "$category_name",
				"category_index": "$category_index",
			},
			"client_id":  bson.M{"$first": "$client_id"},
			"site_code":  bson.M{"$first": "$site_code"},
			"brand_info": bson.M{"$first": "$brand_info"},
			"share_meta": bson.M{"$first": "$share_meta"},
			"campaigns": bson.M{
				"$push": bson.M{
					"campaign":       "$campaign",
					"campaign_index": "$campaign_index",
				},
			},
		},
	}

	// Sort campaigns within each category by their original index
	sortCampaignsInCategory := bson.M{
		"$addFields": bson.M{
			"campaigns": bson.M{
				"$map": bson.M{
					"input": bson.M{"$sortArray": bson.M{"input": "$campaigns", "sortBy": bson.M{"campaign_index": 1}}},
					"as":    "camp",
					"in":    "$$camp.campaign",
				},
			},
		},
	}

	// Group by document to collect categories, preserving order
	groupByDoc := bson.M{
		"$group": bson.M{
			"_id":        "$_id.doc_id",
			"client_id":  bson.M{"$first": "$client_id"},
			"site_code":  bson.M{"$first": "$site_code"},
			"brand_info": bson.M{"$first": "$brand_info"},
			"share_meta": bson.M{"$first": "$share_meta"},
			"categories": bson.M{
				"$push": bson.M{
					"name":           "$_id.category_name",
					"campaigns":      "$campaigns",
					"category_index": "$_id.category_index",
				},
			},
		},
	}

	// Sort categories by their original index
	sortCategories := bson.M{
		"$addFields": bson.M{
			"categories": bson.M{
				"$map": bson.M{
					"input": bson.M{"$sortArray": bson.M{"input": "$categories", "sortBy": bson.M{"category_index": 1}}},
					"as":    "cat",
					"in": bson.M{
						"name":      "$$cat.name",
						"campaigns": "$$cat.campaigns",
					},
				},
			},
		},
	}

	// Final projection to match DTO structure
	finalProject := bson.M{
		"$project": bson.M{
			"_id":        "$_id",
			"client_id":  1,
			"site_code":  1,
			"brand_info": 1,
			"share_meta": 1,
			"categories": bson.M{
				"$map": bson.M{
					"input": "$categories",
					"as":    "cat",
					"in": bson.M{
						"name":      "$$cat.name",
						"campaigns": "$$cat.campaigns",
					},
				},
			},
		},
	}

	// Build pipeline for full campaign data (when text is not provided)
	pipeline := []bson.M{
		match,
		addCategoryIndex,
		unwindCategories,
		addCampaignIndex,
		unwindCampaigns,
		lookupCampaigns,
		unwindCampaignDoc,
		filterInactiveCampaigns,
		lookupExperiences,
		projectCampaign,
		filterCampaignsWithoutExperiences,
		groupByCategory,
		sortCampaignsInCategory,
		groupByDoc,
		sortCategories,
		finalProject,
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var categories []dtos.CategoryAppResponseDto
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	if len(categories) == 0 {
		return nil, nil
	}

	category = categories[0]
	category.OrderButtonText = "Order"
	return &category, nil
}

func (impl *CategoryDaoImpl) CreateCategoryDao(ctx context.Context, createCategoryDto *dtos.CreateCategoryDto) (*models.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	clientId, err := primitive.ObjectIDFromHex(createCategoryDto.ClientID)
	if err != nil {
		return nil, err
	}
	var category models.Category
	category.ID = primitive.NewObjectID()
	category.ClientID = clientId
	category.SiteCode = createCategoryDto.SiteCode
	category.BrandInfo = createCategoryDto.BrandInfo
	category.ShareMeta = createCategoryDto.ShareMeta
	category.Categories = createCategoryDto.Categories
	category.IsActive = true
	category.Name = createCategoryDto.Name
	// category.Status = consts.Processing
	category.CreatedBy = createCategoryDto.CreatedBy
	category.CreatedAt = time.Duration(time.Now().UnixMilli())
	category.UpdatedAt = time.Duration(time.Now().UnixMilli())

	coll := impl.db.Collection(consts.CategoryCollection)
	_, err = coll.InsertOne(ctx, category)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, fmt.Errorf("Name already taken")
		}
		return nil, err
	}
	return &category, nil
}

func (impl *CategoryDaoImpl) GetCategoryByCampaignShortCodeDao(campaignShortCode string) ([]models.Category, error) {
	coll := impl.db.Collection(consts.CategoryCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{"categories.campaigns": bson.M{"$in": []string{campaignShortCode}}, "is_active": true}
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var categories []models.Category
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

func (impl *CategoryDaoImpl) UpdateCategoryDao(ctx context.Context, sessionCtx *mongo.SessionContext, ID string, updateCategoryDto *dtos.UpdateCategoryDto) (*models.Category, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	categoryID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}

	updateMap := impl.buildMongoUpdate(updateCategoryDto)
	if len(updateMap) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	ClientObjId, err := primitive.ObjectIDFromHex(updateCategoryDto.ClientID)
	if err != nil {
		return nil, fmt.Errorf("invalid client ID")
	}

	coll := impl.db.Collection(consts.CategoryCollection)
	filter := bson.M{"_id": categoryID, "is_active": true, "client_id": ClientObjId}
	update := bson.M{"$set": updateMap}
	if sessionCtx != nil {
		ctx = *sessionCtx
	}
	result := coll.FindOneAndUpdate(ctx, filter, update)
	if err = result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("category not found")
		}
		return nil, err
	}
	var updatedCategory models.Category
	if err := result.Decode(&updatedCategory); err != nil {
		return nil, err
	}
	return &updatedCategory, nil

}

func (impl *CategoryDaoImpl) ClientCategoriesDao(ID string) ([]models.Category, error) {
	clientObjID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		return nil, err
	}
	coll := impl.db.Collection(consts.CategoryCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{"client_id": clientObjID, "is_active": true}
	opt := &options.FindOptions{
		Sort: bson.M{"updated_at": -1},
	}
	cursor, err := coll.Find(ctx, filter, opt)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	categories := []models.Category{}
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

func (impl *CategoryDaoImpl) GetCategoryByNameDao(name string, ClientID string) (*models.Category, error) {
	coll := impl.db.Collection(consts.CategoryCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ClientObjId, err := primitive.ObjectIDFromHex(ClientID)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"name": name, "client_id": ClientObjId, "is_active": true}
	result := coll.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	var category models.Category
	if err := result.Decode(&category); err != nil {
		return nil, err
	}
	return &category, nil
}

func (impl *CategoryDaoImpl) GetCategoryByID(ctx context.Context, clientObjID, categoryObjID primitive.ObjectID) (map[string]any, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"client_id": clientObjID,
				"is_active": true,
				"_id":       categoryObjID,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "campaigns",
				"localField":   "categories.campaigns",
				"foreignField": "short_code",
				"as":           "camps",
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"is_active": true,
						},
					},
					{
						"$project": bson.M{
							"name":       1,
							"short_code": 1,
							"group_id":   1,
							"group_name": 1,
							"status":     1,
						},
					},
				},
			},
		},
		// getting experience status
		{
			"$lookup": bson.M{
				"from":         "experiences",
				"localField":   "camps._id",
				"foreignField": "campaign_id",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"campaign_id": 1,
							"status":      1,
						},
					},
				},
				"as": "exps",
			},
		},
		// setting experience status to campaigns
		{
			"$set": bson.M{
				"camps": bson.M{
					"$map": bson.M{
						"input": "$camps",
						"as":    "cp",
						"in": bson.M{
							"$mergeObjects": bson.A{
								"$$cp",
								bson.M{
									"status": bson.M{
										"$let": bson.M{
											"vars": bson.M{
												"expp": bson.M{
													"$first": bson.M{
														"$filter": bson.M{
															"input": "$exps",
															"as":    "ex",
															"cond": bson.M{
																"$eq": bson.A{
																	"$$ex.campaign_id",
																	"$$cp._id",
																},
															},
														},
													},
												},
											},
											"in": "$$expp.status",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"$set": bson.M{
				"categories": bson.M{
					"$map": bson.M{
						"input": "$categories",
						"as":    "ctg",
						"in": bson.M{
							"name": "$$ctg.name",
							"campaigns": bson.M{
								"$map": bson.M{
									"input": "$$ctg.campaigns",
									"as":    "shortCode",
									"in": bson.M{
										"$first": bson.M{
											"$filter": bson.M{
												"input": "$camps",
												"as":    "cp",
												"cond": bson.M{
													"$eq": bson.A{
														"$$cp.short_code",
														"$$shortCode",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"$project": bson.M{
				"camps": 0,
				"exps":  0,
			},
		},
	}

	coll := impl.db.Collection(consts.CategoryCollection)

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []map[string]any
	if err := cursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	if len(categories) == 0 {
		return nil, nil
	}

	return categories[0], nil
}
