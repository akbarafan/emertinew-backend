package seed

import (
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

func SeedGeofences(db *gorm.DB) {
	log.Println("[SEED] Seeding geofences...")

	geofences := []model.MGeofence{
		{
			Name:        "Jalur Pelayaran Madura",
			Description: "Jalur pelayaran utama di perairan Madura",
			Points: model.Points{
				{Latitude: -7.231784, Longitude: 113.299102},
				{Latitude: -7.260868, Longitude: 113.311996},
				{Latitude: -7.289952, Longitude: 113.324890},
				{Latitude: -7.319036, Longitude: 113.337784},
				{Latitude: -7.348120, Longitude: 113.350678},
				{Latitude: -7.377204, Longitude: 113.363572},
				{Latitude: -7.406288, Longitude: 113.376466},
				{Latitude: -7.435372, Longitude: 113.389360},
				{Latitude: -7.464456, Longitude: 113.402254},
				{Latitude: -7.493266, Longitude: 113.415041},
			},
			IsActive:    true,
			RadiusMiles: 5,
		},
	}

	var compID *uint
	var hcmlCompany model.MCompany
	if err := db.Where("code = ?", "hcml").First(&hcmlCompany).Error; err == nil {
		compID = &hcmlCompany.ID
	}

	for _, geofence := range geofences {
		geofence.CompanyID = compID
		var existing model.MGeofence
		result := db.Where("name = ?", geofence.Name).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new geofence
			if err := db.Create(&geofence).Error; err != nil {
				log.Printf("[SEED] Error creating geofence %s: %v", geofence.Name, err)
			} else {
				log.Printf("[SEED] Created geofence: %s", geofence.Name)
			}
		} else {
			if existing.CompanyID == nil && compID != nil {
				existing.CompanyID = compID
				db.Save(&existing)
			}
			log.Printf("[SEED] Skip geofence %s - exists", geofence.Name)
		}
	}

	log.Println("[SEED] Geofence seeding completed")
}
