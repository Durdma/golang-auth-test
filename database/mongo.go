// Подключение к mongoDB и выполнение функций поиска, добавления, удаления записей
package database

import (
	"context"
	"test-auth/auth"
	"test-auth/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Необходимые константы. НА ПРОДЕ КОНСТАНТЫ ДОЛЖНЫ БЫТЬ ПЕРЕМЕЩЕНЫ В ПЕРЕМЕННЫЕ ОКРУЖЕНИЯ
//
//	timeout - предельное время выполнения операции на БД
//	mongoURI - путь подключения к БД
const (
	timeout  = 10 * time.Second
	mongoURI = "mongodb://127.0.0.1:27017"
)

// Обертка для более удобного использования mongo.Client - подключения к БД
type Database struct {
	DatabaseConnection *mongo.Client
}

// Создание нового подключения к mongoDB
func NewClient() (*Database, error) {
	opts := options.Client().ApplyURI(mongoURI)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		// Возвращает ошибку связанную с подключеним к БД
		return nil, err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		// Возвращает ошибку связанную с проверкой подключения к БД
		return nil, err
	}

	return &Database{DatabaseConnection: client}, nil
}

// Возвращает колекцию документов("Sessions") mongoDB из заданной БД
func (db *Database) getCollection() *mongo.Collection {
	collection := db.DatabaseConnection.Database("authDB").Collection("Sessions")
	return collection
}

// Добавляет запись о новой сессии пользователя в БД.
//
// Сигнатура документа соответствует структуре models.Record
func (db *Database) AddToken(token models.Record) (*mongo.InsertOneResult, error) {
	collection := db.getCollection()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return collection.InsertOne(ctx, token)
}

// Функция осуществляет поиск по refreshToken запись о нужной сессии и выполняет проверку
// токена на соответствие условиям действительности токена(не превышено ли времяы жизни токена).
//
// Если условия выполнены, то возвращаются "_id" документа и "guid" сессии
//
// Если сессия с таким refreshToken не найдена, возвращает иформацию об этом
func (db *Database) CheckToken(decodedToken string) (string, string, primitive.ObjectID, error) {
	collection := db.getCollection()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		// Возвращает ошибку связанную с поиском записей в БД
		return "", "", primitive.NilObjectID, err
	}

	var current models.Record

	for cursor.Next(ctx) {
		if err := cursor.Decode(&current); err != nil {
			// Возвращает ошибку связанную с декодированием полученной записи из БД  структуру models.Record
			return "", "", primitive.NilObjectID, err
		}

		if auth.CheckRefreshToken(decodedToken, current.RefreshToken) {
			if time.Now().Before(current.ExpiresAt) || time.Now().Equal(current.ExpiresAt) {
				return "OK", current.Guid, current.ID, nil
			} else {
				// Возвращает информацию о том, что refreshToken просрочен
				return "token expired!", "", current.ID, nil
			}
		}
	}
	// Возвращает информацию о том, что сессия с таким refreshToken не найдена
	return "wrong token!", "", primitive.NilObjectID, nil
}

// Функция для удаления документа из БД по его "_id"
func (db *Database) DelToken(tokenId primitive.ObjectID) error {
	collection := db.getCollection()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.M{"_id": tokenId})
	if err != nil {
		// Возвращает ошибку связанную с удалением документа из БД
		return err
	}

	return nil
}
