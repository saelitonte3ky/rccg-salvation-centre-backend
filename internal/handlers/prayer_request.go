// internal/handlers/prayer_request.go
package handlers

import (
	"net/http"

	"rccg-salvation-centre-backend/internal/database"
	"rccg-salvation-centre-backend/internal/middleware"
	"rccg-salvation-centre-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// Public: Submit prayer request
func CreatePrayerRequest(c *gin.Context) {
	var input struct {
		Name    string `json:"name" binding:"required"`
		Email   string `json:"email" binding:"required,email"`
		Request string `json:"request" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prayer := models.PrayerRequest{
		Name:    input.Name,
		Email:   input.Email,
		Request: input.Request,
		Status:  "pending",
	}

	if err := database.DB.Create(&prayer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save prayer request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Prayer request submitted successfully",
		"data":    prayer,
	})
}

// Admin: Get all prayer requests
func AdminGetPrayerRequests(c *gin.Context) {
	var requests []models.PrayerRequest
	database.DB.Order("submitted_at DESC").Find(&requests)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    requests,
	})
}

// Admin: Update prayer request status (or details)
func UpdatePrayerRequest(c *gin.Context) {
	id := c.Param("id")
	var request models.PrayerRequest
	if err := database.DB.First(&request, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prayer request not found"})
		return
	}

	var input struct {
		Name    *string `json:"name"`
		Email   *string `json:"email"`
		Request *string `json:"request"`
		Status  *string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name != nil {
		request.Name = *input.Name
	}
	if input.Email != nil {
		request.Email = *input.Email
	}
	if input.Request != nil {
		request.Request = *input.Request
	}
	if input.Status != nil {
		allowed := map[string]bool{"pending": true, "prayed": true, "archived": true}
		if !allowed[*input.Status] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
			return
		}
		request.Status = *input.Status
	}

	if err := database.DB.Save(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prayer request"})
		return
	}

	middleware.LogActivity(c, "Updated prayer request", request.Name)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Prayer request updated",
		"data":    request,
	})
}

// Admin: Delete prayer request
func DeletePrayerRequest(c *gin.Context) {
	id := c.Param("id")

	var request models.PrayerRequest
	if err := database.DB.First(&request, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prayer request not found"})
		return
	}

	database.DB.Delete(&request)
	middleware.LogActivity(c, "Deleted prayer request", request.Name)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Prayer request deleted",
	})
}
