// internal/handlers/auth.go
package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

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

	// Set HTTP-only secure cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     os.Getenv("SESSION_COOKIE_NAME"),
		Value:    sessionCookie,
		Expires:  time.Now().Add(14 * 24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   os.Getenv("ENVIRONMENT") == "production",
		SameSite: http.SameSiteLaxMode,
	})

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
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     os.Getenv("SESSION_COOKIE_NAME"),
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     "/",
		HttpOnly: true,
		Secure:   os.Getenv("ENVIRONMENT") == "production",
		SameSite: http.SameSiteLaxMode,
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}
