package dao

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/campaign-svc/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CampaignDao interface {
	GetClientCampaignDao(ClientId string, filterDto *dtos.CampaignFilterDto) (interface{}, error)
	GetClientCampaignV2Dao(ClientId string, filterDto *dtos.CampaignFilterDto) (interface{}, error)
	GetAllCampaignDao(GetCampaignDto *dtos.GetCampaignDto) (interface{}, error)
	GetCampaignGroupsDao(clientID string) ([]*models.CampaignGroup, error)
	CreateCampaignDao(ctx context.Context, sessionCtx *mongo.SessionContext, campaignReqDto *dtos.CampaignRequestDto) (*models.Campaign, error)
	CreateCampaignV2Dao(ctx context.Context, sessionCtx *mongo.SessionContext, campaignReqDto *dtos.CampaignRequestV2Dto) (*models.Campaign, error)
	CreateCampaignGroupDao(ctx context.Context, sessionCtx *mongo.SessionContext, dto *dtos.CampaignGroupCreateDto) (*models.CampaignGroup, error)
	CreateBulkCampaignDao(sessionCtx *mongo.SessionContext, bulkCampaignReqDto []*models.Campaign) (string, error)
	GetCampaignDao(ID string, clientId string) (*models.Campaign, error)
	UpdateCampaignDao(objID primitive.ObjectID, campaignUpdateDto *dtos.CampaignUpdateDto) (*models.Campaign, error)
	DeleteCampaignDao(objID primitive.ObjectID) (*models.Campaign, error)
	GetCampaignExperiencesDao(ctx context.Context, ID string, optionalParams ...string) (interface{}, error)
	GetAppExperiencesDao(ctx context.Context, ID string) (interface{}, error)
	PostbackCampaignDao(shortCode string, updateMap map[string]interface{}) (*models.Campaign, error)
	CheckProjectExists(ID string) (*models.Project, error)
	FdbCampaignDao(expID, fdbUrl string) error
	GetExperincesCampaignDao(campaignID string) ([]models.Experience, error)
	UpdateExperienceCreditDeduct(ExpId primitive.ObjectID, CreditAllowanceID string) error
	GetCampaignDaoByShortCode(ID string) (*models.Campaign, error)
	GetClientDemoCampaignsCount(clientID string) (int64, error)
	CheckCampaingsByName(Names []string, clientId primitive.ObjectID) ([]string, error)
	GetBulkCampaignDao(clientID, shortCode string, page, pageSize int, publish *bool) (interface{}, error)
	GetCampaignBySourceCodeDao(shortCode string) ([]models.Campaign, error)
	GetCampaignWithExperienceBySourceCodeDao(shortcode string, pending bool) (*dtos.CampaignsExperiences, error)
	GetClientCampaignsDAO(clientID primitive.ObjectID) ([]dtos.ClientCampaignsInfo, error)
	GetCampaignByShortCodesDao(shortCodes []string, clientId string) ([]models.Campaign, error)
	GetShortcodesByMilvusRefID(IDs []string) ([]string, error)
}
