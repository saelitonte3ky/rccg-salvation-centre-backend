package main

import (
	"log"
	"os"
	"github.com/gin-gonic/gin"
	"rccg-salvation-centre-backend/internal/router"
	"github.com/joho/godotenv"
)

func main() {
	// Load env
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	r := gin.Default()
	r = router.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
