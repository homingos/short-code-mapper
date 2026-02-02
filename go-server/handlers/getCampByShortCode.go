package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"github.com/homingos/flam-go-common/errors"
	redisStorage "github.com/homingos/campaign-svc/lib/redis"
	"go.uber.org/zap"
	dao "github.com/homingos/campaign-svc/daos"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/homingos/campaign-svc/lib/transaction"
	"github.com/homingos/flam-go-common/authz"
	"github.com/homingos/campaign-svc/lib/nats"
)

type CategorySvcImpl struct {
	lgr          *zap.SugaredLogger
	redisClient  *redisStorage.RedisClient
	MilvusClient client.Client
	categoryDao  dao.CategoryDao
	campaignDao  dao.CampaignDao
	txManager    transaction.TransactionManager
	fgaClient    *authz.OpenFGAClient
	expDao       dao.ExperienceDao
	templateDao  dao.TemplateDao
	natsClient   *nats.Client
	milvusDao    dao.MilvusDao
}

func NewCategorySvc(
	lgr *zap.SugaredLogger,
	redisClient *redisStorage.RedisClient,
	milvusClient client.Client,
	categoryDao dao.CategoryDao,
	campaignDao dao.CampaignDao,
	expDao dao.ExperienceDao,
	templateDao dao.TemplateDao,
	natsClient *nats.Client,
	txManager transaction.TransactionManager,
	fgaClient *authz.OpenFGAClient,
	milvusDao dao.MilvusDao,
) *CategorySvcImpl {
	return &CategorySvcImpl{
		lgr:          lgr,
		categoryDao:  categoryDao,
		campaignDao:  campaignDao,
		redisClient:  redisClient,
		MilvusClient: milvusClient,
		expDao:       expDao,
		templateDao:  templateDao,
		natsClient:   natsClient,
		txManager:    txManager,
		fgaClient:    fgaClient,
		milvusDao:    milvusDao,
	}
}

func (impl *CategorySvcImpl) GetCategoriesBySiteCodeSvc(ctx context.Context, siteCode string, text string) (interface{}, *errors.AppError) {
	data, err := impl.redisClient.GetCampaignExperiences(ctx, siteCode, false, true)
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}
	var ShortCodes []string
	if text != "" || data == nil {
		//make embeding api call: returns float array
		if text != "" {
			embeddings, err := GetEmbeddings(text)
			if err != nil {
				return nil, errors.InternalServerError(err.Error())
			}
			milvusDocs, err := impl.milvusDao.Search(ctx, embeddings, siteCode)
			if err != nil {
				return nil, errors.InternalServerError(err.Error())
			}

			var ids []string
			for _, doc := range milvusDocs {
				ids = append(ids, doc.ID)
			}

			// filter campaigns shortcodes from campaigns collection using campaign.milvus_ref_id.
			if len(ids) > 0 {
				shortCodes, err := impl.campaignDao.GetShortcodesByMilvusRefID(ids) // todo: change here input
				if err != nil {
					return nil, errors.InternalServerError(err.Error())
				}
				ShortCodes = shortCodes
			}

		}

		categoryData, err := impl.categoryDao.GetCategoriesBySiteCodeDao(ctx, siteCode, ShortCodes, text)
		if err != nil {
			return nil, errors.InternalServerError(err.Error())
		}
		if categoryData == nil {
			return nil, errors.BadRequest(fmt.Sprintf("No categories found for this site code: %s", siteCode))
		}
		data = categoryData
		barr, err := json.Marshal(data)
		if err != nil {
			return data, nil
		}
		replacedStr := strings.ReplaceAll(string(barr), "storage.googleapis.com/bucket-fi-production-apps-0672ab2d", "instant.cdn.flamapis.com")
		replacedStr = strings.ReplaceAll(replacedStr, "storage.googleapis.com/zingcam/", "zingcam.cdn.flamapp.com/")

		err = json.Unmarshal([]byte(replacedStr), &data)
		if err != nil {
			return data, nil
		}
		if text == "" {
			go impl.redisClient.SetCampaignExperiences(siteCode, data, false, true)
		}
	}
	return data, nil
}

func (impl *CategorySvcImpl) InvalidateCategorySvc(ctx context.Context, siteCode string) (string, *errors.AppError) {
	data, err := impl.redisClient.GetCampaignExperiences(ctx, siteCode, false, true)
	if err != nil {
		return "", errors.InternalServerError(err.Error())
	}
	if data == nil {
		return "Category not found in Cache:" + siteCode, nil
	}
	go impl.redisClient.ExpireCampaignExperiences(siteCode, false, true)
	return "cache invalidated successfully for category:" + siteCode, nil
}
