package redisStorage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/homingos/campaign-svc/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisClient struct {
	lgr *zap.SugaredLogger
	cli *redis.Client
}

func NewRedisClient(lgr *zap.SugaredLogger) *RedisClient {
	connectionURL := config.GetAppConfig().REDIS.URI
	client := redis.NewClient(&redis.Options{
		Addr:            connectionURL,
		Password:        "",
		DB:              0,
		DisableIdentity: true,
	})
	redisTracingErr := redisotel.InstrumentTracing(client)
	if redisTracingErr != nil {
		lgr.Info("Failed to instrument Redis client for tracing", zap.Error(redisTracingErr))
	}
	return &RedisClient{lgr: lgr, cli: client}
}

func (redisCli *RedisClient) SetCampaignExperiences(campaignID string, val interface{}, universalCampaign bool, category bool) error {
	var key string
	if universalCampaign {
		key = universalCampaignKey(campaignID)
	} else if category {
		key = categoryExperiencesKey(campaignID)
	} else {
		key = campaignExperiencesKey(campaignID)
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	barr, err := json.Marshal(&val)
	if err != nil {
		return err
	}
	expiry := 7 * 24 * time.Hour
	err = redisCli.cli.Set(ctx, key, barr, expiry).Err()
	return err
}

func (redisCli *RedisClient) GetCampaignExperiences(ctx context.Context, campaignID string, universalCampaign bool, category bool) (interface{}, error) {
	var key string
	if universalCampaign {
		key = universalCampaignKey(campaignID)
	} else if category {
		key = categoryExperiencesKey(campaignID)
	} else {
		key = campaignExperiencesKey(campaignID)
	}
	var result interface{}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	val, redErr := redisCli.cli.Get(ctx, key).Result()
	if redErr == redis.Nil {
		return nil, nil // Key not found
	}

	// Refresh TTL to 7 days on successful read
	go func() {
		expiry := 7 * 24 * time.Hour
		if err := redisCli.cli.Expire(context.TODO(), key, expiry).Err(); err != nil {
			// Log but don't fail the call if TTL update fails
			redisCli.lgr.Warnf("Failed to update TTL for key %s: %v", key, err)
		}
	}()

	_, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("redis-client").Start(
		ctx,
		"RedisMsg.Unmarshall",
	)
	defer span.End()

	if err := json.Unmarshal([]byte(val), &result); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return result, nil
}

func (redisCli *RedisClient) ExpireCampaignExperiences(campaignID string, universalCampaign bool, category bool) error {
	var key string
	if universalCampaign {
		key = universalCampaignKey(campaignID)
	} else if category {
		key = categoryExperiencesKey(campaignID)
	} else {
		key = campaignExperiencesKey(campaignID)
	}
	redisCli.lgr.Info("expiring key: ", key)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := redisCli.cli.Expire(ctx, key, 0).Result()
	if err != nil {
		redisCli.lgr.Errorf("Error expiring key: %v", key)
		return err
	}
	return nil
}
