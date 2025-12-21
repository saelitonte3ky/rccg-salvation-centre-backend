// internal/middleware/request_limiter.go
package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestSizeLimiter limits the maximum size of the request body
func RequestSizeLimiter(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)

		// Trigger early enforcement
		if err := c.Request.ParseForm(); err != nil {
			if err.Error() == "http: request body too large" {
				c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
					"error":   "Payload too large",
					"message": fmt.Sprintf("Request body must not exceed %s", formatBytes(maxBytes)),
				})
				return
			}
		}

		c.Next()
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
