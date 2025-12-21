// internal/handlers/first_timer.go
package handlers

import (
	"net/http"
	"time"

	"rccg-salvation-centre-backend/internal/database"
	"rccg-salvation-centre-backend/internal/middleware"
	"rccg-salvation-centre-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// Public: Submit first-timer information
func CreateFirstTimer(c *gin.Context) {
	var input struct {
		FirstName              string `json:"firstName" binding:"required"`
		LastName               string `json:"lastName" binding:"required"`
		Email                  string `json:"email"`
		Phone                  string `json:"phone"`
		Address                string `json:"address"`
		City                   string `json:"city"`
		State                  string `json:"state"`
		DateOfBirth            string `json:"dateOfBirth"`
		Gender                 string `json:"gender"`
		MaritalStatus          string `json:"maritalStatus"`
		Occupation             string `json:"occupation"`
		VisitDate              string `json:"visitDate" binding:"required"` // YYYY-MM-DD
		HowDidYouHear          string `json:"howDidYouHear"`
		PrayerRequest          string `json:"prayerRequest"`
		InterestedInMembership bool   `json:"interestedInMembership"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsedVisitDate, err := time.Parse("2006-01-02", input.VisitDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid visit date format. Use YYYY-MM-DD"})
		return
	}

	firstTimer := models.FirstTimer{
		FirstName:              input.FirstName,
		LastName:               input.LastName,
		Email:                  input.Email,
		Phone:                  input.Phone,
		Address:                input.Address,
		City:                   input.City,
		State:                  input.State,
		DateOfBirth:            input.DateOfBirth,
		Gender:                 input.Gender,
		MaritalStatus:          input.MaritalStatus,
		Occupation:             input.Occupation,
		VisitDate:              parsedVisitDate,
		HowDidYouHear:          input.HowDidYouHear,
		PrayerRequest:          input.PrayerRequest,
		InterestedInMembership: input.InterestedInMembership,
		FollowUpStatus:         "pending",
		Status:                 "new",
	}

	if err := database.DB.Create(&firstTimer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save first-timer"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "First-timer information submitted successfully. Thank you!",
	})
}

// Admin: Get all first-timers (sorted latest visit first)
func AdminGetFirstTimers(c *gin.Context) {
	var firstTimers []models.FirstTimer
	database.DB.Order("visit_date DESC").Find(&firstTimers)
	c.JSON(http.StatusOK, gin.H{"data": firstTimers})
}

// Admin: Update first-timer (follow-up status, etc.)
func UpdateFirstTimer(c *gin.Context) {
	id := c.Param("id")

	var firstTimer models.FirstTimer
	if err := database.DB.First(&firstTimer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "First-timer not found"})
		return
	}

	var input struct {
		FollowUpStatus *string `json:"followUpStatus"`
		Status         *string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.FollowUpStatus != nil {
		firstTimer.FollowUpStatus = *input.FollowUpStatus
	}
	if input.Status != nil {
		firstTimer.Status = *input.Status
	}

	if err := database.DB.Save(&firstTimer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update first-timer"})
		return
	}

	middleware.LogActivity(c, "Updated first-timer", firstTimer.FirstName+" "+firstTimer.LastName)

	c.JSON(http.StatusOK, gin.H{
		"message":    "First-timer updated",
		"firstTimer": firstTimer,
	})
}

// Admin: Delete first-timer
func DeleteFirstTimer(c *gin.Context) {
	id := c.Param("id")

	var firstTimer models.FirstTimer
	if err := database.DB.First(&firstTimer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "First-timer not found"})
		return
	}

	database.DB.Delete(&firstTimer)
	middleware.LogActivity(c, "Deleted first-timer", firstTimer.FirstName+" "+firstTimer.LastName)

	c.JSON(http.StatusOK, gin.H{"message": "First-timer deleted"})
}
