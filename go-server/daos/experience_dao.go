package dao

import (
	"context"

	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/campaign-svc/models"
	"github.com/homingos/flam-go-common/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ExperienceDao interface {
	CreateExperienceDao(expCreationDto *dtos.ExperienceCreationDto) (*models.MediaProcess, error)
	CreateBulkExperienceDao(sessionCtx *mongo.SessionContext, Experiences []*models.Experience) (string, error)
	ResetExperienceDao(expResetDto *dtos.ExperienceResetDto, EditedBy *models.User) (*models.Experience, *errors.AppError)
	DuplicateExperienceWithNewCampaignIdDao(ctx context.Context, model *models.Experience) (*models.Experience, error)
	GetExperienceDao(objID primitive.ObjectID) (interface{}, error)
	UpdateExperienceDao(objID primitive.ObjectID, updates map[string]interface{}, OptionalVariant ...dtos.ExperienceUpdateDto) (*dtos.UpdateResponseDto, error)
	PostbackExperienceDao(objID primitive.ObjectID, updateMap map[string]interface{}) (*dtos.PostbackResponseDto, error)
	GetElementBannerVariant(bannerDto *dtos.GetBannerDto) int
	// IsImageHashUnique(uniqHash *dtos.ImageHashUnique) (bool, error)
	GetExperienceByID(objID primitive.ObjectID) (*models.Experience, error)
	GetExperienceByCampaignID(campaignID primitive.ObjectID) (*models.Experience, error)
	GetImageUrls(uniqueImgDto *dtos.UniqueImageDto) ([]string, error)
	GetCampaignExperiencesDao(campaignID string) (*dtos.CampaignExperienceDto, error)
	GetLogsDao(filter *dtos.EditLogsFilterDto, optionalParams ...string) (map[string]interface{}, error)
	UpdateExperienceAssetsDao(WfResult dtos.WorkflowFinalPubResult) error
	UpdateExperienceAssetsWithQrImageDao(WfResult dtos.WorkflowFinalPubResult) error
	UpdateExperienceWorkflowData(ID primitive.ObjectID, WorkflowID string, StitchWfID string, TaskLength int32) error
	GetCampaignExperiencesStatus(CampaignId string) (*bool, error)
	ConsumeCreditAndPublishCampaign(CampaignID string, EditedBy *models.User, ExperienceId primitive.ObjectID) (*bool, error)
	GetExpByProductId(productId []string, clientId string) ([]string, error)
	UpdateRegenerateExperienceAssetsDao(WfResult dtos.WorkflowFinalPubResult) error
}
