package dtos

import (
	"time"

	"github.com/homingos/campaign-svc/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CampaignRequestDto struct {
	Name          string               `json:"name" validate:"required,min=1"`
	ClientId      string               `json:"client_id"`
	ShortCode     string               `json:"short_code"`
	TrackType     string               `json:"track_type" validate:"required,oneof=GROUND CARD POSE"`
	AirTracking   *bool                `json:"air_tracking,omitempty"`
	FeatureFlags  *models.FeatureFlags `json:"feature_flags,omitempty"`
	CreatedBy     *models.User         `json:"user,omitempty"`
	ShowCopyRight bool                 `json:"show_copyright,omitempty"`
	Publish       bool                 `json:"publish,omitempty"`
	ExpiresAt     int64                `json:"expires_at,omitempty" bson:"expires_at"`
	Share         *models.Share        `json:"share,omitempty"`
	NetworkInfo   *NetworkInfo         `json:"network_info,omitempty"`
}

type CampaignRequestV2Dto struct {
	CampaignRequestDto
	GroupName       string             `json:"group_name,omitempty" validate:"required"`
	CampaignGroupId primitive.ObjectID `json:"campaign_group_id,omitempty" bson:"campaign_group_id,omitempty"`
	GoLiveAt        time.Duration      `bson:"golive_at,omitempty" json:"golive_at,omitempty"`
}

type CampaignGroupCreateDto struct {
	ClientID  string       `json:"client_id"`
	GroupName string       `json:"group_name" validate:"required,min=1"`
	CreatedBy *models.User `json:"created_by,omitempty"`
}

type CampaignBulkCreateDto struct {
	GroupName      string  `json:"group_name" validate:"required,min=1"`
	Name           string  `json:"name" validate:"required,min=1"`
	QrID           *string `json:"qr_id,omitempty"`
	CTA            *bool   `json:"cta" validate:"required"`
	Title          string  `json:"title,omitempty"`
	RedirectionUrl string  `json:"redirection_url,omitempty"`
	BodyText       *string `json:"body_text,omitempty"`
	QrBGColor      *string `json:"qr_bg_color,omitempty"`
	QrTextColor    *string `json:"qr_text_color,omitempty"`
}

type BulkCampaignRequestDto struct {
	Campaigns     []CampaignBulkCreateDto `json:"campaigns" validate:"required,min=1,dive"`
	ClientId      string                  `json:"client_id"`
	ShortCode     string                  `json:"short_code" validate:"required"`
	ExperienceID  primitive.ObjectID      `json:"experience_id"  validate:"required"`
	CreatedBy     *models.User            `json:"user,omitempty"`
	QRCoordinates []QRCoordinates         `json:"qr_coordinates" validate:"required,dive"`
	QRImageURl    string                  `json:"qr_image_url" validate:"required"`
}

type QRCoordinates struct {
	X float32 `json:"x" validate:"required"`
	Y float32 `json:"y" validate:"required"`
}

type CampaignUpdateDto struct {
	Name          string               `json:"name,omitempty"`
	Status        string               `json:"status,omitempty"`
	FeatureFlags  *models.FeatureFlags `json:"feature_flags,omitempty"`
	Scan          *models.Scan         `json:"scan,omitempty"`
	IsActive      *bool                `json:"is_active,omitempty"`
	ExpiresAt     time.Duration        `json:"expires_at,omitempty"`
	ExpiryDate    string               `json:"expiry_date,omitempty" validate:"omitempty,expirydate"`
	GoliveAt      time.Duration        `json:"golive_at,omitempty"`
	Publish       *bool                `json:"publish,omitempty"`
	ShowCopyRight *bool                `json:"show_copyright,omitempty"`
	EditedBy      *models.User         `json:"edited_by,omitempty"`
	Share         *models.Share        `json:"share,omitempty"`
	NetworkInfo   *NetworkInfo         `json:"network_info,omitempty"`
	QRConfig      *models.QrConfig     `json:"qr_config,omitempty"`
	GroupID       primitive.ObjectID   `json:"group_id,omitempty"`
	GroupName     string               `json:"group_name,omitempty"`
}

type PostbackCampaignDto struct {
	ShortCode              string `json:"short_code"`
	ScanCompressedImageUrl string `json:"scan_compressed_image_url,omitempty"`
}

type ClaimCampaignDto struct {
	ShortCode string       `json:"short_code" validate:"required"`
	User      *models.User `json:"user"`
	ClientID  string       `json:"client_id"`
}

type SendCampaignMailDto struct {
	CampaignId       string       `json:"campaign_id"`
	CreatedAt        int64        `json:"created_at"`
	UpdatedAt        int64        `json:"updated_at"`
	Name             string       `json:"name"`
	ClientId         string       `json:"client_id"`
	Email            string       `json:"email"`
	CreatedBy        *models.User `json:"created_by,omitempty"`
	TriggerImage     string       `json:"trigger_image"`
	TrackType        string       `json:"track_type"`
	NotificationType string       `json:"notification_type"`
	CreditType       string       `json:"credit_type"`
	Balance          int32        `json:"balance"`
	Unlimited        bool         `json:"unlimited"`
}

type SendCampaignPushNotifDto struct {
	SendPush      bool                   `json:"send_push"`
	CampaignId    string                 `json:"campaign_id"`
	ShortCode     string                 `json:"short_code"`
	UserID        string                 `json:"user_id"`
	CreatedAt     int64                  `json:"created_at"`
	Name          string                 `json:"name"`
	ClientId      string                 `json:"client_id"`
	Variables     map[string]interface{} `json:"variables"`
	Email         string                 `json:"email"`
	RecipientType string                 `json:"recipient_type"`
	NotifType     string                 `json:"notif_type"`
}

type GetCampaignDto struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	ShortCode string `json:"short_code"`
}

type CampaignFilterDto struct {
	Name     string `json:"name"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Offset   int    `json:"offset"`
	Publish  *bool  `json:"publish"`
	Draft    *bool  `json:"draft"`
}

type UpdateBulkCampaignDto struct {
	ClientId           primitive.ObjectID    `json:"client_id,omitempty"`
	EditedBy           *models.User          `json:"edited_by,omitempty"`
	SourceShortCode    string                `json:"source_short_code" validate:"required"`
	BulkCampaignUpdate *[]BulkCampaignUpdate `json:"campaign_update" validated:"dive,min=1"`
	QRCoordinates      []QRCoordinates       `json:"qr_coordinates" validate:"required,dive"`
	QRImageURl         string                `json:"qr_image_url" validate:"required"`
}

type BulkCampaignUpdate struct {
	ShortCode      string `json:"short_code" validate:"required"`
	CampaignName   string `json:"campaign_name,omitempty"`
	GroupName      string `json:"group_name,omitempty"`
	QrId           string `json:"qr_id,omitempty"`
	QrBGColor      string `json:"qr_bg_color,omitempty"`
	QrTextColor    string `json:"qr_text_color,omitempty"`
	Title          string `json:"title,omitempty"`
	BodyText       string `json:"body_text,omitempty"`
	RedirectionUrl string `json:"redirection_url,omitempty"`
	CTA            *bool  `json:"cta" validate:"required"`
}

type ClientCampaignsInfo struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	Name      string             `json:"name" bson:"name"`
	ShortCode string             `json:"short_code" bson:"short_code"`
}
