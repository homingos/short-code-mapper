package dtos

import (
	"github.com/go-playground/validator/v10"
	"github.com/homingos/campaign-svc/models"
)

type Event struct {
	Name   string                 `json:"name" validate:"required"`
	Type   models.Type            `json:"type" validate:"required,oneof=MODULE INTERACTION SCAN VIDEO RECORDING"`
	Device models.DeviceDetails   `json:"device" validate:"required"`
	App    models.AppDetails      `json:"app,omitempty"`
	Asset  models.AssetDetails    `json:"asset"`
	Action models.ActionDetails   `json:"action" validate:"omitempty"`
	Track  models.TrackingDetails `json:"track,omitempty"`
	Meta   models.Meta            `json:"meta,omitempty"`
}

type DevicePropDto2 struct {
	DeviceID      string          `json:"device_id" binding:"required"`
	Platform      models.Platform `json:"Platform" binding:"required" validate:"required,oneof=ANDROID IOS"`
	DeviceModel   string          `json:"DeviceModel" `
	OsVersion     string          `json:"OsVersion"`
	ArCoreEnabled bool            `json:"ar_core_enabled" binding:"required"`
	BuildVersion  string          `json:"build_version"  binding:"required"`
	Runtime       string          `json:"runtime,omitempty" validate:"omitempty,oneof=WEB NATIVE"`
	AdId          string          `json:"ad_id,omitempty"`
	PushToken     string          `json:"push_token,omitempty"`
}

type PostHogEvent struct {
	DistinctId string                 `json:"distinct_id"`
	Event      string                 `json:"event"`
	Properties PostHogEventProperties `json:"properties"`
}

// PostHog event.
type PostHogEventProperties struct {
	SetOnce        PostHogSetOnceProperties `json:"$set_once"`
	Set            PostHogSetProperties     `json:"$set,omitempty"`
	ActionType     string                   `json:"action_type,omitempty"`
	ActionName     string                   `json:"action_name,omitempty"`
	ActionAction   string                   `json:"action,omitempty"`
	ActionURL      string                   `json:"action_url,omitempty"`
	AssetID        string                   `json:"asset_id,omitempty"`
	AssetType      string                   `json:"asset_type,omitempty"`
	AssetURL       string                   `json:"asset_url,omitempty"`
	Runtime        models.Runtime           `json:"runtime,omitempty"`
	Name           string                   `json:"name,omitempty"`
	IP             string                   `json:"$ip,omitempty"`
	Duration       int                      `json:"duration,omitempty"`
	PublishState   string                   `json:"publish_state,omitempty"`
	ExperienceId   string                   `json:"experience_id,omitempty"`
	ExperienceType string                   `json:"experience_type,omitempty"`
}

type PostHogSetOnceProperties struct {
	DeviceID string          `json:"device_id"`
	Platform models.Platform `json:"platform" validate:"required,oneof=ANDROID IOS"`
	Model    string          `json:"device_model"`
	OS       string          `json:"device_os"`
}

type PostHogSetProperties struct {
	AdID       string `json:"ad_id,omitempty"`
	PushToken  string `json:"push_token,omitempty"`
	AppID      string `json:"app_id,omitempty"`
	AppVersion string `json:"app_version,omitempty"`
	NativeAR   bool   `json:"native_ar"`
	AppType    string `json:"app_type"`
}

func (e *Event) Validate() error {
	validate := validator.New()
	return validate.Struct(e)
}

func (e *Event) ToPostHogsEvent() (PostHogEvent, error) {
	if err := e.Validate(); err != nil {
		return PostHogEvent{}, err.(validator.ValidationErrors)
	}
	var publishState string
	if !e.Asset.PublishState {
		publishState = "false"
	} else if e.Asset.PublishState {
		publishState = "true"
	}

	return PostHogEvent{
		DistinctId: e.Device.ID,
		Event:      string(e.Type),
		Properties: PostHogEventProperties{
			SetOnce: PostHogSetOnceProperties{
				DeviceID: e.Device.ID,
				Platform: e.Device.Platform,
				Model:    e.Device.Model,
				OS:       e.Device.OS,
			},
			Set: PostHogSetProperties{
				AdID:       e.Track.AdId,
				PushToken:  e.Track.PushToken,
				AppID:      e.App.AppId,
				AppVersion: e.App.AppVersion,
				NativeAR:   e.Device.NativeAR,
				AppType:    e.App.AppType,
			},
			Name:           e.Name,
			ActionType:     e.Action.Type,
			ActionName:     e.Action.Name,
			ActionAction:   e.Action.Action,
			ActionURL:      e.Action.URL,
			AssetID:        e.Asset.ID,
			AssetType:      e.Asset.Type,
			AssetURL:       e.Asset.URL,
			Runtime:        e.Device.Runtime,
			Duration:       e.Meta.Duration,
			PublishState:   publishState,
			ExperienceId:   e.Asset.ExperienceId,
			ExperienceType: e.Asset.ExperienceType,
		},
	}, nil
}
