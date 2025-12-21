// internal/middleware/security_headers.go
package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders returns a Gin middleware that adds essential security response headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), midi=(), sync-xhr=(), microphone=(), camera=(), magnetometer=(), gyroscope=(), fullscreen=(self), payment=()")
		c.Header("Server", "")
		c.Header("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none';")

		c.Next()
	}
}
