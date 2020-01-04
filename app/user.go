package app

import (
	"context"
	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func userInfo(c *gin.Context) {
	claims := jwt.ExtractClaims(c)

	userCollection := connectDB("user")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result bson.M
	err := userCollection.FindOne(ctx, bson.M{"tel": claims[identityKey]}).Decode(&result)
	handleError(err)

	c.JSON(200, result)
}
