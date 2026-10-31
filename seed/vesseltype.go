package seed

import (
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

// SeedVesselTypes membuat data vessel type master
func SeedVesselTypes(db *gorm.DB) {
	log.Println("[SEED] Seeding vessel types...")

	vesselTypes := []model.MVesselType{
		{Name: "Tugboat"},
		{Name: "Cargo"},
		{Name: "Tanker"},
		{Name: "Fishing"},
		{Name: "Patrol"},
	}

	for _, vesselType := range vesselTypes {
		var existing model.MVesselType
		result := db.Where("name = ?", vesselType.Name).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&vesselType).Error; err != nil {
				log.Printf("[SEED] Error creating vessel type %s: %v", vesselType.Name, err)
			} else {
				log.Printf("[SEED] Created vessel type: %s", vesselType.Name)
			}
		} else {
			log.Printf("[SEED] Skip vessel type %s - exists", vesselType.Name)
		}
	}

	log.Println("[SEED] Vessel type seeding completed")
}
