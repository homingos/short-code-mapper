package dao

import (
	"context"
	"time"

	"github.com/homingos/campaign-svc/dtos"
	"github.com/homingos/campaign-svc/models"
	"github.com/homingos/campaign-svc/types/consts"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type RemotionDaoImpl struct {
	lgr *zap.SugaredLogger
	db  *mongo.Database
}

func NewRemotionDao(lgr *zap.SugaredLogger, db *mongo.Database) *RemotionDaoImpl {
	return &RemotionDaoImpl{lgr, db}
}

func (impl *RemotionDaoImpl) CreateRemotionRequest(req *dtos.RemotionRequest, workflowId string) (string, error) {
	// Convert string UserID to ObjectID
	userId, err := primitive.ObjectIDFromHex(req.UserId)
	if err != nil {
		impl.lgr.Errorw("Failed to convert user ID", "error", err)
		return "", err
	}

	// Create a new Remotion instance
	remotion := &models.Remotion{
		ID:         primitive.NewObjectID(),
		WorkflowID: workflowId,
		UserID:     userId,
		CreatedAt:  time.Duration(time.Now().UnixMilli()),
		UpdatedAt:  time.Duration(time.Now().UnixMilli()),
		Status:     consts.Processing,
	}

	if req.ProjectID != "" {
		ProjectObjId, err := primitive.ObjectIDFromHex(req.ProjectID)
		if err != nil {
			impl.lgr.Errorw("Failed to convert project ID", "error", err)
			return "", err
		}
		remotion.ProjectID = ProjectObjId
	}
	// Insert into MongoDB
	_, err = impl.db.Collection(consts.RemotionCollection).InsertOne(context.Background(), remotion)
	if err != nil {
		impl.lgr.Errorw("Failed to insert remotion", "error", err)
		return "", err
	}

	// Return the ID as a string
	return remotion.ID.Hex(), nil
}

func (impl *RemotionDaoImpl) UpdateRemotionResult(ID, videoUrl, maskedUrl string) error {
	ctx := context.Background()

	filter := map[string]interface{}{"workflow_id": ID}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":     consts.Processed,
			"video_url":  videoUrl, // Assuming the workflow ID is used as or maps to the video URL
			"mask_url":   maskedUrl,
			"updated_at": time.Duration(time.Now().UnixMilli()),
		},
	}

	result, err := impl.db.Collection(consts.RemotionCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		impl.lgr.Errorw("Failed to update remotion document", "error", err, "workflow_id", ID)
		return err
	}

	if result.MatchedCount == 0 {
		impl.lgr.Warnw("No remotion document found with workflow ID", "workflow_id", ID)
		return mongo.ErrNoDocuments
	}

	impl.lgr.Infow("Remotion document updated successfully", "workflow_id", ID)
	return nil
}

func (impl *RemotionDaoImpl) GetRemotion(ID primitive.ObjectID) (*[]models.Remotion, error) {
	filter := bson.M{}

	if ID != primitive.NilObjectID {
		filter = bson.M{"_id": ID}
	}

	cursor, err := impl.db.Collection(consts.RemotionCollection).Find(context.Background(), filter)
	if err != nil {
		impl.lgr.Errorw("Failed to find remotion document", "error", err, "ID", ID)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var remotions []models.Remotion
	if err = cursor.All(context.Background(), &remotions); err != nil {
		impl.lgr.Errorw("Failed to decode remotion documents", "error", err)
		return nil, err
	}

	return &remotions, nil
}
