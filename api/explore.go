package api

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"time"
)

func explore(c *gin.Context) {

	guildCollection := ConnectDB("guild")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	options := options.Find()
	options.SetSort(bson.M{"membercount": -1})
	options.SetLimit(10)

	cur, err := guildCollection.Find(ctx, filter, options)
	handleError(err)

	var results []bson.M
	for cur.Next(ctx) {
		var guild bson.M
		err := cur.Decode(&guild)
		handleError(err)

		guildInfo := extractGuildInfo(guild)

		results = append(results, guildInfo)
	}
	if err := cur.Err(); err != nil {
		handleError(err)
	}

	c.JSON(200, results)

}

func extractGuildInfo(guild bson.M) bson.M {
	// guildInfo := bson.M{
	// 	"_id":         guild["_id"],
	// 	"avatar":      guild["avatar"],
	// 	"isavatar":    guild["isavatar"],
	// 	"cover":       guild["cover"],
	// 	"iscover":     guild["iscover"],
	// 	"name":        guild["name"],
	// 	"membercount": guild["membercount"],
	// 	"description": guild["description"],
	// }
	// return guildInfo

	return guild
}
