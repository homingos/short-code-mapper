package dtos

type RemotionRequest struct {
	CompositionID string      `json:"composition_id" validate:"required"`
	EntryPoint    string      `json:"entrypoint" validate:"required"`
	InputProps    interface{} `json:"input_props" validate:"required"`
	Codec         string      `json:"codec" validate:"required"`
	UserId        string      `json:"user_id" validate:"required"`
	Mask          bool        `json:"mask"`
	ProjectID     string      `json:"project_id"`
}
