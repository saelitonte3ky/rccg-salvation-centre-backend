// internal/handlers/attendance.go
package handlers

import (
	"net/http"
	"time"

	"rccg-salvation-centre-backend/internal/database"
	"rccg-salvation-centre-backend/internal/middleware"
	"rccg-salvation-centre-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// Admin: Get all attendance records (sorted latest date first)
func AdminGetAttendance(c *gin.Context) {
	var attendance []models.Attendance
	database.DB.Order("date DESC").Find(&attendance)
	c.JSON(http.StatusOK, gin.H{"data": attendance})
}

// Admin: Create attendance record
func CreateAttendance(c *gin.Context) {
	adminEmail := c.GetString("adminEmail")

	var input struct {
		Date        string `json:"date" binding:"required"` // YYYY-MM-DD
		ServiceType string `json:"serviceType" binding:"required"`
		Adults      int    `json:"adults"`
		Children    int    `json:"children"`
		Total       int    `json:"total"`
		FirstTimers int    `json:"firstTimers"`
		Visitors    int    `json:"visitors"`
		Members     int    `json:"members"`
		Notes       string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsedDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	attendance := models.Attendance{
		Date:        parsedDate,
		ServiceType: input.ServiceType,
		Adults:      input.Adults,
		Children:    input.Children,
		Total:       input.Total,
		FirstTimers: input.FirstTimers,
		Visitors:    input.Visitors,
		Members:     input.Members,
		Notes:       input.Notes,
		RecordedBy:  adminEmail,
	}

	if err := database.DB.Create(&attendance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create attendance record"})
		return
	}

	middleware.LogActivity(c, "Created attendance record", input.ServiceType+" on "+input.Date)

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Attendance record created",
		"attendance": attendance,
	})
}

// Admin: Update attendance record
func UpdateAttendance(c *gin.Context) {
	id := c.Param("id")
	adminEmail := c.GetString("adminEmail")

	var attendance models.Attendance
	if err := database.DB.First(&attendance, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attendance record not found"})
		return
	}

	var input struct {
		ServiceType *string `json:"serviceType"`
		Adults      *int    `json:"adults"`
		Children    *int    `json:"children"`
		Total       *int    `json:"total"`
		FirstTimers *int    `json:"firstTimers"`
		Visitors    *int    `json:"visitors"`
		Members     *int    `json:"members"`
		Notes       *string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.ServiceType != nil {
		attendance.ServiceType = *input.ServiceType
	}
	if input.Adults != nil {
		attendance.Adults = *input.Adults
	}
	if input.Children != nil {
		attendance.Children = *input.Children
	}
	if input.Total != nil {
		attendance.Total = *input.Total
	}
	if input.FirstTimers != nil {
		attendance.FirstTimers = *input.FirstTimers
	}
	if input.Visitors != nil {
		attendance.Visitors = *input.Visitors
	}
	if input.Members != nil {
		attendance.Members = *input.Members
	}
	if input.Notes != nil {
		attendance.Notes = *input.Notes
	}
	attendance.RecordedBy = adminEmail

	if err := database.DB.Save(&attendance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update attendance record"})
		return
	}

	middleware.LogActivity(c, "Updated attendance record", attendance.ServiceType+" on "+attendance.Date.Format("2006-01-02"))

	c.JSON(http.StatusOK, gin.H{
		"message":    "Attendance record updated",
		"attendance": attendance,
	})
}

// Admin: Delete attendance record
func DeleteAttendance(c *gin.Context) {
	id := c.Param("id")

	var attendance models.Attendance
	if err := database.DB.First(&attendance, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attendance record not found"})
		return
	}

	database.DB.Delete(&attendance)
	middleware.LogActivity(c, "Deleted attendance record", attendance.ServiceType+" on "+attendance.Date.Format("2006-01-02"))

	c.JSON(http.StatusOK, gin.H{"message": "Attendance record deleted"})
}
