package seed

import (
	"log"

	"gorm.io/gorm"
)

// RunAll menjalankan semua seed functions
func RunAll(db *gorm.DB) {
	log.Println("[SEED] Starting database seeding...")
	log.Println("================================")

	// Seed users (superadmin, admin, guest)
	SeedUsers(db)

	// Seed companies (must be before vessels)
	SeedCompanies(db)

	// Seed master data for vessel lookup tables
	SeedClassifications(db)
	SeedVesselTypes(db)
	SeedOperationStatuses(db)
	SeedMaterialHulls(db)

	// Seed MCUs first
	SeedMCUs(db)

	// Seed sensors with mcu_id
	SeedSensors(db)

	// Seed vessels (will assign MCUs to vessels)
	SeedVessels(db)

	// Seed offshore locations
	SeedOffshores(db)

	// Seed ports
	SeedPorts(db)

	// Seed geofences
	SeedGeofences(db)

	log.Println("=================================")
	log.Println("[SEED] Database seeding completed!")
}
