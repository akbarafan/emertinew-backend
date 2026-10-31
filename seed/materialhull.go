package seed

import (
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

// SeedMaterialHulls membuat data material hull master
func SeedMaterialHulls(db *gorm.DB) {
	log.Println("[SEED] Seeding material hulls...")

	materialHulls := []model.MMaterialHull{
		{Name: "Steel"},
		{Name: "Alumunium"},
		{Name: "FRP"},
	}

	for _, materialHull := range materialHulls {
		var existing model.MMaterialHull
		result := db.Where("name = ?", materialHull.Name).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&materialHull).Error; err != nil {
				log.Printf("[SEED] Error creating material hull %s: %v", materialHull.Name, err)
			} else {
				log.Printf("[SEED] Created material hull: %s", materialHull.Name)
			}
		} else {
			log.Printf("[SEED] Skip material hull %s - exists", materialHull.Name)
		}
	}

	log.Println("[SEED] Material hull seeding completed")
}
