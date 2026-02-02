package dtos

type ClientExpiryRespDto struct {
	Validity Validity `json:"expiry_duration"`
	Name     string   `json:"name"`
}

type AdjustEscrowCreditsResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		CreditAllowanceID string `json:"credit_allowance_id"`
	} `json:"data"`
	Error bool `json:"error"`
}

type ConsumeCreditsResponse struct {
	Status  int                 `json:"status"`
	Message string              `json:"message"`
	Data    ConsumeCreditResDto `json:"data"`
	Error   bool                `json:"error"`
}

type Validity struct {
	Unit  string `json:"unit"`
	Value int    `json:"value"`
}

type GatewayResponse[T any] struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Error   bool   `json:"error"`
}

type ExpiryDurationWithUserName struct {
	Duration int64  `json:"duration"`
	UserName string `json:"user_name"`
}

type ConsumeCreditResDto struct {
	Balance    int32  `json:"balance"`
	Unlimited  bool   `json:"unlimited"`
	CreditType string `json:"credit_type"`
}

type CheckEscrowDto struct {
	Balance      int32  `json:"balance"`
	Unlimited    bool   `json:"unlimited"`
	ExpType      string `json:"exp_type"`
	EscrowCredit int32  `json:"escrow_credit"`
}

type CheckEscrowResDto struct {
	Status  int              `json:"status"`
	Message string           `json:"message"`
	Data    []CheckEscrowDto `json:"data"`
	Error   bool             `json:"error"`
}

type CreditVisibilityDto struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		FeatureFlags FeatureFlags `json:"feature_flags"`
	} `json:"data"`
	Error bool `json:"error"`
}

type FeatureFlags struct {
	CreditVisibility CreditsVisibility `json:"credits_visibility"`
}

type CreditsVisibility struct {
	PRISM               *bool `bson:"PRISM" json:"PRISM"`
	PRISM_INTERACTIVE   *bool `bson:"PRISM_INTERACTIVE" json:"PRISM_INTERACTIVE"`
	ALPHA               *bool `bson:"ALPHA" json:"ALPHA"`
	ALPHA_INTERACTIVE   *bool `bson:"ALPHA_INTERACTIVE" json:"ALPHA_INTERACTIVE"`
	SPATIAL_ALPHA       *bool `bson:"SPATIAL_ALPHA" json:"SPATIAL_ALPHA"`
	FANDOM_AI           *bool `bson:"FANDOM_AI" json:"FANDOM_AI"`
	FANDOM_VIDEO        *bool `bson:"FANDOM_VIDEO" json:"FANDOM_VIDEO"`
	SPATIAL             *bool `bson:"SPATIAL" json:"SPATIAL"`
	SPATIAL_INTERACTIVE *bool `bson:"SPATIAL_INTERACTIVE" json:"SPATIAL_INTERACTIVE"`
	THREE_D             *bool `bson:"THREE_D" json:"THREE_D"`
}
