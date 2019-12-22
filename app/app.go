package app

import (
	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"

	"log"
	"time"
)

var identityKey = "id"

type user struct {
	Tel      string
	UserName string
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Key:         []byte("FE880BE1D7558D62B1DFDAE3E7D4F82ED9E987FA12D9195A18312741A1F87858"),
		IdentityKey: identityKey,
		Timeout:     time.Hour * 10,
		MaxRefresh:  time.Hour * 1000,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*user); ok {
				return jwt.MapClaims{
					identityKey: v.Tel,
				}
			}
			return jwt.MapClaims{}
		},
		Authenticator: authCallback,
	})

	r.POST("/auth/login", authMiddleware.LoginHandler)

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	return r
}

// Run is the root fucntion to excute
func Run() {

	r := setupRouter()

	r.Run()
}
