// internal/database/database.go
package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"rccg-salvation-centre-backend/internal/models"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get underlying SQL database to configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	// Configure connection pool settings
	sqlDB.SetMaxIdleConns(10)                  // Maximum idle connections in pool
	sqlDB.SetMaxOpenConns(100)                 // Maximum open connections to database
	sqlDB.SetConnMaxLifetime(time.Hour)        // Maximum lifetime of a connection
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Maximum idle time before closing

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Database connected successfully with connection pool configured")
	log.Printf("Connection Pool Settings: MaxIdle=%d, MaxOpen=%d, MaxLifetime=%v, MaxIdleTime=%v",
		10, 100, time.Hour, 10*time.Minute)

	// Auto-migrate
	err = DB.AutoMigrate(
		&models.Admin{},
		&models.Sermon{},
		&models.ServiceType{},
		&models.Testimony{},
		&models.RegularProgram{},
		&models.SpecialEvent{},
		&models.FirstTimer{},
		&models.Attendance{},
		&models.PrayerRequest{},
		//&models.ActivityLog{},
	)
	if err != nil {
		log.Fatal("Failed to auto migrate:", err)
	}

	log.Println("Auto migration completed successfully for all models")
}

// Close closes the database connection
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
