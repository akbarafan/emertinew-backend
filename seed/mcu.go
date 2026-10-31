package seed

import (
	"log"

	"gorm.io/gorm"
)

// SeedMCUs membuat data MCU master
// Note: MCUs are now created automatically via SeedVessels -> assignMCUsToVessel
// This function is kept for backward compatibility
func SeedMCUs(db *gorm.DB) {
	log.Println("[SEED] Seeding MCUs...")
	log.Println("[SEED] MCUs are created automatically when seeding vessels")
	log.Println("[SEED] MCU seeding completed")
}
