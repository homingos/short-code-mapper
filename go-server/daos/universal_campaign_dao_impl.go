package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/campaign-svc/models"
	"github.com/homingos/campaign-svc/types/consts"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type UniversalCampaignDaoImpl struct {
	lgr *zap.SugaredLogger
	db  *mongo.Database
}

func createUniversalCampaignIndexes(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := db.Collection(consts.UniversalCampaignCollection)
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "short_code", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := coll.Indexes().CreateMany(ctx, indexes, opts)
	if err != nil {
		fmt.Println(err)
	}
}

func NewUniversalCampaignDao(lgr *zap.SugaredLogger, db *mongo.Database) *UniversalCampaignDaoImpl {
	createUniversalCampaignIndexes(db)
	return &UniversalCampaignDaoImpl{lgr, db}
}

func (impl *UniversalCampaignDaoImpl) buildMongoUpdateMap(universalCampaignReq *dtos.UpdateUniversalCampaignRequestDto) bson.M {
	updateMap := bson.M{}
	if universalCampaignReq.Name != "" {
		updateMap["name"] = universalCampaignReq.Name
	}
	if universalCampaignReq.CampaignShortCode != "" {
		updateMap["campaign_short_code"] = universalCampaignReq.CampaignShortCode
	}
	if universalCampaignReq.AppType != "" {
		updateMap["app_type"] = universalCampaignReq.AppType
	}
	if universalCampaignReq.IsActive != nil {
		updateMap["is_active"] = universalCampaignReq.IsActive
	}
	if universalCampaignReq.Share != nil {
		if universalCampaignReq.Share.Title != "" {
			updateMap["share.title"] = universalCampaignReq.Share.Title
		}
		if universalCampaignReq.Share.Image != "" {
			updateMap["share.image"] = universalCampaignReq.Share.Image
		}
		if universalCampaignReq.Share.OgImage != "" {
			updateMap["share.og_image"] = universalCampaignReq.Share.OgImage
		}
	}
	updateMap["updated_at"] = time.Duration(time.Now().UnixMilli())
	return updateMap
}

func (impl *UniversalCampaignDaoImpl) GetUniversalCampaignByShortCode(shortCode string) (*models.UniversalCampaign, error) {
	coll := impl.db.Collection(consts.UniversalCampaignCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{"short_code": shortCode, "is_active": true}
	result := coll.FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	var universalCampaign models.UniversalCampaign
	if err := result.Decode(&universalCampaign); err != nil {
		return nil, err
	}
	return &universalCampaign, nil
}

func (impl *UniversalCampaignDaoImpl) CreateUniversalCampaignDao(ctx context.Context, universalCampaignReq *models.UniversalCampaign) (*models.UniversalCampaign, error) {
	coll := impl.db.Collection(consts.UniversalCampaignCollection)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err := coll.InsertOne(ctx, universalCampaignReq)
	if err != nil {
		return nil, err
	}
	return universalCampaignReq, nil
}

func (impl *UniversalCampaignDaoImpl) UpdateUniversalCampaignDao(ctx context.Context, universalCampaignID primitive.ObjectID, universalCampaignReq *dtos.UpdateUniversalCampaignRequestDto) (*models.UniversalCampaign, error) {
	coll := impl.db.Collection(consts.UniversalCampaignCollection)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	updateMap := impl.buildMongoUpdateMap(universalCampaignReq)
	filter := bson.M{"_id": universalCampaignID, "client_id": universalCampaignReq.ClientID, "is_active": true}
	update := bson.M{"$set": updateMap}
	result := coll.FindOneAndUpdate(ctx, filter, update)
	if result.Err() != nil {
		return nil, result.Err()
	}
	var universalCampaign models.UniversalCampaign
	if err := result.Decode(&universalCampaign); err != nil {
		return nil, err
	}
	return &universalCampaign, nil

}

func (impl *UniversalCampaignDaoImpl) GetUniversalClientCampaignDao(ctx context.Context, clientID primitive.ObjectID) ([]models.UniversalCampaign, error) {
	coll := impl.db.Collection(consts.UniversalCampaignCollection)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	filter := bson.M{"client_id": clientID, "is_active": true}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	var universalCampaigns []models.UniversalCampaign
	if err = cursor.All(ctx, &universalCampaigns); err != nil {
		return nil, err
	}
	return universalCampaigns, nil
}
