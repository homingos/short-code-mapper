package dtos

import (
	"github.com/homingos/campaign-svc/models"
)

type ProcessType string

const (
	TypeImage   ProcessType = "image"
	TypeVideo   ProcessType = "video"
	TypeHls     ProcessType = "hls"
	TypeDash    ProcessType = "dash"
	TypeOverlay ProcessType = "overlay"
)

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	Pending   TaskStatus = "pending"
	Running   TaskStatus = "running"
	Completed TaskStatus = "completed"
	Failed    TaskStatus = "failed"
	Cancelled TaskStatus = "cancelled"
	NoCredit  TaskStatus = "no_credit"
	TimedOut  TaskStatus = "timed_out"
)

// Task represents a single unit of work
type Task struct {
	WorkflowId string        `json:"workflow_id"`
	Id         string        `json:"id"`
	Subject    string        `json:"subject"`
	Body       WorkflowInput `json:"body"`
}

type Workflow struct {
	ID           string              `json:"id"`
	Tasks        map[string]Task     `json:"tasks"`
	Dependencies map[string][]string `json:"dependencies"`
	ReplySubject string              `json:"reply_subject"`
	Publish      bool                `json:"publish"`
}
type WorkflowFinalPubResult struct {
	WorkflowId    string               `json:"workflow_id"`
	Status        TaskStatus           `json:"status"`
	TaskResults   []TaskResult         `json:"task_results"`
	WorkflowError models.WorkflowError `json:"workflow_error,omitempty"`
	Publish       bool                 `json:"publish"`
}

type TaskResult struct {
	WorkflowId string                `json:"workflow_id"`
	TaskId     string                `json:"task_id"`
	Status     TaskStatus            `json:"status"`
	Payload    PostbackExperienceDto `json:"payload"`
}

type WorkflowInput struct {
	URL                      string                           `json:"url" validate:"required"`
	MaskURL                  string                           `json:"mask_url,omitempty"`
	SpawnImage               string                           `json:"spawn_image,omitempty"`
	Variant                  models.Variant                   `json:"variant,omitempty"`
	Overlay                  models.Overlay                   `json:"overlay,omitempty"`
	Template                 map[string]interface{}           `json:"template_details,omitempty"`
	ShortCode                string                           `json:"short_code,omitempty"`
	QRCode                   bool                             `json:"qrcode,omitempty"`
	ExperienceID             string                           `json:"experience_id,omitempty"`
	ScanUrl                  string                           `json:"scan_url,omitempty"`
	Publish                  bool                             `json:"publish,omitempty"`
	RemotionJob              *Remotion                        `json:"remotion,omitempty"`
	QRGenerate               bool                             `json:"qr_generate,omitempty"`
	QrID                     string                           `json:"qr_id,omitempty"`
	QrBGColor                string                           `json:"qr_bg_color,omitempty"`
	QrTextColor              string                           `json:"qr_text_color,omitempty"`
	QRCoordinates            []QRCoordinates                  `json:"qr_coordinates" validate:"required,dive"`
	QRImageURl               string                           `json:"qr_image_url" validate:"required"`
	Stitch                   bool                             `json:"stitch,omitempty"`
	Segments                 []SegmentVideo                   `json:"segments,omitempty"`
	GenerateGreenScreen      bool                             `json:"generate_green_screen,omitempty"`
	ImageVectorLLMProductJob *models.ImageVectorLLMProductJob `json:"image_vector_llm_product_job,omitempty"`
	AlphaVideoJob            *models.AlphaVideoJob            `json:"alpha_video_job,omitempty"`
	GenStudioJob             *models.FalVideoJob              `json:"genstudio_job,omitempty"`
}

type Remotion struct {
	CompositionId string      `json:"compositionId,omitempty"`
	EntryPoint    string      `json:"entryPoint,omitempty"`
	InputProps    interface{} `json:"inputProps,omitempty"`
	Codec         string      `json:"codec,omitempty"`
	Mask          bool        `json:"mask,omitempty"`
}

type WorkflowCancel struct {
	WorkflowId string `json:"workflow_id"`
}

type SegmentVideo struct {
	OriginalURL string `json:"original_url"`
	MaskURL     string `json:"mask_url"`
	MarkedID    string `json:"marker_id"`
}
