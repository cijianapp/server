package api

import (
	"go.mongodb.org/mongo-driver/mongo"
)

//ConnectDB return the db collection
func ConnectDB(collectionName string) *mongo.Collection {

	collection := client.Database("testing").Collection(collectionName)

	return collection
}
