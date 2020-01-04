package app

import (
	"context"
	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"log"
	"time"
)

type postPost struct {
	Title   string      `form:"title" json:"title" `
	Guild   string      `form:"guild" json:"guild" binding:"required"`
	Content interface{} `form:"content" json:"content" binding:"required"`
}

type post struct {
	Title       string      `form:"title" json:"title"`
	Content     interface{} `form:"content" json:"content"`
	Owner       interface{} `form:"owner" json:"owner"`
	OwnerName   interface{} `form:"ownername" json:"ownername"`
	OwnerAvatar interface{} `form:"owneravatar" json:"owneravatar"`
	Guild       string      `form:"guild" json:"guild"`
	GuildName   interface{} `form:"guildname" json:"guildname"`
	GuildAvatar interface{} `form:"guildavatar" json:"guildavatar"`
	Time        time.Time   `form:"time" json:"time"`
	Vote        int64       `form:"vote" json:"vote"`
}

type getPost struct {
	Guild string `form:"guild" json:"guild" binding:"required"`
}

func newPost(c *gin.Context) {
	claims := jwt.ExtractClaims(c)

	var postPostVals postPost

	if err := c.ShouldBind(&postPostVals); err != nil {

		c.JSON(401, gin.H{"error": "cannot generate a post"})
		return
	}

	var postVals post
	postVals.Title = postPostVals.Title
	postVals.Guild = postPostVals.Guild
	postVals.Content = postPostVals.Content

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userCollection := connectDB("user")
	var user bson.M
	err := userCollection.FindOne(ctx, bson.M{"tel": claims[identityKey]}).Decode(&user)
	handleError(err)
	postVals.Owner = user["_id"]
	postVals.OwnerName = user["username"]
	postVals.OwnerAvatar = ""

	guildsCollection := connectDB("guilds")
	var guild bson.M

	guildID, err := primitive.ObjectIDFromHex(postVals.Guild)
	handleError(err)

	err = guildsCollection.FindOne(ctx, bson.M{"_id": guildID}).Decode(&guild)
	handleError(err)
	postVals.GuildName = guild["name"]
	if guild["isavatar"] == true {
		postVals.GuildAvatar = guild["avatar"]
	} else {
		postVals.GuildAvatar = ""
	}

	log.Println(guild)

	postVals.Time = time.Now()
	postVals.Vote = 0

	postCollection := connectDB("post")

	postRes, err := postCollection.InsertOne(ctx, postVals)
	handleError(err)

	c.JSON(200, gin.H{"code": 200, "post": postRes})
}

func getPosts(c *gin.Context) {

	var getPostVals getPost

	if err := c.ShouldBind(&getPostVals); err != nil {

		c.JSON(401, gin.H{"error": "cannot get the posts"})
		return
	}

	postCollection := connectDB("post")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := postCollection.Find(ctx, bson.M{"guild": getPostVals.Guild})
	handleError(err)

	var results []bson.M
	for cur.Next(ctx) {
		var result bson.M
		err := cur.Decode(&result)
		handleError(err)

		results = append(results, result)
	}
	if err := cur.Err(); err != nil {
		handleError(err)
	}

	c.JSON(200, results)
}
