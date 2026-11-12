package seed

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TrackingSeedConfig configuration for tracking seed
type TrackingSeedConfig struct {
	StartDate time.Time
	EndDate   time.Time
	StartLat  float64
	StartLon  float64
	VesselID  uint
	SensorID  uint
}

// RunTrackingSeed runs the tracking and sensor_data seed
func RunTrackingSeed(db *gorm.DB) {
	log.Println("[SEED] Starting tracking data seeding...")
	log.Println("=====================================")

	// Load timezone
	loc, _ := time.LoadLocation("Asia/Jakarta")

	// Configuration
	config := TrackingSeedConfig{
		StartDate: time.Date(2026, 10, 24, 0, 0, 0, 0, loc),
		EndDate:   time.Date(2026, 10, 30, 23, 59, 0, 0, loc),
		StartLat:  -7.231800,
		StartLon:  113.299201,
		VesselID:  319, // CB Marine Hawk
		SensorID:  1,   // First sensor (GPS)
	}

	// Check if vessel exists
	var vessel model.MVessel
	if err := db.First(&vessel, config.VesselID).Error; err != nil {
		log.Printf("[SEED] Vessel ID %d not found. Please run `make run-seed` first!", config.VesselID)
		return
	}
	log.Printf("[SEED] Found vessel: %s (ID: %d)", vessel.Name, vessel.ID)

	// Generate tracking data
	generateTrackingData(db, config)

	log.Println("=====================================")
	log.Println("[SEED] Tracking data seeding completed!")
}

// generateTrackingData generates realistic ship movement data
func generateTrackingData(db *gorm.DB, cfg TrackingSeedConfig) {
	currentLat := cfg.StartLat
	currentLon := cfg.StartLon
	currentTime := cfg.StartDate

	// Ship movement parameters
	baseSpeed := 8.0 // Base speed in knots
	maxSpeed := 14.0 // Max speed in knots

	// Direction changes (simulate realistic ship route)
	// Ship will generally move northeast then loop back
	directions := []float64{45, 60, 30, 90, 120, 150, 180, 210, 240, 270, 300, 330}
	directionIndex := 0
	currentHeading := directions[directionIndex]

	totalMinutes := int(cfg.EndDate.Sub(cfg.StartDate).Minutes())
	minutesProcessed := 0
	trackingsCreated := 0
	sensorDataCreated := 0

	log.Printf("[SEED] Generating data for %d minutes (%.1f days)", totalMinutes, float64(totalMinutes)/1440.0)

	// Process in batches to avoid memory issues
	batchSize := 1000
	var trackingBatch []model.Tracking
	var sensorDataBatch []model.SensorData

	for currentTime.Before(cfg.EndDate) {
		// Skip some minutes randomly to simulate real-world gaps
		if rand.Float64() < 0.05 { // 5% chance to skip a minute
			currentTime = currentTime.Add(time.Minute)
			minutesProcessed++
			continue
		}

		// Calculate speed with some variation
		speed := baseSpeed + rand.Float64()*(maxSpeed-baseSpeed)
		if rand.Float64() < 0.1 { // 10% chance of slower speed (anchored/waiting)
			speed = rand.Float64() * 3
		}

		// Change direction periodically
		if minutesProcessed%60 == 0 && rand.Float64() < 0.3 { // 30% chance to change direction every hour
			directionIndex = (directionIndex + 1) % len(directions)
			currentHeading = directions[directionIndex] + (rand.Float64()*30 - 15) // Add some randomness
		}

		// Small heading adjustments
		currentHeading += rand.Float64()*4 - 2 // ±2 degrees
		if currentHeading < 0 {
			currentHeading += 360
		}
		if currentHeading >= 360 {
			currentHeading -= 360
		}

		// Calculate new position based on speed and heading
		// Speed is in knots (nautical miles per hour), we need distance per minute
		distanceNM := speed / 60.0 // Distance traveled in 1 minute

		// Convert heading to radians
		headingRad := currentHeading * math.Pi / 180.0

		// Calculate lat/lon change
		// 1 nautical mile = 1 minute of latitude = 1/60 degree
		latChange := distanceNM * math.Cos(headingRad) / 60.0
		// Longitude change depends on latitude
		lonChange := distanceNM * math.Sin(headingRad) / (60.0 * math.Cos(currentLat*math.Pi/180.0))

		currentLat += latChange
		currentLon += lonChange

		// Keep within reasonable bounds (around Indonesia waters)
		if currentLat < -8.5 || currentLat > -6.0 {
			currentHeading = 180 - currentHeading // Reverse latitude direction
			if currentHeading < 0 {
				currentHeading += 360
			}
		}
		if currentLon < 112.0 || currentLon > 115.0 {
			currentHeading = 360 - currentHeading // Reverse longitude direction
		}

		// Generate RPM values (realistic range for ship engines)
		rpm1 := 800 + rand.Intn(400) // 800-1200 RPM
		rpm2 := 800 + rand.Intn(400)
		if speed < 3 { // Idle/slow speed
			rpm1 = 400 + rand.Intn(200)
			rpm2 = 400 + rand.Intn(200)
		}

		// Generate flow values (fuel consumption)
		//
		// `trackings.flow1..4` holds LITRES PER MINUTE: the scheduler adds up
		// every 5-second sample of the minute before storing it. The rates
		// follow the command-center prototype — about 65 L/h for the port
		// engine and 62 L/h for the starboard one — and ease off with the
		// speed.
		flow1 := ((55.0 + speed*1.2) / 60) * (0.9 + rand.Float64()*0.2)
		flow2 := ((52.0 + speed*1.2) / 60) * (0.9 + rand.Float64()*0.2)

		// Generate realistic BMKG weather data
		waveHeight := math.Round((0.4+rand.Float64()*1.3)*100) / 100
		waveCat := "Rendah"
		if waveHeight < 0.5 {
			waveCat = "Tenang"
		} else if waveHeight >= 1.25 {
			waveCat = "Sedang"
		}
		currentSpeed := math.Round((0.4+rand.Float64()*1.8)*10) / 10
		currentDirs := []string{"Barat", "Barat Laut", "Timur", "Tenggara", "Utara", "Selatan"}
		currentTo := currentDirs[rand.Intn(len(currentDirs))]
		windSpeed := math.Round((7.0+rand.Float64()*14.0)*10) / 10
		windGust := math.Round((windSpeed+3.0+rand.Float64()*5.0)*10) / 10
		windDirs := []string{"Timur", "Tenggara", "Selatan", "Barat Daya"}
		windFrom := windDirs[rand.Intn(len(windDirs))]
		weatherCond := "Cerah Berawan"
		if rand.Float64() < 0.2 {
			weatherCond = "Hujan Ringan"
		}

		// Generate environmental sensor data (Bridge Deck & Engine Room)
		topTemp := math.Round((24.0+rand.Float64()*3.5)*10) / 10
		topHumidity := math.Round((55.0+rand.Float64()*12.0)*10) / 10
		botTemp := math.Round((38.5+rand.Float64()*5.5)*10) / 10
		botHumidity := math.Round((60.0+rand.Float64()*14.0)*10) / 10

		// Create tracking record
		tracking := model.Tracking{
			Time:         currentTime,
			VesselID:     cfg.VesselID,
			Lat:          &currentLat,
			Lon:          &currentLon,
			Speed:        &speed,
			Heading:      &currentHeading,
			RPM1:         &rpm1,
			RPM2:         &rpm2,
			Flow1:        &flow1,
			Flow2:        &flow2,
			TopTemp:      &topTemp,
			TopHumidity:  &topHumidity,
			BotTemp:      &botTemp,
			BotHumidity:  &botHumidity,
			WaveHeight:   &waveHeight,
			WaveCat:      &waveCat,
			CurrentSpeed: &currentSpeed,
			CurrentTo:    &currentTo,
			WindSpeed:    &windSpeed,
			WindGust:     &windGust,
			WindFrom:     &windFrom,
			Weather:      &weatherCond,
		}

		trackingBatch = append(trackingBatch, tracking)

		// Create sensor_data record (raw GPS data)
		sensorValue := map[string]interface{}{
			"lat":     currentLat,
			"lon":     currentLon,
			"speed":   speed,
			"heading": currentHeading,
		}
		jsonValue, _ := json.Marshal(sensorValue)

		sensorData := model.SensorData{
			Time:     currentTime,
			SensorID: cfg.SensorID,
			Value:    datatypes.JSON(jsonValue),
			// TrackingID will be set after tracking is created
		}

		sensorDataBatch = append(sensorDataBatch, sensorData)

		// Save batch
		if len(trackingBatch) >= batchSize {
			if err := saveBatch(db, trackingBatch, sensorDataBatch); err != nil {
				log.Printf("[SEED] Error saving batch: %v", err)
			} else {
				trackingsCreated += len(trackingBatch)
				sensorDataCreated += len(sensorDataBatch)
			}
			trackingBatch = nil
			sensorDataBatch = nil

			// Progress log
			progress := float64(minutesProcessed) / float64(totalMinutes) * 100
			log.Printf("[SEED] Progress: %.1f%% (%d trackings, %d sensor_data)", progress, trackingsCreated, sensorDataCreated)
		}

		currentTime = currentTime.Add(time.Minute)
		minutesProcessed++
	}

	// Save remaining batch
	if len(trackingBatch) > 0 {
		if err := saveBatch(db, trackingBatch, sensorDataBatch); err != nil {
			log.Printf("[SEED] Error saving final batch: %v", err)
		} else {
			trackingsCreated += len(trackingBatch)
			sensorDataCreated += len(sensorDataBatch)
		}
	}

	log.Printf("[SEED] Created %d tracking records", trackingsCreated)
	log.Printf("[SEED] Created %d sensor_data records", sensorDataCreated)
}

// saveBatch saves tracking and sensor_data in a transaction
func saveBatch(db *gorm.DB, trackings []model.Tracking, sensorData []model.SensorData) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Insert trackings
		if err := tx.Create(&trackings).Error; err != nil {
			return err
		}

		// Link sensor_data to tracking (only the last one per minute gets the tracking_id)
		for i := range sensorData {
			sensorData[i].TrackingID = &trackings[i].ID
		}

		// Insert sensor_data
		if err := tx.Create(&sensorData).Error; err != nil {
			return err
		}

		return nil
	})
}
