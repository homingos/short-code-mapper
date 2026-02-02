package dtos

import (
	"time"

	"github.com/homingos/campaign-svc/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExperienceCreationDto struct {
	CampaignID       string                   `json:"campaign_id" validate:"required"`
	Name             string                   `bson:"name,omitempty" json:"name,omitempty"`
	Variant          *models.Variant          `json:"variant,omitempty" validate:"required"`
	SegmentData      SegmentData              `json:"segment_data,omitempty" validate:"omitempty"`
	ScanText         string                   `json:"scan_text"`
	ImageURL         string                   `json:"image_url,omitempty" validate:"omitempty,url"`
	OriginalInputUrl string                   `json:"original_input_url,omitempty" validate:"omitempty,url"` //doubt 1: in case of button segments, nothing will be stored in origin video url at root level
	MaskedPhotoURL   string                   `json:"masked_photo_url,omitempty" validate:"omitempty,url"`
	VideoURL         string                   `json:"video_url,omitempty" validate:"omitempty,url"` // doubt 1
	AudioURL         string                   `json:"audio_url,omitempty" validate:"omitempty,url"`
	GLB              string                   `json:"glb,omitempty" validate:"omitempty,url"`
	USDZ             string                   `json:"usdz,omitempty" validate:"omitempty,url"`
	OBJ              string                   `json:"obj,omitempty" validate:"omitempty,url"`
	TextureFile      string                   `json:"texture_file,omitempty" validate:"omitempty,url"`
	BlendFile        string                   `json:"blend_file,omitempty" validate:"omitempty,url"`
	UIElements       *models.UIElements       `json:"ui_elements,omitempty"`
	QrCode           bool                     `json:"qr_code,omitempty"`
	WebmUrl          string                   `json:"webm_url,omitempty"`
	PlaybackScale    float64                  `json:"playback_scale,omitempty"`
	EngagmentOptions *models.EngagmentOptions `json:"engagment_options,omitempty"`
	ImageHash        string                   `json:"image_hash,omitempty"`
	FeatureImageURL  string                   `json:"feature_image_url,omitempty"`
	RewardEnabled    *bool                    `json:"reward_enabled,omitempty"`
	Publish          bool                     `json:"publish,omitempty"`
	CreatedBy        *models.User             `json:"user,omitempty"`
	NetworkInfo      *NetworkInfo             `json:"network_info,omitempty"`
	Status           string                   `json:"status,omitempty"`
	Rewards          *models.Rewards          `json:"rewards,omitempty"`
	MaskUrl          string                   `json:"mask_url,omitempty"`
	Overlay          *models.Overlay          `json:"overlay,omitempty"`
	SpawnImage       string                   `json:"spawn_image,omitempty"`
	TemplateDetails  map[string]interface{}   `json:"template_details,omitempty"`
	TemplateCategory *models.TemplateCategory `json:"template_category,omitempty"`
	Scene            *models.Scene            `json:"scene,omitempty"`
	QrPanel          *models.QrPanel          `json:"qr_panel,omitempty"`
}

type SegmentData struct {
	UseSegmentedElement bool              `json:"use_segmented_element"`
	ButtonConfig        *models.BtnConfig `json:"button_config,omitempty" validate:"omitempty"`
	ButtonSegments      []ButtonSegment   `json:"button_segments" validate:"omitempty,dive"`
}

type ButtonSegment struct {
	ButtonType       string                   `json:"button_type" validate:"omitempty,oneof=image"` // image,video,gif
	AssetFileName    string                   `json:"asset_file_name"`
	AssetURL         string                   `json:"asset_url" validate:"omitempty,url"`
	Color            string                   `json:"color"` // for border color, hexcode
	MarkerID         string                   `json:"marker_id"`
	Position         models.ThreeDCoordinates `json:"position"`
	Scale            models.TwoDCoordinates   `json:"scale"`
	MaskID           *int                     `json:"mask_id" validate:"omitempty,oneof=0 1 2"` // (0 - Rectangle, 1 - Square,2 - Circle)
	ShowElements     []int                    `json:"show_elements"`                            // which ui elements to show when button is clicked
	OriginalVideoURL string                   `json:"original_video_url,omitempty" validate:"omitempty,url"`
	MergeVideo       string                   `json:"merge_video,omitempty" validate:"omitempty,url"`
	Orientation      *string                  `json:"orientation,omitempty"`
	MaskURL          string                   `json:"mask_url,omitempty" validate:"omitempty,url"`
}

type ExperienceResetDto struct {
	ExperienceCreationDto
	ExperienceID string `json:"experience_id" validate:"required"`
}

type Image struct {
	K string `json:"k"`
	V string `json:"v"`
}

type Video struct {
	K string `json:"k"`
	V string `json:"v"`
}

type ExperienceUpdateDto struct {
	Variant          *models.Variant          `json:"variant,omitempty"`
	SegmentData      SegmentData              `json:"segment_data,omitempty" validate:"omitempty"`
	Name             string                   `bson:"name,omitempty" json:"name,omitempty"`
	ImageURL         string                   `json:"image_url,omitempty" validate:"omitempty,url"`
	OriginalInputUrl string                   `json:"original_input_url,omitempty" validate:"omitempty,url"`
	FeatureImageUrl  string                   `json:"feature_image_url,omitempty"`
	MaskedPhotoURL   string                   `json:"masked_photo_url,omitempty" validate:"omitempty,url"`
	PlaybackURL      string                   `json:"playback_url,omitempty" validate:"omitempty,url"`
	VideoURL         string                   `json:"video_url,omitempty" validate:"omitempty,url"`
	AudioURL         string                   `json:"audio_url,omitempty" validate:"omitempty,url"`
	GLB              string                   `json:"glb,omitempty" validate:"omitempty,url"`
	USDZ             string                   `json:"usdz,omitempty" validate:"omitempty,url"`
	OBJ              string                   `json:"obj,omitempty" validate:"omitempty,url"`
	TextureFile      string                   `json:"texture_file,omitempty" validate:"omitempty,url"`
	BlendFile        string                   `json:"blend_file,omitempty" validate:"omitempty,url"`
	UIElements       *models.UIElements       `json:"ui_elements,omitempty"`
	Status           string                   `json:"status,omitempty"`
	QrCode           bool                     `json:"qr_code,omitempty"`
	WebmUrl          string                   `json:"webm_url,omitempty"`
	PlaybackScale    float64                  `json:"playback_scale,omitempty"`
	EngagmentOptions *models.EngagmentOptions `json:"engagment_options,omitempty"`
	Canvas           *models.Canvas           `json:"canvas,omitempty"`
	EditedBy         *models.User             `json:"edited_by,omitempty"`
	RewardEnabled    *bool                    `json:"reward_enabled,omitempty"`
	DeleteImage      bool                     `json:"delete_image,omitempty"`
	DeleteVideo      bool                     `json:"delete_video,omitempty"`
	DeleteMask       *bool                    `json:"delete_mask,omitempty"`
	DeleteSpawn      bool                     `json:"delete_spawn,omitempty"`
	DeleteTexture    bool                     `json:"delete_texture,omitempty"`
	DeleteObj        bool                     `json:"delete_obj,omitempty"`
	DeleteGlb        bool                     `json:"delete_glb,omitempty"`
	DeleteUsdz       bool                     `json:"delete_usdz,omitempty"`
	DeleteBlend      bool                     `json:"delete_blend,omitempty"`
	Publish          bool                     `json:"publish,omitempty"`
	Rewards          *models.Rewards          `json:"rewards,omitempty"`
	MaskUrl          string                   `json:"mask_url,omitempty"`
	Overlay          *models.Overlay          `json:"overlay,omitempty"`
	SpawnImage       string                   `json:"spawn_image,omitempty"`
	Scene            *models.Scene            `json:"scene,omitempty"`
	QrPanel          *models.QrPanel          `json:"qr_panel,omitempty"`
	ShareMeta        *models.ShareMeta        `json:"share_meta,omitempty"`
	DeleteShareMeta  bool                     `json:"delete_share_meta"`
	TemplateCategory *models.TemplateCategory `json:"template_category,omitempty"`
}

// Add fields for different use cases ...
type PostbackExperienceDto struct {
	ID                        string              `json:"id"`
	MarkerID                  string              `json:"marker_id,omitempty"` // for segment markers
	MaskVideo                 string              `json:"mask,omitempty"`      // for segment videos
	CompressedVideo           string              `json:"compressed_video,omitempty"`
	CompressedPlaybackVideo   string              `json:"compressed_playback_video,"`
	CompressedImage           string              `json:"compressed_image,omitempty"`
	OverlayCompressed         string              `json:"overlay_compressed,omitempty"`
	ShareUrl                  string              `json:"share_url,omitempty"`
	ShortCode                 string              `json:"short_code,omitempty"`
	HlsUrl                    string              `json:"hls_url,omitempty"`
	DashUrl                   string              `json:"dash_url,omitempty"`
	OriginalVideo             string              `json:"original_video,omitempty"`
	PlaybackUrl               string              `json:"playback_url,omitempty"`
	StdCompressedImage        string              `json:"std_compressed_image,omitempty"`
	ColorCompressedImage      string              `json:"color_compressed_image,omitempty"`
	TemplateMaskUrl           string              `json:"template_mask_url,omitempty"`
	VideoAspectRatio          float64             `json:"video_aspect_ratio,omitempty"`
	ImageAspectRatio          float64             `json:"image_aspect_ratio,omitempty"`
	IsEdited                  bool                `json:"is_edited,omitempty"`
	FeatureImage              string              `json:"feature_image,omitempty"`
	EditedBy                  *models.User        `json:"edited_by,omitempty"`
	StreamChange              bool                `json:"stream_change,omitempty"`
	IsHorizontal              *bool               `json:"is_horizontal,omitempty"`
	SpawnCompressedImage      string              `json:"spawn_compressed_image,omitempty"`
	Plane                     *models.Plane       `json:"plane,omitempty"`
	Publish                   bool                `json:"publish"`
	RemotionVideoUrl          string              `json:"remotion_video_url,omitempty"`
	RemotionMaskedVideoUrl    string              `json:"remotion_masked_video_url,omitempty"`
	ScanCompressedImage       string              `json:"scan_compressed_image_url,omitempty"`
	OgImageWithQR             string              `json:"og_image_with_qr,omitempty"`
	ErrorMessage              string              `json:"error_message,omitempty"`
	SegmentInfo               []SegmentMarkerInfo `json:"segment_info,omitempty"` // used when all segments will be stitched
	WebMUrl                   string              `json:"webm_url,omitempty"`
	MilvusRefID               string              `json:"milvus_ref_id,omitempty"`
	OriginalGreenScreenIMGURL string              `json:"original_green_screen_img_url,omitempty"`
	RGBVideoUrl               string              `json:"rgb_video_url,omitempty"`
	MaskVideoUrl              string              `json:"mask_video_url,omitempty"`
	GenStudioOutput           *GenStudioOutputDto `json:"genstudio_output,omitempty"`
	ProductDescription        string              `json:"product_description,omitempty"`
}

type GenStudioOutputDto struct {
	Value string `json:"value"`
}
type UniqueImageDto struct {
	CampaignID   string `json:"campaign_id" validate:"required"`
	ImageURL     string `json:"image_url" validate:"required"`
	ExperienceID string `json:"experience_id"`
}

type ImageHashResponse struct {
	IsUnique  bool   `json:"is_unique"`
	Message   string `json:"message"`
	ImageHash string `json:"image_hash"`
}

type ImageHashUnique struct {
	CampaignID         string `json:"campaign_id"`
	ImageHash          string `json:"image_hash"`
	MaxHammingDistance int    `json:"max_hamming"`
	ExperienceID       string `json:"experience_id,omitempty"`
}

type StreamResolutionDto struct {
	AvailableHlsResolutions  []int  `json:"available_hls_resolutions"`
	ActiveHlsResolutions     []int  `json:"active_hls_resolutions"`
	AvailableDashResolutions []int  `json:"available_dash_resolutions"`
	ActiveDashResolutions    []int  `json:"active_dash_resolutions"`
	DashURL                  string `json:"dash_url"`
	HlsURL                   string `json:"hls_url"`
	ActiveHlsUrl             string `json:"active_hls_url"`
	ActiveDashUrl            string `json:"active_dash_url"`
}

type UpdateResponseDto struct {
	MediaProcess *models.MediaProcess `json:"media_process"`
	EditLog      EditLogDto           `json:"edit_log"`
	SegmentInfo  *SegmentInfo         `json:"segment_update_info,omitempty"`
}

type PostbackResponseDto struct {
	EditLog    EditLogDto         `json:"edit_log"`
	Campaign   *models.Campaign   `json:"campaign"`
	Experience *models.Experience `json:"experience"`
}

type EditLogsFilterDto struct {
	ClientId     primitive.ObjectID `json:"client_id" validate:"required"`
	StartDate    time.Time          `json:"start_date"`
	EndDate      time.Time          `json:"end_date"`
	ShortCode    string             `json:"short_code"`
	Page         int                `json:"page" default:"0" validate:"min=1"`
	PageSize     int                `json:"page_size" default:"10"`
	ExperienceId primitive.ObjectID `json:"experience_id"`
}

type EditLogDto struct {
	ClientId       primitive.ObjectID `json:"client_id"`
	ExperienceId   primitive.ObjectID `json:"experience_id"`
	CampaignId     primitive.ObjectID `json:"campaign_id"`
	ShortCode      string             `json:"short_code"`
	Before         *models.Experience `json:"before"`
	After          *models.Experience `json:"after,omitempty"`
	BeforeCampaign *models.Campaign   `json:"before_campaign"`
	AfterCampaign  *models.Campaign   `json:"after_campaign"`
	NetworkInfo    *NetworkInfo       `json:"network_info,omitempty"`
}
type CloudFunctionRequest struct {
	QueryURL   string   `json:"query_url"`
	SearchURLs []string `json:"search_urls"`
}

type ImageSimilarityResponse struct {
	IsSimilar bool    `json:"is_similar"`
	Message   string  `json:"message"`
	Score     float64 `json:"score"`
}

type ConsumeCreditRequest struct {
	RefID             string `json:"ref_id"`
	RefName           string `json:"ref_name"`
	RefType           string `json:"ref_type"`
	CreditAllowanceID string `json:"credit_allowance_id"`
	UserID            string `json:"user_id"`
}

type AdjustEscrowCreditsRequest struct {
	ClientId          string `json:"client_id"`
	CreditType        string `json:"credit_type"`
	Reverse           bool   `json:"reverse"`
	CreditAllowanceID string `json:"credit_allowance_id"`
}

type GetBannerDto struct {
	Element  models.Element `json:"element"`
	QrCode   bool           `json:"qr_code"`
	ImageUrl string         `json:"image_url"`
}

type CampaignExperienceDto struct {
	Campaign    models.Campaign     `json:"campaign"`
	Experiences []models.Experience `json:"experiences"`
	Experience  *models.Experience  `json:"experience"`
}

type NetworkInfo struct {
	IPAddress  string `bson:"ip_address" json:"ip_address"`
	MACAddress string `bson:"mac_address,omitempty" json:"mac_address,omitempty"`
}

type RewardTokenDto struct {
	DeviceId string `json:"device_id"`
	Type     string `json:"type"`
	Code     string `json:"code"`
}

type CustomCampaign struct {
	ID           string                 `json:"_id,omitempty"`
	DID          string                 `json:"id,omitempty"` //id without _
	Name         string                 `json:"name,omitempty"`
	ShortCode    string                 `json:"short_code,omitempty"`
	TrackType    string                 `valid:"required,in(POSE|GROUND|CARD)" json:"track_type,omitempty"`
	AirTracking  *bool                  `json:"air_tracking,omitempty"`
	Scan         *models.Scan           `json:"scan,omitempty"`
	Status       string                 `valid:"in(CREATED|PROCESSING|PROCESSED)" json:"status,omitempty"`
	IsActive     bool                   `json:"is_active,omitempty"`
	Publish      *bool                  `json:"publish,omitempty"`
	CopyRight    *models.CopyRight      `json:"copyright,omitempty"`
	FeatureFlags map[string]interface{} `json:"feature_flags,omitempty"`
	CreatedAt    time.Duration          `json:"created_at,omitempty"`
	UpdatedAt    time.Duration          `json:"updated_at,omitempty"`
	GroupName    string                 `json:"group_name,omitempty"`
	ExpiresAt    time.Duration          `json:"expires_at,omitempty"`
	GoLiveAt     time.Duration          `json:"golive_at,omitempty"`
	Share        *models.Share          `json:"share,omitempty"`
	Experiences  *CustomExperience      `json:"experiences"`
	FlamLogo     string                 `json:"flam_logo,omitempty"`
	LogoWidth    int32                  `json:"logo_width,omitempty"`
	QrConfig     *models.QrConfig       `json:"qr_config,omitempty"`
	ClientId     primitive.ObjectID     `json:"client_id,omitempty"`
}
type Campaigns struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	ClientId    primitive.ObjectID `bson:"client_id" json:"client_id"`
	ShortCode   string             `bson:"short_code" json:"short_code"`
	QrConfig    *models.QrConfig   `bson:"qr_config,omitempty" json:"qr_config,omitempty"`
	Experiences *models.Experience `json:"experiences"`
}

type CampaignsExperiences struct {
	Status    string      `json:"status"`
	Campaigns []Campaigns `json:"campaigns" bson:"campaigns"`
}

type Videos struct {
	Compressed         string `json:"compressed,omitempty" bson:"compressed,omitempty"`
	CompressedPlayback string `json:"compressed_playback,omitempty" bson:"compressed_playback,omitempty"`
	DASH               string `json:"dash,omitempty" bson:"dash,omitempty"`
	HLS                string `json:"hls,omitempty" bson:"hls,omitempty"`
	Mask               string `json:"mask,omitempty" bson:"mask,omitempty"`
	MergedVideo        string `json:"merged_video,omitempty" bson:"merged_video,omitempty"`
	Orientation        string `json:"orientation,omitempty" bson:"orientation,omitempty"`
	Original           string `json:"original,omitempty" bson:"original,omitempty"`
}

type Images struct {
	Original        string `json:"original,omitempty" bson:"original,omitempty"`
	OriginalInput   string `json:"original_input,omitempty" bson:"original_input,omitempty"`
	Compressed      string `json:"compressed,omitempty" bson:"compressed,omitempty"`
	ColorCompressed string `json:"color_compressed,omitempty" bson:"color_compressed,omitempty"`
	StdCompressed   string `json:"std_compressed,omitempty" bson:"std_compressed,omitempty"`
	FeatureImage    string `json:"feature_image,omitempty" bson:"feature_image,omitempty"`
	Spawn           string `json:"spawn,omitempty" bson:"spawn,omitempty"`
	CompressedSpawn string `json:"compressed_spawn,omitempty" bson:"compressed_spawn,omitempty"`
}

type GLBs struct {
	OriginalGLB  string `json:"original_glb,omitempty"`
	OriginalUSDZ string `json:"original_usdz,omitempty"`
	OriginalOBJ  string `json:"original_obj,omitempty"`
	BlendFile    string `json:"blend_file,omitempty"`
	TextureFile  string `json:"texture_file,omitempty"`
}

type CustomExperience struct {
	AspectRatio       float32                  `json:"aspect_ratio,omitempty"`
	CampaignID        string                   `json:"campaign_id,omitempty"`
	Canvas            models.Canvas            `json:"canvas,omitempty"`
	CreatedAt         int64                    `json:"created_at,omitempty"`
	EngagementOptions models.EngagmentOptions  `json:"engagment_options,omitempty"`
	ID                string                   `json:"id,omitempty"`
	Images            Images                   `json:"images,omitempty"`
	IsActive          bool                     `json:"is_active,omitempty"`
	PlaybackScale     float64                  `json:"playback_scale,omitempty"`
	Rewards           models.Rewards           `json:"rewards,omitempty"`
	Status            string                   `json:"status,omitempty"`
	TemplateDetails   map[string]interface{}   `json:"template_details,omitempty"`
	TotalTask         int                      `json:"total_task,omitempty"`
	UIElements        models.UIElements        `json:"ui_elements,omitempty"`
	UpdatedAt         int64                    `json:"updated_at,omitempty"`
	Variant           models.Variant           `json:"variant,omitempty"`
	Videos            Videos                   `json:"videos,omitempty"`
	GLBs              GLBs                     `json:"3d_assets,omitempty"`
	WorkflowID        string                   `json:"workflow_id,omitempty"`
	WorkflowError     *models.WorkflowError    `json:"workflow_error,omitempty"`
	Overlay           models.Overlay           `json:"overlay,omitempty"`
	CreditDeduct      bool                     `json:"credit_deduct"`
	Mask              models.Mask              `json:"mask,omitempty"`
	Name              string                   `json:"name,omitempty"`
	QrPanel           *models.QrPanel          `json:"qr_panel,omitempty"`
	ShareMeta         *models.ShareMeta        `json:"share_meta,omitempty"`
	TemplateCategory  *models.TemplateCategory `json:"template_category,omitempty"`
	VideoGeneration   *models.VideoGeneration  `json:"video_generation,omitempty"`
}

type ReprocessBulkCampaignDto struct {
	Pending           bool               `json:"pending,omitempty"`
	ExperienceID      primitive.ObjectID `json:"experience_id,omitempty"`
	SourceShortCode   string             `json:"source_short_code" validate:"required"`
	CampaignShortCode string             `json:"campaign_short_code"`
	QRCoordinates     []QRCoordinates    `json:"qr_coordinates" validate:"required,dive"`
	QRImageURl        string             `json:"qr_image_url" validate:"required"`
}

type UpdatedVariantInfo struct {
	Type     string
	MarkerId string
	VideoURL string
	MaskURL  string
	AssetURL string
}

type SegmentInfo struct {
	ImageInfo          []UpdatedVariantInfo `json:"image_info"`
	VideoInfo          []UpdatedVariantInfo `json:"video_info"`
	VideoUrls          []SegmentVideo       `json:"video_url"`
	ProcessStitchVideo bool                 `json:"process_stitch_video"`
}
type SegmentMarkerInfo struct {
	MarkerID  string `json:"marker_id,omitempty"`
	StartTime int64  `json:"stime"`
	EndTime   int64  `json:"etime"`
}

type RegenerateExperienceAssetsDto struct {
	Prompt   string       `json:"prompt" validate:"required"`
	EditedBy *models.User `json:"edited_by,omitempty"`
}
