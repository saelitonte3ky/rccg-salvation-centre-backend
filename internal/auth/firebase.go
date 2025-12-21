// internal/auth/firebase.go
package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

var FirebaseAuth *auth.Client

func InitFirebase() error {
	ctx := context.Background()
	var opt option.ClientOption

	// Priority 1: Try FIREBASE_CREDENTIALS_JSON environment variable (for production)
	firebaseCredsJSON := os.Getenv("FIREBASE_CREDENTIALS_JSON")

	if firebaseCredsJSON != "" {
		// Validate JSON format
		var js json.RawMessage
		if err := json.Unmarshal([]byte(firebaseCredsJSON), &js); err != nil {
			return fmt.Errorf("invalid FIREBASE_CREDENTIALS_JSON format: %v", err)
		}

		opt = option.WithCredentialsJSON([]byte(firebaseCredsJSON))
		log.Println("✅ Using Firebase credentials from FIREBASE_CREDENTIALS_JSON environment variable")
	} else if credsPath := os.Getenv("FIREBASE_CREDENTIALS_PATH"); credsPath != "" {
		// Priority 2: Try FIREBASE_CREDENTIALS_PATH (alternative approach)
		if _, err := os.Stat(credsPath); err == nil {
			opt = option.WithCredentialsFile(credsPath)
			log.Printf("✅ Using Firebase credentials from path: %s", credsPath)
		} else {
			return fmt.Errorf("FIREBASE_CREDENTIALS_PATH set but file not found: %s", credsPath)
		}
	} else if _, err := os.Stat("firebase-adminsdk.json"); err == nil {
		// Priority 3: Default file for local development
		opt = option.WithCredentialsFile("firebase-adminsdk.json")
		log.Println("✅ Using Firebase credentials from firebase-adminsdk.json file")
	} else {
		return fmt.Errorf("no Firebase credentials found. Set FIREBASE_CREDENTIALS_JSON env var, FIREBASE_CREDENTIALS_PATH, or provide firebase-adminsdk.json file")
	}

	// Initialize Firebase app
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return fmt.Errorf("failed to initialize Firebase app: %v", err)
	}

	// Get Auth client
	FirebaseAuth, err = app.Auth(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Firebase Auth client: %v", err)
	}

	log.Println("✅ Firebase Auth initialized successfully")
	return nil
}

// Generate session cookie (valid 14 days)
func CreateSessionCookie(idToken string) (string, error) {
	if FirebaseAuth == nil {
		return "", fmt.Errorf("Firebase Auth not initialized")
	}

	expiresIn := time.Hour * 24 * 14 // 14 days max
	sessionCookie, err := FirebaseAuth.SessionCookie(context.Background(), idToken, expiresIn)
	if err != nil {
		return "", fmt.Errorf("failed to create session cookie: %v", err)
	}

	return sessionCookie, nil
}

// Verify session cookie and get admin role
func VerifySessionCookie(cookie string) (*auth.Token, error) {
	if FirebaseAuth == nil {
		return nil, fmt.Errorf("Firebase Auth not initialized")
	}

	token, err := FirebaseAuth.VerifySessionCookieAndCheckRevoked(context.Background(), cookie)
	if err != nil {
		return nil, fmt.Errorf("failed to verify session cookie: %v", err)
	}

	return token, nil
}

// Generate fingerprint for extra security
func GenerateFingerprint(c *gin.Context) string {
	userAgent := c.Request.UserAgent()
	ip := c.ClientIP()
	input := fmt.Sprintf("%s|%s|%s", userAgent, ip, os.Getenv("JWT_SECRET"))
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
