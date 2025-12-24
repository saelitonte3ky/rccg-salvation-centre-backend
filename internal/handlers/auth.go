package handlers

import (
	"context"
	"log"
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

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid idToken"})
		return
	}

	sessionCookie, err := auth.CreateSessionCookie(req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired Firebase token"})
		return
	}

	cookieName := os.Getenv("SESSION_COOKIE_NAME")

	// Detect if request is from localhost
	isDev := strings.Contains(c.Request.Header.Get("Origin"), "localhost") ||
		strings.Contains(c.Request.Header.Get("Referer"), "localhost") ||
		strings.Contains(c.Request.Host, "localhost")

	// For cross-origin requests (frontend and backend on different domains),
	// we MUST NOT set Domain, and MUST use Secure=true + SameSite=None
	var domain string
	var secure bool
	var sameSite http.SameSite

	if isDev {
		domain = ""
		secure = false
		sameSite = http.SameSiteLaxMode
	} else {
		// Production: Cross-origin setup (Vercel frontend + Render backend)
		// CRITICAL: Do NOT set domain for cross-origin cookies
		domain = ""
		secure = true                    // REQUIRED for SameSite=None
		sameSite = http.SameSiteNoneMode // REQUIRED for cross-origin

		// Log for debugging
		log.Printf("Setting cookie for production: Origin=%s, Secure=%v, SameSite=%v",
			c.Request.Header.Get("Origin"), secure, sameSite)
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieName,
		Value:    sessionCookie,
		Path:     "/",
		Domain:   domain,
		MaxAge:   14 * 24 * 60 * 60, // 14 days
		Secure:   secure,
		HttpOnly: true,
		SameSite: sameSite,
	})

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

	var admin models.Admin
	if err := database.DB.Where("email = ?", email).First(&admin).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin account not found. Contact superadmin."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": gin.H{
			"email": admin.Email,
			"role":  admin.Role,
		},
	})
}

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

func Logout(c *gin.Context) {
	cookieName := os.Getenv("SESSION_COOKIE_NAME")

	isDev := strings.Contains(c.Request.Header.Get("Origin"), "localhost") ||
		strings.Contains(c.Request.Header.Get("Referer"), "localhost") ||
		strings.Contains(c.Request.Host, "localhost")

	var domain string
	var secure bool
	var sameSite http.SameSite

	if isDev {
		domain = ""
		secure = false
		sameSite = http.SameSiteLaxMode
	} else {
		domain = "" // Empty for cross-origin
		secure = true
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		Domain:   domain,
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: true,
		SameSite: sameSite,
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}
