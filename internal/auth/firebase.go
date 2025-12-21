// internal/auth/firebase.go
package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

func InitFirebase() {
	opt := option.WithCredentialsFile("firebase-adminsdk.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatal("Error initializing Firebase app:", err)
	}

	FirebaseAuth, err = app.Auth(context.Background())
	if err != nil {
		log.Fatal("Error initializing Firebase Auth:", err)
	}

	log.Println("Firebase Auth initialized successfully")
}

// Generate session cookie (valid 14 days)
func CreateSessionCookie(idToken string) (string, error) {
	expiresIn := time.Hour * 24 * 14 // 14 days max
	return FirebaseAuth.SessionCookie(context.Background(), idToken, expiresIn)
}

// Verify session cookie and get admin role
func VerifySessionCookie(cookie string) (*auth.Token, error) {
	return FirebaseAuth.VerifySessionCookieAndCheckRevoked(context.Background(), cookie)
}

// Generate fingerprint for extra security
func GenerateFingerprint(c *gin.Context) string {
	userAgent := c.Request.UserAgent()
	ip := c.ClientIP()
	input := fmt.Sprintf("%s|%s|%s", userAgent, ip, os.Getenv("JWT_SECRET"))
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
