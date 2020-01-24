package api

import (
	"context"
	"github.com/appleboy/gin-jwt/v2"
	"github.com/cijianapp/server/oss"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"errors"
	"time"
)

type postUpdateUser struct {
	Username string `form:"username" json:"username" binding:"required"`
	Icon     string `form:"icon" json:"icon"`
}

func userInfo(c *gin.Context) {
	claims := jwt.ExtractClaims(c)

	userCollection := ConnectDB("user")
	guildCollection := ConnectDB("guild")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user bson.M
	err := userCollection.FindOne(ctx, bson.M{"tel": claims[identityKey]}).Decode(&user)
	handleError(err)

	var guilds []bson.M

	if guildList, ok := user["guild"].(primitive.A); ok {
		for _, guildID := range guildList {
			var guild bson.M
			err := guildCollection.FindOne(ctx, bson.M{"_id": guildID}).Decode(&guild)
			handleError(err)
			guilds = append(guilds, guild)
		}
	}

	user["guild"] = guilds

	c.JSON(200, user)
}

func updateUser(c *gin.Context) {
	claims := jwt.ExtractClaims(c)

	var postUpdateUserVals postUpdateUser

	if err := c.ShouldBind(&postUpdateUserVals); err != nil {

		c.JSON(401, gin.H{"error": "cannot update the user info"})
		return
	}

	userCollection := ConnectDB("user")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"tel": claims[identityKey]}

	if postUpdateUserVals.Icon != "" {

		if postUpdateUserVals.Icon == "remove" {
			Isavatar, Avatar := false, ""

			update := bson.M{"$set": bson.M{"avatar": Avatar, "isavatar": Isavatar, "username": postUpdateUserVals.Username}}

			_, err := userCollection.UpdateOne(ctx, filter, update)

			if err != nil {
				c.JSON(401, gin.H{"error": "cannot update the user info"})
				return
			}

		} else {
			Isavatar, Avatar := oss.PutImageResize(postUpdateUserVals.Icon)

			update := bson.M{"$set": bson.M{"avatar": Avatar, "isavatar": Isavatar, "username": postUpdateUserVals.Username}}

			_, err := userCollection.UpdateOne(ctx, filter, update)

			if err != nil {
				c.JSON(401, gin.H{"error": "cannot update the user info"})
				return
			}
		}

	} else {
		update := bson.M{"$set": bson.M{"username": postUpdateUserVals.Username}}

		_, err := userCollection.UpdateOne(ctx, filter, update)

		if err != nil {
			c.JSON(401, gin.H{"error": "cannot update the user info"})
			return
		}
	}

	c.JSON(200, gin.H{"code": 200})
}

func userInfo1(user interface{}) (bson.M, error) {

	if userID, ok := user.(primitive.ObjectID); ok {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var user bson.M
		guildCollection := ConnectDB("user")
		err := guildCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
		handleError(err)

		return user, nil
	}
	return bson.M{}, errors.New("wrong type")

}
