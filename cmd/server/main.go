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

	"github.com/gin-gonic/gin"
)

func main() {
	// Always run in production/release mode (backend is always deployed)
	gin.SetMode(gin.ReleaseMode)

	if err := validateEnv(); err != nil {
		log.Fatal("Environment validation failed:", err)
	}

	database.Connect()
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	auth.InitFirebase()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimiter())

	// 1. Lightweight health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

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
		log.Printf("Server starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 2. Start the background routine targeting your live domain
	ctxKeepAlive, cancelKeepAlive := context.WithCancel(context.Background())
	go startKeepAliveTicker(ctxKeepAlive)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 3. Stop the keep-alive routine smoothly on server termination
	cancelKeepAlive()

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

// 4. Background task to ping the live API every 5 minutes
func startKeepAliveTicker(ctx context.Context) {
	// Give the server 10 seconds to fully deploy and start up before firing the first request
	time.Sleep(10 * time.Second)

	url := "https://api.rccgsalvationcentre.org/health"
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Initial ping
	pingServer(url)

	for {
		select {
		case <-ticker.C:
			pingServer(url)
		case <-ctx.Done():
			log.Println("Stopping keep-alive ticker...")
			return
		}
	}
}

func pingServer(url string) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Keep-alive ping to %s failed: %v", url)
		return
	}
	resp.Body.Close()
}