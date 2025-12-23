// internal/middleware/security_headers.go
package middleware

import (
	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic protection
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), midi=(), sync-xhr=(), microphone=(), camera=(), magnetometer=(), gyroscope=(), fullscreen=(self), payment=()")

		// Hide server info
		c.Header("Server", "")

		// Relaxed but safe CSP for SPA + separate API backend
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data: https:; "+
				"font-src 'self' data:; "+
				"connect-src 'self' "+
				"https://rccg-salvation-centre.vercel.app "+
				"http://localhost:3000 "+
				"https://api.rccgsalvationcentre.org; "+
				"frame-ancestors 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self';",
		)

		c.Next()
	}
}
