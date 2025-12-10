package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func SetupRoutes(r *gin.Engine) *gin.Engine {
	api := r.Group("/api")

	api.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "RCCG Salvation Centre API"})
	})

	return r
}
