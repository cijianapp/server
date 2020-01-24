package api

import (
	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"context"
	"errors"
	"log"
	"time"
)

type postChannel struct {
	Name    string `form:"name" json:"name" binding:"required"`
	GuildID string `form:"guildid" json:"guildid" binding:"required"`
	Type    string `form:"type" json:"type" binding:"required"`
}

type channel struct {
	Name    string             `form:"name" json:"name" binding:"required"`
	GuildID primitive.ObjectID `form:"guildid" json:"guildid" binding:"required"`
	Type    string             `form:"type" json:"type" binding:"required"`
}

func insertChannel(guildID primitive.ObjectID, ChannelID primitive.ObjectID) error {

	guildCollection := ConnectDB("guild")
	channelCollection := ConnectDB("channel")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var channel bson.M
	err := channelCollection.FindOne(ctx, bson.M{"_id": ChannelID}).Decode(&channel)
	if err != nil {
		log.Fatal(err)
	}

	filter := bson.D{bson.E{Key: "_id", Value: guildID}}
	update := bson.D{bson.E{Key: "$push", Value: bson.D{bson.E{Key: "channel", Value: channel}}}}

	_, err = guildCollection.UpdateOne(context.TODO(), filter, update)

	return err
}

func newChannel(c *gin.Context) {
	claims := jwt.ExtractClaims(c)

	var postChannelVals postChannel

	if err := c.ShouldBind(&postChannelVals); err != nil {

		c.JSON(401, gin.H{"error": "cannot generate a channel"})
		return
	}

	var channelVals channel

	guildID, err := primitive.ObjectIDFromHex(postChannelVals.GuildID)
	if err != nil {
		c.JSON(401, gin.H{"error": "cannot generate a channel"})
		return
	}

	channelVals.Name = postChannelVals.Name
	channelVals.Type = postChannelVals.Type
	channelVals.GuildID = guildID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userCollection := ConnectDB("user")

	var user bson.M
	err = userCollection.FindOne(ctx, bson.M{"tel": claims[identityKey]}).Decode(&user)
	handleError(err)

	// check user is the guild owner here

	channelCollection := ConnectDB("channel")

	channelRes, err := channelCollection.InsertOne(ctx, channelVals)
	handleError(err)

	if channelID, ok := channelRes.InsertedID.(primitive.ObjectID); ok {
		err = insertChannel(channelVals.GuildID, channelID)
		if err != nil {
			c.JSON(401, gin.H{"error": "cannot generate a channel"})
			return
		}
	}

	c.JSON(200, gin.H{"code": 200, "channelid": channelRes.InsertedID})
}

// generateChannel generate default channel based on channel type
func generateChannel(channelType string, guildID primitive.ObjectID) interface{} {
	var channelVals channel

	if channelType == "post" {
		channelVals.Name = "主频道"
		channelVals.Type = "post"
		channelVals.GuildID = guildID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channelCollection := ConnectDB("channel")

	channelRes, err := channelCollection.InsertOne(ctx, channelVals)
	handleError(err)

	return channelRes.InsertedID
}

func channelInfo(channel interface{}) (bson.M, error) {
	if channelID, ok := channel.(primitive.ObjectID); ok {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var channel bson.M
		guildCollection := ConnectDB("channel")
		err := guildCollection.FindOne(ctx, bson.M{"_id": channelID}).Decode(&channel)
		handleError(err)

		return channel, nil
	}

	return bson.M{}, errors.New("wrong type")

}
