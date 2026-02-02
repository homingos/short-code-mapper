package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Category struct {
	ID         primitive.ObjectID `bson:"_id" json:"_id"`
	ClientID   primitive.ObjectID `bson:"client_id" json:"client_id"`
	Name       string             `bson:"name" json:"name" validate:"required"`
	SiteCode   string             `bson:"site_code" json:"site_code"`
	IsActive   bool               `bson:"is_active" json:"is_active"`
	BrandInfo  BrandInfo          `bson:"brand_info" json:"brand_info" validate:"required"`
	ShareMeta  CategoryShareMeta  `bson:"share_meta,omitempty" json:"share_meta,omitempty" validate:"required"`
	Categories []Categories       `bson:"categories" json:"categories" validate:"dive"`
	Status     string             `bson:"status" json:"status"`
	CreatedBy  *User              `bson:"created_by" json:"created_by"`
	CreatedAt  time.Duration      `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Duration      `bson:"updated_at" json:"updated_at"`
}

type BrandInfo struct {
	Title             string `bson:"title" json:"title" validate:"required"`
	Description       string `bson:"description" json:"description" validate:"required"`
	RedirectionURL    string `bson:"redirection_url" json:"redirection_url" validate:"required,url"`
	ForegroundColor   string `bson:"foreground_color" json:"foreground_color" validate:"required"`
	LogoURL           string `bson:"logo_url" json:"logo_url" validate:"required,url"`
	BackgroundColor   string `bson:"background_color" json:"background_color" validate:"required"`
	BackgroundOpacity int    `bson:"background_opacity" json:"background_opacity" validate:"required,min=0,max=100"`
	SelectionColor    string `bson:"selection_color" json:"selection_color" validate:"required"`
}

type CategoryShareMeta struct {
	Title       string `bson:"title,omitempty" json:"title,omitempty" validate:"required"`
	Description string `bson:"description,omitempty" json:"description,omitempty" validate:"required"`
	SiteURL     string `bson:"site_url,omitempty" json:"site_url,omitempty" validate:"required,url"`
	ImageURL    string `bson:"image_url,omitempty" json:"image_url,omitempty" validate:"required,url"`
}

type Categories struct {
	Name      string   `bson:"name" json:"name" validate:"required"`
	Campaigns []string `bson:"campaigns" json:"campaigns" validate:"required,min=1"`
}
