package api

import (
	"github.com/appleboy/gin-jwt/v2"
	"github.com/cijianapp/server/oss"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"context"
	"log"
	"time"
)

type postGuild struct {
	Name    string `form:"name" json:"name" binding:"required"`
	Privacy string `form:"privacy" json:"privacy" binding:"required"`
	Icon    string `form:"icon" json:"icon"`
}

type joinGuildType struct {
	Guild string `form:"guild" json:"guild" binding:"required"`
}

type guild struct {
	Name     string      `form:"name" json:"name"`
	Isavatar bool        `form:"isavatar" json:"isavatar"`
	Avatar   string      `form:"avatar" json:"avatar"`
	Privacy  bool        `form:"privacy" json:"privacy"`
	Owner    interface{} `form:"owner" json:"owner"`
}

func generateGuild(name string, owner interface{}) interface{} {
	guildsCollection := ConnectDB("guilds")

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
	userCollection := ConnectDB("user")
	guildsCollection := ConnectDB("guilds")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var guild bson.M
	err := guildsCollection.FindOne(ctx, bson.M{"_id": guildID}).Decode(&guild)
	handleError(err)

	filter := bson.M{"_id": owner, "guilds": bson.M{"$elemMatch": bson.M{"_id": guildID}}}

	var user bson.M
	err = userCollection.FindOne(context.TODO(), filter).Decode(&user)

	if err != nil {
		filter = bson.M{"_id": owner}
		update := bson.M{"$push": bson.M{"guilds": guild}}

		_, err = userCollection.UpdateOne(context.TODO(), filter, update)
		handleError(err)

		filter = bson.M{"_id": owner}
		err = userCollection.FindOne(context.TODO(), filter).Decode(&user)
		handleError(err)

		filter = bson.M{"_id": guildID}
		update = bson.M{"$push": bson.M{"members": bson.M{"_id": user["_id"], "username": user["username"]}}}
		_, err = guildsCollection.UpdateOne(context.TODO(), filter, update)
		handleError(err)
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

	userCollection := ConnectDB("user")

	var user bson.M
	err := userCollection.FindOne(ctx, bson.M{"tel": claims[identityKey]}).Decode(&user)
	handleError(err)

	guildVals.Owner = user["_id"]

	guildVals.Isavatar, guildVals.Avatar = oss.PutObject(postGuildVals.Icon)

	guildsCollection := ConnectDB("guilds")

	guildRes, err := guildsCollection.InsertOne(ctx, guildVals)
	handleError(err)

	insertGuild(guildVals.Owner, guildRes.InsertedID)

	c.JSON(200, gin.H{"code": 200, "guild": guildRes.InsertedID})
}

func joinGuild(c *gin.Context) {

	claims := jwt.ExtractClaims(c)

	var joinGuildVals joinGuildType

	if err := c.ShouldBind(&joinGuildVals); err != nil {

		c.JSON(401, gin.H{"error": "cannot join the guild"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userCollection := ConnectDB("user")

	var user bson.M
	err := userCollection.FindOne(ctx, bson.M{"tel": claims[identityKey]}).Decode(&user)
	handleError(err)

	guildID, err := primitive.ObjectIDFromHex(joinGuildVals.Guild)
	handleError(err)

	insertGuild(user["_id"], guildID)

	c.JSON(200, gin.H{"code": 200, "guild": guildID})

}
