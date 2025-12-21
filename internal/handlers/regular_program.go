// internal/handlers/regular_program.go
package handlers

import (
	"net/http"

	"rccg-salvation-centre-backend/internal/database"
	"rccg-salvation-centre-backend/internal/middleware"
	"rccg-salvation-centre-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// Public: Get active regular programs
func GetRegularPrograms(c *gin.Context) {
	var programs []models.RegularProgram
	database.DB.Where("active = ?", true).Find(&programs)
	c.JSON(http.StatusOK, gin.H{"data": programs})
}

// Admin: Get all regular programs
func AdminGetRegularPrograms(c *gin.Context) {
	var programs []models.RegularProgram
	database.DB.Find(&programs)
	c.JSON(http.StatusOK, gin.H{"data": programs})
}

// Admin: Create regular program
func CreateRegularProgram(c *gin.Context) {
	var input struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Day         string `json:"day" binding:"required"`
		Frequency   string `json:"frequency" binding:"required"`
		Time        string `json:"time"`
		Location    string `json:"location"`
		Type        string `json:"type" binding:"required"` // ADDED
		Active      bool   `json:"active"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	program := models.RegularProgram{
		Title:       input.Title,
		Description: input.Description,
		Day:         input.Day,
		Frequency:   input.Frequency,
		Time:        input.Time,
		Location:    input.Location,
		Type:        input.Type, // ADDED
		Active:      input.Active,
	}

	if err := database.DB.Create(&program).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create program"})
		return
	}

	middleware.LogActivity(c, "Created regular program", program.Title)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Regular program created",
		"program": program,
	})
}

// Admin: Update regular program
func UpdateRegularProgram(c *gin.Context) {
	id := c.Param("id")

	var program models.RegularProgram
	if err := database.DB.First(&program, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Program not found"})
		return
	}

	var input struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Day         *string `json:"day"`
		Frequency   *string `json:"frequency"`
		Time        *string `json:"time"`
		Location    *string `json:"location"`
		Type        *string `json:"type"` // ADDED
		Active      *bool   `json:"active"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Title != nil {
		program.Title = *input.Title
	}
	if input.Description != nil {
		program.Description = *input.Description
	}
	if input.Day != nil {
		program.Day = *input.Day
	}
	if input.Frequency != nil {
		program.Frequency = *input.Frequency
	}
	if input.Time != nil {
		program.Time = *input.Time
	}
	if input.Location != nil {
		program.Location = *input.Location
	}
	if input.Type != nil { // ADDED
		program.Type = *input.Type
	}
	if input.Active != nil {
		program.Active = *input.Active
	}

	if err := database.DB.Save(&program).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update program"})
		return
	}

	middleware.LogActivity(c, "Updated regular program", program.Title)

	c.JSON(http.StatusOK, gin.H{
		"message": "Regular program updated",
		"program": program,
	})
}

// Admin: Delete regular program
func DeleteRegularProgram(c *gin.Context) {
	id := c.Param("id")

	var program models.RegularProgram
	if err := database.DB.First(&program, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Program not found"})
		return
	}

	database.DB.Delete(&program)
	middleware.LogActivity(c, "Deleted regular program", program.Title)

	c.JSON(http.StatusOK, gin.H{"message": "Regular program deleted"})
}
