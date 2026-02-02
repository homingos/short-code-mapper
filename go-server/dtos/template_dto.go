package dtos

import "github.com/homingos/campaign-svc/models"

type CreateTemplateDto struct {
	Name             string   `json:"name" validate:"required,min=1"`
	Class            *int     `json:"class" validate:"required,oneof=0 1 2 3"`
	TrackType        string   `json:"track_type" validate:"required,oneof=POSE GROUND CARD"`
	ThumbnailUrl     string   `json:"thumbnail_url" validate:"required,min=1"`
	HelpUrl          string   `json:"help_url,omitempty" validate:"required,min=1"`
	Description      string   `json:"description" bson:"description"`
	Tags             []string `json:"tags" bson:"tags" validate:"required,min=1"`
	EnableAlpha      string   `json:"enable_alpha" validate:"required,oneof=NOT_REQUIRED REQUIRED OPTIONAL"`
	EnableBackground bool     `json:"enable_background"`
	EnableMask       bool     `json:"enable_mask"`
	HaveSegments     bool     `json:"have_segments,omitempty"`
	ViewPriority     int      `json:"view_priority"`
	FitType          *string  `json:"fit_type,omitempty" bson:"fit_type,omitempty"`
	CreditType       string   `json:"credit_type" bson:"credit_type" validate:"required,oneof=PRISM PRISM_INTERACTIVE ALPHA ALPHA_INTERACTIVE FANDOM_AI FANDOM_VIDEO SPATIAL SPATIAL_ALPHA SPATIAL_INTERACTIVE THREE_D"`
}

type WebPrefrences struct {
	Photo PhotoPref `json:"photo"`
	Video VideoPref `json:"video"`
}

type VideoPref struct {
	MaxDuration   int `json:"max_duration"`
	MaxResolution int `json:"max_resolution"`
	VideoSize     int `json:"video_size"`
}

type PhotoPref struct {
	MaxResolution int `json:"max_resolution"`
	PhotoSize     int `json:"photo_size"`
}

type ClientRespData struct {
	WebPrefrences WebPrefrences `json:"web_prefrences"`
}

type CreditAllowance struct {
	ExpType      string `json:"exp_type"`
	TotalCredits int    `json:"total_credits"`
	Balance      int    `json:"balance"`
	Unlimited    bool   `json:"unlimited"`
}

type UpdateTemplateDto struct {
	Name             string `json:"name" validate:"omitempty,min=1"`
	VideoUrl         string `json:"video_url,omitempty" validate:"omitempty,min=1"`
	HelpUrl          string `json:"help_url,omitempty" validate:"omitempty,min=1"`
	TrackType        string `json:"track_type,omitempty" validate:"omitempty,oneof=POSE GROUND CARD"`
	Class            *int   `json:"class,omitempty" validate:"omitempty,oneof=0 1 2 3"`
	EnableAlpha      string `json:"enable_alpha,omitempty" validate:"omitempty,oneof=NOT_REQUIRED REQUIRED OPTIONAL"`
	EnableBackground bool   `json:"enable_background,omitempty"`
	EnableMask       bool   `json:"enable_mask,omitempty"`
	Steps            string `json:"steps,omitempty"`
}

type TemplateFiltersDto struct {
	Tags        []string `json:"tags"`
	CreditTypes []string `json:"credit_types"`
}

type AllTemplatesWithTrackType struct {
	Templates  []models.Template `json:"templates" bson:"templates"`
	TrackType  string            `json:"track_type" bson:"track_type"`
	CreditType string            `json:"credit_type" bson:"credit_type"`
	Balance    int               `json:"balance" bson:"balance"`
	Unlimited  bool              `json:"unlimited" bson:"unlimited"`
}
