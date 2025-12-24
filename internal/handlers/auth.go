package handlers

import (
	"context"
	"log"
	"net/http"
	"os"

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
		log.Printf("[LOGIN] Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid idToken"})
		return
	}

	log.Printf("[LOGIN] Received login request from origin: %s", c.Request.Header.Get("Origin"))

	// Create session cookie from Firebase ID token
	sessionCookie, err := auth.CreateSessionCookie(req.IDToken)
	if err != nil {
		log.Printf("[LOGIN] Failed to create session cookie: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired Firebase token"})
		return
	}

	log.Println("[LOGIN] Session cookie created successfully")

	// Verify the ID token and extract claims
	token, err := auth.FirebaseAuth.VerifyIDToken(context.Background(), req.IDToken)
	if err != nil {
		log.Printf("[LOGIN] Failed to verify Firebase token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify Firebase token"})
		return
	}

	email, ok := token.Claims["email"].(string)
	if !ok || email == "" {
		log.Println("[LOGIN] Email not found in token claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in token"})
		return
	}

	log.Printf("[LOGIN] Looking up admin with email: %s", email)

	// Find admin in database
	var admin models.Admin
	if err := database.DB.Where("email = ?", email).First(&admin).Error; err != nil {
		log.Printf("[LOGIN] Admin not found in database for email %s: %v", email, err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin account not found. Contact superadmin."})
		return
	}

	log.Printf("[LOGIN] Admin found: Email=%s, Role=%s", admin.Email, admin.Role)

	// Cookie name with fallback
	cookieName := os.Getenv("SESSION_COOKIE_NAME")
	if cookieName == "" {
		cookieName = "rccg_session"
	}

	secure := true
	origin := c.Request.Header.Get("Origin")
	if origin == "http://localhost:1200" {
		secure = false
		log.Println("[TEMP DEBUG] Secure=false allowed for local frontend testing[](http://localhost:8080)")
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieName,
		Value:    sessionCookie,
		Path:     "/",
		MaxAge:   14 * 24 * 60 * 60, // 14 days
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})

	log.Printf("[LOGIN] Cookie '%s' set: Secure=%v, SameSite=None", cookieName, secure)

	// Return success response
	response := gin.H{
		"success": true,
		"user": gin.H{
			"email": admin.Email,
			"role":  admin.Role,
		},
	}

	log.Printf("[LOGIN] Login successful! Returning: %+v", response)
	c.JSON(http.StatusOK, response)
}

func Me(c *gin.Context) {
	email := c.GetString("adminEmail")
	role := c.GetString("adminRole")

	log.Printf("[ME] Request from origin: %s, adminEmail=%s", c.Request.Header.Get("Origin"), email)

	if email == "" || role == "" {
		log.Println("[ME] Not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	response := gin.H{
		"email": email,
		"role":  role,
	}

	log.Printf("[ME] Authenticated user: %+v", response)
	c.JSON(http.StatusOK, response)
}

func Logout(c *gin.Context) {
	log.Printf("[LOGOUT] Logout request from origin: %s", c.Request.Header.Get("Origin"))

	cookieName := os.Getenv("SESSION_COOKIE_NAME")
	if cookieName == "" {
		cookieName = "rccg_session"
	}

	secure := true
	origin := c.Request.Header.Get("Origin")
	if origin == "http://localhost:1200" {
		secure = false
		log.Println("[TEMP DEBUG] Secure=false allowed for local frontend testing[](http://localhost:8080)")
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})

	log.Printf("[LOGOUT] Cookie '%s' cleared (Secure=%v)", cookieName, secure)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}
