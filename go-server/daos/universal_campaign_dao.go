package dao

import (
	"context"

	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/campaign-svc/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UniversalCampaignDao interface {
	GetUniversalCampaignByShortCode(shortCode string) (*models.UniversalCampaign, error)
	CreateUniversalCampaignDao(ctx context.Context, universalCampaignReq *models.UniversalCampaign) (*models.UniversalCampaign, error)
	UpdateUniversalCampaignDao(ctx context.Context, universalCampaignID primitive.ObjectID, universalCampaignReq *dtos.UpdateUniversalCampaignRequestDto) (*models.UniversalCampaign, error)
	GetUniversalClientCampaignDao(ctx context.Context, clientID primitive.ObjectID) ([]models.UniversalCampaign, error)
}
