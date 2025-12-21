// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rccg-salvation-centre-backend/internal/auth"
	"rccg-salvation-centre-backend/internal/database"
	"rccg-salvation-centre-backend/internal/middleware"
	"rccg-salvation-centre-backend/internal/routes"
	"rccg-salvation-centre-backend/seed"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	env := os.Getenv("ENVIRONMENT")
	if env != "production" {
		if err := godotenv.Load(".env.local"); err != nil {
			if err := godotenv.Load(); err != nil {
				log.Println("No .env file found – using system environment variables")
			}
		}
	}

	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := validateEnv(); err != nil {
		log.Fatal("Environment validation failed:", err)
	}

	database.Connect()
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	if env != "production" {
		seed.SeedAdmins()
		seed.SeedServiceTypes()
	}

	auth.InitFirebase()

	r := gin.Default()
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RequestSizeLimiter(10 << 20))
	r.Use(middleware.RateLimiter())
	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}

func validateEnv() error {
	required := map[string]string{
		"DATABASE_URL": "Database connection string",
		"JWT_SECRET":   "JWT signing secret",
	}

	missing := []string{}
	for key := range required {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	if len(os.Getenv("JWT_SECRET")) < 32 {
		log.Println("Warning: JWT_SECRET should be at least 32 characters for better security")
	}

	return nil
}
