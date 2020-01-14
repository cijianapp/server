package api

import (
	"go.mongodb.org/mongo-driver/bson"

	"context"
	"log"
	"time"
)

type channel struct {
	Name  string      `form:"name" json:"name" binding:"required"`
	Owner interface{} `form:"owner" json:"owner" `
}

//owner should be the ID of guild
func generateChannel(name string, owner interface{}) interface{} {
	channelssCollection := ConnectDB("channels")

	var channelVals channel

	channelVals.Name = name

	channelVals.Owner = owner

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := channelssCollection.InsertOne(ctx, channelVals)
	if err != nil {
		log.Fatal(err)
	}

	return res.InsertedID
}

func insertChannel(owner interface{}, ChannelID interface{}) {

	guildCollection := ConnectDB("guilds")
	channelssCollection := ConnectDB("channels")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var channel bson.M
	err := channelssCollection.FindOne(ctx, bson.M{"_id": ChannelID}).Decode(&channel)
	if err != nil {
		log.Fatal(err)
	}

	filter := bson.D{bson.E{Key: "_id", Value: owner}}
	update := bson.D{bson.E{Key: "$push", Value: bson.D{bson.E{Key: "channels", Value: channel}}}}

	_, err = guildCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}
}
