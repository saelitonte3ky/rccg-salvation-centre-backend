// internal/handlers/auth.go
package handlers

import (
	"context"
	"net/http"
	"os"
	"strings"

	"rccg-salvation-centre-backend/internal/auth"
	"rccg-salvation-centre-backend/internal/database"
	"rccg-salvation-centre-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	IDToken string `json:"idToken" binding:"required"`
}

// POST /api/auth/login
// Receives Firebase ID token from frontend and creates secure session cookie
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid idToken"})
		return
	}

	// Create secure session cookie from Firebase ID token (valid 14 days)
	sessionCookie, err := auth.CreateSessionCookie(req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired Firebase token"})
		return
	}

	// Get configuration from environment
	cookieName := os.Getenv("SESSION_COOKIE_NAME")
	if cookieName == "" {
		cookieName = "rccg_session" // fallback default
	}

	// Detect if request is from local development (for testing local frontend against prod backend)
	isDev := strings.Contains(c.Request.Header.Get("Origin"), "localhost") ||
		strings.Contains(c.Request.Header.Get("Referer"), "localhost") ||
		strings.Contains(c.Request.Host, "localhost")

	var domain string
	secure := os.Getenv("ENVIRONMENT") == "production"

	if isDev {
		domain = ""    // no domain for local cross-port
		secure = false // allow HTTP for local
	} else {
		domain = os.Getenv("COOKIE_DOMAIN") // e.g., ".rccgsalvationcentre.org"
	}

	// Use Gin's built-in SetCookie
	c.SetCookie(
		cookieName,    // name
		sessionCookie, // value
		14*24*60*60,   // maxAge: 14 days in seconds
		"/",           // path
		domain,        // domain (empty in dev)
		secure,        // secure (false in dev)
		true,          // httpOnly
	)

	// Verify the ID token to extract email
	token, err := auth.FirebaseAuth.VerifyIDToken(context.Background(), req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify Firebase token"})
		return
	}

	email, ok := token.Claims["email"].(string)
	if !ok || email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in token"})
		return
	}

	// Check if this email exists in our admins table
	var admin models.Admin
	if err := database.DB.Where("email = ?", email).First(&admin).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin account not found. Contact superadmin."})
		return
	}

	// Success — return user info
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"email": admin.Email,
			"role":  admin.Role,
		},
	})
}

// GET /api/auth/me
// Returns current logged-in admin (protected by AuthRequired middleware)
func Me(c *gin.Context) {
	email := c.GetString("adminEmail")
	role := c.GetString("adminRole")

	if email == "" || role == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email": email,
		"role":  role,
	})
}

// POST /api/auth/logout
// Clears the session cookie
func Logout(c *gin.Context) {
	cookieName := os.Getenv("SESSION_COOKIE_NAME")
	if cookieName == "" {
		cookieName = "rccg_session"
	}

	// Detect local dev
	isDev := strings.Contains(c.Request.Header.Get("Origin"), "localhost") ||
		strings.Contains(c.Request.Header.Get("Referer"), "localhost") ||
		strings.Contains(c.Request.Host, "localhost")

	var domain string
	secure := os.Getenv("ENVIRONMENT") == "production"

	if isDev {
		domain = ""
		secure = false
	} else {
		domain = os.Getenv("COOKIE_DOMAIN")
	}

	// Properly clear the cookie using Gin's SetCookie with maxAge = -1
	c.SetCookie(
		cookieName,
		"",
		-1, // delete cookie
		"/",
		domain,
		secure,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}
