package app

import (
	"github.com/appleboy/gin-jwt/v2"
	"github.com/cijianapp/server/oss"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"context"
	"log"
	"time"
)

type postGuild struct {
	Name    string `form:"name" json:"name" binding:"required"`
	Privacy string `form:"privacy" json:"privacy" binding:"required"`
	Icon    string `form:"icon" json:"icon"`
}

type guild struct {
	Name     string      `form:"name" json:"name"`
	Isavatar bool        `form:"isavatar" json:"isavatar"`
	Avatar   string      `form:"avatar" json:"avatar"`
	Privacy  bool        `form:"privacy" json:"privacy"`
	Owner    interface{} `form:"owner" json:"owner"`
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

func newGuild(c *gin.Context) {
	claims := jwt.ExtractClaims(c)

	var postGuildVals postGuild

	if err := c.ShouldBind(&postGuildVals); err != nil {

		c.JSON(401, gin.H{"error": "cannot generate a guild"})
		return
	}

	var guildVals guild
	guildVals.Name = postGuildVals.Name
	if postGuildVals.Privacy == "privacy" {
		guildVals.Privacy = true
	} else {
		guildVals.Privacy = false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userCollection := connectDB("user")

	var user bson.M
	err := userCollection.FindOne(ctx, bson.M{"tel": claims[identityKey]}).Decode(&user)
	handleError(err)

	guildVals.Owner = user["_id"]

	guildVals.Isavatar, guildVals.Avatar = oss.PutObject(postGuildVals.Icon)

	guildsCollection := connectDB("guilds")

	guildRes, err := guildsCollection.InsertOne(ctx, guildVals)
	handleError(err)

	insertGuild(guildVals.Owner, guildRes.InsertedID)

	c.JSON(200, gin.H{"code": 200, "guild": guildRes.InsertedID})
}
