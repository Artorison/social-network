package mongodb

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongoDB(ctx context.Context, address string) *mongo.Client {
	opt := options.Client().SetServerSelectionTimeout(time.Second * 5)
	client, err := mongo.Connect(ctx, opt.ApplyURI(address))
	if err != nil {
		slog.Error("FAILED TO connect mongo", "ERROR", err.Error())
		panic(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		slog.Error("FAILED TO check connection to mongo", "ERROR", err.Error())
		panic(err)
	}

	slog.Info("connection mongo db success")

	return client
}
