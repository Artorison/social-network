package helpers

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetTime() time.Time {
	return time.Now().UTC().Round(time.Millisecond)
}

func GenerateID() string {
	pID := primitive.NewObjectID()
	return pID.Hex()
}
