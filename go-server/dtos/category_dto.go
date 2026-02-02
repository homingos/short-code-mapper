package dtos

import "github.com/homingos/campaign-svc/models"

type CategoryAppResponseDto struct {
	ID              string                   `bson:"_id" json:"_id"`
	ClientID        string                   `bson:"client_id" json:"client_id"`
	SiteCode        string                   `bson:"site_code" json:"site_code"`
	BrandInfo       models.BrandInfo         `bson:"brand_info" json:"brand_info"`
	ShareMeta       models.CategoryShareMeta `bson:"share_meta" json:"share_meta"`
	Categories      []CategoriesResponseDto  `bson:"categories" json:"categories" validate:"dive"`
	OrderButtonText string                   `json:"order_button_text"`
}

type CategoriesResponseDto struct {
	Name      string                `bson:"name" json:"name"`
	Campaigns []CategoryCampaignDto `bson:"campaigns" json:"campaigns"`
}

// CategorySearchResponseDto - simplified response for text search
type CategorySearchResponseDto struct {
	ID              string                        `bson:"_id" json:"_id"`
	ClientID        string                        `bson:"client_id" json:"client_id"`
	SiteCode        string                        `bson:"site_code" json:"site_code"`
	BrandInfo       models.BrandInfo              `bson:"brand_info" json:"brand_info"`
	ShareMeta       models.CategoryShareMeta      `bson:"share_meta" json:"share_meta"`
	Categories      []CategoriesSearchResponseDto `bson:"categories" json:"categories"`
	OrderButtonText string                        `json:"order_button_text"`
}

type CategoriesSearchResponseDto struct {
	Name      string   `bson:"name" json:"name"`
	Campaigns []string `bson:"campaigns" json:"campaigns"`
}

type CategoryCampaignDto struct {
	ID          string                `bson:"_id" json:"_id"`
	IconURL     string                `bson:"icon_url" json:"icon_url"`
	Name        string                `bson:"name" json:"name"`
	ShortCode   string                `bson:"short_code" json:"short_code"`
	Experiences CategoryExperienceDto `bson:"experiences" json:"experiences"`
}

type CategoryExperienceDto struct {
	ID               string              `bson:"_id" json:"_id"`
	AspectRatio      float64             `bson:"aspect_ratio" json:"aspect_ratio"`
	Canvas           models.Canvas       `bson:"canvas" json:"canvas"`
	Images           Images              `bson:"images" json:"images"`
	PlaybackScale    float64             `bson:"playback_scale" json:"playback_scale"`
	Videos           Videos              `bson:"videos" json:"videos"`
	Variant          CategoryVariantDto  `bson:"variant" json:"variant"`
	CatalogueDetails CatalogueDetailsDto `bson:"catalogue_details" json:"catalogue_details"`
}

type CatalogueDetailsDto struct {
	Name     string `bson:"name" json:"name"`
	Currency string `bson:"currency" json:"currency"`
	Price    string `bson:"price" json:"price"`
}

type CategoryVariantDto struct {
	AndroidHapticPattern string                   `bson:"android_haptic_pattern" json:"android_haptic_pattern"`
	Class                int                      `bson:"class" json:"class"`
	EnableVoiceModule    bool                     `bson:"enable_voice_module" json:"enable_voice_module"`
	IOSHapticPattern     string                   `bson:"ios_haptic_pattern" json:"ios_haptic_pattern"`
	IsAlpha              bool                     `bson:"is_alpha" json:"is_alpha"`
	IsHorizontal         bool                     `bson:"is_horizontal" json:"is_horizontal"`
	Offset               models.ThreeDCoordinates `bson:"offset" json:"offset"`
	ScaleAxis            models.ThreeDCoordinates `bson:"scale_axis" json:"scale_axis"`
	TrackType            string                   `bson:"track_type" json:"track_type"`
	Segments             CategorySegmentsDto      `bson:"segments" json:"segments"`
}

type CategorySegmentsDto struct {
	BackColor  string                     `bson:"back_color" json:"back_color"`
	FlushColor string                     `bson:"flush_color" json:"flush_color"`
	Default    string                     `bson:"default" json:"default"`
	Markers    []CategorySegmentMarkerDto `bson:"markers" json:"markers"`
}

type CategorySegmentMarkerDto struct {
	ID             string `bson:"id" json:"id"`
	Color          string `bson:"color" json:"color"`
	ColorMap       string `bson:"color_map" json:"color_map"`
	AllowSelection bool   `bson:"allow_selection" json:"allow_selection"`
	Stime          int64  `bson:"stime" json:"stime"`
	Etime          int64  `bson:"etime" json:"etime"`
	Next           string `bson:"next" json:"next"`
	Multiplier     string `bson:"multiplier" json:"multiplier"`
	RedirectionURL string `bson:"redirection_url" json:"redirection_url"`
	WebRedirect    bool   `bson:"web_redirect" json:"web_redirect"`
	Name           string `bson:"name" json:"name"`
}

type CreateCategoryDto struct {
	Name       string                   `json:"name" validate:"required"`
	SiteCode   string                   `json:"site_code"`
	BrandInfo  models.BrandInfo         `json:"brand_info" validate:"required"`
	ShareMeta  models.CategoryShareMeta `json:"share_meta" validate:"required"`
	Categories []models.Categories      `json:"categories" validate:"required,min=1,dive"`
	ClientID   string                   `json:"client_id" validate:"required"`
	CreatedBy  *models.User             `json:"created_by" validate:"required"`
}

type UpdateCategoryDto struct {
	Name       string                    `json:"name"`
	ClientID   string                    `json:"client_id"`
	BrandInfo  *models.BrandInfo         `json:"brand_info" validate:"omitempty"`
	ShareMeta  *models.CategoryShareMeta `json:"share_meta" validate:"omitempty"`
	Categories []models.Categories       `json:"categories" validate:"omitempty,dive"`
	UpdatedBy  *models.User              `json:"updated_by"`
}

type ProductCatalogueDto struct {
	Title       string       `json:"title" validate:"required"`
	Description string       `json:"description" validate:"required"`
	LogoURL     string       `json:"logo_url" validate:"omitempty,url"`
	WebsiteLink string       `json:"website_link" validate:"required,url"`
	Products    []ProductDto `json:"products" validate:"required,dive"`
	ClientID    string       `json:"client_id"`
	CreatedBy   *models.User `json:"created_by"`
	SiteCode    string       `json:"site_code"`
	CategoryID  string       `json:"category_id"`
}

type ProductDto struct {
	ID          string `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Category    string `json:"category" validate:"required"`
	Price       string `json:"price" validate:"required"`
	Currency    string `json:"currency" validate:"required"`
	ImageUrl    string `json:"image_url" validate:"required,url"`
	ProductUrl  string `json:"product_url" validate:"required,url"`
	Description string `json:"description" validate:"required"`
}

type Document struct {
	ID          string `json:"id"`
	Name 	  	string `json:"name"`
	CatalogID   string `json:"catalog_id"`
	ClientID    string `json:"client_id"`
	Description string `json:"description"`
}

type SearchResult struct {
	Document
	Score float32 `json:"score,omitempty"`
}

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding,omitempty"`
	ModelName string    `json:"model_name,omitempty"`
	Dimension int       `json:"dimension,omitempty"`
	DType     string    `json:"d_type,omitempty"`
}

type ReprocessCatalogueDto struct {
	ClientID   string `json:"client_id"`
	CampaignID string `json:"campaign_id" validate:"required"`
	ShortCode  string `json:"short_code" validate:"required"`
	SiteCode   string `json:"site_code" validate:"required"`
}
