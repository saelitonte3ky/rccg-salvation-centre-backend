// internal/middleware/auth.go
package middleware

import (
	"net/http"
	"time"

	"rccg-salvation-centre-backend/internal/auth"
	"rccg-salvation-centre-backend/internal/database"
	"rccg-salvation-centre-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// AuthRequired - verifies session cookie and loads admin into context
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionCookie, err := c.Cookie("rccg_session")
		if err != nil || sessionCookie == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No session found"})
			c.Abort()
			return
		}

		token, err := auth.VerifySessionCookie(sessionCookie)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid or expired session"})
			c.Abort()
			return
		}

		email, ok := token.Claims["email"].(string)
		if !ok || email == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Email missing"})
			c.Abort()
			return
		}

		var admin models.Admin
		if err := database.DB.Where("email = ?", email).First(&admin).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Admin not found"})
			c.Abort()
			return
		}

		c.Set("adminEmail", admin.Email)
		c.Set("adminID", admin.ID)
		c.Set("adminRole", admin.Role)

		c.Next()
	}
}

// RequireRoles - allows one or more roles
func RequireRoles(allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("adminRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Role not found"})
			c.Abort()
			return
		}

		for _, r := range allowed {
			if role.(string) == r {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Insufficient permissions"})
		c.Abort()
	}
}

// Shortcut for superadmin only
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRoles("superadmin")
}

// Log admin actions (create, update, delete, etc.)
func LogActivity(c *gin.Context, action string, details string) {
	email, _ := c.Get("adminEmail")
	adminID, _ := c.Get("adminID")

	if email == nil || adminID == nil {
		return
	}

	go database.DB.Create(&models.ActivityLog{
		AdminID:    adminID.(uint),
		AdminEmail: email.(string),
		Action:     action,
		Details:    details,
		CreatedAt:  time.Now(),
	})
}
