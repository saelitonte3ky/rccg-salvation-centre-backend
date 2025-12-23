package middleware

import (
	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME-type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Restrict browser features
		c.Header(
			"Permissions-Policy",
			"geolocation=(), midi=(), sync-xhr=(), microphone=(), camera=(), magnetometer=(), gyroscope=(), fullscreen=(self), payment=()",
		)

		// Enforce HTTPS
		c.Header(
			"Strict-Transport-Security",
			"max-age=63072000; includeSubDomains; preload",
		)

		// Hide server information
		c.Header("Server", "")

		// Content Security Policy
		c.Header(
			"Content-Security-Policy",
			"default-src 'self'; frame-ancestors 'none';",
		)

		c.Next()
	}
}
