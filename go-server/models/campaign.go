package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FeatureFlags struct {
	EnableNextar            bool `json:"enable_nextar" bson:"enable_nextar"`
	EnableGeoVideos         bool `json:"enable_geo_videos" bson:"enable_geo_videos"`
	EnableVideoFullscreen   bool `json:"enable_video_fullscreen" bson:"enable_video_fullscreen"`
	EnableQRButton          bool `json:"enable_qr_button" bson:"enable_qr_button"`
	EnableRecording         bool `json:"enable_recording" bson:"enable_recording"`
	EnableIosStreaming      bool `json:"enable_ios_streaming" bson:"enable_ios_streaming"`
	EnableAndroidStreaming  bool `json:"enable_android_streaming" bson:"enable_android_streaming"`
	EnableAdaptiveStreaming bool `json:"enable_adaptive_streaming" bson:"enable_adaptive_streaming"`
	EnableScreenCapture     bool `json:"enable_screen_capture" bson:"enable_screen_capture"`
	EnableAirBoard          bool `json:"enable_airboard" bson:"enable_airboard"`
	EnableAutoPlay          bool `json:"enable_auto_play" bson:"enable_auto_play"`
}

type User struct {
	ID    string `json:"id" bson:"id"`
	Email string `json:"email" bson:"email"`
	Name  string `json:"name" bson:"name,omitempty"`
}

type Scan struct {
	ScanText           string `json:"scan_text" bson:"scan_text"`
	ImageUrl           string `json:"image_url" bson:"image_url"`
	CompressedImageUrl string `json:"compressed_image_url" bson:"compressed_image_url"`
}

type CopyRight struct {
	Show    bool   `json:"show" bson:"show"`
	Content string `json:"content" bson:"content"`
}

type Campaign struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	ClientId     primitive.ObjectID `bson:"client_id" json:"client_id"`
	GroupID      primitive.ObjectID `bson:"group_id,omitempty" json:"group_id,omitempty"`
	GroupName    string             `bson:"group_name,omitempty" json:"group_name,omitempty"`
	Name         string             `valid:"required" bson:"name" json:"name"`
	ShortCode    string             `bson:"short_code" json:"short_code"`
	TrackType    string             `valid:"required,in(POSE|GROUND|CARD)" bson:"track_type" json:"track_type"`
	AirTracking  *bool              `bson:"air_tracking,omitempty" json:"air_tracking,omitempty"`
	Scan         Scan               `bson:"scan" json:"scan"`
	Status       string             `valid:"in(CREATED|PROCESSING|PROCESSED)" bson:"status" json:"status"`
	IsActive     bool               `bson:"is_active" json:"is_active"`
	Publish      bool               `bson:"publish" json:"publish"`
	CopyRight    CopyRight          `bson:"copyright,omitempty" json:"copyright,omitempty"`
	FeatureFlags FeatureFlags       `bson:"feature_flags" json:"feature_flags"`
	CreatedBy    *User              `bson:"created_by,omitempty" json:"created_by,omitempty"`
	EditedBy     *User              `bson:"edited_by,omitempty" json:"edited_by,omitempty"`
	CreatedAt    time.Duration      `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Duration      `bson:"updated_at" json:"updated_at"`
	ExpiresAt    time.Duration      `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	GoLiveAt     time.Duration      `bson:"golive_at,omitempty" json:"golive_at,omitempty"`
	Share        *Share             `bson:"share,omitempty" json:"share,omitempty"`
	QrConfig     *QrConfig          `bson:"qr_config,omitempty" json:"qr_config,omitempty"`
	Source       string             `bson:"source,omitempty" json:"source,omitempty"`
}

type QrConfig struct {
	QrId        *string `bson:"qr_id,omitempty" json:"qr_id,omitempty"`
	QrBGColor   *string `bson:"qr_bg_color,omitempty" json:"qr_bg_color,omitempty"`
	QrTextColor *string `bson:"qr_text_color,omitempty" json:"qr_text_color,omitempty"`
}

type CampaignGroup struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	ClientID  primitive.ObjectID `bson:"client_id" json:"client_id"`
	Name      string             `bson:"name" json:"name"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
	CreatedAt time.Duration      `bson:"created_at" json:"created_at"`
	CreatedBy *User              `bson:"created_by,omitempty" json:"created_by,omitempty"`
	EditedBy  *User              `bson:"edited_by,omitempty" json:"edited_by,omitempty"`
	UpdatedAt time.Duration      `bson:"updated_at" json:"updated_at"`
}

type Share struct {
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
}
