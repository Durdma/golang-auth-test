package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Record struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Guid         string             `bson:"guid"`
	RefreshToken string             `json:"refreshToken" bson:"refresh_token"`
	ExpiresAt    time.Time          `json:"expiresAt" bson:"expires_at"`
}

type Session struct {
	AccessToken  string `json:"accessToken" bson:"acess_token"`
	RefreshToken string `json:"refreshToken" bson:"refresh_token"`
}
