package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const timeout = 10 * time.Second
const mongoURI = "mongodb://127.0.0.1:27017"

func NewClient() (*mongo.Client, error) {
	opts := options.Client().ApplyURI(mongoURI)

	client, err := mongo.NewClient(opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func GetCollection(client *mongo.Client) *mongo.Collection {
	collection := client.Database("authDB").Collection("Users")
	return collection
}
