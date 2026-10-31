package seed

import (
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

// SeedOffshores membuat data offshore master
func SeedOffshores(db *gorm.DB) {
	log.Println("[SEED] Seeding offshores...")

	offshores := []model.MOffshore{
		{
			Name: "BD North",
			Code: "BD-NORTH",
			Type: "buoys",
			Lat:  -7.364350,
			Long: 113.276267,
		},
		{
			Name: "BD South",
			Code: "BD-SOUTH",
			Type: "buoys",
			Lat:  -7.379883,
			Long: 113.279367,
		},
		{
			Name: "FPSO",
			Code: "FPSO",
			Type: "field",
			Lat:  -7.374217,
			Long: 113.277278,
		},
		{
			Name: "BD WHP",
			Code: "BD-WHP",
			Type: "field",
			Lat:  -7.372619,
			Long: 113.274389,
		},
		{
			Name: "MAC/MOPU",
			Code: "MAC-MOPU",
			Type: "field",
			Lat:  -7.385469,
			Long: 113.843269,
		},
		{
			Name: "MAC",
			Code: "MAC",
			Type: "buoys",
			Lat:  -7.381500,
			Long: 113.838200,
		},
		{
			Name: "MBH",
			Code: "MBH",
			Type: "buoys",
			Lat:  -7.319067,
			Long: 114.305867,
		},
		{
			Name: "FPU",
			Code: "FPU",
			Type: "field",
			Lat:  -7.312033,
			Long: 114.307694,
		},
		{
			Name: "MBH Field",
			Code: "MBH-FIELD",
			Type: "field",
			Lat:  -7.312344,
			Long: 114.309806,
		},
		{
			Name: "MDA",
			Code: "MDA",
			Type: "buoys",
			Lat:  -7.329583,
			Long: 114.548033,
		},
		{
			Name: "MDA Field",
			Code: "MDA-FIELD",
			Type: "field",
			Lat:  -7.325056,
			Long: 114.548944,
		},
	}

	var compID *uint
	var hcmlCompany model.MCompany
	if err := db.Where("code = ?", "hcml").First(&hcmlCompany).Error; err == nil {
		compID = &hcmlCompany.ID
	}

	for _, offshore := range offshores {
		offshore.CompanyID = compID
		// Check if offshore already exists
		var existingOffshore model.MOffshore
		result := db.Where("name = ?", offshore.Name).First(&existingOffshore)

		if result.Error == gorm.ErrRecordNotFound {
			// Offshore doesn't exist, create it
			if err := db.Create(&offshore).Error; err != nil {
				log.Printf("[SEED] Error creating offshore %s: %v", offshore.Name, err)
			} else {
				log.Printf("[SEED] Created offshore: %s (Code: %s, Lat: %.6f, Long: %.6f)",
					offshore.Name, offshore.Code, offshore.Lat, offshore.Long)
			}
		} else {
			if existingOffshore.CompanyID == nil && compID != nil {
				existingOffshore.CompanyID = compID
				db.Save(&existingOffshore)
			}
			log.Printf("[SEED] Skip offshore %s - exists", offshore.Name)
		}
	}

	log.Println("[SEED] Offshore seeding completed")
}
