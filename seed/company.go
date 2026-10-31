package seed

import (
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

// SeedCompanies creates company master data
func SeedCompanies(db *gorm.DB) {
	log.Println("[SEED] Seeding companies...")

	companies := []model.MCompany{
		{
			ID:   1,
			Code: "hcml",
			Name: "Husky-CNOOC Madura Limited",
		},
		{
			ID:   2,
			Code: "meratus",
			Name: "PT Meratus Shipping Line",
		},
	}

	for _, company := range companies {
		var existing model.MCompany
		result := db.Where("code = ?", company.Code).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&company).Error; err != nil {
				log.Printf("[SEED] Error creating company %s: %v", company.Name, err)
			} else {
				log.Printf("[SEED] Created company: %s (Code: %s, ID: %d)", company.Name, company.Code, company.ID)
			}
		} else {
			log.Printf("[SEED] Skip company %s - exists", company.Code)
		}
	}

	log.Println("[SEED] Company seeding completed")
}
