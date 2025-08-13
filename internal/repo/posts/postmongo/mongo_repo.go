package postmongo

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type PostMongoDB struct {
	Posts    *mongo.Collection
	Comments *mongo.Collection
}

func NewModgoDB(client *mongo.Client, dbName string) *PostMongoDB {
	return &PostMongoDB{
		Posts:    client.Database(dbName).Collection("posts"),
		Comments: client.Database(dbName).Collection("comments"),
	}
}
