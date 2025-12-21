// internal/handlers/testimony.go
package handlers

import (
	"net/http"
	"time"

	"rccg-salvation-centre-backend/internal/database"
	"rccg-salvation-centre-backend/internal/middleware"
	"rccg-salvation-centre-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// Public: Get approved testimonies, sorted by approved date (latest first)
func GetTestimonies(c *gin.Context) {
	var testimonies []models.Testimony
	database.DB.Where("status = ?", models.Approved).
		Order("COALESCE(approved_at, submitted_at) DESC").
		Find(&testimonies)
	c.JSON(http.StatusOK, gin.H{"data": testimonies})
}

// Admin: Get all testimonies, sorted by submission date (latest first)
func AdminGetTestimonies(c *gin.Context) {
	var testimonies []models.Testimony
	database.DB.Order("submitted_at DESC").Find(&testimonies)
	c.JSON(http.StatusOK, gin.H{"data": testimonies})
}

// Public: Submit new testimony (status = pending)
func CreateTestimony(c *gin.Context) {
	var input struct {
		Name    string `json:"name" binding:"required"`
		Title   string `json:"title" binding:"required"`
		Message string `json:"message" binding:"required"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	testimony := models.Testimony{
		Name:    input.Name,
		Title:   input.Title,
		Message: input.Message,
		Email:   input.Email,
		Phone:   input.Phone,
		Status:  models.Pending,
	}

	if err := database.DB.Create(&testimony).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit testimony"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Testimony submitted successfully. It will be reviewed soon.",
	})
}

// Admin: Update testimony (approve/reject)
func UpdateTestimony(c *gin.Context) {
	id := c.Param("id")

	var testimony models.Testimony
	if err := database.DB.First(&testimony, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Testimony not found"})
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"` // "approved" or "rejected"
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	switch input.Status {
	case "approved":
		testimony.Status = models.Approved
		testimony.ApprovedAt = &now
		testimony.RejectedAt = nil
	case "rejected":
		testimony.Status = models.Rejected
		testimony.RejectedAt = &now
		testimony.ApprovedAt = nil
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Use 'approved' or 'rejected'"})
		return
	}

	if err := database.DB.Save(&testimony).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update testimony"})
		return
	}

	action := "Approved"
	if input.Status == "rejected" {
		action = "Rejected"
	}
	middleware.LogActivity(c, action+" testimony", testimony.Title)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Testimony " + input.Status,
		"testimony": testimony,
	})
}

// Admin: Delete testimony
func DeleteTestimony(c *gin.Context) {
	id := c.Param("id")

	var testimony models.Testimony
	if err := database.DB.First(&testimony, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Testimony not found"})
		return
	}

	database.DB.Delete(&testimony)
	middleware.LogActivity(c, "Deleted testimony", testimony.Title)

	c.JSON(http.StatusOK, gin.H{"message": "Testimony deleted"})
}
