package dao

import (
	"context"
	"time"

	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/campaign-svc/models"
	"github.com/homingos/campaign-svc/types/consts"
	"github.com/homingos/flam-go-common/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type TemplateDaoImpl struct {
	lgr *zap.SugaredLogger
	db  *mongo.Database
}

func createTemplateIndex(db *mongo.Database) {

}

func NewTemplateDao(lgr *zap.SugaredLogger, db *mongo.Database) *TemplateDaoImpl {
	createTemplateIndex(db)
	return &TemplateDaoImpl{lgr, db}
}

func (impl *TemplateDaoImpl) CreateTemplateDao(createTemplateDto dtos.CreateTemplateDto) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var template models.Template
	template.ID = primitive.NewObjectID()
	template.Name = createTemplateDto.Name
	template.Class = *createTemplateDto.Class
	template.TrackType = createTemplateDto.TrackType
	template.ThumbnailUrl = createTemplateDto.ThumbnailUrl
	template.HelpUrl = createTemplateDto.HelpUrl
	template.EnableAlpha = createTemplateDto.EnableAlpha
	template.EnableBackground = createTemplateDto.EnableBackground
	template.EnableMask = createTemplateDto.EnableMask
	template.HaveSegments = createTemplateDto.HaveSegments
	template.IsActive = true
	template.Description = createTemplateDto.Description
	template.Tags = createTemplateDto.Tags
	template.CreditType = createTemplateDto.CreditType
	template.ViewPriority = createTemplateDto.ViewPriority
	template.CreatedAt = time.Duration(time.Now().UnixMilli())
	template.UpdatedAt = time.Duration(time.Now().UnixMilli())

	if createTemplateDto.FitType != nil {
		template.FitType = createTemplateDto.FitType
	}

	coll := impl.db.Collection(consts.TemplateCollection)
	_, err := coll.InsertOne(ctx, template)
	if err != nil {
		return "", err
	}

	return "Template created successfully", nil
}

func (impl *TemplateDaoImpl) GetTemplateByIdDao(ID primitive.ObjectID) (*models.Template, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var template models.Template
	coll := impl.db.Collection(consts.TemplateCollection)
	err := coll.FindOne(ctx, bson.M{"_id": ID, "is_active": true}).Decode(&template)
	if err != nil {
		return nil, err
	}

	return &template, nil
}

func (impl *TemplateDaoImpl) GetAllTemplateDao(filters dtos.TemplateFiltersDto) ([]*dtos.AllTemplatesWithTrackType, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	coll := impl.db.Collection(consts.TemplateCollection)

	match := bson.M{
		"$match": bson.M{
			"is_active": true,
		},
	}

	if len(filters.CreditTypes) > 0 {
		match["$match"].(bson.M)["credit_type"] = bson.M{"$in": filters.CreditTypes}
	}

	if len(filters.Tags) > 0 {
		var regexFilters []bson.M
		for _, tag := range filters.Tags {
			regexFilters = append(regexFilters, bson.M{"tags": bson.M{"$regex": tag, "$options": "i"}})
		}

		match["$match"].(bson.M)["$or"] = regexFilters
	}

	sortBeforeGroup := bson.M{
		"$sort": bson.M{
			"credit_type":   1,
			"track_type":    1,
			"view_priority": 1,
		},
	}

	group := bson.M{
		"$group": bson.M{
			"_id": bson.M{
				"credit_type": "$credit_type",
				"track_type":  "$track_type",
			},
			"templates": bson.M{"$push": "$$ROOT"},
		},
	}

	project := bson.M{
		"$project": bson.M{
			"credit_type": "$_id.credit_type",
			"track_type":  "$_id.track_type",
			"templates":   1,
			"_id":         0,
		},
	}

	customSortField := bson.M{
		"$addFields": bson.M{
			"custom_sort": bson.M{
				"$switch": bson.M{
					"branches": []bson.M{
						{
							"case": bson.M{"$eq": []string{"$credit_type", consts.CreditTypePrism}},
							"then": 1,
						},
						{
							"case": bson.M{"$eq": []string{"$credit_type", consts.CreditTypePrismInteractive}},
							"then": 2,
						},
						{
							"case": bson.M{"$eq": []string{"$credit_type", consts.CreditTypeAlpha}},
							"then": 3,
						},
						{
							"case": bson.M{"$eq": []string{"$credit_type", consts.CreditTypeAlphaInteractive}},
							"then": 4,
						},
						{
							"case": bson.M{"$eq": []string{"$credit_type", consts.CreditTypeFandomAI}},
							"then": 5,
						},
						{
							"case": bson.M{"$eq": []string{"$credit_type", consts.CreditTypeFandomVideo}},
							"then": 6,
						},
						{
							"case": bson.M{"$eq": []string{"$credit_type", consts.CreditTypeSpatial}},
							"then": 7,
						},
						{
							"case": bson.M{"$eq": []string{"$credit_type", consts.CreditTypeAlphaSpatial}},
							"then": 8,
						},
						{
							"case": bson.M{"$eq": []string{"$credit_type", consts.CreditTypeSpatialInteractive}},
							"then": 9,
						},
						{
							"case": bson.M{"$eq": []string{"$credit_type", consts.CreditTypeThreeD}},
							"then": 10,
						},
					},
					"default": 11,
				},
			},
		},
	}

	custom_sort := bson.M{
		"$sort": bson.M{
			"custom_sort": 1,
		},
	}

	projectFinal := bson.M{
		"$project": bson.M{
			"custom_sort": 0,
		},
	}

	// Execute the pipeline
	cursor, err := coll.Aggregate(ctx, []bson.M{match, sortBeforeGroup, group, project, customSortField, custom_sort, projectFinal})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*dtos.AllTemplatesWithTrackType
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return []*dtos.AllTemplatesWithTrackType{}, nil
	}
	return results, nil
}

func (impl *TemplateDaoImpl) UpdateTemplateDao(updateTemplateDto dtos.UpdateTemplateDto) (string, error) {
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return "Nothing's been updated: Not developed yet", nil
}

func (impl *TemplateDaoImpl) GetAllRawTemplatesDao() ([]models.Template, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var template []models.Template
	coll := impl.db.Collection(consts.TemplateCollection)
	cursor, err := coll.Find(ctx, bson.M{"is_active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &template)
	if err != nil {
		return nil, err
	}

	return template, nil
}

func (impl *TemplateDaoImpl) GetTemplateCredit() ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	coll := impl.db.Collection(consts.TemplateCollection)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"is_active": true,
			},
		},
		{
			"$group": bson.M{
				"_id":   "$credit_type",
				"names": bson.M{"$push": "$name"},
			},
		},
		{
			"$project": bson.M{
				"credit_type": "$_id",
				"names":       1,
				"_id":         0,
			},
		},
	}
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (impl *TemplateDaoImpl) GetAlphaTemplateDao() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	coll := impl.db.Collection(consts.TemplateCollection)
	template := map[string]interface{}{}
	err := coll.FindOne(ctx, bson.M{"is_active": true, "class": 1, "track_type": "CARD"}).Decode(&template)
	if err != nil {
		return nil, errors.InternalServerError("No alpha template found")
	}
	return template, nil
}
