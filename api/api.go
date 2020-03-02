package api

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/cijianapp/server/model"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"fmt"
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

	// r.Use(cors.Default())

	r.Use(cors.New(cors.Config{
		// AllowOrigins: []string{"http://cijian.net:3000", "http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		// AllowCredentials: true,
		MaxAge:          12 * time.Hour,
		AllowAllOrigins: true,
	}))

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
		// SendCookie:     true,
		// CookieHTTPOnly: true,
		// SecureCookie:   true,
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	r.POST("auth/verify", authVerify)

	r.POST("auth/login", authMiddleware.LoginHandler)

	r.POST("auth/register", authMiddleware.LoginHandler)

	hub := model.NewHub()
	go hub.Run()

	r.GET("/ws", func(c *gin.Context) {
		connectWebSocket(c, hub)
	})

	api := r.Group("/api")

	api.Use(authMiddleware.MiddlewareFunc())
	{
		api.POST("/updateuser", updateUser)
		api.GET("/info", userInfo)
		api.POST("/guild", newGuild)
		api.POST("/updateguild", updateGuild)
		api.POST("/join", joinGuild)
		api.POST("/post", newPost)
		api.POST("/deletepost", deletePost)
		api.GET("/posts", getPosts)
		api.GET("/post", getPost)
		api.GET("/homeposts", getHomePosts)
		api.POST("/upload", newUpload)
		api.POST("/channel", newChannel)
		api.POST("/vote", vote)
		api.GET("/explore", explore)
	}

	guest := r.Group("/guest")
	{
		guest.GET("/explore", explore)
		guest.GET("/homeposts", guestPosts)
		guest.GET("/posts", getPosts)
		guest.GET("/post", getPost)
		guest.GET("/guild", getGuild)

	}

	upload := r.Group("/upload")

	upload.Use(authMiddleware.MiddlewareFunc())
	{
		upload.POST("/image", uploadImage)
	}

	return r
}

func handleError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	}
}

// Run is the root fucntion to excute
func Run() {

	r := setupRouter()

	r.Run()
}
