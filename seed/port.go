package seed

import (
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

// SeedPorts membuat data port master
func SeedPorts(db *gorm.DB) {
	log.Println("[SEED] Seeding ports...")

	ports := []model.MPort{
		{
			Name: "KALIANGET",
			Code: "KALIANGET",
			Lat:  -7.054889,
			Long: 113.945389,
		},
		{
			Name: "TADDAN",
			Code: "TADDAN",
			Lat:  -7.231800,
			Long: 113.299201,
		},
	}

	for _, port := range ports {
		// Check if port already exists
		var existingPort model.MPort
		result := db.Where("name = ?", port.Name).First(&existingPort)

		if result.Error == gorm.ErrRecordNotFound {
			// Port doesn't exist, create it
			if err := db.Create(&port).Error; err != nil {
				log.Printf("[SEED] Error creating port %s: %v", port.Name, err)
			} else {
				log.Printf("[SEED] Created port: %s (Code: %s, Lat: %.6f, Long: %.6f)",
					port.Name, port.Code, port.Lat, port.Long)
			}
		} else {
			log.Printf("[SEED] Skip port %s - exists", port.Name)
		}
	}

	log.Println("[SEED] Port seeding completed")
}
