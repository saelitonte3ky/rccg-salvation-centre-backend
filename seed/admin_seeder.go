// seed/admin_seeder.go
package seed

import (
	"log"
	"rccg-salvation-centre-backend/internal/database"
	"rccg-salvation-centre-backend/internal/models"

	"gorm.io/gorm"
)

type AdminRole struct {
	Email string
	Role  string
}

var admins = []AdminRole{
	{Email: "admin@rccgsalvationcentre.org", Role: "superadmin"},
	{Email: "media@rccgsalvationcentre.org", Role: "media_team"},
	{Email: "secretariat@rccgsalvationcentre.org", Role: "secretariat"},
	{Email: "followup@rccgsalvationcentre.org", Role: "visitors_welfare"},
}

func SeedAdmins() {
	for _, admin := range admins {
		var existing models.Admin
		result := database.DB.Where("email = ?", admin.Email).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			newAdmin := models.Admin{
				Email: admin.Email,
				Role:  admin.Role,
			}
			if err := database.DB.Create(&newAdmin).Error; err != nil {
				log.Printf("Failed to seed admin %s: %v", admin.Email, err)
			} else {
				log.Printf("Seeded admin: %s (%s)", admin.Email, admin.Role)
			}
		} else {
			log.Printf("Admin already exists: %s", admin.Email)
		}
	}
}
