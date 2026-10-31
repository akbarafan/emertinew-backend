package seed

import (
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

// SeedClassifications membuat data classification master
func SeedClassifications(db *gorm.DB) {
	log.Println("[SEED] Seeding classifications...")

	classifications := []model.MClassification{
		{Name: "BKI"},
		{Name: "ABS"},
		{Name: "LR"},
		{Name: "DNV"},
	}

	for _, classification := range classifications {
		var existing model.MClassification
		result := db.Where("name = ?", classification.Name).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&classification).Error; err != nil {
				log.Printf("[SEED] Error creating classification %s: %v", classification.Name, err)
			} else {
				log.Printf("[SEED] Created classification: %s", classification.Name)
			}
		} else {
			log.Printf("[SEED] Skip classification %s - exists", classification.Name)
		}
	}

	log.Println("[SEED] Classification seeding completed")
}
