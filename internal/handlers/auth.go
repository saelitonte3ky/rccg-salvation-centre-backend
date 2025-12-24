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

	log.Printf("[LOGIN] ✓ Session cookie created successfully")

	// Verify the ID token and extract claims
	token, err := auth.FirebaseAuth.VerifyIDToken(context.Background(), req.IDToken)
	if err != nil {
		log.Printf("[LOGIN] Failed to verify Firebase token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify Firebase token"})
		return
	}

	email, ok := token.Claims["email"].(string)
	if !ok || email == "" {
		log.Printf("[LOGIN] Email not found in token claims")
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

	log.Printf("[LOGIN] ✓ Admin found: Email=%s, Role=%s", admin.Email, admin.Role)

	// Set cookie
	cookieName := os.Getenv("SESSION_COOKIE_NAME")
	if cookieName == "" {
		cookieName = "session"
		log.Printf("[LOGIN] Warning: SESSION_COOKIE_NAME not set, using default: %s", cookieName)
	}

	// Detect environment
	origin := c.Request.Header.Get("Origin")
	isDev := strings.Contains(origin, "localhost") ||
		strings.Contains(c.Request.Header.Get("Referer"), "localhost") ||
		strings.Contains(c.Request.Host, "localhost")

	var domain string
	var secure bool
	var sameSite http.SameSite

	if isDev {
		// Pure local development
		domain = ""
		secure = false
		sameSite = http.SameSiteNoneMode
		log.Printf("[LOGIN] Development mode (localhost detected): Secure=false, SameSite=None")
	} else {
		// Production environment (deployed on Render)
		domain = ".rccgsalvationcentre.org" // Allows sharing across subdomains (api., www., etc.)
		secure = true
		sameSite = http.SameSiteNoneMode

		// TEMPORARY DEV FIX: Allow insecure cookie if origin is localhost/ngrok (for local frontend testing)
		lowerOrigin := strings.ToLower(origin)
		if strings.Contains(lowerOrigin, "localhost") ||
			strings.Contains(lowerOrigin, "127.0.0.1") ||
			strings.Contains(lowerOrigin, "ngrok") ||
			strings.Contains(lowerOrigin, "localtunnel") ||
			strings.Contains(lowerOrigin, ".vercel.app") { // Vercel previews if needed
			secure = false
			log.Printf("[LOGIN] *** WARNING: Insecure cookie (Secure=false) allowed for local/tunnel development. Origin: %s ***", origin)
			log.Printf("[LOGIN] *** REMOVE THIS OVERRIDE BEFORE FULL PRODUCTION LAUNCH ***")
		}

		log.Printf("[LOGIN] Production mode: Origin=%s, Secure=%v, SameSite=None, Domain=%s", origin, secure, domain)
	}

	// Set the cookie
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

	log.Printf("[LOGIN] ✓ Cookie set: Name=%s, Path=/, MaxAge=14days, Secure=%v, HttpOnly=true, SameSite=%v, Domain=%s",
		cookieName, secure, sameSite, domain)

	// Return success response
	response := gin.H{
		"success": true,
		"user": gin.H{
			"email": admin.Email,
			"role":  admin.Role,
		},
	}

	log.Printf("[LOGIN] ✓ Login successful! Returning: %+v", response)

	c.JSON(http.StatusOK, response)
}

func Me(c *gin.Context) {
	email := c.GetString("adminEmail")
	role := c.GetString("adminRole")

	log.Printf("[ME] Request from: %s, adminEmail=%s, adminRole=%s",
		c.Request.Header.Get("Origin"), email, role)

	if email == "" || role == "" {
		log.Printf("[ME] Not authenticated: email=%s, role=%s", email, role)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	response := gin.H{
		"email": email,
		"role":  role,
	}

	log.Printf("[ME] ✓ Authenticated user: %+v", response)

	c.JSON(http.StatusOK, response)
}

func Logout(c *gin.Context) {
	log.Printf("[LOGOUT] Logout request from: %s", c.Request.Header.Get("Origin"))

	cookieName := os.Getenv("SESSION_COOKIE_NAME")
	if cookieName == "" {
		cookieName = "session"
	}

	origin := c.Request.Header.Get("Origin")
	isDev := strings.Contains(origin, "localhost") ||
		strings.Contains(c.Request.Header.Get("Referer"), "localhost") ||
		strings.Contains(c.Request.Host, "localhost")

	var domain string
	var secure bool
	var sameSite http.SameSite

	if isDev {
		domain = ""
		secure = false
		sameSite = http.SameSiteNoneMode
	} else {
		domain = ".rccgsalvationcentre.org"
		secure = true
		sameSite = http.SameSiteNoneMode

		// TEMPORARY DEV FIX: Match Login behavior
		lowerOrigin := strings.ToLower(origin)
		if strings.Contains(lowerOrigin, "localhost") ||
			strings.Contains(lowerOrigin, "127.0.0.1") ||
			strings.Contains(lowerOrigin, "ngrok") ||
			strings.Contains(lowerOrigin, "localtunnel") ||
			strings.Contains(lowerOrigin, ".vercel.app") {
			secure = false
			log.Printf("[LOGOUT] *** WARNING: Insecure cookie clear (Secure=false) for local development. Origin: %s ***", origin)
		}
	}

	// Clear the cookie by setting MaxAge to -1
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

	log.Printf("[LOGOUT] ✓ Cookie cleared: %s (Secure=%v, Domain=%s)", cookieName, secure, domain)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}
