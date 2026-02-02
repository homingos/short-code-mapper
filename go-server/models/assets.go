package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Asset struct {
	ID         primitive.ObjectID `bson:"_id" json:"_id"`
	Url        string             `bson:"url" json:"url"`
	MaskedUrl  *string            `bson:"masked_url,omitempty" json:"masked_url,omitempty"`
	Type       string             `bson:"type" json:"type"`
	ShortCode  string             `bson:"short_code" json:"short_code"`
	CampaignID primitive.ObjectID `bson:"campaign_id" json:"campaign_id"`
	CreatedAt  time.Duration      `bson:"created_at" json:"created_at"`
	ClientID   primitive.ObjectID `bson:"client_id" json:"client_id"`
	CreatedBy  *User              `bson:"created_by" json:"created_by"`
}
