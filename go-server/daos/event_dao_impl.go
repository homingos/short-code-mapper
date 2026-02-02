package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/homingos/campaign-svc/dtos"
	common "github.com/homingos/campaign-svc/lib/posthog"
	"github.com/homingos/campaign-svc/types/consts"
	"github.com/posthog/posthog-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type EventDaoImpl struct {
	lgr *zap.SugaredLogger
	db  *mongo.Database
}

func NewEventDao(lgr *zap.SugaredLogger, db *mongo.Database) *EventDaoImpl {
	return &EventDaoImpl{lgr, db}
}

func (impl *EventDaoImpl) ProcessEventDao(eventReq dtos.Event, sourceIp string) error {
	posthogEvent, posthogEventErr := eventReq.ToPostHogsEvent()
	if posthogEventErr != nil {
		return posthogEventErr
	}
	event, posthogError := impl.emitInstantPosthogEvent(posthogEvent, sourceIp)
	if posthogError != nil {
		return posthogError
	}
	fmt.Print("OkEvent DID: " + event.Properties.SetOnce.DeviceID + " Name: " + event.Properties.Name)
	dbErr := impl.emitInstantDbEvent(event)
	if dbErr != nil {
		return dbErr
	}
	return nil
}

func (impl *EventDaoImpl) emitInstantPosthogEvent(event dtos.PostHogEvent, sourceIp string) (dtos.PostHogEvent, error) {
	postHogClient := common.GetPostHogClient()
	event.Properties.IP = sourceIp
	eventPropBytes, err := json.Marshal(event.Properties)
	if err != nil {
		return dtos.PostHogEvent{}, err
	}
	var data = make(map[string]interface{})
	e := json.Unmarshal(eventPropBytes, &data)
	if e != nil {
		return dtos.PostHogEvent{}, e
	}
	posthogError := postHogClient.Enqueue(posthog.Capture{
		DistinctId: event.DistinctId,
		Event:      event.Event,
		Properties: data,
	})
	if posthogError != nil {
		return dtos.PostHogEvent{}, posthogError
	}

	return event, posthogError
}

func (impl *EventDaoImpl) emitInstantDbEvent(event dtos.PostHogEvent) error {
	var data map[string]interface{}
	dataBytes, _ := json.Marshal(event.Properties)
	json.Unmarshal(dataBytes, &data)
	fmt.Print(event.Properties.Set)
	data["set"] = event.Properties.Set
	data["set_once"] = event.Properties.SetOnce
	data["ip"] = event.Properties.IP
	data["created_at"] = time.Now()
	data["updated_at"] = time.Now()

	delete(data, "$set")
	delete(data, "$set_once")
	delete(data, "$ip")
	coll := impl.db.Collection(consts.EventCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := coll.InsertOne(ctx, data)
	if err != nil {
		return err
	}
	if event.Properties.Set.PushToken != "" {
		upsert := true
		coll = impl.db.Collection(consts.DevicesCollection)
		_, err := coll.UpdateOne(
			ctx,
			bson.M{"_id": event.DistinctId},
			bson.M{
				"$set": bson.M{
					"Platform":        event.Properties.SetOnce.Platform,
					"DeviceModel":     event.Properties.SetOnce.Model,
					"OsVersion":       event.Properties.SetOnce.OS,
					"ar_core_enabled": event.Properties.Set.NativeAR,
					"device_id":       event.DistinctId,
					"push_token":      event.Properties.Set.PushToken,
					"ad_id":           event.Properties.Set.AdID,
					"updated_at":      time.Now(),
				},
				"$setOnInsert": bson.M{
					"created_at": time.Now(),
				},
			},
			&options.UpdateOptions{Upsert: &upsert},
		)
		if err != nil {
			fmt.Print("error in db", err.Error())
			return err
		}
	}

	return nil
}
