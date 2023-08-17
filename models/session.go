package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Session struct {
	ID           primitive.ObjectID `bson:"_id"`
	Guid         string             `bson:"guid"`
	RefreshToken string             `json:"refreshToken" bson:"refresh_token"`
	ExpiresAt    time.Time          `json:"expiresAt" bson:"expired_at"`
}
