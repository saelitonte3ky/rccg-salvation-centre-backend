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
	isDev := strings.Contains(c.Request.Header.Get("Origin"), "localhost") ||
		strings.Contains(c.Request.Header.Get("Referer"), "localhost") ||
		strings.Contains(c.Request.Host, "localhost")

	domain := os.Getenv("COOKIE_DOMAIN")
	secure := !isDev
	sameSite := http.SameSiteLaxMode
	if !isDev {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieName,
		Value:    sessionCookie,
		Path:     "/",
		Domain:   domain,
		MaxAge:   14 * 24 * 60 * 60,
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

	domain := os.Getenv("COOKIE_DOMAIN")
	secure := !isDev
	sameSite := http.SameSiteLaxMode
	if !isDev {
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
