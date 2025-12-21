// internal/handlers/sermon.go
package handlers

import (
	"net/http"
	"time"

	"rccg-salvation-centre-backend/internal/database"
	"rccg-salvation-centre-backend/internal/middleware"
	"rccg-salvation-centre-backend/internal/models"

	"github.com/gin-gonic/gin"
)

/*
PUBLIC ENDPOINTS
*/

// GET /api/sermons
func GetSermons(c *gin.Context) {
	var sermons []models.Sermon
	database.DB.Where("published = ?", true).
		Order("date DESC, created_at DESC"). // ADD created_at DESC as secondary sort
		Find(&sermons)

	c.JSON(http.StatusOK, gin.H{
		"data":  sermons,
		"count": len(sermons),
	})
}

// GET /api/sermons/latest
func GetLatestSermon(c *gin.Context) {
	var sermon models.Sermon
	err := database.DB.Where("published = ?", true).
		Order("created_at DESC"). // Changed from "date DESC" to "created_at DESC"
		First(&sermon).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No published sermon found"})
		return
	}
	c.JSON(http.StatusOK, sermon)
}

// GET /api/sermons/search?q=faith
// Search published sermons by title, pastor, or description
func SearchSermons(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query 'q' is required"})
		return
	}

	var sermons []models.Sermon
	db := database.DB

	// Public: only published sermons
	if c.GetString("adminEmail") == "" {
		db = db.Where("published = ?", true)
	}

	db = db.Where("title ILIKE ? OR pastor ILIKE ? OR description ILIKE ?",
		"%"+query+"%", "%"+query+"%", "%"+query+"%")

	db.Order("date DESC").Limit(20).Find(&sermons)

	c.JSON(http.StatusOK, gin.H{
		"query": query,
		"count": len(sermons),
		"data":  sermons,
	})
}

/*
ADMIN ENDPOINTS (Protected)
*/

// GET /api/admin/sermons
// Admin sees all sermons (including drafts)
func AdminGetSermons(c *gin.Context) {
	var sermons []models.Sermon
	database.DB.Order("date DESC").Find(&sermons)
	c.JSON(http.StatusOK, gin.H{
		"data":  sermons,
		"count": len(sermons),
	})
}

// POST /api/admin/sermons
// Create new sermon (media_team or superadmin)
func CreateSermon(c *gin.Context) {
	adminEmail := c.GetString("adminEmail")

	var input struct {
		Title       string `json:"title" binding:"required"`
		Pastor      string `json:"pastor" binding:"required"`
		Service     string `json:"service" binding:"required"`
		Date        string `json:"date" binding:"required"` // YYYY-MM-DD
		YoutubeID   string `json:"youtubeId" binding:"required"`
		Duration    string `json:"duration"`
		Description string `json:"description"`
		Published   bool   `json:"published"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate date
	parsedDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	// Prevent duplicate YouTube ID
	var existing models.Sermon
	if err := database.DB.Where("youtube_id = ?", input.YoutubeID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Sermon with this YouTube video already exists"})
		return
	}

	sermon := models.Sermon{
		Title:       input.Title,
		Pastor:      input.Pastor,
		Service:     input.Service,
		Date:        parsedDate,
		YoutubeID:   input.YoutubeID,
		Duration:    input.Duration,
		Description: input.Description,
		Published:   input.Published,
	}

	if err := database.DB.Create(&sermon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save sermon"})
		return
	}

	middleware.LogActivity(c, adminEmail, "Created sermon")

	c.JSON(http.StatusCreated, gin.H{
		"message": "Sermon created successfully",
		"sermon":  sermon,
	})
}

// PUT /api/admin/sermons/:id
func UpdateSermon(c *gin.Context) {
	id := c.Param("id")
	adminEmail := c.GetString("adminEmail")

	var sermon models.Sermon
	if err := database.DB.First(&sermon, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sermon not found"})
		return
	}

	var input struct {
		Title       *string `json:"title"`
		Pastor      *string `json:"pastor"`
		Service     *string `json:"service"`
		Date        *string `json:"date"`
		YoutubeID   *string `json:"youtubeId"`
		Duration    *string `json:"duration"`
		Description *string `json:"description"`
		Published   *bool   `json:"published"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update only provided fields
	if input.Title != nil {
		sermon.Title = *input.Title
	}
	if input.Pastor != nil {
		sermon.Pastor = *input.Pastor
	}
	if input.Service != nil {
		sermon.Service = *input.Service
	}
	if input.Date != nil {
		if parsed, err := time.Parse("2006-01-02", *input.Date); err == nil {
			sermon.Date = parsed
		}
	}
	if input.YoutubeID != nil {
		// Prevent duplicate YouTube ID
		var dup models.Sermon
		if err := database.DB.Where("youtube_id = ? AND id != ?", *input.YoutubeID, id).First(&dup).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Another sermon uses this YouTube video"})
			return
		}
		sermon.YoutubeID = *input.YoutubeID
	}
	if input.Duration != nil {
		sermon.Duration = *input.Duration
	}
	if input.Description != nil {
		sermon.Description = *input.Description
	}
	if input.Published != nil {
		sermon.Published = *input.Published
	}

	database.DB.Save(&sermon)
	middleware.LogActivity(c, adminEmail, "Updated sermon")

	c.JSON(http.StatusOK, gin.H{
		"message": "Sermon updated",
		"sermon":  sermon,
	})
}

// DELETE /api/admin/sermons/:id (Superadmin only)
func DeleteSermon(c *gin.Context) {
	id := c.Param("id")
	adminEmail := c.GetString("adminEmail")

	var sermon models.Sermon
	if err := database.DB.First(&sermon, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sermon not found"})
		return
	}

	database.DB.Delete(&sermon)
	middleware.LogActivity(c, adminEmail, "Deleted sermon")

	c.JSON(http.StatusOK, gin.H{"message": "Sermon deleted permanently"})
}
