package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Template struct {
	ID               primitive.ObjectID `bson:"_id" json:"_id"`
	Name             string             `json:"name" bson:"name"`
	Class            int                `bson:"class" json:"class"`
	TrackType        string             `bson:"track_type" json:"track_type"`
	ThumbnailUrl     string             `json:"thumbnail_url" bson:"thumbnail_url"`
	HelpUrl          string             `json:"help_url,omitempty" bson:"help_url,omitempty"`
	Description      string             `json:"description" bson:"description"`
	EnableAlpha      string             `json:"enable_alpha" bson:"enable_alpha"`
	EnableBackground bool               `json:"enable_background" bson:"enable_background"`
	EnableMask       bool               `json:"enable_mask" bson:"enable_mask"`
	IsActive         bool               `bson:"is_active" json:"is_active"`
	HaveSegments     bool               `json:"have_segments,omitempty" bson:"have_segments,omitempty"`
	CreatedAt        time.Duration      `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Duration      `bson:"updated_at" json:"updated_at"`
	Tags             []string           `json:"tags,omitempty" bson:"tags"`
	FitType          *string            `json:"fit_type,omitempty" bson:"fit_type,omitempty"`
	CreditType       string             `json:"credit_type" bson:"credit_type"`
	ViewPriority     int                `json:"view_priority" bson:"view_priority"`
}
