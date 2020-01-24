package api

import (
	"context"

	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"fmt"
	"time"
)

type postVote struct {
	Post string `form:"post" json:"post" binding:"required"`
	Vote string `form:"vote" json:"vote"`
}

func vote(c *gin.Context) {

	claims := jwt.ExtractClaims(c)
	userCollection := ConnectDB("user")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var postVotetVals postVote

	if err := c.ShouldBind(&postVotetVals); err != nil {
		c.JSON(401, gin.H{"error": "cannot vote"})
		return
	}

	postCollection := ConnectDB("post")

	postID, err := primitive.ObjectIDFromHex(postVotetVals.Post)
	handleError(err)

	var update bson.M
	filter := bson.M{"_id": postID}

	var user bson.M
	err = userCollection.FindOne(ctx, bson.M{"tel": claims[identityKey]}).Decode(&user)
	if err != nil {
		c.JSON(401, gin.H{"error": "cannot vote"})
		return
	}

	var userUpdate bson.M
	var userFilter bson.M

	isVoted := false

	if voteList, ok := user["vote"].(primitive.A); ok {
		for _, element := range voteList {

			if vote, ok := element.(bson.M); ok {

				fmt.Println(vote)

				if vote["post"] == postID {

					isVoted = true

					if vote["vote"] == "up" {

						if postVotetVals.Vote == "" {
							userFilter = bson.M{"_id": user["_id"]}
							userUpdate = bson.M{"$pull": bson.M{"vote": bson.M{"post": postID}}}
							update = bson.M{"$inc": bson.M{"vote": -1}}
						}

						if postVotetVals.Vote == "down" {
							userFilter = bson.M{"_id": user["_id"], "vote.post": postID}
							userUpdate = bson.M{"$set": bson.M{"vote.$.vote": "down"}}
							update = bson.M{"$inc": bson.M{"vote": -2}}
						}

					}

					if vote["vote"] == "down" {

						if postVotetVals.Vote == "" {
							userFilter = bson.M{"_id": user["_id"]}
							userUpdate = bson.M{"$pull": bson.M{"vote": bson.M{"post": postID}}}
							update = bson.M{"$inc": bson.M{"vote": 1}}
						}

						if postVotetVals.Vote == "up" {
							userFilter = bson.M{"_id": user["_id"], "vote.post": postID}
							userUpdate = bson.M{"$set": bson.M{"vote.$.vote": "up"}}
							update = bson.M{"$inc": bson.M{"vote": 2}}
						}

					}
				}
			}

		}

	}

	if isVoted == false {
		userFilter = bson.M{"_id": user["_id"]}
		userUpdate = bson.M{"$push": bson.M{"vote": bson.M{"post": postID, "vote": postVotetVals.Vote}}}

		if postVotetVals.Vote == "up" {
			update = bson.M{"$inc": bson.M{"vote": 1}}
		}

		if postVotetVals.Vote == "down" {
			update = bson.M{"$inc": bson.M{"vote": -1}}
		}

	}

	_, err = postCollection.UpdateOne(ctx, filter, update)

	if err != nil {
		c.JSON(401, gin.H{"error": "cannot vote"})
		return
	}

	_, err = userCollection.UpdateOne(ctx, userFilter, userUpdate)

	if err != nil {
		c.JSON(401, gin.H{"error": "cannot vote"})
		return
	}

	c.JSON(200, gin.H{"code": 200})
}
