// seed/service_type_seeder.go
package seed

import (
	"log"
	"rccg-salvation-centre-backend/internal/database"
	"rccg-salvation-centre-backend/internal/models"

	"gorm.io/gorm"
)

var serviceTypes = []string{
	"Sunday Service",
	"Midweek Service",
	"Bible Study",
	"Special Service",
	"Youth Service",
	"Workers Meeting",
	"Revival",
	"Crusade",
	"Convention",
	"Holy Ghost Service",
	"Thanksgiving Service",
	"Anointing Service",
}

func SeedServiceTypes() {
	for _, name := range serviceTypes {
		var existing models.ServiceType
		if err := database.DB.Where("name = ?", name).First(&existing).Error; err == gorm.ErrRecordNotFound {
			st := models.ServiceType{Name: name}
			if err := database.DB.Create(&st).Error; err != nil {
				log.Printf("Failed to seed service type %s: %v", name, err)
			} else {
				log.Printf("Seeded service type: %s", name)
			}
		}
	}
}
