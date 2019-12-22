package app

import (
	"go.mongodb.org/mongo-driver/bson"

	"context"
	"log"
	"time"
)

type guild struct {
	Name  string      `form:"name" json:"name" binding:"required"`
	Owner interface{} `form:"owner" json:"owner" `
}

func generateGuild(name string, owner interface{}) interface{} {
	guildsCollection := connectDB("guilds")

	var guildVals guild

	guildVals.Name = name

	guildVals.Owner = owner

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := guildsCollection.InsertOne(ctx, guildVals)
	if err != nil {
		log.Fatal(err)
	}

	guildID := res.InsertedID

	channelID := generateChannel("动态", guildID)

	insertChannel(guildID, channelID)

	channelID = generateChannel("群聊", guildID)

	insertChannel(guildID, channelID)

	return guildID
}

func insertGuild(owner interface{}, guildID interface{}) {
	userCollection := connectDB("user")
	guildsCollection := connectDB("guilds")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var guild bson.M
	err := guildsCollection.FindOne(ctx, bson.M{"_id": guildID}).Decode(&guild)
	if err != nil {
		log.Fatal(err)
	}

	filter := bson.D{bson.E{Key: "_id", Value: owner}}
	update := bson.D{bson.E{Key: "$push", Value: bson.D{bson.E{Key: "guilds", Value: guild}}}}

	_, err = userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}
}
