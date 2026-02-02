package consts

const (

	// Thresholds
	SimilarityThreshold = 0.11

	// routing prefix
	RoutePrefix            = "campaign-svc"
	ResourceSvcRoutePrefix = "resource-svc"
	UserSvcRoutePrefix     = "userv2-svc"
	PaymentSvcRoutePrefix  = "payment-svc"
	SimilarityRoutePrefix  = "similarity"
	FdbRoutePrefix         = "fdb-svc"

	// Collections
	CampaignCollection          = "campaigns"
	CampaignGroupCollection     = "campaign_groups"
	UniversalCampaignCollection = "universal_campaigns"
	ExperienceCollection        = "experiences"
	EventCollection             = "events"
	DevicesCollection           = "devices"
	GalleryCollection           = "gallery"
	ProjectCollection           = "projects"
	PortfolioCollection         = "portfolio"
	EditLogsCollection          = "edit_logs"
	TemplateCollection          = "templates"
	RemotionCollection          = "remotion"
	CategoryCollection          = "category"

	// Status
	Created       = "CREATED"
	Processing    = "PROCESSING"
	Processed     = "PROCESSED"
	Draft         = "DRAFT"
	Portfolio     = "PORTFOLIO"
	AntiPortfolio = "ANTI_PORTFOLIO"
	Failed        = "FAILED"
	NoCredit      = "NO_CREDIT"
	TimedOut      = "TIMED_OUT"
	Cancelled     = "CANCELLED"

	// PresignedURL Expiry
	Expires = 15

	//Role
	SuperAdmin    = "SUPER_ADMIN"
	OrgSuperAdmin = "ORG_SUPER_ADMIN"

	// Default Scan Text
	DefaultCardScanText   = "SCAN THIS IMAGE"
	DefaultGroundScanText = "FIND SPOT TO PLACE THE EXPERIENCE"

	// Topics
	AssetsCompression = "assets-compression"
	DashGeneration    = "dash-generation"
	HlsGeneration     = "hls-generation"
	EditLogs          = "edit-logs"

	// Storage
	GCPStoragePrefix       = "https://storage.googleapis.com/"
	LOGO                   = "https://storage.googleapis.com/bucket-fi-production-apps-0672ab2d/logo/Watermark-extended.png"
	OldLogo                = "https://storage.googleapis.com/zingcam/gixib4gxg8vsjirl0bxdo27y.png"
	CopyRightLOGO          = "https://storage.googleapis.com/zingcam/original/images/ccaqa87y4zm6oec2wihdr8em.png"
	LogoWidth        int32 = 1

	// Trace
	SERVICE_NAME = "campaign-svc"
	PublicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA9Pl0EKLL9XDowA6DjNv/
J92WJTDhIPckgNW46q/B/9lFczKZoaS40r+1Mj1u8/eHPHSqkOvZYPYwXRBF03Ia
jOKSCkTW0Dgiym8XattNkqyJgzKDiVInw9SjH7y+niQo612tO6CnY0JLkXZC6pof
AnLQ+YLJhA7YXDDYf9nqoxWiHG8ZM0Ktej2JIHVa49B7cTtSbwy2algitimI1F5D
kb1r7taVm8axH565tDXAuZrDMT+9UFDU12QgAnSh5h0gIxkDVqhz9HliV5bYRoOy
6wQ7aUQKXuKqTDSe0zddfOuIbnqN63ux9FLl+dwL4P6jDIvBbo2PAnqEF1Y5clnJ
FwIDAQAB
-----END PUBLIC KEY-----`

	//CopyRight Content
	CopyRightContent = "© Flamingos Technologies Inc., 2024. All Rights Reserved."

	//NATS
	WorkflowStreamName           = "WORKFLOW"
	WorkflowStreamNameStreamName = "MEDIAPROCESSOR"
	AssetCompressionSubject      = "assetCompression"
	EditLogsStramName            = "EDITLOGS"
	SubjectHlsProcessor          = "MEDIAPROCESSOR.hls.process"
	SubjectDashProcessor         = "MEDIAPROCESSOR.dash.process"
	SubjectImageProcessor        = "MEDIAPROCESSOR.image.process"
	SubjectVideoProcessor        = "MEDIAPROCESSOR.video.process"
	SubjectOverlayProcessor      = "MEDIAPROCESSOR.overlay.process"
	WorkflowCompleteSubject      = "workflow.completed"
	WorkflowCancelSubject        = "workflow.cancel"
	EditLogsSubject              = "EDITLOGS.logs"
	ProductCatalogueSubject      = "product.catalogue.create"
	ProductCatalogueStreamName   = "PRODUCT_CATALOGUE"

	//overlay
	OverlayImage      = "IMAGE"
	OverlayTrasparent = "TRANSPARENT"
	OverlayBlur       = "BLUR"
	OverlayColor      = "COLOR"

	//Video task lenght
	VideoTaskLength = 4

	// Header keys
	Authorization   = "Authorization"
	ContentType     = "Content-Type"
	ApplicationJson = "application/json"

	// Campaign Groups
	CampaignGroupDemo = "Demo Campaigns"

	// Periods
	Day   = 24 * 60 * 60
	Week  = 7 * Day
	Month = 30 * Day

	//Credit Types
	CreditTypePrism              = "PRISM"
	CreditTypePrismInteractive   = "PRISM_INTERACTIVE"
	CreditTypeAlpha              = "ALPHA"
	CreditTypeAlphaInteractive   = "ALPHA_INTERACTIVE"
	CreditTypeAlphaSpatial       = "SPATIAL_ALPHA"
	CreditTypeFandomAI           = "FANDOM_AI"
	CreditTypeFandomVideo        = "FANDOM_VIDEO"
	CreditTypeSpatial            = "SPATIAL"
	CreditTypeSpatialInteractive = "SPATIAL_INTERACTIVE"
	CreditTypeThreeD             = "3D"

	// Credit Types
	PRISM               = "PRISM"
	PRISM_INTERACTIVE   = "PRISM_INTERACTIVE"
	ALPHA               = "ALPHA"
	ALPHA_INTERACTIVE   = "ALPHA_INTERACTIVE"
	SPATIAL_ALPHA       = "SPATIAL_ALPHA"
	FANDOM_AI           = "FANDOM_AI"
	FANDOM_VIDEO        = "FANDOM_VIDEO"
	SPATIAL             = "SPATIAL"
	SPATIAL_INTERACTIVE = "SPATIAL_INTERACTIVE"
	THREE_D             = "THREE_D"
)
