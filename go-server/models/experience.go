package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// const (
// 	nose int = iota
// 	leftEyeInner
// 	leftEye
// 	leftEyeOuter
// 	rightEyeInner
// 	rightEye
// 	rightEyeOuter
// 	leftEar
// 	rightEar
// 	mouthLeft
// 	mouthRight
// 	leftShoulder
// 	rightShoulder
// 	leftElbow
// 	rightElbow
// 	leftWrist
// 	rightWrist
// 	leftPinky
// 	rightPinky
// 	leftIndex
// 	rightIndex
// 	leftThumb
// 	rightThumb
// 	leftHip
// 	rightHip
// 	leftKnee
// 	rightKnee
// 	leftAnkle
// 	rightAnkle
// 	leftHeel
// 	rightHeel
// 	leftFootIndex
// 	rightFootIndex
// )

type Marker struct {
	Color          string `bson:"color" json:"color"`
	Stime          int64  `bson:"stime" json:"stime"`
	Etime          int64  `bson:"etime" json:"etime"`
	RedirectionURL string `bson:"redirection_url" json:"redirection_url"`
	Default        bool   `bson:"default" json:"default"`
}

type Canvas struct {
	IOS     int64 `bson:"ios" json:"ios"`
	Android int64 `bson:"android" json:"android"`
}

type SegmentMarker struct {
	Id             string        `bson:"id" json:"id"`
	Color          string        `bson:"color" json:"color"`
	ColorMap       string        `bson:"color_map" json:"color_map"`
	AllowSelection bool          `bson:"allow_selection" json:"allow_selection"`
	Stime          int64         `bson:"stime" json:"stime"`
	Etime          int64         `bson:"etime" json:"etime"`
	Next           string        `bson:"next" json:"next"`
	Multiplier     string        `bson:"multiplier" json:"multiplier"`
	RedirectionURL string        `bson:"redirection_url" json:"redirection_url"`
	WebRedirect    bool          `bson:"web_redirect" json:"web_redirect"`
	ShowElements   []int         `bson:"show_elements" json:"show_elements"` // IDs of elements to show
	Videos         VideoObject   `bson:"videos,omitempty" json:"videos,omitempty"`
	IsHorizontal   *bool         `bson:"is_horizontal,omitempty" json:"is_horizontal,omitempty"`
	MarkerButtons  []interface{} `bson:"marker_buttons,omitempty" json:"marker_buttons,omitempty"`
	Name           string        `bson:"name,omitempty" json:"name,omitempty"`
}

type Loop struct {
	Stime *int64 `bson:"stime,omitempty" json:"stime,omitempty"`
	Etime *int64 `bson:"etime,omitempty" json:"etime,omitempty"`
}

type GroundMeta struct {
	Scale            float64           `bson:"scale,omitempty" json:"scale,omitempty"`
	ScaleUnit        string            `bson:"scale_unit,omitempty" json:"scale_unit,omitempty" validate:"omitempty,oneof=ft m"`
	Rotation         []float64         `bson:"rotation,omitempty" json:"rotation,omitempty"`
	LookAt           *bool             `bson:"lookat,omitempty" json:"lookat,omitempty"`
	Offset           ThreeDCoordinates `bson:"offset,omitempty" json:"offset,omitempty"`
	OffsetUnit       string            `bson:"offset_unit,omitempty" json:"offset_unit,omitempty" validate:"omitempty,oneof=ft m"`
	Timeout          int64             `bson:"timeout,omitempty" json:"timeout,omitempty"`
	TimeoutHelpText  string            `bson:"timeout_help_text,omitempty" json:"timeout_help_text,omitempty"`
	Distance         float64           `bson:"distance,omitempty" json:"distance,omitempty"`
	DistanceHelpText string            `bson:"distance_help_text,omitempty" json:"distance_help_text,omitempty"`
	InstantPlacement *bool             `bson:"instant_placement,omitempty" json:"instant_placement"`
	Shadow           Shadow            `bson:"shadow,omitempty" json:"shadow,omitempty"`
}
type Shadow struct {
	Softness     uint32            `bson:"softness,omitempty" json:"softness,omitempty"`
	Opacity      float64           `bson:"opacity,omitempty" json:"opacity,omitempty"`
	Offset       ThreeDCoordinates `bson:"offset,omitempty" json:"offset,omitempty"`
	EnableShadow *bool             `bson:"enable_shadow,omitempty" json:"enable_shadow,omitempty"`
}

type Anchor struct {
	ID           int               `bson:"id" json:"id"`
	ReadableName string            `bson:"readable_name" json:"readable_name"`
	Offset       ThreeDCoordinates `bson:"offset" json:"offset"`
	Rotation     []float64         `bson:"rotation" json:"rotation"`
}

// type Offset struct {
// 	X float64 `json:"x" bson:"x"`
// 	Y float64 `json:"y" bson:"y"`
// 	Z float64 `json:"z" bson:"z"`
// }

type ThreeDCoordinates struct {
	TwoDCoordinates `bson:",inline" json:",inline"`
	Z               float64 `json:"z" bson:"z"`
}

type TwoDCoordinates struct {
	X float64 `json:"x" bson:"x"`
	Y float64 `json:"y" bson:"y"`
}

// type Rotation struct {
// 	X float64 `json:"x" bson:"x"`
// 	Y float64 `json:"y" bson:"y"`
// 	Z float64 `json:"z" bson:"z"`
// }

// type ScaleAxis struct {
// 	X float64 `bson:"x" json:"x"`
// 	Y float64 `bson:"y" json:"y"`
// 	Z float64 `bson:"z" json:"z"`
// }

type PoseMeta struct {
	Scale       float64  `bson:"scale,omitempty" json:"scale,omitempty"`
	Anchors     []Anchor `bson:"anchors" json:"anchors"`
	MultiSelect bool     `bson:"multiselect" json:"multiselect"`
}

// type EngagmentOptions struct {
// 	Sharable          bool   `bson:"sharable" json:"sharable"`
// 	ProcessingOptions string `bson:"processing_options" json:"processing_options"`
// }

type EngagmentOptions struct {
	Sharable           *bool               `bson:"sharable" json:"sharable"`
	Name               string              `bson:"name,omitempty" json:"name,omitempty"`
	ShadowGeneration   *ShadowGeneration   `bson:"shadow_generation,omitempty" json:"shadow_generation,omitempty"`
	ImageHarmonization *ImageHarmonization `bson:"image_harmonization,omitempty" json:"image_harmonization,omitempty"`
	BorderSmoothing    *BorderSmoothing    `bson:"border_smoothing,omitempty" json:"border_smoothing,omitempty"`
	PostProcessFilter  *PostProcessFilter  `bson:"post_process_filter,omitempty" json:"post_process_filter,omitempty"`
}

type ShadowGeneration struct {
	Enabled *bool `bson:"enabled" json:"enabled"`
}

type ImageHarmonization struct {
	Enabled   *bool    `bson:"enabled" json:"enabled"`
	ModelList []string `bson:"model_list" json:"model_list"`
}

type BorderSmoothing struct {
	Enabled *bool `bson:"enabled" json:"enabled"`
}

type PostProcessFilter struct {
	Enabled    *bool    `bson:"enabled" json:"enabled"`
	FilterList []string `bson:"filter_list" json:"filter_list"`
}

type Segments struct {
	BackColor            string          `bson:"back_color" json:"back_color"`
	FlushColor           string          `bson:"flush_color" json:"flush_color"`
	Default              string          `bson:"default" json:"default"`
	UseMarkerVideo       bool            `bson:"use_marker_video" json:"use_marker_video"`
	UseSegmentedElements bool            `bson:"use_segmented_elements" json:"use_segmented_elements"`
	Markers              []SegmentMarker `bson:"markers" json:"markers"`
}

type Variant struct {
	TrackType            string              `valid:"in(POSE|GROUND|CARD)" bson:"track_type" json:"track_type" validate:"required,oneof=POSE GROUND CARD"`
	Class                int                 `valid:"in(0|1|2|3)" bson:"class" json:"class"`
	Buttons              []InteractiveButton `bson:"buttons,omitempty" json:"buttons,omitempty" validate:"omitempty,dive"`
	EnableVoiceModule    *bool               `bson:"enable_voice_module,omitempty" json:"enable_voice_module,omitempty"`
	Segments             *Segments           `bson:"segments,omitempty" json:"segments,omitempty"`
	VoiceModule          *VoiceModule        `bson:"voice_module,omitempty" json:"voice_module,omitempty"`
	GroundMeta           *GroundMeta         `bson:"ground_meta,omitempty" json:"ground_meta,omitempty"`
	PoseMeta             *PoseMeta           `bson:"pose_meta,omitempty" json:"pose_meta,omitempty"`
	IsAlpha              *bool               `bson:"is_alpha,omitempty" json:"is_alpha,omitempty"`
	Offset               ThreeDCoordinates   `bson:"offset,omitempty" json:"offset,omitempty"`
	IsHorizontal         *bool               `bson:"is_horizontal,omitempty" json:"is_horizontal,omitempty"`
	ScaleAxis            ThreeDCoordinates   `bson:"scale_axis,omitempty" json:"scale_axis,omitempty"`
	ButtonConfig         *BtnConfig          `bson:"button_config,omitempty" json:"button_config,omitempty" validate:"omitempty"`
	IOSHapticPattern     *string             `bson:"ios_haptic_pattern,omitempty" json:"ios_haptic_pattern,omitempty"`
	AndroidHapticPattern *string             `bson:"android_haptic_pattern,omitempty" json:"android_haptic_pattern,omitempty"`
	StitchedVideo        *bool               `bson:"stitched_video,omitempty" json:"stitched_video,omitempty"`
}

type BtnConfig struct {
	ButtonLayout      string      `bson:"button_layout" json:"button_layout" validate:"required,oneof=top bottom left right"`
	ButtonAspectRatio HeightWidth `bson:"button_aspect_ratio" json:"button_aspect_ratio"`
	VideoAspectRation HeightWidth `bson:"video_aspect_ratio" json:"video_aspect_ratio"`
	ButtonAlignment   string      `bson:"button_alignment" json:"button_alignment" validate:"required,oneof=left right center top bottom middle"`
	ButtonGap         float64     `bson:"button_gap" json:"button_gap"`
	ButtonOffset      float64     `bson:"button_offset" json:"button_offset"`
}

type HeightWidth struct {
	Height int `bson:"height" json:"height"`
	Width  int `bson:"width" json:"width"`
}

type Image struct {
	K string `valid:"in(original|compressed|color_compressed|masked_photo|compressed_masked_photo|feature_image|spawn|compressed_spawn|fdb|original_input)" bson:"k" json:"k"`
	V string `bson:"v" json:"v"`
}

type Video struct {
	K string `valid:"in(original|playback|compressed|compressed_playback|webm|hls|dash|mask)" bson:"k" json:"k"`
	V string `bson:"v" json:"v"`
}

type Audio struct {
	K string `valid:"in(original)" bson:"k" json:"k"`
	V string `bson:"v" json:"v"`
}

type GLB struct {
	K string `valid:"in(original)" bson:"k" json:"k"`
	V string `bson:"v" json:"v"`
}

type Banners struct {
	// Think of it as display flags
	Variant        int    `bson:"variant" json:"variant"`
	Title          string `bson:"title,omitempty" json:"title,omitempty"`
	SubTitle       string `bson:"sub_title,omitempty" json:"sub_title,omitempty"`
	ShareText      string `bson:"share_text,omitempty" json:"share_text,omitempty"`
	RedirectionUrl string `bson:"redirection_url,omitempty" json:"redirection_url,omitempty"`
	PrimaryColor   string `bson:"primary_color,omitempty" json:"primary_color,omitempty"`
	ShareURL       string `bson:"share_url,omitempty" json:"share_url,omitempty"`
	SecondaryColor string `bson:"secondary_color,omitempty" json:"secondary_color,omitempty"`
}

type Buttons struct {
	Title          string `bson:"title" json:"title"`
	SubTitle       string `bson:"sub_title" json:"sub_title"`
	ShareText      string `bson:"share_text" json:"share_text"`
	RedirectionUrl string `bson:"redirection_url" json:"redirection_url"`
	Icon           string `bson:"icon,omitempty" json:"icon,omitempty"`
}
type BottomSheet struct {
	Variant        *int      `bson:"variant" json:"variant"`
	PrimaryColor   string    `bson:"primary_color" json:"primary_color"`
	SecondaryColor string    `bson:"secondary_color" json:"secondary_color"`
	TitleText      string    `bson:"title_text,omitempty" json:"title_text,omitempty"`
	Buttons        []Buttons `bson:"buttons" json:"buttons"`
}

type BottomSheetItems struct {
	Title          string `bson:"title,omitempty" json:"title,omitempty"`
	SubTitle       string `bson:"sub_title,omitempty" json:"sub_title,omitempty"`
	ShareText      string `bson:"share_text,omitempty" json:"share_text,omitempty"`
	RedirectionUrl string `bson:"redirection_url,omitempty" json:"redirection_url,omitempty" validate:"omitempty,url"`
	IconUrl        string `bson:"icon_url,omitempty" json:"icon_url,omitempty" validate:"omitempty,url"`
}

type StoryTrayItems struct {
	Type           string `bson:"type" json:"type" validate:"required,oneof=SEGMENT REDIRECTION"`
	SegmentId      string `bson:"segment_id,omitempty" json:"segment_id,omitempty"`
	ImageUrl       string `bson:"image_url,omitempty" json:"image_url,omitempty" validate:"omitempty,url"`
	Text           string `bson:"text,omitempty" json:"text,omitempty"`
	RedirectionUrl string `bson:"redirection_url,omitempty" json:"redirection_url,omitempty" validate:"omitempty,url"`
}

type Element struct {
	Type             string             `bson:"type" json:"type" validate:"required,oneof=BANNER BOTTOM_SHEET STORY_TRAY"`
	Variant          int                `bson:"variant" json:"variant"` // todo: ask its definition
	PrimaryColor     string             `bson:"primary_color,omitempty" json:"primary_color,omitempty"`
	SecondaryColor   string             `bson:"secondary_color,omitempty" json:"secondary_color,omitempty"`
	Title            string             `bson:"title,omitempty" json:"title,omitempty"`
	SubTitle         string             `bson:"sub_title,omitempty" json:"sub_title,omitempty"`
	ShareText        string             `bson:"share_text,omitempty" json:"share_text,omitempty"`
	RedirectionUrl   string             `bson:"redirection_url,omitempty" json:"redirection_url,omitempty" validate:"omitempty,url"`
	ShareUrl         string             `bson:"share_url,omitempty" json:"share_url,omitempty"`
	Icon             string             `bson:"icon,omitempty" json:"icon,omitempty"`
	BottomSheetItems []BottomSheetItems `bson:"bottom_sheet_items,omitempty" json:"bottom_sheet_items,omitempty" validate:"omitempty,dive"`
	StoryTrayItems   []StoryTrayItems   `bson:"story_tray_items,omitempty" json:"story_tray_items,omitempty" validate:"omitempty,dive"`
}

type UIElements struct {
	Elements []Element `bson:"elements,omitempty" json:"elements,omitempty" validate:"omitempty,dive"`
}

type MediaProcess struct {
	Experience               Experience                `json:"experience,omitempty"`
	Plane                    *Plane                    `json:"plane,omitempty"`
	ShortCode                string                    `json:"short_code"`
	ScanUrl                  string                    `json:"scan_url"`
	IsEdited                 bool                      `json:"is_edited"`
	Publish                  bool                      `json:"publish"`
	Type                     string                    `json:"type"`
	Name                     string                    `json:"name"`
	CreatedBy                User                      `json:"created_by"`
	ClientId                 primitive.ObjectID        `json:"client_id"`
	GenerateGreenScreen      bool                      `json:"generate_green_screen,omitempty"`
	ImageVectorLLMProductJob *ImageVectorLLMProductJob `json:"image_vector_llm_product_job,omitempty"`
	GenStudiJob              *FalVideoJob              `json:"genstudio_job,omitempty"`
	AlphaVideoJob            *AlphaVideoJob            `json:"alpha_video_job,omitempty"`
}

type ImageVectorLLMProductJob struct {
	ID          string `json:"id"`
	SiteCode    string `json:"site_code"`
	ClientID    string `json:"client_id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Price       string `json:"price"`
	Currency    string `json:"currency"`
	ImageURL    string `json:"image"`
	ProductURL  string `json:"product_url"`
	Description string `json:"description"`
}

type FalVideoJob struct {
	Type                  string           `json:"type"`
	Prompt                string           `json:"prompt"`
	VideoGenerationSource string           `json:"video_generation_source"`
	MediaReferences       []MediaReference `json:"media_references"`
	UserID                string           `json:"user_id"`
	EnableAudio           bool             `json:"enable_audio,omitempty"`
	VideoCategory         string           `json:"video_category,omitempty"`
	LowResolution         bool             `json:"low_resolution,omitempty"`
}

type MediaReference struct {
	FrameType string `json:"frame_type"`
	Type      string `json:"type"`
	URL       string `json:"url"`
}

type AlphaVideoJob struct {
	VideoUrl  string `json:"video_url"`
	ColorType string `json:"color_type"`
}

type Rewards struct {
	Type    string `bson:"type,omitempty" json:"type,omitempty" validate:"omitempty,oneof=PAYOUT COUPON"`
	Code    string `bson:"code,omitempty" json:"code,omitempty"`
	Enabled bool   `json:"enabled"`
}

type ShareMeta struct {
	Title               string `bson:"title,omitempty" json:"title,omitempty"`
	Description         string `bson:"description,omitempty" json:"description,omitempty"`
	OgImageUrl          string `bson:"og_image_url,omitempty" json:"og_image_url,omitempty"`
	RedirectionImageUrl string `bson:"redirection_image_url,omitempty" json:"redirection_image_url,omitempty"`
}

type Experience struct {
	ID                primitive.ObjectID     `bson:"_id" json:"id"`
	Name              string                 `bson:"name,omitempty" json:"name,omitempty"`
	CampaignID        primitive.ObjectID     `bson:"campaign_id" json:"campaign_id"`
	Canvas            Canvas                 `bson:"canvas,omitempty" json:"canvas,omitempty"`
	IsActive          bool                   `bson:"is_active" json:"is_active"`
	Variant           Variant                `bson:"variant,omitempty" json:"variant,omitempty"`
	Status            string                 `valid:"in(CREATED|PROCESSING|PROCESSED)" bson:"status" json:"status"`
	ImageHash         string                 `bson:"image_hash,omitempty" json:"image_hash,omitempty"`
	Images            []Image                `bson:"images" json:"images"`
	Videos            []Video                `bson:"videos" json:"videos"`
	PlaybackScale     float64                `bson:"playback_scale,omitempty" json:"playback_scale,omitempty"`
	Audios            []Audio                `bson:"audios,omitempty" json:"audios,omitempty"`
	GLBs              []GLB                  `bson:"3d_assets,omitempty" json:"3d_assets,omitempty"`
	QrCode            bool                   `bson:"qr_code" json:"qr_code"`
	UIElements        *UIElements            `bson:"ui_elements" json:"ui_elements"`
	Rewards           Rewards                `bson:"rewards,omitempty" json:"rewards,omitempty" validate:"omitempty,dive"`
	CreatedAt         time.Duration          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Duration          `bson:"updated_at" json:"updated_at"`
	CreatedBy         *User                  `bson:"created_by,omitempty" json:"created_by,omitempty"`
	EditedBy          *User                  `bson:"edited_by,omitempty" json:"edited_by,omitempty"`
	AspectRatio       float64                `bson:"aspect_ratio" json:"aspect_ratio"`
	EngagmentOptions  *EngagmentOptions      `bson:"engagment_options,omitempty" json:"engagment_options,omitempty"`
	ShareMeta         *ShareMeta             `bson:"share_meta,omitempty" json:"share_meta,omitempty"`
	Overlay           *Overlay               `bson:"overlay,omitempty" json:"overlay,omitempty"`
	Mask              *Mask                  `bson:"mask,omitempty" json:"mask,omitempty"`
	TemplateDetails   map[string]interface{} `bson:"template_details,omitempty" json:"template_details,omitempty"`
	Scene             *Scene                 `bson:"scene,omitempty" json:"scene,omitempty"`
	TemplateCategory  *TemplateCategory      `bson:"template_category,omitempty" json:"template_category,omitempty"`
	WorkflowError     WorkflowError          `bson:"workflow_error,omitempty" json:"workflow_error,omitempty"`
	CreditDeduct      bool                   `bson:"credit_deduct,omitempty" json:"credit_deduct,omitempty"`
	TotalTask         int32                  `bson:"total_task,omitempty" json:"total_task,omitempty"`
	WorkflowID        string                 `bson:"workflow_id,omitempty" json:"workflow_id,omitempty"`
	StitchWorkflowID  string                 `bson:"stitch_workflow_id,omitempty" json:"stitch_workflow_id,omitempty"`
	QrPanel           *QrPanel               `bson:"qr_panel,omitempty" json:"qr_panel,omitempty"`
	CreditAllowanceID primitive.ObjectID     `bson:"credit_allowance_id,omitzero" json:"credit_allowance_id,omitzero"`
	CatalogueDetails  *CatalogueDetails      `bson:"catalogue_details,omitempty" json:"catalogue_details,omitempty"`
	VideoGeneration   *VideoGeneration       `bson:"video_generation,omitempty" json:"video_generation,omitempty"`
}

type TemplateCategory struct {
	LicenseID       string                  `json:"license_id,omitempty" bson:"license_id,omitempty"`
	TemplateGroupID string                  `json:"template_group_id,omitempty" bson:"template_group_id,omitempty"`
	TemplateID      string                  `json:"template_id,omitempty" bson:"template_id,omitempty"`
	TemplateName    string                  `json:"template_name,omitempty" bson:"template_name,omitempty"`
	Canvas          *TemplateCategoryCanvas `json:"canvas,omitempty" bson:"canvas,omitempty"`
}

type TemplateCategoryCanvas struct {
	Type   string `json:"type,omitempty" bson:"type,omitempty"`
	Height int64  `json:"height,omitempty" bson:"height,omitempty"`
	Width  int64  `json:"width,omitempty" bson:"width,omitempty"`
}

type Overlay struct {
	Type            string `bson:"type,omitempty" json:"type,omitempty" validate:"required,oneof=COLOR BLUR IMAGE TRANSPARENT"`
	Value           string `bson:"value" json:"value"`
	CompressedImage string `bson:"compressed_image" json:"compressed_image"`
}

type Mask struct {
	URL           string            `bson:"url" json:"url" validate:"required,url"`
	Offset        ThreeDCoordinates `bson:"offset,omitempty" json:"offset,omitempty"`
	Rotation      ThreeDCoordinates `bson:"rotation,omitempty" json:"rotation,omitempty"`
	Scale         float32           `bson:"scale" json:"scale"`
	CompressedUrl string            `bson:"compressed_url,omitempty" json:"compressed_url,omitempty"`
}

type Scene struct {
	WindowRatio float64    `bson:"window_ratio" json:"window_ratio"`
	Parallax    []Parallax `bson:"parallax" json:"parallax" validate:"required,min=1,dive"`
}

type Parallax struct {
	ID     primitive.ObjectID `bson:"id" json:"id"`
	Mask   *Mask              `bson:"mask,omitempty" json:"mask,omitempty"`
	Planes []Plane            `bson:"planes" json:"planes" validate:"required,min=1,dive"`
}

type Plane struct {
	ID           primitive.ObjectID `bson:"id" json:"id"`
	ParallaxID   primitive.ObjectID `bson:"parallax_id,omitempty" json:"parallax_id,omitempty"`
	Type         *int               `bson:"type" json:"type" validate:"required,oneof=0 1 2 3 4"` // 0 is image, 1 is video, 2 for alpha video, 3 for 3d model, 4 for webview
	URL          string             `bson:"url" json:"url" validate:"required,url"`
	Compressed   string             `bson:"compressed,omitempty" json:"compressed,omitempty"`
	Mask         string             `bson:"mask,omitempty" json:"mask,omitempty"`
	Hls          string             `bson:"hls,omitempty" json:"hls,omitempty"`
	Dash         string             `bson:"dash,omitempty" json:"dash,omitempty"`
	Offset       ThreeDCoordinates  `bson:"offset" json:"offset"`
	Rotation     ThreeDCoordinates  `bson:"rotation" json:"rotation"`
	Scale        float64            `bson:"scale" json:"scale"`
	IsHorizontal *bool              `bson:"isHorizontal,omitempty" json:"isHorizontal,omitempty"`
}

type WorkflowError struct {
	ConsumerType string `json:"consumer_type,omitempty" bson:"consumer_type,omitempty"`
	TaskID       string `json:"task_id,omitempty" bson:"task_id,omitempty"`
	Msg          string `json:"msg,omitempty" bson:"msg,omitempty"`
	Filename     string `json:"filename,omitempty"`
}

type QrPanel struct {
	Active *bool   `json:"active,omitempty" bson:"active,omitempty"`
	QrData *QrData `json:"qr_data,omitempty" bson:"qr_data,omitempty"`
	Text   *Text   `json:"text,omitempty" bson:"text,omitempty"`
}

type QrData struct {
	Position *XYObject `json:"position,omitempty" bson:"position,omitempty"`
	Size     *HWObject `json:"size,omitempty" bson:"size,omitempty"`
}

type XYObject struct {
	X *float64 `json:"x,omitempty" bson:"x,omitempty"`
	Y *float64 `json:"y,omitempty" bson:"y,omitempty"`
}

type HWObject struct {
	H float64 `json:"h,omitempty" bson:"h,omitempty"`
	W float64 `json:"w,omitempty" bson:"w,omitempty"`
}

type Text struct {
	Value        string    `json:"value" bson:"value"`
	FontSize     int       `json:"font_size,omitempty" bson:"font_size,omitempty"`
	FontColor    string    `json:"font_color,omitempty" bson:"font_color,omitempty"`
	FontFamily   string    `json:"font_family,omitempty" bson:"font_family,omitempty"`
	BgColor      string    `json:"bg_color,omitempty" bson:"bg_color,omitempty"`
	BorderRadius int       `json:"border_radius,omitempty" bson:"border_radius,omitempty"`
	Position     XYObject  `json:"position,omitempty" bson:"position,omitempty"`
	Size         *HWObject `json:"size,omitempty" bson:"size,omitempty"`
	IsBold       *bool     `json:"is_bold,omitempty" bson:"is_bold,omitempty"`
	IsItalic     *bool     `json:"is_italic,omitempty" bson:"is_italic,omitempty"`
	Alignment    string    `json:"alignment,omitempty" bson:"alignment,omitempty"`
}

type InteractiveButton struct {
	ID                 string            `bson:"id" json:"id"`
	Type               string            `bson:"type" json:"type" validate:"required,oneof=image video gif"`
	AssetFileName      string            `bson:"asset_file_name,omitempty" json:"asset_file_name"`
	AssetUrl           string            `bson:"asset_url" json:"asset_url" validate:"omitempty,url"`
	CompressedAssetUrl string            `bson:"compressed_asset_url,omitempty" json:"compressed_asset_url,omitempty" validate:"omitempty,url"`
	Color              string            `bson:"color,omitempty" json:"color,omitempty"`
	MarkerId           string            `bson:"marker_id,omitempty" json:"marker_id,omitempty"`
	Position           ThreeDCoordinates `bson:"position,omitempty" json:"position,omitempty"`
	Scale              TwoDCoordinates   `bson:"scale,omitempty" json:"scale,omitempty"`
	MaskId             *int              `bson:"mask_id,omitempty" json:"mask_id,omitempty"` //(0 - Rectangle, 1 - Square,2 - Circle)
}

type VideoObject struct {
	Original    string `bson:"original" json:"original"`
	Compressed  string `bson:"compressed,omitempty" json:"compressed,omitempty"`
	Mask        string `bson:"mask,omitempty" json:"mask,omitempty"`
	Hls         string `bson:"hls,omitempty" json:"hls,omitempty"`
	Dash        string `bson:"dash,omitempty" json:"dash,omitempty"`
	WebM        string `bson:"webm,omitempty" json:"webm,omitempty"`
	MergeVideo  string `bson:"merge_video,omitempty" json:"merge_video,omitempty"`
	Orientation string `bson:"orientation,omitempty" json:"orientation,omitempty"`
}

type VoiceModule struct {
	FailureText         string         `bson:"failure_text,omitempty" json:"failure_text,omitempty"`
	FailureTime         int64          `bson:"failure_time,omitempty" json:"failure_time,omitempty"`
	JumpSegmentTime     int64          `bson:"jump_segment_time,omitempty" json:"jump_segment_time,omitempty"`
	PostSayingText      string         `bson:"post_saying_text,omitempty" json:"post_saying_text,omitempty"`
	PreSayingText       string         `bson:"pre_saying_text,omitempty" json:"pre_saying_text,omitempty"`
	TrySayingText       string         `bson:"try_saying_text,omitempty" json:"try_saying_text,omitempty"`
	VoicePermissionText string         `bson:"voice_permission_text,omitempty" json:"voice_permission_text,omitempty"`
	VoiceSegments       []VoiceSegment `bson:"voice_segments,omitempty" json:"voice_segments,omitempty"`
}
type VoiceSegment struct {
	EnableAutoListen bool          `bson:"enable_auto_listen,omitempty" json:"enable_auto_listen,omitempty"`
	ListenTime       int64         `bson:"listen_time,omitempty" json:"listen_time,omitempty"`
	SegmentId        string        `bson:"segment_id,omitempty" json:"segment_id,omitempty"`
	SuggestionPhrase []string      `bson:"suggestion_phrase,omitempty" json:"suggestion_phrase,omitempty"`
	VoiceOptions     []VoiceOption `bson:"voice_options,omitempty" json:"voice_options,omitempty"`
}

type VoiceOption struct {
	ID              string   `bson:"id" json:"id"`
	Phrases         []string `bson:"phrases,omitempty" json:"phrases,omitempty"`
	TargetSegmentId string   `bson:"target_segment_id,omitempty" json:"target_segment_id,omitempty"`
}

type CatalogueDetails struct {
	ProductID   string `bson:"product_id" json:"product_id"`
	Name        string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
	Currency    string `bson:"currency" json:"currency"`
	Price       string `bson:"price" json:"price"`
	ImageURL    string `bson:"image_url" json:"image_url"`
	ProductUrl  string `bson:"product_url" json:"product_url"`
	Category    string `bson:"category" json:"category"`
}

type VideoGeneration struct {
	Prompt     string `json:"prompt" bson:"prompt"`
	WorkflowID string `json:"workflow_id" bson:"workflow_id"`
	Status     string `json:"status" bson:"status"`
	VidoeUrl   string `json:"video_url" bson:"video_url"`
}
