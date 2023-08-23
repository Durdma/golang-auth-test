// Основные модели для работы API
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Структура записи в БД
//
// GUID хранится для выполнения обновления токенов
type Record struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Guid         string             `bson:"guid"`
	RefreshToken string             `json:"refreshToken" bson:"refresh_token"`
	ExpiresAt    time.Time          `json:"expiresAt" bson:"expires_at"`
}

// Структура для возвращения accessToken через json
type Response struct {
	AccessToken string `json:"accessToken" bson:"access_token"`
}
