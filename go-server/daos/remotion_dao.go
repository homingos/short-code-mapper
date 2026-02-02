package dao

import (
	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/campaign-svc/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RemotionDao interface {
	CreateRemotionRequest(req *dtos.RemotionRequest, workflowId string) (string, error)
	UpdateRemotionResult(ID, videoUrl, maskedUrl string) error
	GetRemotion(ID primitive.ObjectID) (*[]models.Remotion, error)
}
