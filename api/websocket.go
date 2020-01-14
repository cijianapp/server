package api

import (
	"github.com/gin-gonic/gin"

	"net/http"

	"github.com/cijianapp/server/model"
)

func connectWebSocket(c *gin.Context, hub *model.Hub) {

	w := c.Writer
	r := c.Request

	model.ServeWs(hub, w, r)

}

func originChecker(*http.Request) bool { return true }
