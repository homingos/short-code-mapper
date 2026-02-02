package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Type string

const (
	INTERACTION Type = "INTERACTION"
	SCAN        Type = "SCAN"
	VIDEO       Type = "VIDEO"
	RECORDING   Type = "RECORDING"
	MODULE      Type = "MODULE"
)

type Platform string

const (
	ANDROID Platform = "ANDROID"
	IOS     Platform = "IOS"
)

type Runtime string

const (
	WEB    Runtime = "WEB"
	NATIVE Runtime = "NATIVE"
)

type TrackingDetails struct {
	AdId      string `json:"ad_id,omitempty"`
	PushToken string `json:"push_token,omitempty"`
}

type AppDetails struct {
	AppId      string `json:"app_id"`
	AppVersion string `json:"app_version"`
	AppType    string `json:"app_type"`
}

type AssetDetails struct {
	ID             string `json:"id" validate:"required"`
	Type           string `json:"type" validate:"required,oneof=CARD GROUND"`
	URL            string `json:"url,omitempty"`
	ExperienceId   string `json:"experience_id,omitempty"`
	PublishState   bool   `json:"publish_state,omitempty"`
	ExperienceType string `json:"experience_type,omitempty" validate:"omitempty,oneof=CARD ALPHA INTERACTIVE"`
}

type ActionDetails struct {
	Type   string `json:"type" validate:"omitempty"`
	Name   string `json:"name" validate:"omitempty"`
	Action string `json:"action" validate:"omitempty"`
	URL    string `json:"url,omitempty"`
}
type DeviceDetails struct {
	ID       string   `json:"id"`
	Platform Platform `json:"platform" validate:"required,oneof=ANDROID IOS"`
	Model    string   `json:"model"`
	OS       string   `json:"os"`
	NativeAR bool     `json:"native_ar"`
	Runtime  Runtime  `json:"runtime,omitempty" validate:"omitempty,oneof=WEB NATIVE"`
}

type Meta struct {
	Duration int `json:"duration,omitempty"`
}

type Event struct {
	ID     primitive.ObjectID `bson:"_id" json:"id"`
	Name   string             `json:"name" bson:"name"`
	Type   Type               `json:"type" bson:"type"`
	Device DeviceDetails      `json:"device" bson:"device"`
	App    AppDetails         `json:"app,omitempty" bson:"app,omitempty"`
	Asset  AssetDetails       `json:"asset" bson:"asset"`
	Action ActionDetails      `json:"action" bson:"action"`
	Track  TrackingDetails    `json:"track,omitempty" bson:"track,omitempty"`
	Meta   Meta               `json:"meta,omitempty" bson:"meta,omitempty"`
}
