package dao

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type AssetsDaoImpl struct {
	lgr *zap.SugaredLogger
	db  *mongo.Database
}

func createAssetIndexes(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	coll := db.Collection("assets")
	indexes := []mongo.IndexModel{{
		Keys: bson.D{
			{Key: "type", Value: 1},
			{Key: "url", Value: 1},
			{Key: "masked_url", Value: 1},
		},
	}}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	coll.Indexes().CreateMany(ctx, indexes, opts)
}

func NewAssetDao(lgr *zap.SugaredLogger, db *mongo.Database) *AssetsDaoImpl {
	createAssetIndexes(db)
	return &AssetsDaoImpl{
		lgr: lgr,
		db:  db,
	}
}
