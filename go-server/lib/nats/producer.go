package nats

import (
	"context"
	"fmt"
	"time"

	redisStorage "github.com/homingos/campaign-svc/lib/redis"

	dao "github.com/homingos/campaign-svc/daos"
	"github.com/homingos/campaign-svc/lib/transaction"
	"github.com/homingos/campaign-svc/types/consts"
	"github.com/homingos/flam-go-common/authz"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type Client struct {
	nc          *nats.Conn
	js          jetstream.JetStream
	Stream      jetstream.Stream
	KV          jetstream.KeyValue
	expDao      dao.ExperienceDao
	redisClient *redisStorage.RedisClient
	campDao     dao.CampaignDao
	remotionDao dao.RemotionDao
	categoryDao dao.CategoryDao
	templateDao dao.TemplateDao
	txManager   transaction.TransactionManager
	fgaClient   *authz.OpenFGAClient
	lgr         *zap.SugaredLogger
}

func NewClient(expDao dao.ExperienceDao, campDao dao.CampaignDao, remDao dao.RemotionDao, redisClient *redisStorage.RedisClient, categoryDao dao.CategoryDao, templateDao dao.TemplateDao, txManager transaction.TransactionManager, fgaClient *authz.OpenFGAClient, lgr *zap.SugaredLogger) (*Client, error) {
	nc, err := nats.Connect(NatsServerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	stream, err := js.Stream(context.Background(), consts.WorkflowStreamName)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to get stream: %w", err)
	}

	kv, err := newStore(context.Background(), js, WorkflowStore)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to fetch/create store: %w", err)
	}

	return &Client{
		nc:          nc,
		js:          js,
		Stream:      stream,
		KV:          kv,
		expDao:      expDao,
		campDao:     campDao,
		remotionDao: remDao,
		redisClient: redisClient,
		categoryDao: categoryDao,
		templateDao: templateDao,
		txManager:   txManager,
		fgaClient:   fgaClient,
		lgr:         lgr,
	}, nil
}

func newStore(ctx context.Context, js jetstream.JetStream, bucket string) (jetstream.KeyValue, error) {
	kv, err := js.KeyValue(ctx, bucket)
	if err != nil {
		if err == jetstream.ErrBucketNotFound {
			kv, err = js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
				Bucket:  bucket,
				History: 5,
			})
			if err != nil {
				return nil, err
			}
			return kv, nil
		} else {
			return nil, err
		}
	}
	return kv, nil
}

func (cli *Client) PublishMsg(subject string, payload []byte) error {
	err := cli.nc.Publish(subject, payload)
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	fmt.Println("Published message to subject:", subject, "at:", time.Now())
	return nil
}
