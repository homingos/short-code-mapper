package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Project struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	ClientId  primitive.ObjectID `bson:"client_id" json:"client_id"`
	Name      string             `valid:"required" bson:"name" json:"name"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
	Publish   bool               `bson:"publish" json:"publish"`
	CreatedBy *User              `bson:"created_by" json:"created_by"`
	CreatedAt time.Duration      `bson:"created_at" json:"created_at"`
	UpdatedAt time.Duration      `bson:"updated_at" json:"updated_at"`
}
