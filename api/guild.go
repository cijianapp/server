package api

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/cijianapp/server/oss"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"context"
	"errors"
	"time"
)

type postGuild struct {
	Name    string `form:"name" json:"name" binding:"required"`
	Privacy string `form:"privacy" json:"privacy" binding:"required"`
	Icon    string `form:"icon" json:"icon"`
}

type getGuildType struct {
	ID string `form:"id" json:"id" binding:"required"`
}

type joinGuildType struct {
	Guild string `form:"guild" json:"guild" binding:"required"`
}

type guild struct {
	Name        string             `form:"name" json:"name"`
	Isavatar    bool               `form:"isavatar" json:"isavatar"`
	Avatar      string             `form:"avatar" json:"avatar"`
	Privacy     bool               `form:"privacy" json:"privacy"`
	Owner       interface{}        `form:"owner" json:"owner"`
	PostChannel primitive.ObjectID `form:"postchannel" json:"postchannel"`
}

type postUpdateGuild struct {
	ID          string `form:"id" json:"id" binding:"required"`
	Name        string `form:"name" json:"name" binding:"required"`
	Avatar      string `form:"avatar" json:"avatar"`
	Cover       string `form:"cover" json:"cover"`
	Description string `form:"description" json:"description"`
}

// When user generate a guild, this function tell user register this guild
func insertGuild(owner interface{}, guildID primitive.ObjectID) {
	userCollection := ConnectDB("user")
	guildCollection := ConnectDB("guild")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": owner, "guild": bson.M{"$elemMatch": bson.M{"_id": guildID}}}

	var user bson.M
	err := userCollection.FindOne(ctx, filter).Decode(&user)

	if err != nil {
		filter = bson.M{"_id": owner}
		update := bson.M{"$addToSet": bson.M{"guild": guildID}}

		_, err = userCollection.UpdateOne(ctx, filter, update)
		handleError(err)

		filter = bson.M{"_id": owner}
		err = userCollection.FindOne(ctx, filter).Decode(&user)
		handleError(err)

		filter = bson.M{"_id": guildID}
		update = bson.M{"$addToSet": bson.M{"members": bson.M{"_id": user["_id"], "username": user["username"]}}}
		_, err = guildCollection.UpdateOne(ctx, filter, update)
		handleError(err)

		updateMemberCount(guildID)
	}

}

// generate a new guild
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

	guildVals.Isavatar, guildVals.Avatar = oss.PutImageResize(postGuildVals.Icon, "avatar")

	guildCollection := ConnectDB("guild")

	guildRes, err := guildCollection.InsertOne(ctx, guildVals)
	handleError(err)

	if guildID, ok := guildRes.InsertedID.(primitive.ObjectID); ok {
		insertGuild(guildVals.Owner, guildID)

		channel := generateChannel("post", guildID)

		if channelID, ok := channel.(primitive.ObjectID); ok {

			err = insertChannel(guildID, channelID)
			if err != nil {
				c.JSON(401, gin.H{"error": "cannot generate a guild"})
				return
			}

			filter := bson.M{"_id": guildID}
			update := bson.M{"$set": bson.M{"postchannel": channelID}}
			_, err := guildCollection.UpdateOne(ctx, filter, update)

			if err != nil {
				c.JSON(401, gin.H{"error": "cannot update the user info"})
				return
			}

			c.JSON(200, gin.H{"code": 200, "guild": guildRes.InsertedID, "channel": channelID})

		}

	} else {
		c.JSON(401, gin.H{"error": "cannot generate a guild"})

	}

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

func guildInfo(guild interface{}) (bson.M, error) {

	if guildID, ok := guild.(primitive.ObjectID); ok {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var guild bson.M
		guildCollection := ConnectDB("guild")
		err := guildCollection.FindOne(ctx, bson.M{"_id": guildID}).Decode(&guild)
		handleError(err)

		return guild, nil

	}

	return bson.M{}, errors.New("wrong type")

}

// update the memberCount in a guild
func updateMemberCount(guildID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	guildCollection := ConnectDB("guild")
	filter := bson.M{"_id": guildID}

	var guild bson.M
	err := guildCollection.FindOne(ctx, filter).Decode(&guild)
	handleError(err)

	if guildMembers, ok := guild["members"].(bson.A); ok {
		memberCount := len(guildMembers)
		update := bson.M{"$set": bson.M{"membercount": memberCount}}
		_, err = guildCollection.UpdateOne(ctx, filter, update)
		handleError(err)
	}

}

// update the guild by guild settings
func updateGuild(c *gin.Context) {

	var postUpdateGuildVals postUpdateGuild

	if err := c.ShouldBind(&postUpdateGuildVals); err != nil {

		c.JSON(401, gin.H{"error": "cannot update the guild info"})
		return
	}

	guildCollection := ConnectDB("guild")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	guildID, err := primitive.ObjectIDFromHex(postUpdateGuildVals.ID)
	if err != nil {
		c.JSON(401, gin.H{"error": "cannot update the guild info"})
		return
	}

	filter := bson.M{"_id": guildID}

	update := bson.M{"$set": bson.M{"name": postUpdateGuildVals.Name, "description": postUpdateGuildVals.Description}}
	_, err = guildCollection.UpdateOne(ctx, filter, update)

	if err != nil {
		c.JSON(401, gin.H{"error": "cannot update the user info"})
		return
	}

	if postUpdateGuildVals.Avatar != "" {

		if postUpdateGuildVals.Avatar == "remove" {
			Isavatar, Avatar := false, ""

			update := bson.M{"$set": bson.M{"avatar": Avatar, "isavatar": Isavatar}}

			_, err := guildCollection.UpdateOne(ctx, filter, update)

			if err != nil {
				c.JSON(401, gin.H{"error": "cannot update the guild info"})
				return
			}

		} else {
			Isavatar, Avatar := oss.PutImageResize(postUpdateGuildVals.Avatar, "avatar")

			update := bson.M{"$set": bson.M{"avatar": Avatar, "isavatar": Isavatar}}

			_, err := guildCollection.UpdateOne(ctx, filter, update)

			if err != nil {
				c.JSON(401, gin.H{"error": "cannot update the user info"})
				return
			}
		}

	}

	if postUpdateGuildVals.Cover != "" {
		isCover, Cover := oss.PutImageResize(postUpdateGuildVals.Cover, "cover")

		update := bson.M{"$set": bson.M{"cover": Cover, "iscover": isCover}}

		_, err := guildCollection.UpdateOne(ctx, filter, update)

		if err != nil {
			c.JSON(401, gin.H{"error": "cannot update the user info"})
			return
		}

	}

	c.JSON(200, gin.H{"code": 200})
}

func getGuild(c *gin.Context) {
	var getGuildVals getGuildType

	if err := c.ShouldBind(&getGuildVals); err != nil {
		c.JSON(401, gin.H{"error": "cannot get the guild"})
		return
	}

	guildCollection := ConnectDB("guild")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	guildID, err := primitive.ObjectIDFromHex(getGuildVals.ID)
	handleError(err)

	var guild bson.M
	err = guildCollection.FindOne(ctx, bson.M{"_id": guildID}).Decode(&guild)
	handleError(err)

	c.JSON(200, guild)
}
