package dao

import (
	"context"

	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/campaign-svc/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CategoryDao interface {
	GetCategoriesBySiteCodeDao(ctx context.Context, siteCode string, shortCodes []string, text string) (interface{}, error)
	CreateCategoryDao(ctx context.Context, createCategoryDto *dtos.CreateCategoryDto) (*models.Category, error)
	GetCategoryByCampaignShortCodeDao(campaignShortCode string) ([]models.Category, error)
	UpdateCategoryDao(ctx context.Context, sessionCtx *mongo.SessionContext, ID string, updateCategoryDto *dtos.UpdateCategoryDto) (*models.Category, error)
	ClientCategoriesDao(ID string) ([]models.Category, error)
	GetCategoryByNameDao(name string, ClientID string) (*models.Category, error)
	GetCategoryByID(ctx context.Context, clientObjID, categoryObjID primitive.ObjectID) (map[string]any, error)
}
