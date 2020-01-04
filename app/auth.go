package app

import (
	"context"
	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"time"
)

type login struct {
	Tel      string `form:"tel" json:"tel" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	Username string `form:"username" json:"username"`
}

func authCallback(c *gin.Context) (interface{}, error) {

	userCollection := connectDB("user")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var loginVals login
	if err := c.ShouldBind(&loginVals); err != nil {
		return "", jwt.ErrMissingLoginValues
	}
	tel := loginVals.Tel
	password := loginVals.Password
	username := loginVals.Username

	//Register
	if username != "" {
		var result bson.M

		err := userCollection.FindOne(ctx, bson.M{"tel": tel}).Decode(&result)

		if err != nil {

			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err := userCollection.InsertOne(ctx, loginVals)

			if err != nil {
				log.Print(err)
			}

			return &user{
				Tel:      tel,
				UserName: username,
			}, nil

		}
		return nil, jwt.ErrForbidden

	}
	var result bson.M

	err := userCollection.FindOne(ctx, bson.M{"tel": tel, "password": password}).Decode(&result)

	if err != nil {
		return nil, jwt.ErrFailedAuthentication
	}

	return &user{
		Tel:      tel,
		UserName: username,
	}, nil
}

type tel struct {
	Tel string `form:"tel" json:"tel" binding:"required"`
}

func authVerify(c *gin.Context) {

	var telephone tel
	if err := c.ShouldBind(&telephone); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "wrong format"})
		return
	}

	userCollection := connectDB("user")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result bson.M
	err := userCollection.FindOne(ctx, bson.M{"tel": telephone.Tel}).Decode(&result)
	if err == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "already registed"})
		return
	}

	//verify here
	c.JSON(http.StatusOK, gin.H{"verify": "0000"})

}
