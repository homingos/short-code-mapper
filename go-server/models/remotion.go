package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Remotion struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id"`
	VideoURL   string             `json:"video_url,omitempty" bson:"video_url,omitempty"`
	MaskURL    string             `json:"mask_url,omitempty" bson:"mask_url,omitempty"`
	WorkflowID string             `json:"workflow_id" bson:"workflow_id"`
	UserID     primitive.ObjectID `json:"user_id" bson:"user_id"`
	CreatedAt  time.Duration      `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Duration      `json:"updated_at" bson:"updated_at"`
	Status     string             `json:"status" bson:"status"`
	ProjectID  primitive.ObjectID `json:"project_id,omitempty" bson:"project_id,omitempty"`
}
