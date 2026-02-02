package publisher

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/pubsub"
	"github.com/homingos/campaign-svc/config"
	"go.uber.org/zap"
)

type GCPPubSubClient struct {
	lgr *zap.SugaredLogger
	cli *pubsub.Client
}

func NewClient(lgr *zap.SugaredLogger) (*GCPPubSubClient, error) {
	conf := config.GetAppConfig()
	projectID := conf.GCP.ProjectID
	client, err := pubsub.NewClient(context.Background(), projectID)
	if err != nil {
		return nil, err
	}
	return &GCPPubSubClient{lgr: lgr, cli: client}, nil
}

func (client *GCPPubSubClient) PublishMessage(topic string, message interface{}) error {
	messageStr, err := json.Marshal(&message)
	if err != nil {
		client.lgr.Error("marshalling error: ", err)
		return err
	}
	m := pubsub.Message{Data: messageStr}
	id, err := client.cli.Topic(topic).Publish(context.Background(), &m).Get(context.Background())
	if err != nil {
		return err
	}
	client.lgr.Info("message ID: ", id)
	return nil
}
