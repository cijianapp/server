package api

import (
	"context"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/cijianapp/server/oss"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"fmt"
	"strings"
	"time"
)

type postPost struct {
	Title   string        `form:"title" json:"title" `
	Guild   string        `form:"guild" json:"guild" binding:"required"`
	Channel string        `form:"channel" json:"channel" binding:"required"`
	Content []interface{} `form:"content" json:"content" binding:"required"`
}

type post struct {
	Title   string             `form:"title" json:"title"`
	Content interface{}        `form:"content" json:"content"`
	Owner   interface{}        `form:"owner" json:"owner"`
	Guild   primitive.ObjectID `form:"guild" json:"guild"`
	Channel primitive.ObjectID `form:"channel" json:"channel" binding:"required"`
	Time    time.Time          `form:"time" json:"time"`
	Vote    int64              `form:"vote" json:"vote"`
}

type getPostsType struct {
	Guild   string `form:"guild" json:"guild" binding:"required"`
	Channel string `form:"channel" json:"channel" binding:"required"`
}

type getPostType struct {
	Post string `form:"post" json:"post" binding:"required"`
}

type deletePostType struct {
	Post string `form:"post" json:"post" binding:"required"`
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
	postVals.Guild, _ = primitive.ObjectIDFromHex(postPostVals.Guild)
	postVals.Channel, _ = primitive.ObjectIDFromHex(postPostVals.Channel)

	var content []interface{}

	// change base64 to url
	for _, v := range postPostVals.Content {
		element, ok := v.(map[string]interface{})
		if ok {
			if element["type"] == "image" {
				str := fmt.Sprintf("%v", element["url"])

				if strings.Contains(str, "base64") {
					_, ossImage := oss.PutObject(str)
					element["url"] = oss.OssURL + ossImage
				}
			}
		}

		fmt.Println(element)

		if element["type"] == "paragraph" {

			children, ok := element["children"].([]interface{})

			if ok {

				text, ok := children[0].(map[string]interface{})
				if ok {
					if text["text"] != "" {
						content = append(content, element)
					}
				}

			}
		} else {
			content = append(content, element)
		}

	}
	postVals.Content = content

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userCollection := ConnectDB("user")
	var user bson.M
	err := userCollection.FindOne(ctx, bson.M{"tel": claims[identityKey]}).Decode(&user)
	handleError(err)
	postVals.Owner = user["_id"]

	postVals.Time = time.Now()
	postVals.Vote = 0

	postCollection := ConnectDB("post")

	postRes, err := postCollection.InsertOne(ctx, postVals)
	handleError(err)

	c.JSON(200, gin.H{"code": 200, "post": postRes})
}

func getPost(c *gin.Context) {
	var getPostVals getPostType

	if err := c.ShouldBind(&getPostVals); err != nil {
		c.JSON(401, gin.H{"error": "cannot get the post"})
		return
	}

	postCollection := ConnectDB("post")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	postID, err := primitive.ObjectIDFromHex(getPostVals.Post)
	handleError(err)

	var post bson.M
	err = postCollection.FindOne(ctx, bson.M{"_id": postID}).Decode(&post)
	handleError(err)

	post = completePost(post)

	c.JSON(200, post)
}

func getPosts(c *gin.Context) {

	var getPostsVals getPostsType

	if err := c.ShouldBind(&getPostsVals); err != nil {

		c.JSON(401, gin.H{"error": "cannot get the posts"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	guildCollection := ConnectDB("guild")

	guildID, err := primitive.ObjectIDFromHex(getPostsVals.Guild)
	if err != nil {
		c.JSON(401, gin.H{"error": "cannot get the posts"})
		return
	}

	var guild bson.M
	err = guildCollection.FindOne(ctx, bson.M{"_id": guildID}).Decode(&guild)

	postCollection := ConnectDB("post")

	channelID, err := primitive.ObjectIDFromHex(getPostsVals.Channel)
	if err != nil {
		c.JSON(401, gin.H{"error": "cannot get the posts"})
		return
	}

	var cur *mongo.Cursor

	if guild["postchannel"] == channelID {
		cur, err = postCollection.Find(ctx, bson.M{"guild": guildID})

	} else {
		cur, err = postCollection.Find(ctx, bson.M{"channel": channelID})
	}

	handleError(err)

	var results []bson.M
	for cur.Next(ctx) {
		var post bson.M
		err := cur.Decode(&post)
		handleError(err)

		post = completePost(post)

		results = append(results, post)
	}
	if err := cur.Err(); err != nil {
		handleError(err)
	}

	c.JSON(200, results)
}

func getHomePosts(c *gin.Context) {

	claims := jwt.ExtractClaims(c)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userCollection := ConnectDB("user")

	var user bson.M
	err := userCollection.FindOne(ctx, bson.M{"tel": claims[identityKey]}).Decode(&user)
	handleError(err)

	postCollection := ConnectDB("post")

	var cur *mongo.Cursor
	cur, err = postCollection.Find(ctx, bson.M{"guild": bson.M{"$in": user["guild"]}})
	handleError(err)

	var results []bson.M
	for cur.Next(ctx) {
		var post bson.M
		err := cur.Decode(&post)
		handleError(err)

		post = completePost(post)

		results = append(results, post)
	}
	if err := cur.Err(); err != nil {
		handleError(err)
	}

	c.JSON(200, results)
}

func guestPosts(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	postCollection := ConnectDB("post")

	options := options.Find()
	options.SetSort(bson.M{"vote": -1})
	options.SetLimit(50)

	var cur *mongo.Cursor
	cur, err := postCollection.Find(ctx, bson.M{}, options)
	handleError(err)

	var results []bson.M
	for cur.Next(ctx) {
		var post bson.M
		err := cur.Decode(&post)
		handleError(err)

		post = completePost(post)

		results = append(results, post)
	}
	if err := cur.Err(); err != nil {
		handleError(err)
	}

	c.JSON(200, results)
}

func completePost(post bson.M) bson.M {
	guild, err := guildInfo(post["guild"])
	if err == nil {
		post["guildname"] = guild["name"]
	}

	channel, err := channelInfo(post["channel"])
	if err == nil {
		post["channelname"] = channel["name"]
	}

	user, err := userInfo1(post["owner"])
	if err == nil {
		post["username"] = user["username"]
		post["useravatar"] = user["avatar"]
	}

	return post
}

func deletePost(c *gin.Context) {
	var deletePostVals deletePostType

	if err := c.ShouldBind(&deletePostVals); err != nil {
		c.JSON(401, gin.H{"error": "cannot delete the post"})
		return
	}

	postCollection := ConnectDB("post")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	postID, err := primitive.ObjectIDFromHex(deletePostVals.Post)
	handleError(err)

	deleteResult, err := postCollection.DeleteOne(ctx, bson.M{"_id": postID})
	handleError(err)

	c.JSON(200, deleteResult)
}
