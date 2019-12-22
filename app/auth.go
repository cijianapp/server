package app

import (
	"context"
	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"log"
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
			res, err := userCollection.InsertOne(ctx, loginVals)

			guildID := generateGuild("动态", res.InsertedID)

			insertGuild(res.InsertedID, guildID)

			if err != nil {
				log.Fatal(err)
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
