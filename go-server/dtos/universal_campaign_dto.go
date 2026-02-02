package dtos

import (
	"github.com/homingos/campaign-svc/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UniversalCampaignResponseDto struct {
	ID                primitive.ObjectID     `bson:"_id" json:"id"`
	Name              string                 `bson:"name" json:"name"`
	ShortCode         string                 `bson:"short_code" json:"short_code"`
	CampaignShortCode string                 `bson:"campaign_short_code" json:"campaign_short_code"`
	AppType           string                 `bson:"app_type" json:"app_type"`
	Campaign          interface{}            `bson:"campaign" json:"campaign"`
	Share             *models.UniversalShare `bson:"share,omitempty" json:"share,omitempty"`
}

type UniversalCampaignRequestDto struct {
	Name              string                 `json:"name" validate:"required,min=1"`
	CampaignShortCode string                 `json:"campaign_short_code" validate:"required,min=1"`
	AppType           string                 `json:"app_type" validate:"required,min=1"`
	ClientID          string                 `json:"client_id"`
	Share             *models.UniversalShare `json:"share,omitempty"`
}

type UpdateUniversalCampaignRequestDto struct {
	ClientID          primitive.ObjectID     `json:"client_id" validate:"omitempty"`
	Name              string                 `json:"name" validate:"omitempty"`
	CampaignShortCode string                 `json:"campaign_short_code" validate:"omitempty"`
	AppType           string                 `json:"app_type" validate:"omitempty"`
	IsActive          *bool                  `json:"is_active" validate:"omitempty"`
	Share             *models.UniversalShare `json:"share,omitempty"`
}
