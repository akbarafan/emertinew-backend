package seed

import (
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

// SeedOperationStatuses membuat data operation status master
func SeedOperationStatuses(db *gorm.DB) {
	log.Println("[SEED] Seeding operation statuses...")

	operationStatuses := []model.MOperationStatus{
		{Name: "Active"},
		{Name: "Lay-up"},
		{Name: "Charter"},
	}

	for _, operationStatus := range operationStatuses {
		var existing model.MOperationStatus
		result := db.Where("name = ?", operationStatus.Name).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&operationStatus).Error; err != nil {
				log.Printf("[SEED] Error creating operation status %s: %v", operationStatus.Name, err)
			} else {
				log.Printf("[SEED] Created operation status: %s", operationStatus.Name)
			}
		} else {
			log.Printf("[SEED] Skip operation status %s - exists", operationStatus.Name)
		}
	}

	log.Println("[SEED] Operation status seeding completed")
}
