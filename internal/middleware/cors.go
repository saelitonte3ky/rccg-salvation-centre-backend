package middleware

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	// Get frontend URL from environment
	frontendURL := os.Getenv("FRONTEND_URL")

	// List of allowed origins (exact match required)
	allowedOrigins := map[string]bool{
		"http://localhost:3000":                    true,
		"http://127.0.0.1:3000":                    true,
		"http://localhost:8080":                    true,
		"http://127.0.0.1:8080":                    true,
		"https://rccg-salvation-centre.vercel.app": true,
		"https://www.rccgsalvationcentre.org":      true,
	}

	// Add frontend URL from env if it exists
	if frontendURL != "" {
		allowedOrigins[frontendURL] = true
		log.Printf("[CORS] Added frontend URL from env: %s", frontendURL)
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Always set Vary header for proper caching
		c.Writer.Header().Set("Vary", "Origin")

		if origin != "" && allowedOrigins[origin] {
			log.Printf("[CORS] Allowing origin: %s", origin)
			// CRITICAL: Set the exact origin, never use "*" with credentials
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			// CRITICAL: Must be "true" for cookies to work cross-origin
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if origin != "" {
			log.Printf("[CORS] Blocked origin: %s", origin)
		}

		// These headers are safe to always send
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Writer.Header().Set("Access-Control-Max-Age", "3600") // Cache preflight for 1 hour

		// Handle preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			log.Printf("[CORS] Handling preflight request for: %s", c.Request.URL.Path)
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
