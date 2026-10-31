package seed

import (
	"fmt"
	"log"
	"strings"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

// SeedVessels membuat data vessel master
func SeedVessels(db *gorm.DB) {
	log.Println("[SEED] Seeding vessels...")

	// Get company ID for HCML
	var company model.MCompany
	db.Where("code = ?", "hcml").First(&company)
	companyID := company.ID

	// Test vessels (ID 1-6)
	testVessels := []model.MVessel{
		{ID: 1, Code: "TS001", Name: "Test Vessel 1", IsActive: true, CompanyID: &companyID},
		{ID: 2, Code: "TS002", Name: "Test Vessel 2", IsActive: true, CompanyID: &companyID},
		{ID: 3, Code: "TS003", Name: "Test Vessel 3", IsActive: true, CompanyID: &companyID},
		{ID: 4, Code: "TS004", Name: "Test Vessel 4", IsActive: true, CompanyID: &companyID},
		{ID: 5, Code: "TS005", Name: "Test Vessel 5", IsActive: true, CompanyID: &companyID},
		{ID: 6, Code: "TS006", Name: "Test Vessel 6", IsActive: true, CompanyID: &companyID},
	}

	for _, vessel := range testVessels {
		var existingVessel model.MVessel
		result := db.Where("name = ?", vessel.Name).First(&existingVessel)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&vessel).Error; err != nil {
				log.Printf("[SEED] Error creating test vessel %s: %v", vessel.Name, err)
			} else {
				log.Printf("[SEED] Created test vessel: %s (ID: %d)", vessel.Name, vessel.ID)
				if vessel.IsActive {
					assignMCUsToVessel(db, vessel.ID, vessel.Name, vessel.Code, true)
				}
			}
		} else {
			log.Printf("[SEED] Skip test vessel %s - exists", vessel.Name)
		}
	}

	// Production vessels
	vessels := []model.MVessel{
		{
			ID:        1,
			Code:      "VS100",
			Name:      "Test Alugara 1",
			IsActive:  false,
			CompanyID: &companyID,
		},
		{
			ID:        2,
			Code:      "VS101",
			Name:      "Test Alugara 2",
			IsActive:  false,
			CompanyID: &companyID,
		},
		{
			ID:        3,
			Code:      "VS102",
			Name:      "Test Alugara 3",
			IsActive:  false,
			CompanyID: &companyID,
		},
		{
			ID:        4,
			Code:      "VS103",
			Name:      "Test Alugara 4",
			IsActive:  false,
			CompanyID: &companyID,
		},
		{
			ID:        5,
			Code:      "VS104",
			Name:      "Test Alugara 5",
			IsActive:  false,
			CompanyID: &companyID,
		},
		{
			ID:        319,
			Code:      "VS001",
			Name:      "CB Marine Hawk",
			IsActive:  true,
			CompanyID: &companyID,
		},
		{
			ID:        320,
			Code:      "VS002",
			Name:      "NMS ACHIEVE",
			IsActive:  true,
			CompanyID: &companyID,
		},
		{
			ID:        321,
			Code:      "VS003",
			Name:      "AHT. LIMIN",
			IsActive:  true,
			CompanyID: &companyID,
		},
		{
			ID:        322,
			Code:      "VS004",
			Name:      "CB ALDWIN",
			IsActive:  true,
			CompanyID: &companyID,
		},
		{
			ID:        323,
			Code:      "VS005",
			Name:      "OPS ALHENA",
			IsActive:  true,
			CompanyID: &companyID,
		},
		{
			ID:        324,
			Code:      "VS006",
			Name:      "AHT KYLIE S",
			IsActive:  true,
			CompanyID: &companyID,
		},
	}

	for _, vessel := range vessels {

		// Check if vessel already exists
		var existingVessel model.MVessel
		result := db.Where("name = ?", vessel.Name).First(&existingVessel)

		if result.Error == gorm.ErrRecordNotFound {
			// Vessel doesn't exist, create it
			if err := db.Create(&vessel).Error; err != nil {
				log.Printf("[SEED] Error creating vessel %s: %v", vessel.Name, err)
			} else {
				log.Printf("[SEED] Created vessel: %s (ID: %d, Active: %v)", vessel.Name, vessel.ID, vessel.IsActive)

				// Assign MCUs to vessels (only to active vessels)
				if vessel.IsActive {
					assignMCUsToVessel(db, vessel.ID, vessel.Name, vessel.Code, true)
				}
			}
		} else if vessel.IsActive {
			log.Printf("[SEED] Vessel %s exists, ensuring MCUs and sensors...", vessel.Name)
			assignMCUsToVessel(db, existingVessel.ID, vessel.Name, vessel.Code, true)
		}
	}

	log.Println("[SEED] Vessel seeding completed")
}

// assignMCUsToVessel creates and assigns MCUs to vessel
// MCU types: bd (bridge deck), er (engine room)
// flowKFactor returns the calibration factor of a flow channel, in pulse per
// litre.
//
// The values come from the hardware configuration of the vessels
// (`FLOW_USE=A_BIRU,B_BIRU,C_ORANGE,D_ORANGE` in the vessel .env):
//
//	FLOW_A_BIRU    36.34
//	FLOW_B_BIRU    33.79
//	FLOW_C_ORANGE  24.97
//	FLOW_D_ORANGE  25.58
//
// They are indexed by channel, so channel 1 and 2 are the blue pair and
// channel 3 and 4 the orange pair, exactly like the GPIO order on the Pi.
func flowKFactor(channel int) *float64 {
	factors := map[int]float64{
		1: 36.34,
		2: 33.79,
		3: 24.97,
		4: 25.58,
	}
	if value, ok := factors[channel]; ok {
		return &value
	}
	return nil
}

func assignMCUsToVessel(db *gorm.DB, vesselID uint, vesselName string, vesselCode string, withSensors bool) {
	mcuTypes := []struct {
		Location string
		Name     string
	}{
		{Location: "bd", Name: "Bridge Deck MCU " + vesselName},
		{Location: "er", Name: "Engine Room MCU " + vesselName},
	}

	for _, mcuType := range mcuTypes {
		var existing model.MMCU
		if err := db.Where("vessel_id = ? AND location = ?", vesselID, mcuType.Location).First(&existing).Error; err == gorm.ErrRecordNotFound {
			mcu := model.MMCU{
				VesselID: &vesselID,
				Name:     mcuType.Name,
				Location: mcuType.Location,
			}
			if err := db.Create(&mcu).Error; err != nil {
				log.Printf("[SEED] Error creating MCU %s: %v", mcuType.Name, err)
			} else {
				log.Printf("[SEED] Created MCU: %s (ID: %d)", mcu.Name, mcu.ID)
				// Create sensors for this MCU
				if withSensors {
					assignSensorsToMCU(db, mcu.ID, mcuType.Location, vesselName, vesselCode)
				}
			}
		} else {
			log.Printf("[SEED] MCU %s exists (ID: %d), ensuring sensors...", existing.Name, existing.ID)
			if withSensors {
				assignSensorsToMCU(db, existing.ID, mcuType.Location, vesselName, vesselCode)
			}
		}
	}
}

// assignSensorsToMCU creates sensors for an MCU
// Topic format: <company>/<vessel>/<location>/<sensor_group>/<number>
// Example: hcml/vs001/bd/gps/1, hcml/vs001/er/rpm/1
// Locations: bd (bridge deck), er (engine room)
func assignSensorsToMCU(db *gorm.DB, mcuID uint, mcuType, vesselName string, vesselCode string) {
	vesselCode = strings.ToLower(vesselCode) // e.g. vs001, ts001
	companyCode := "hcml"

	var sensors []model.MSensor

	if mcuType == "bd" {
		// Bridge Deck sensors
		location := "bd"
		sensors = []model.MSensor{
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/gps/1", companyCode, vesselCode, location), Name: "GPS Sensor " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/ac/1", companyCode, vesselCode, location), Name: "AC Sensor BD " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/dc/1", companyCode, vesselCode, location), Name: "DC Kapal Sensor BD " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/dc/2", companyCode, vesselCode, location), Name: "DC Baterai Sensor BD " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/door/1", companyCode, vesselCode, location), Name: "Door Sensor 1 BD " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/door/2", companyCode, vesselCode, location), Name: "Door Sensor 2 BD " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/temp/1", companyCode, vesselCode, location), Name: "Temperature Sensor BD " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/humidity/1", companyCode, vesselCode, location), Name: "Humidity Sensor BD " + vesselName},
		}
	} else {
		// Engine Room sensors
		location := "er"
		sensors = []model.MSensor{
			// 4 RPM sensors
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/rpm/1", companyCode, vesselCode, location), Name: "RPM Sensor 1 " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/rpm/2", companyCode, vesselCode, location), Name: "RPM Sensor 2 " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/rpm/3", companyCode, vesselCode, location), Name: "RPM Sensor 3 " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/rpm/4", companyCode, vesselCode, location), Name: "RPM Sensor 4 " + vesselName},
			// 4 Flow sensors
			//
			// The calibration factor is the vessel's own, taken from the
			// hardware .env (FLOW_USE=A_BIRU,B_BIRU,C_ORANGE,D_ORANGE):
			// 36.34, 33.79, 24.97 and 25.58 pulse per litre. Without it the
			// extractor skips every reading, so a seeded meter would look
			// healthy and report nothing.
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/flow/1", companyCode, vesselCode, location), Name: "Flow Sensor 1 " + vesselName, KFactor: flowKFactor(1)},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/flow/2", companyCode, vesselCode, location), Name: "Flow Sensor 2 " + vesselName, KFactor: flowKFactor(2)},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/flow/3", companyCode, vesselCode, location), Name: "Flow Sensor 3 " + vesselName, KFactor: flowKFactor(3)},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/flow/4", companyCode, vesselCode, location), Name: "Flow Sensor 4 " + vesselName, KFactor: flowKFactor(4)},
			// Electrical
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/ac/1", companyCode, vesselCode, location), Name: "AC Sensor ER " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/dc/1", companyCode, vesselCode, location), Name: "DC Kapal Sensor ER " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/dc/2", companyCode, vesselCode, location), Name: "DC Baterai Sensor ER " + vesselName},
			// Environment
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/door/1", companyCode, vesselCode, location), Name: "Door Sensor 1 ER " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/door/2", companyCode, vesselCode, location), Name: "Door Sensor 2 ER " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/temp/1", companyCode, vesselCode, location), Name: "Temperature Sensor ER " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/humidity/1", companyCode, vesselCode, location), Name: "Humidity Sensor ER " + vesselName},
			{MCUID: &mcuID, Topic: fmt.Sprintf("%s/%s/%s/accel/1", companyCode, vesselCode, location), Name: "Accelerometer Sensor " + vesselName},
		}
	}

	for _, sensor := range sensors {
		// TopicKey must be set to satisfy unique constraint idx_m_sensors_topic_key.
		// Format: company:vessel_code:location:sensor_group:sensor_number
		sensor.TopicKey = strings.ReplaceAll(sensor.Topic, "/", ":")

		var existing model.MSensor
		if err := db.Where("topic = ?", sensor.Topic).First(&existing).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&sensor).Error; err != nil {
				log.Printf("[SEED] Error creating sensor %s: %v", sensor.Topic, err)
			} else {
				log.Printf("[SEED] Created sensor: %s - %s", sensor.Topic, sensor.Name)
			}
		} else if existing.TopicKey == "" {
			// Fix topic_key for sensors created before the topic_key fix
			db.Model(&existing).Update("topic_key", sensor.TopicKey)
			log.Printf("[SEED] Fixed topic_key for existing sensor: %s -> %s", sensor.Topic, sensor.TopicKey)
		}
	}
}
