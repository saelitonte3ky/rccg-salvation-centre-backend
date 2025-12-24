// middleware/cors.go

package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		c.Writer.Header().Set("Vary", "Origin")

		if origin != "" {
			log.Printf("[CORS] Allowing origin: %s", origin)
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Always allow these headers and methods
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours

		// Handle preflight OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			log.Printf("[CORS] Handling preflight request for: %s", c.Request.URL.Path)
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
