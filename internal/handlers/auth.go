// internal/handlers/auth.go
package handlers

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
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
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid idToken"})
		return
	}

	// Create session cookie from Firebase ID token
	sessionCookie, err := auth.CreateSessionCookie(req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired Firebase token"})
		return
	}

	// Set HttpOnly session cookie - improved settings
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "rccg_session",
		Value:    sessionCookie,
		Expires:  time.Now().Add(14 * 24 * time.Hour),
		MaxAge:   14 * 24 * 3600, // 14 days
		Path:     "/",
		Domain:   "", // Leave empty unless using subdomains
		HttpOnly: true,
		Secure:   os.Getenv("ENVIRONMENT") == "production",
		SameSite: http.SameSiteLaxMode,
	})

	// Verify the original Firebase ID token to get claims
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

	// Find admin in database
	var admin models.Admin
	if err := database.DB.Where("email = ?", email).First(&admin).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin account not found. Contact superadmin."})
		return
	}

	// Return user data
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"id":    strconv.FormatUint(uint64(admin.ID), 10),
			"name":  strings.Split(admin.Email, "@")[0], // Replace with real name field later
			"email": admin.Email,
			"role":  admin.Role,
		},
	})
}

// GET /api/auth/me
func Me(c *gin.Context) {
	email := c.GetString("adminEmail")
	role := c.GetString("adminRole")
	adminID := c.GetUint("adminID")

	if email == "" || role == "" || adminID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    strconv.FormatUint(uint64(adminID), 10),
		"name":  strings.Split(email, "@")[0],
		"email": email,
		"role":  role,
	})
}

// POST /api/auth/logout
func Logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "rccg_session",
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
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
