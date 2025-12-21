// internal/handlers/service_type.go
package handlers

import (
	"net/http"
	"rccg-salvation-centre-backend/internal/database"
	"rccg-salvation-centre-backend/internal/models"

	"github.com/gin-gonic/gin"
)

func GetServiceTypes(c *gin.Context) {
	var types []models.ServiceType
	if err := database.DB.Order("name ASC").Find(&types).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load service types"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    types,
	})
}
