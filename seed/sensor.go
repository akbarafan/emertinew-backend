package seed

import (
	"log"

	"gorm.io/gorm"
)

// SeedSensors membuat data sensor master dengan mcu_id
// Note: Sensors are now created automatically via SeedVessels -> assignMCUsToVessel -> assignSensorsToMCU
// This function is kept for backward compatibility
func SeedSensors(db *gorm.DB) {
	log.Println("[SEED] Seeding sensors...")
	log.Println("[SEED] Sensors are created automatically when seeding vessels")
	log.Println("[SEED] Sensor seeding completed")
}
