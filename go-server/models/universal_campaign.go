package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UniversalCampaign struct {
	ID                primitive.ObjectID `bson:"_id" json:"_id"`
	ClientID          primitive.ObjectID `bson:"client_id" json:"client_id"`
	Name              string             `bson:"name" json:"name"`
	ShortCode         string             `bson:"short_code" json:"short_code"`
	CampaignShortCode string             `bson:"campaign_short_code" json:"campaign_short_code"`
	Share             *UniversalShare    `bson:"share,omitempty" json:"share,omitempty"`
	IsActive          bool               `bson:"is_active" json:"is_active"`
	AppType           string             `bson:"app_type" json:"app_type"`
	CreatedAt         time.Duration      `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Duration      `bson:"updated_at" json:"updated_at"`
}

type UniversalShare struct {
	Title   string `json:"title,omitempty" bson:"title,omitempty"`
	Image   string `json:"image,omitempty" bson:"image,omitempty"`
	OgImage string `json:"og_image,omitempty" bson:"og_image,omitempty"`
}
