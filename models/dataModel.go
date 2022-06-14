package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Data struct {
	ID         primitive.ObjectID `bson:"_id"`
	
	Name       *string            `json:"name" validate:"required"`
	Phone      *string            `json:"phone" validate:"required"`
	Movie      *string            `json:"movie" validate:"required"`
	Created_at time.Time          `json:"created_at"`
	Updated_at time.Time          `json:"updated_at"`
	Count      string             `json:"count"`
	User_id    string             `json:"user_id"`
}
