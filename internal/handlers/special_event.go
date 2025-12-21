// internal/handlers/special_event.go
package handlers

import (
	"net/http"
	"time"

	"rccg-salvation-centre-backend/internal/database"
	"rccg-salvation-centre-backend/internal/middleware"
	"rccg-salvation-centre-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// Public: Get published special events (latest first)
func GetSpecialEvents(c *gin.Context) {
	var events []models.SpecialEvent
	database.DB.Where("published = ?", true).
		Order("date DESC, start_time DESC").
		Find(&events)
	c.JSON(http.StatusOK, gin.H{"data": events})
}

// Admin: Get all special events (latest first)
func AdminGetSpecialEvents(c *gin.Context) {
	var events []models.SpecialEvent
	database.DB.Order("date DESC, start_time DESC").Find(&events)
	c.JSON(http.StatusOK, gin.H{"data": events})
}

// Admin: Create special event
func CreateSpecialEvent(c *gin.Context) {
	var input struct {
		Title       string `json:"title" binding:"required"`
		Type        string `json:"type" binding:"required"`
		Description string `json:"description"`
		Date        string `json:"date" binding:"required"`
		StartTime   string `json:"startTime"`
		EndTime     string `json:"endTime"`
		Location    string `json:"location"`
		Published   bool   `json:"published"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsedDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	event := models.SpecialEvent{
		Title:       input.Title,
		Type:        input.Type,
		Description: input.Description,
		Date:        parsedDate,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		Location:    input.Location,
		Published:   input.Published,
	}

	if err := database.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	middleware.LogActivity(c, "Created special event", event.Title)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Special event created",
		"event":   event,
	})
}

// Admin: Update special event
func UpdateSpecialEvent(c *gin.Context) {
	id := c.Param("id")

	var event models.SpecialEvent
	if err := database.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var input struct {
		Title       *string `json:"title"`
		Type        *string `json:"type"`
		Description *string `json:"description"`
		Date        *string `json:"date"`
		StartTime   *string `json:"startTime"`
		EndTime     *string `json:"endTime"`
		Location    *string `json:"location"`
		Published   *bool   `json:"published"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Title != nil {
		event.Title = *input.Title
	}
	if input.Type != nil {
		event.Type = *input.Type
	}
	if input.Description != nil {
		event.Description = *input.Description
	}
	if input.Date != nil {
		if parsed, err := time.Parse("2006-01-02", *input.Date); err == nil {
			event.Date = parsed
		}
	}
	if input.StartTime != nil {
		event.StartTime = *input.StartTime
	}
	if input.EndTime != nil {
		event.EndTime = *input.EndTime
	}
	if input.Location != nil {
		event.Location = *input.Location
	}
	if input.Published != nil {
		event.Published = *input.Published
	}

	if err := database.DB.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	middleware.LogActivity(c, "Updated special event", event.Title)

	c.JSON(http.StatusOK, gin.H{
		"message": "Special event updated",
		"event":   event,
	})
}

// Admin: Delete special event
func DeleteSpecialEvent(c *gin.Context) {
	id := c.Param("id")

	var event models.SpecialEvent
	if err := database.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	database.DB.Delete(&event)
	middleware.LogActivity(c, "Deleted special event", event.Title)

	c.JSON(http.StatusOK, gin.H{"message": "Special event deleted"})
}
