package tracking

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/tracking/extractor"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TrackingScheduler processes sensor_data and creates tracking records
type TrackingScheduler struct {
	db       *gorm.DB
	stopChan chan struct{}
	label    string
}

// NewTrackingScheduler creates a new scheduler.
func NewTrackingScheduler(db *gorm.DB, label string) *TrackingScheduler {
	return &TrackingScheduler{
		db:       db,
		stopChan: make(chan struct{}),
		label:    label,
	}
}

// parseTopic extracts sensor type, location, and number from topic.
// Supports two topic formats:
//
//	Old (5-part): hcml/vs001/bd/gps/1  → location=parts[2], type=parts[3], num=parts[4]
//	New (7-part): hcml/flt001/sqd001/vs001/brd/fuel/01f → location=parts[4], type=parts[5], num=parts[6]
//
// Location mapping: bd/brd (bridge deck) → isTop, er/eng (engine room) → isBot
func parseTopic(topic string) (sensorType string, sensorNumber int, isTop bool, isBot bool) {
	parts := strings.Split(strings.ToLower(topic), "/")
	if len(parts) < 5 {
		return "", 0, false, false
	}

	var location string
	switch len(parts) {
	case 5:
		location = parts[2]
		sensorType = parts[3]
		if n, err := strconv.Atoi(parts[4]); err == nil {
			sensorNumber = n
		}
	case 7:
		location = parts[4]
		sensorType = parts[5]
		if n, err := strconv.Atoi(parts[6]); err == nil {
			sensorNumber = n
		}
	default:
		location = parts[2]
		sensorType = parts[3]
		if n, err := strconv.Atoi(parts[4]); err == nil {
			sensorNumber = n
		}
	}

	// Location mapping
	isTop = location == "bd" || location == "brd" // bridge deck
	isBot = location == "er" || location == "eng" // engine room

	return sensorType, sensorNumber, isTop, isBot
}

// maxLatePerCycle limits how many late vessel-minutes are processed per scheduler tick.
// ORDER BY minute ASC ensures the oldest pending data (e.g. GPS backfill after
// hardware reconnection) is processed first, preventing data from being
// perpetually starved by newer late arrivals.
const maxLatePerCycle = 2000

// Start begins the scheduler that runs every minute
func (s *TrackingScheduler) Start() {
	go func() {
		s.processUnlinkedSensorData()

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.processUnlinkedSensorData()
			case <-s.stopChan:
				log.Printf("[SCHEDULER-%s] Stopped", s.label)
				return
			}
		}
	}()
	log.Printf("[SCHEDULER-%s] Started (runs every 1 minute)", s.label)
}

// Stop gracefully stops the scheduler
func (s *TrackingScheduler) Stop() {
	close(s.stopChan)
}

// processUnlinkedSensorData finds all sensor_data without tracking and creates tracking for completed minutes.
// Processing is split into two phases to prevent late data from blocking actual (current) data:
//
//	Phase 1: Process ACTUAL data (most recent completed minute) — always processed first
//	Phase 2: Process LATE data (older minutes) — limited to maxLatePerCycle per tick
func (s *TrackingScheduler) processUnlinkedSensorData() {
	currentMinute := time.Now().Truncate(time.Minute)
	recentMinute := currentMinute.Add(-1 * time.Minute)

	type VesselMinute struct {
		VesselID uint
		Minute   time.Time
	}

	// ── Phase 1: ACTUAL data (recent minute, processed first) ──
	var actualMinutes []VesselMinute
	err := s.db.Raw(`
		SELECT DISTINCT 
			m.vessel_id,
			date_trunc('minute', sd.time) as minute
		FROM sensor_data sd
		JOIN m_sensors s ON sd.sensor_id = s.id
		JOIN m_mcu m ON s.mcu_id = m.id
		WHERE sd.tracking_id IS NULL
		AND sd.time BETWEEN ? AND ?
		AND date_trunc('minute', sd.time) = ?
	`, recentMinute.Add(-5*time.Minute), recentMinute.Add(5*time.Minute), recentMinute).Scan(&actualMinutes).Error

	if err != nil {
		log.Printf("[SCHEDULER-%s] Error querying actual data: %v", s.label, err)
		return
	}

	actualOK, actualErr := 0, 0
	for _, vm := range actualMinutes {
		if err := s.createTrackingForMinute(vm.VesselID, vm.Minute); err != nil {
			actualErr++
			log.Printf("[SCHEDULER-%s] ACTUAL error vessel=%d minute=%s: %v", s.label, vm.VesselID, vm.Minute.Format("15:04"), err)
		} else {
			actualOK++
		}
	}
	if len(actualMinutes) > 0 {
		log.Printf("[SCHEDULER-%s] ACTUAL: %d processed, %d errors", s.label, actualOK, actualErr)
	}

	// ── Phase 2: LATE data (older minutes, process newest first for fast catch-up) ──
	var lateMinutes []VesselMinute
	err = s.db.Raw(`
		SELECT DISTINCT 
			m.vessel_id,
			date_trunc('minute', sd.time) as minute
		FROM sensor_data sd
		JOIN m_sensors s ON sd.sensor_id = s.id
		JOIN m_mcu m ON s.mcu_id = m.id
		WHERE sd.tracking_id IS NULL
		AND date_trunc('minute', sd.time) < ?
		ORDER BY minute DESC
		LIMIT ?
	`, recentMinute, maxLatePerCycle).Scan(&lateMinutes).Error

	if err != nil {
		log.Printf("[SCHEDULER-%s] Error querying late data: %v", s.label, err)
		return
	}

	lateOK, lateErr := 0, 0
	for _, vm := range lateMinutes {
		if err := s.createTrackingForMinute(vm.VesselID, vm.Minute); err != nil {
			lateErr++
			log.Printf("[SCHEDULER-%s] LATE error vessel=%d minute=%s: %v", s.label, vm.VesselID, vm.Minute.Format("15:04"), err)
		} else {
			lateOK++
		}
	}
	if len(lateMinutes) > 0 {
		log.Printf("[SCHEDULER-%s] LATE: %d processed, %d errors (batch limit %d)", s.label, lateOK, lateErr, maxLatePerCycle)
	}
}

// createTrackingForMinute creates or updates a tracking record from sensor_data of a specific minute
// Handles late-arriving data by updating existing tracking if it already exists
func (s *TrackingScheduler) createTrackingForMinute(vesselID uint, minute time.Time) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	minuteEnd := minute.Add(time.Minute)

	// Check if tracking already exists for this minute (silenced: "record not found" is expected)
	var existingTracking model.Tracking
	trackingExists := false
	if err := tx.Session(&gorm.Session{Logger: logger.Discard}).Where("vessel_id = ? AND time = ?", vesselID, minute).First(&existingTracking).Error; err == nil {
		trackingExists = true
	}

	// Get all sensors for this vessel
	var sensors []model.MSensor
	if err := tx.Joins("JOIN m_mcu ON m_sensors.mcu_id = m_mcu.id").
		Where("m_mcu.vessel_id = ?", vesselID).
		Find(&sensors).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(sensors) == 0 {
		tx.Rollback()
		return nil
	}

	// Collect sensor IDs
	sensorIDs := make([]uint, len(sensors))
	for i, s := range sensors {
		sensorIDs[i] = s.ID
	}

	// Get ALL sensor_data for this minute
	var sensorDataList []model.SensorData
	err := tx.Raw(`
		SELECT * FROM sensor_data
		WHERE sensor_id IN ?
		AND time BETWEEN ? AND ?
		AND date_trunc('minute', time) >= ?
		AND date_trunc('minute', time) < ?
		ORDER BY time DESC
	`, sensorIDs, minute.Add(-5*time.Minute), minuteEnd.Add(5*time.Minute), minute, minuteEnd).Scan(&sensorDataList).Error

	if err != nil || len(sensorDataList) == 0 {
		tx.Rollback()
		return err
	}

	// Build sensor map for topic lookup (sensor_id -> topic)
	sensorMap := make(map[uint]string)
	sensorKFactorMap := make(map[uint]float64)
	for _, s := range sensors {
		sensorMap[s.ID] = s.Topic
		if s.KFactor != nil {
			sensorKFactorMap[s.ID] = *s.KFactor
		}
	}

	// Aggregate latest values from all sensor_data in this minute
	tracking := model.Tracking{
		Time:     minute,
		VesselID: vesselID,
	}

	// Process each sensor_data (already sorted by time DESC, so first occurrence is latest)
	processedFields := make(map[string]bool)

	for _, sd := range sensorDataList {
		var sensorValue map[string]interface{}
		if err := json.Unmarshal(sd.Value, &sensorValue); err != nil {
			continue
		}

		// Get topic for this sensor
		topic, ok := sensorMap[sd.SensorID]
		if !ok {
			continue
		}

		// Parse sensor type, number, and location from topic
		// Topic format: hcml/vs001/bd/gps/1 → sensorType="gps", sensorNumber=1, location="bd"
		sensorType, sensorNumber, isTopDevice, isBotDevice := parseTopic(topic)

		// Call only the appropriate extractor based on sensor type
		switch sensorType {
		case "gps":
			extractor.GPS(&tracking, sensorValue, processedFields)
		case "rpm":
			extractor.RPM(&tracking, sensorValue, processedFields, sensorNumber)
		case "flow":
			// Accumulate, do not overwrite: a `trackings` row represents one
			// minute, and the hardware publishes one sample every 5 seconds, so
			// the row has to hold the litres that passed during that whole
			// minute. Using the latest sample only would store 1/12 of the
			// real volume and make every daily total read low by that factor.
			extractor.FlowAccumulate(&tracking, sensorValue, processedFields, sensorNumber, sensorKFactorMap[sd.SensorID])
		case "ac":
			extractor.AC(&tracking, sensorValue, processedFields, isTopDevice, isBotDevice)
		case "dc":
			extractor.DC(&tracking, sensorValue, processedFields, isTopDevice, isBotDevice)
		case "bat":
			extractor.Battery(&tracking, sensorValue, processedFields, isTopDevice, isBotDevice)
		case "accel":
			extractor.Accelerometer(&tracking, sensorValue, processedFields)
		case "temp":
			extractor.Temperature(&tracking, sensorValue, processedFields, isTopDevice, isBotDevice)
		case "humidity":
			extractor.Humidity(&tracking, sensorValue, processedFields, isTopDevice, isBotDevice)
		case "door":
			extractor.Door(&tracking, sensorValue, processedFields, isTopDevice, isBotDevice, sensorNumber)
		}
	}

	// Validate coordinate range — nullify if out of bounds before spike check
	if tracking.Lat != nil && tracking.Lon != nil {
		if *tracking.Lon > 180 || *tracking.Lon < -180 || *tracking.Lat > 90 || *tracking.Lat < -90 {
			log.Printf("[GPS-FILTER] Invalid coords: vessel=%d minute=%s (%.4f, %.4f)",
				vesselID, minute.Format("15:04"), *tracking.Lat, *tracking.Lon)
			tracking.Lat = nil
			tracking.Lon = nil
		}
	}

	// Get LATEST previous tracking for speed/heading calculation (not just previous minute)
	var prevTracking model.Tracking
	tx.Where("vessel_id = ? AND time < ?", vesselID, minute).
		Order("time DESC").
		First(&prevTracking)

	// Calculate speed & heading from previous tracking
	if prevTracking.VesselID > 0 && prevTracking.Lat != nil && prevTracking.Lon != nil &&
		tracking.Lat != nil && tracking.Lon != nil {
		timeDiff := minute.Sub(prevTracking.Time).Hours()

		// GPS anomaly filter — reject if position jumped unrealistically far
		if isGPSSpike(*prevTracking.Lat, *prevTracking.Lon, *tracking.Lat, *tracking.Lon, timeDiff) {
			log.Printf("[GPS-FILTER] Spike rejected: vessel=%d minute=%s (%.4f,%.4f) → (%.4f,%.4f) in %.1f min",
				vesselID, minute.Format("15:04"), *prevTracking.Lat, *prevTracking.Lon, *tracking.Lat, *tracking.Lon, timeDiff*60)
			tracking.Lat = nil
			tracking.Lon = nil
		} else {
			// Calculate speed (includes 0 for stationary vessel)
			speed := CalculateSpeed(*prevTracking.Lat, *prevTracking.Lon, *tracking.Lat, *tracking.Lon, timeDiff)
			tracking.Speed = &speed

			heading := RoundToTwoDecimals(CalculateBearing(*prevTracking.Lat, *prevTracking.Lon, *tracking.Lat, *tracking.Lon))
			tracking.Heading = &heading
		}
	}

	var trackingID uint
	if trackingExists {
		// UPDATE existing tracking with late-arriving data
		updates := mergeTrackingUpdates(&existingTracking, &tracking)
		if err := tx.Model(&existingTracking).Updates(updates).Error; err != nil {
			tx.Rollback()
			return err
		}
		trackingID = existingTracking.ID
	} else {
		// CREATE new tracking
		if err := tx.Create(&tracking).Error; err != nil {
			tx.Rollback()
			return err
		}
		trackingID = tracking.ID
	}

	// Link the latest sensor_data to this tracking
	if err := s.linkLatestSensorData(tx, vesselID, minute, minuteEnd, trackingID); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Broadcast tracking via WebSocket after successful commit
	// Use merged tracking for broadcast (existing + new data)
	if trackingExists {
		s.broadcastTracking(&existingTracking, vesselID)
	} else {
		s.broadcastTracking(&tracking, vesselID)
	}

	return nil
}

// linkLatestSensorData links all unlinked sensor_data of each sensor to the tracking
// Prevents orphan records from keeping the minute alive in scheduler queries.
func (s *TrackingScheduler) linkLatestSensorData(tx *gorm.DB, vesselID uint, minute, minuteEnd time.Time, trackingID uint) error {
	var sensors []model.MSensor
	if err := tx.Joins("JOIN m_mcu ON m_sensors.mcu_id = m_mcu.id").
		Where("m_mcu.vessel_id = ?", vesselID).
		Find(&sensors).Error; err != nil {
		return err
	}

	for _, sensor := range sensors {
		err := tx.Model(&model.SensorData{}).
			Where("sensor_id = ?", sensor.ID).
			Where("time BETWEEN ? AND ?", minute.Add(-5*time.Minute), minuteEnd.Add(5*time.Minute)).
			Where("date_trunc('minute', time) >= ?", minute).
			Where("date_trunc('minute', time) < ?", minuteEnd).
			Where("tracking_id IS NULL").
			Update("tracking_id", trackingID).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// broadcastTracking sends tracking data via WebSocket
func (s *TrackingScheduler) broadcastTracking(tracking *model.Tracking, vesselID uint) {
	if Hub == nil || SchedulerService == nil {
		return
	}

	// Use service to convert tracking to DTO
	svc, ok := SchedulerService.(*service)
	if !ok {
		log.Printf("[SCHEDULER] Broadcast skipped: cannot cast SchedulerService")
		return
	}

	dto := svc.convertTrackingToDTO(tracking, vesselID)
	if dto == nil {
		return
	}

	// Wrap with type marker for scheduled data (per-minute aggregate)
	scheduledData := map[string]interface{}{
		"type": "scheduled",
		"data": dto,
	}
	jsonData, err := json.Marshal(scheduledData)
	if err != nil {
		return
	}

	// Use TrackingBroadcast for filtering (only vessel_id)
	broadcastFilter := &TrackingBroadcast{
		VesselID: vesselID,
	}

	Hub.Broadcast(broadcastFilter, jsonData)
}

// mergeTrackingUpdates creates a map of updates from new tracking data
// Only non-nil fields in newTracking will be included (late-arriving data)
// This preserves existing data while adding new sensor values
func mergeTrackingUpdates(existing, newData *model.Tracking) map[string]interface{} {
	updates := make(map[string]interface{})

	// GPS data
	if newData.Lat != nil {
		updates["lat"] = newData.Lat
		existing.Lat = newData.Lat
	}
	if newData.Lon != nil {
		updates["lon"] = newData.Lon
		existing.Lon = newData.Lon
	}
	if newData.Speed != nil {
		updates["speed"] = newData.Speed
		existing.Speed = newData.Speed
	}
	if newData.Heading != nil {
		updates["heading"] = newData.Heading
		existing.Heading = newData.Heading
	}

	// RPM data
	if newData.RPM1 != nil {
		updates["rpm1"] = newData.RPM1
		existing.RPM1 = newData.RPM1
	}
	if newData.RPM2 != nil {
		updates["rpm2"] = newData.RPM2
		existing.RPM2 = newData.RPM2
	}
	if newData.RPM3 != nil {
		updates["rpm3"] = newData.RPM3
		existing.RPM3 = newData.RPM3
	}
	if newData.RPM4 != nil {
		updates["rpm4"] = newData.RPM4
		existing.RPM4 = newData.RPM4
	}

	// Flow data — litres per minute (sum of the 5-second samples in this minute)
	if newData.Flow1 != nil {
		updates["flow1"] = newData.Flow1
		existing.Flow1 = newData.Flow1
	}
	if newData.Flow2 != nil {
		updates["flow2"] = newData.Flow2
		existing.Flow2 = newData.Flow2
	}
	if newData.Flow3 != nil {
		updates["flow3"] = newData.Flow3
		existing.Flow3 = newData.Flow3
	}
	if newData.Flow4 != nil {
		updates["flow4"] = newData.Flow4
		existing.Flow4 = newData.Flow4
	}

	// Top electrical
	if newData.TopACVolt != nil {
		updates["top_ac_volt"] = newData.TopACVolt
		existing.TopACVolt = newData.TopACVolt
	}
	if newData.TopACAmp != nil {
		updates["top_ac_amp"] = newData.TopACAmp
		existing.TopACAmp = newData.TopACAmp
	}
	if newData.TopACFreq != nil {
		updates["top_ac_freq"] = newData.TopACFreq
		existing.TopACFreq = newData.TopACFreq
	}
	if newData.TopACPower != nil {
		updates["top_ac_power"] = newData.TopACPower
		existing.TopACPower = newData.TopACPower
	}
	if newData.TopDCVolt != nil {
		updates["top_dc_volt"] = newData.TopDCVolt
		existing.TopDCVolt = newData.TopDCVolt
	}
	if newData.TopDCAmp != nil {
		updates["top_dc_amp"] = newData.TopDCAmp
		existing.TopDCAmp = newData.TopDCAmp
	}
	if newData.TopBatVolt != nil {
		updates["top_bat_volt"] = newData.TopBatVolt
		existing.TopBatVolt = newData.TopBatVolt
	}
	if newData.TopBatAmp != nil {
		updates["top_bat_amp"] = newData.TopBatAmp
		existing.TopBatAmp = newData.TopBatAmp
	}

	// Bot electrical
	if newData.BotACVolt != nil {
		updates["bot_ac_volt"] = newData.BotACVolt
		existing.BotACVolt = newData.BotACVolt
	}
	if newData.BotACAmp != nil {
		updates["bot_ac_amp"] = newData.BotACAmp
		existing.BotACAmp = newData.BotACAmp
	}
	if newData.BotACFreq != nil {
		updates["bot_ac_freq"] = newData.BotACFreq
		existing.BotACFreq = newData.BotACFreq
	}
	if newData.BotACPower != nil {
		updates["bot_ac_power"] = newData.BotACPower
		existing.BotACPower = newData.BotACPower
	}
	if newData.BotDCVolt != nil {
		updates["bot_dc_volt"] = newData.BotDCVolt
		existing.BotDCVolt = newData.BotDCVolt
	}
	if newData.BotDCAmp != nil {
		updates["bot_dc_amp"] = newData.BotDCAmp
		existing.BotDCAmp = newData.BotDCAmp
	}
	if newData.BotBatVolt != nil {
		updates["bot_bat_volt"] = newData.BotBatVolt
		existing.BotBatVolt = newData.BotBatVolt
	}
	if newData.BotBatAmp != nil {
		updates["bot_bat_amp"] = newData.BotBatAmp
		existing.BotBatAmp = newData.BotBatAmp
	}

	// Door status
	if newData.TopDoor1 != nil {
		updates["top_door1"] = newData.TopDoor1
		existing.TopDoor1 = newData.TopDoor1
	}
	if newData.TopDoor2 != nil {
		updates["top_door2"] = newData.TopDoor2
		existing.TopDoor2 = newData.TopDoor2
	}
	if newData.BotDoor1 != nil {
		updates["bot_door1"] = newData.BotDoor1
		existing.BotDoor1 = newData.BotDoor1
	}
	if newData.BotDoor2 != nil {
		updates["bot_door2"] = newData.BotDoor2
		existing.BotDoor2 = newData.BotDoor2
	}

	// Environment
	if newData.TopHumidity != nil {
		updates["top_humidity"] = newData.TopHumidity
		existing.TopHumidity = newData.TopHumidity
	}
	if newData.BotHumidity != nil {
		updates["bot_humidity"] = newData.BotHumidity
		existing.BotHumidity = newData.BotHumidity
	}
	if newData.TopTemp != nil {
		updates["top_temp"] = newData.TopTemp
		existing.TopTemp = newData.TopTemp
	}
	if newData.BotTemp != nil {
		updates["bot_temp"] = newData.BotTemp
		existing.BotTemp = newData.BotTemp
	}

	// Accelerometer
	if newData.AccelX != nil {
		updates["accel_x"] = newData.AccelX
		existing.AccelX = newData.AccelX
	}
	if newData.AccelY != nil {
		updates["accel_y"] = newData.AccelY
		existing.AccelY = newData.AccelY
	}
	if newData.AccelZ != nil {
		updates["accel_z"] = newData.AccelZ
		existing.AccelZ = newData.AccelZ
	}

	return updates
}
