package tracking

import (
	"encoding/json"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/tracking/extractor"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/datatypes"
)

// realtimeStates menyimpan state tracking realtime per vessel (memory)
// Diakumulasi dari berbagai sensor (GPS, RPM, Flow, dll)
var (
	realtimeStates = make(map[uint]*RealtimeTrackingState)
	stateMutex     sync.RWMutex
)

// InvalidateRealtimeCache removes the realtime state for a vessel.
// Should be called when MCU vessel assignment changes so stale data
// doesn't persist in the in-memory cache.
func InvalidateRealtimeCache(vesselID uint) {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	if _, exists := realtimeStates[vesselID]; exists {
		delete(realtimeStates, vesselID)
		log.Printf("[REALTIME] Invalidated cache for vessel_id=%d", vesselID)
	}
}

// copyTrackingToState copies non-nil fields from model.Tracking to RealtimeTrackingState
// This ensures extractor logic is reused for realtime state updates
func copyTrackingToState(state *RealtimeTrackingState, t *model.Tracking) {
	if t.Lat != nil {
		state.Lat = t.Lat
	}
	if t.Lon != nil {
		state.Lon = t.Lon
	}
	if t.Speed != nil {
		state.Speed = t.Speed
	}
	if t.Heading != nil {
		state.Heading = t.Heading
	}
	if t.RPM1 != nil {
		state.RPM1 = t.RPM1
	}
	if t.RPM2 != nil {
		state.RPM2 = t.RPM2
	}
	if t.RPM3 != nil {
		state.RPM3 = t.RPM3
	}
	if t.RPM4 != nil {
		state.RPM4 = t.RPM4
	}
	if t.Flow1 != nil {
		state.Flow1 = t.Flow1
	}
	if t.Flow2 != nil {
		state.Flow2 = t.Flow2
	}
	if t.Flow3 != nil {
		state.Flow3 = t.Flow3
	}
	if t.Flow4 != nil {
		state.Flow4 = t.Flow4
	}
	if t.TopACVolt != nil {
		state.TopACVolt = t.TopACVolt
	}
	if t.TopACAmp != nil {
		state.TopACAmp = t.TopACAmp
	}
	if t.TopACFreq != nil {
		state.TopACFreq = t.TopACFreq
	}
	if t.TopACPower != nil {
		state.TopACPower = t.TopACPower
	}
	if t.TopDCVolt != nil {
		state.TopDCVolt = t.TopDCVolt
	}
	if t.TopDCAmp != nil {
		state.TopDCAmp = t.TopDCAmp
	}
	if t.TopBatVolt != nil {
		state.TopBatVolt = t.TopBatVolt
	}
	if t.TopBatAmp != nil {
		state.TopBatAmp = t.TopBatAmp
	}
	if t.BotACVolt != nil {
		state.BotACVolt = t.BotACVolt
	}
	if t.BotACAmp != nil {
		state.BotACAmp = t.BotACAmp
	}
	if t.BotACFreq != nil {
		state.BotACFreq = t.BotACFreq
	}
	if t.BotACPower != nil {
		state.BotACPower = t.BotACPower
	}
	if t.BotDCVolt != nil {
		state.BotDCVolt = t.BotDCVolt
	}
	if t.BotDCAmp != nil {
		state.BotDCAmp = t.BotDCAmp
	}
	if t.BotBatVolt != nil {
		state.BotBatVolt = t.BotBatVolt
	}
	if t.BotBatAmp != nil {
		state.BotBatAmp = t.BotBatAmp
	}
	if t.TopDoor1 != nil {
		state.TopDoor1 = t.TopDoor1
	}
	if t.TopDoor2 != nil {
		state.TopDoor2 = t.TopDoor2
	}
	if t.BotDoor1 != nil {
		state.BotDoor1 = t.BotDoor1
	}
	if t.BotDoor2 != nil {
		state.BotDoor2 = t.BotDoor2
	}
	if t.TopHumidity != nil {
		state.TopHumidity = t.TopHumidity
	}
	if t.BotHumidity != nil {
		state.BotHumidity = t.BotHumidity
	}
	if t.TopTemp != nil {
		state.TopTemp = t.TopTemp
	}
	if t.BotTemp != nil {
		state.BotTemp = t.BotTemp
	}
	if t.AccelX != nil {
		state.AccelX = t.AccelX
	}
	if t.AccelY != nil {
		state.AccelY = t.AccelY
	}
	if t.AccelZ != nil {
		state.AccelZ = t.AccelZ
	}
}

type Service interface {
	// Tracking
	CreateTracking(vesselID uint, sensorCode string, sensorID uint, location string, sensorNumber int, kFactor float64, value map[string]interface{}) error
	CreateTrackingByTopicParts(companyCode, vesselCode, location, sensorGroup string, sensorNumber int, sensorID uint, kFactor float64, value map[string]interface{}) error

	// API Tracking dengan TimescaleDB (per-minute, excludes current minute)
	GetTracking(vesselID uint, startTime, endTime time.Time, sort string, limit int, hasCoords bool) ([]VesselTrackingDTO, error)
	GetAllTracking(startTime, endTime time.Time, sort string, limit int, hasCoords bool) ([]VesselTrackingDTO, error)

	// Cleaning
	CleanupHistoricalSpikes()
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// getOrCreateRealtimeState returns existing state or creates new one for vessel
func (s *service) getOrCreateRealtimeState(vesselID uint) *RealtimeTrackingState {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	state, exists := realtimeStates[vesselID]
	if !exists {
		// Get vessel name
		vessel, err := s.repo.GetVesselByID(vesselID)
		vesselName := ""
		if err == nil && vessel != nil {
			vesselName = vessel.Name
		}

		state = &RealtimeTrackingState{
			VesselID:   vesselID,
			VesselName: vesselName,
		}
		realtimeStates[vesselID] = state
	}
	return state
}

// updateRealtimeStateFromSensorData updates realtime state based on sensor type
// Uses extractor package to ensure consistent extraction logic with scheduler
// location: bd (bridge deck) = top, er (engine room) = bot
// sensorNumber: sensor index from topic (e.g., rpm/1 → sensorNumber=1)
// kFactor: calibration factor for flow sensors (pulse/L). 0 = skip pulse calculation.
func (s *service) updateRealtimeStateFromSensorData(state *RealtimeTrackingState, sensorCode string, location string, sensorNumber int, kFactor float64, value map[string]interface{}) {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	now := time.Now()
	state.LastUpdate = now

	// Normalize sensor code to lowercase for matching
	sensorCodeLower := strings.ToLower(sensorCode)

	// Build position flags from location
	// bd/brd = bridge deck (top), er/eng = engine room (bot)
	locationLower := strings.ToLower(location)
	isTop := locationLower == "bd" || locationLower == "brd"
	isBot := locationLower == "er" || locationLower == "eng"

	// Create temp tracking and processed map for extractors
	tempTracking := &model.Tracking{}
	processed := make(map[string]bool) // For realtime, we don't skip any fields

	// Call appropriate extractor based on sensor type
	switch sensorCodeLower {
	case "gps":
		// GPS needs special handling for speed/heading calculation
		extractor.GPS(tempTracking, value, processed)

		// Handle speed/heading calculation from previous position
		if tempTracking.Lat != nil && tempTracking.Lon != nil {
			state.LastGPS = now
			newLat, newLon := *tempTracking.Lat, *tempTracking.Lon

			if state.Lat != nil && state.Lon != nil && !state.LastUpdate.IsZero() {
				timeDiffHours := now.Sub(state.LastUpdate).Hours()

				// GPS anomaly filter — reject if position jumped unrealistically far
				if isGPSSpike(*state.Lat, *state.Lon, newLat, newLon, timeDiffHours) {
					// Breakthrough: if same coordinates rejected N times consecutively,
					// vessel likely moved there and previous state is stale.
					if state.SpikeCandidateLat != nil && state.SpikeCandidateLon != nil &&
						*state.SpikeCandidateLat == newLat && *state.SpikeCandidateLon == newLon {
						state.SpikeCandidateCount++
						if state.SpikeCandidateCount >= 3 {
							log.Printf("[GPS-FILTER] Spike breakthrough: vessel=%d (%.4f,%.4f) → (%.4f,%.4f) accepted after %d consecutive",
								state.VesselID, *state.Lat, *state.Lon, newLat, newLon, state.SpikeCandidateCount)
							// Reset candidate and let it through
							state.SpikeCandidateLat = nil
							state.SpikeCandidateLon = nil
							state.SpikeCandidateCount = 0
							goto acceptPosition
						}
					} else {
						state.SpikeCandidateLat = &newLat
						state.SpikeCandidateLon = &newLon
						state.SpikeCandidateCount = 1
					}
					log.Printf("[GPS-FILTER] Spike rejected: vessel=%d (%.4f,%.4f) → (%.4f,%.4f) in %.1f min (candidate %d/3)",
						state.VesselID, *state.Lat, *state.Lon, newLat, newLon, timeDiffHours*60, state.SpikeCandidateCount)
					tempTracking.Lat = nil
					tempTracking.Lon = nil
					break // skip speed/heading/prev update
				}
				// Reset candidate on accepted position
				state.SpikeCandidateLat = nil
				state.SpikeCandidateLon = nil
				state.SpikeCandidateCount = 0
			acceptPosition:

				if timeDiffHours > 0 && timeDiffHours < 1 {
					speed := CalculateSpeed(*state.PrevLat, *state.PrevLon, newLat, newLon, timeDiffHours)
					heading := RoundToTwoDecimals(CalculateBearing(*state.PrevLat, *state.PrevLon, newLat, newLon))
					tempTracking.Speed = &speed
					tempTracking.Heading = &heading
				}
			}

			// Store current position as previous for next calculation
			state.PrevLat = state.Lat
			state.PrevLon = state.Lon
			state.PrevTime = state.LastUpdate
		}

		copyTrackingToState(state, tempTracking)

	case "rpm":
		extractor.RPM(tempTracking, value, processed, sensorNumber)
		switch sensorNumber {
		case 1:
			if tempTracking.RPM1 != nil {
				state.LastRPM1 = now
			}
		case 2:
			if tempTracking.RPM2 != nil {
				state.LastRPM2 = now
			}
		case 3:
			if tempTracking.RPM3 != nil {
				state.LastRPM3 = now
			}
		case 4:
			if tempTracking.RPM4 != nil {
				state.LastRPM4 = now
			}
		}
		copyTrackingToState(state, tempTracking)

	case "flow":
		extractor.Flow(tempTracking, value, processed, sensorNumber, kFactor)
		switch sensorNumber {
		case 1:
			if tempTracking.Flow1 != nil {
				state.LastFlow1 = now
			}
		case 2:
			if tempTracking.Flow2 != nil {
				state.LastFlow2 = now
			}
		case 3:
			if tempTracking.Flow3 != nil {
				state.LastFlow3 = now
			}
		case 4:
			if tempTracking.Flow4 != nil {
				state.LastFlow4 = now
			}
		}
		copyTrackingToState(state, tempTracking)

	case "ac":
		extractor.AC(tempTracking, value, processed, isTop, isBot)
		if isTop && tempTracking.TopACVolt != nil {
			state.LastTopAC = now
		}
		if isBot && tempTracking.BotACVolt != nil {
			state.LastBotAC = now
		}
		copyTrackingToState(state, tempTracking)

	case "dc":
		extractor.DC(tempTracking, value, processed, isTop, isBot)
		if isTop && tempTracking.TopDCVolt != nil {
			state.LastTopDC = now
		}
		if isBot && tempTracking.BotDCVolt != nil {
			state.LastBotDC = now
		}
		copyTrackingToState(state, tempTracking)

	case "bat":
		extractor.Battery(tempTracking, value, processed, isTop, isBot)
		if isTop && tempTracking.TopBatVolt != nil {
			state.LastTopBat = now
		}
		if isBot && tempTracking.BotBatVolt != nil {
			state.LastBotBat = now
		}
		copyTrackingToState(state, tempTracking)

	case "door":
		extractor.Door(tempTracking, value, processed, isTop, isBot, sensorNumber)
		if isTop && (tempTracking.TopDoor1 != nil || tempTracking.TopDoor2 != nil) {
			state.LastTopDoor = now
		}
		if isBot && (tempTracking.BotDoor1 != nil || tempTracking.BotDoor2 != nil) {
			state.LastBotDoor = now
		}
		copyTrackingToState(state, tempTracking)

	case "temp":
		extractor.Temperature(tempTracking, value, processed, isTop, isBot)
		if isTop && tempTracking.TopTemp != nil {
			state.LastTopTemp = now
		}
		if isBot && tempTracking.BotTemp != nil {
			state.LastBotTemp = now
		}
		copyTrackingToState(state, tempTracking)

	case "humidity":
		extractor.Humidity(tempTracking, value, processed, isTop, isBot)
		if isTop && tempTracking.TopHumidity != nil {
			state.LastTopHumidity = now
		}
		if isBot && tempTracking.BotHumidity != nil {
			state.LastBotHumidity = now
		}
		copyTrackingToState(state, tempTracking)

	case "accel":
		extractor.Accelerometer(tempTracking, value, processed)
		if tempTracking.AccelX != nil || tempTracking.AccelY != nil || tempTracking.AccelZ != nil {
			state.LastAccel = now
		}
		copyTrackingToState(state, tempTracking)

	default:
		log.Printf("[REALTIME] Unknown sensor type: %s", sensorCodeLower)
	}
}

// convertRealtimeStateToDTO converts realtime state to DTO for WebSocket broadcast
// Applies zero-tolerance/deadband expiry (60s) per sensor group so stale sensors
// cleanly drop to nil without bloating WebSocket payload with extra timestamp fields.
func (s *service) convertRealtimeStateToDTO(state *RealtimeTrackingState) *VesselTrackingDTO {
	stateMutex.RLock()
	defer stateMutex.RUnlock()

	now := time.Now()
	const sensorExpiry = 60 * time.Second

	dto := &VesselTrackingDTO{
		VesselID:     state.VesselID,
		VesselName:   state.VesselName,
		WaveHeight:   state.WaveHeight,
		WaveCat:      state.WaveCat,
		CurrentSpeed: state.CurrentSpeed,
		CurrentTo:    state.CurrentTo,
		WindSpeed:    state.WindSpeed,
		WindGust:     state.WindGust,
		WindFrom:     state.WindFrom,
		Weather:      state.Weather,
		Timestamp:    state.LastUpdate,
	}

	// GPS: speed/heading expire after 60s, coordinates retained as last known position for map display
	if !state.LastGPS.IsZero() && now.Sub(state.LastGPS) <= sensorExpiry {
		dto.Lat = state.Lat
		dto.Lon = state.Lon
		dto.Speed = state.Speed
		dto.Heading = state.Heading
	} else if state.Lat != nil && state.Lon != nil {
		dto.Lat = state.Lat
		dto.Lon = state.Lon
	}

	// RPM
	if !state.LastRPM1.IsZero() && now.Sub(state.LastRPM1) <= sensorExpiry {
		dto.RPM1 = state.RPM1
	}
	if !state.LastRPM2.IsZero() && now.Sub(state.LastRPM2) <= sensorExpiry {
		dto.RPM2 = state.RPM2
	}
	if !state.LastRPM3.IsZero() && now.Sub(state.LastRPM3) <= sensorExpiry {
		dto.RPM3 = state.RPM3
	}
	if !state.LastRPM4.IsZero() && now.Sub(state.LastRPM4) <= sensorExpiry {
		dto.RPM4 = state.RPM4
	}

	// Flow
	if !state.LastFlow1.IsZero() && now.Sub(state.LastFlow1) <= sensorExpiry {
		dto.Flow1 = state.Flow1
	}
	if !state.LastFlow2.IsZero() && now.Sub(state.LastFlow2) <= sensorExpiry {
		dto.Flow2 = state.Flow2
	}
	if !state.LastFlow3.IsZero() && now.Sub(state.LastFlow3) <= sensorExpiry {
		dto.Flow3 = state.Flow3
	}
	if !state.LastFlow4.IsZero() && now.Sub(state.LastFlow4) <= sensorExpiry {
		dto.Flow4 = state.Flow4
	}

	// Top MCU Electrical
	if !state.LastTopAC.IsZero() && now.Sub(state.LastTopAC) <= sensorExpiry {
		dto.TopACVolt = state.TopACVolt
		dto.TopACAmp = state.TopACAmp
		dto.TopACFreq = state.TopACFreq
		dto.TopACPower = state.TopACPower
	}
	if !state.LastTopDC.IsZero() && now.Sub(state.LastTopDC) <= sensorExpiry {
		dto.TopDCVolt = state.TopDCVolt
		dto.TopDCAmp = state.TopDCAmp
	}
	if !state.LastTopBat.IsZero() && now.Sub(state.LastTopBat) <= sensorExpiry {
		dto.TopBatVolt = state.TopBatVolt
		dto.TopBatAmp = state.TopBatAmp
	}

	// Bottom MCU Electrical
	if !state.LastBotAC.IsZero() && now.Sub(state.LastBotAC) <= sensorExpiry {
		dto.BotACVolt = state.BotACVolt
		dto.BotACAmp = state.BotACAmp
		dto.BotACFreq = state.BotACFreq
		dto.BotACPower = state.BotACPower
	}
	if !state.LastBotDC.IsZero() && now.Sub(state.LastBotDC) <= sensorExpiry {
		dto.BotDCVolt = state.BotDCVolt
		dto.BotDCAmp = state.BotDCAmp
	}
	if !state.LastBotBat.IsZero() && now.Sub(state.LastBotBat) <= sensorExpiry {
		dto.BotBatVolt = state.BotBatVolt
		dto.BotBatAmp = state.BotBatAmp
	}

	// Door Status
	if !state.LastTopDoor.IsZero() && now.Sub(state.LastTopDoor) <= sensorExpiry {
		dto.TopDoor1 = state.TopDoor1
		dto.TopDoor2 = state.TopDoor2
	}
	if !state.LastBotDoor.IsZero() && now.Sub(state.LastBotDoor) <= sensorExpiry {
		dto.BotDoor1 = state.BotDoor1
		dto.BotDoor2 = state.BotDoor2
	}

	// Environment (Temp & Humidity)
	if !state.LastTopTemp.IsZero() && now.Sub(state.LastTopTemp) <= sensorExpiry {
		dto.TopTemp = state.TopTemp
	}
	if !state.LastBotTemp.IsZero() && now.Sub(state.LastBotTemp) <= sensorExpiry {
		dto.BotTemp = state.BotTemp
	}
	if !state.LastTopHumidity.IsZero() && now.Sub(state.LastTopHumidity) <= sensorExpiry {
		dto.TopHumidity = state.TopHumidity
	}
	if !state.LastBotHumidity.IsZero() && now.Sub(state.LastBotHumidity) <= sensorExpiry {
		dto.BotHumidity = state.BotHumidity
	}

	// Accelerometer
	if !state.LastAccel.IsZero() && now.Sub(state.LastAccel) <= sensorExpiry {
		dto.AccelX = state.AccelX
		dto.AccelY = state.AccelY
		dto.AccelZ = state.AccelZ
	}

	return dto
}

// parsePayloadTimestamp extracts the `timestamp` field from a sensor payload.
// GPS hardware sends: {"lat": ..., "lon": ..., "timestamp": "2026-05-05 03:25:23"}
// The timestamp has no timezone in the payload; we assume WIB (Asia/Jakarta)
// which is already set as time.Local in main.go.
// Returns the parsed time and true on success, or zero time + false if absent/unparsable.
func parsePayloadTimestamp(value map[string]interface{}) (time.Time, bool) {
	tsStr, ok := value["timestamp"].(string)
	if !ok || tsStr == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", tsStr, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// CreateTracking creates tracking data grouped by vessel
// Also broadcasts realtime data via WebSocket with accumulated sensor state
// location: bd (bridge deck), er (engine room) - derived from topic
// sensorNumber: sensor index (1, 2, 3, 4) for multi-sensor types like RPM
// kFactor: calibration factor for flow sensors (pulse/L). 0 = skip pulse calculation.
func (s *service) CreateTracking(vesselID uint, sensorCode string, sensorID uint, location string, sensorNumber int, kFactor float64, value map[string]interface{}) error {
	ts, ok := parsePayloadTimestamp(value)
	if !ok {
		ts = time.Now()
	}

	// Save sensor data and update tracking in one transaction
	jsonValue, _ := json.Marshal(value)
	sensorDataRecord := &model.SensorData{
		Time:     ts,
		SensorID: sensorID,
		Value:    datatypes.JSON(jsonValue),
	}

	// This will save sensor_data and update tracking with calculation
	_, err := s.repo.CreateSensorDataWithTracking(sensorDataRecord, vesselID)
	if err != nil {
		return err
	}

	// Update realtime state and broadcast via WebSocket
	if Hub != nil {
		// Get or create realtime state for this vessel
		state := s.getOrCreateRealtimeState(vesselID)

		// Update state with new sensor data (accumulate)
		s.updateRealtimeStateFromSensorData(state, sensorCode, location, sensorNumber, kFactor, value)

		// Convert to DTO and broadcast
		dto := s.convertRealtimeStateToDTO(state)
		if dto != nil {
			message := map[string]interface{}{
				"type": "realtime",
				"data": dto,
			}
			jsonData, _ := json.Marshal(message)

			broadcastFilter := &TrackingBroadcast{VesselID: vesselID}
			Hub.Broadcast(broadcastFilter, jsonData)
		}
	}

	return nil
}

// CreateTrackingByTopicParts creates tracking using parsed topic parts
// Topic format: <company>/<vessel>/<location>/<sensor_group>/<number>
// Example: hcml/vs001/bd/gps/1, hcml/vs001/er/rpm/1
//
// IMPORTANT: Vessel is resolved via Sensor → MCU → Vessel chain (MCU.VesselID),
// NOT from the vessel code in the MQTT topic. This ensures that when an MCU is
// reassigned to a different vessel, tracking data correctly routes to the new vessel.
// The topic's vessel code is only used as a fallback if the sensor has no MCU/vessel assignment.
func (s *service) CreateTrackingByTopicParts(companyCode, vesselCode, location, sensorGroup string, sensorNumber int, sensorID uint, kFactor float64, value map[string]interface{}) error {
	// Primary resolution: Sensor → MCU → Vessel (follows MCU.VesselID)
	// This correctly handles MCU reassignment to different vessels.
	// Uses raw SQL JOIN for reliable resolution through nullable foreign keys.
	vesselID, err := s.repo.GetVesselIDBySensorID(sensorID)
	if err == nil && vesselID > 0 {
		return s.CreateTracking(vesselID, sensorGroup, sensorID, location, sensorNumber, kFactor, value)
	}

	// Fallback: Find vessel by code from topic (for sensors not yet assigned to MCU)
	// WARNING: This uses the vessel code embedded in the MQTT topic, which is immutable
	// (hardcoded in firmware). If the MCU was reassigned to a different vessel,
	// this fallback will route data to the WRONG vessel.
	log.Printf("[TRACKING] WARN: MCU chain failed sensor_id=%d, fallback to topic vessel_code=%s", sensorID, vesselCode)
	vessel, err := s.repo.GetVesselByCode(vesselCode)
	if err != nil || vessel == nil {
		log.Printf("[TRACKING] Vessel not found: code=%s, company=%s", vesselCode, companyCode)

		// Only save to sensor_data if vessel is unknown
		ts, ok := parsePayloadTimestamp(value)
		if !ok {
			ts = time.Now()
		}
		jsonValue, _ := json.Marshal(value)
		sensorDataRecord := &model.SensorData{
			Time:     ts,
			SensorID: sensorID,
			Value:    datatypes.JSON(jsonValue),
		}
		return s.repo.CreateSensorData(sensorDataRecord)
	}

	// Optionally validate company code matches vessel's company
	if vessel.CompanyID != nil {
		company, _ := s.repo.GetCompanyByCode(companyCode)
		if company != nil && company.ID != *vessel.CompanyID {
			log.Printf("[TRACKING] Warning: Company mismatch - topic company=%s (ID: %d) != vessel company ID %d", companyCode, company.ID, *vessel.CompanyID)
		}
	}

	log.Printf("[TRACKING] FALLBACK: sensor_id=%d → vessel_code=%s → vessel_id=%d", sensorID, vesselCode, vessel.ID)
	return s.CreateTracking(vessel.ID, sensorGroup, sensorID, location, sensorNumber, kFactor, value)
}

// convertTrackingToDTO converts a Tracking model to VesselTrackingDTO
// Used for API response (historical data from database)
func (s *service) convertTrackingToDTO(tracking *model.Tracking, vesselID uint) *VesselTrackingDTO {
	// Get vessel info
	vessel, err := s.repo.GetVesselByID(vesselID)
	if err != nil || vessel == nil {
		log.Printf("[TRACKING] Failed to get vessel %d: %v", vesselID, err)
		return nil
	}

	// Build DTO with extracted fields
	dto := &VesselTrackingDTO{
		ID:         tracking.ID,
		VesselID:   vesselID,
		VesselName: vessel.Name,
		Timestamp:  tracking.Time,
	}

	// Add lat/lon/speed/heading if available
	if tracking.Lat != nil {
		dto.Lat = tracking.Lat
	}
	if tracking.Lon != nil {
		dto.Lon = tracking.Lon
	}
	if tracking.Speed != nil {
		rounded := RoundToTwoDecimals(*tracking.Speed)
		dto.Speed = &rounded
	}
	if tracking.Heading != nil {
		rounded := RoundToTwoDecimals(*tracking.Heading)
		dto.Heading = &rounded
	}

	// Add individual RPM values
	dto.RPM1 = tracking.RPM1
	dto.RPM2 = tracking.RPM2
	dto.RPM3 = tracking.RPM3
	dto.RPM4 = tracking.RPM4

	// Add individual Flow values
	dto.Flow1 = tracking.Flow1
	dto.Flow2 = tracking.Flow2
	dto.Flow3 = tracking.Flow3
	dto.Flow4 = tracking.Flow4

	// Add Top MCU Electrical values
	dto.TopACVolt = tracking.TopACVolt
	dto.TopACAmp = tracking.TopACAmp
	dto.TopACFreq = tracking.TopACFreq
	dto.TopACPower = tracking.TopACPower
	dto.TopDCVolt = tracking.TopDCVolt
	dto.TopDCAmp = tracking.TopDCAmp
	dto.TopBatVolt = tracking.TopBatVolt
	dto.TopBatAmp = tracking.TopBatAmp

	// Add Bottom MCU Electrical values
	dto.BotACVolt = tracking.BotACVolt
	dto.BotACAmp = tracking.BotACAmp
	dto.BotACFreq = tracking.BotACFreq
	dto.BotACPower = tracking.BotACPower
	dto.BotDCVolt = tracking.BotDCVolt
	dto.BotDCAmp = tracking.BotDCAmp
	dto.BotBatVolt = tracking.BotBatVolt
	dto.BotBatAmp = tracking.BotBatAmp

	// Add Door Status
	dto.TopDoor1 = tracking.TopDoor1
	dto.TopDoor2 = tracking.TopDoor2
	dto.BotDoor1 = tracking.BotDoor1
	dto.BotDoor2 = tracking.BotDoor2

	// Add Environment values
	dto.TopHumidity = tracking.TopHumidity
	dto.BotHumidity = tracking.BotHumidity
	dto.TopTemp = tracking.TopTemp
	dto.BotTemp = tracking.BotTemp

	// Add Accelerometer values
	dto.AccelX = tracking.AccelX
	dto.AccelY = tracking.AccelY
	dto.AccelZ = tracking.AccelZ

	// Add BMKG Maritime Weather values
	dto.WaveHeight = tracking.WaveHeight
	dto.WaveCat = tracking.WaveCat
	dto.CurrentSpeed = tracking.CurrentSpeed
	dto.CurrentTo = tracking.CurrentTo
	dto.WindSpeed = tracking.WindSpeed
	dto.WindGust = tracking.WindGust
	dto.WindFrom = tracking.WindFrom
	dto.Weather = tracking.Weather

	return dto
}

// ========================================
// NEW API: GetTracking dengan TimescaleDB
// ========================================

// GetTracking gets tracking data for a specific vessel
// Returns only completed minutes (excludes current minute)
func (s *service) GetTracking(vesselID uint, startTime, endTime time.Time, sort string, limit int, hasCoords bool) ([]VesselTrackingDTO, error) {
	trackings, err := s.repo.GetTrackingWithTimescale(vesselID, startTime, endTime, sort, limit, hasCoords)
	if err != nil {
		return nil, err
	}

	result := make([]VesselTrackingDTO, 0)
	for _, tracking := range trackings {
		dto := s.convertTrackingToDTO(&tracking, tracking.VesselID)
		if dto != nil {
			result = append(result, *dto)
		}
	}

	return result, nil
}

// GetAllTracking gets tracking data for all vessels
// Returns only completed minutes (excludes current minute)
func (s *service) GetAllTracking(startTime, endTime time.Time, sort string, limit int, hasCoords bool) ([]VesselTrackingDTO, error) {
	trackings, err := s.repo.GetAllTrackingWithTimescale(startTime, endTime, sort, limit, hasCoords)
	if err != nil {
		return nil, err
	}

	result := make([]VesselTrackingDTO, 0)
	for _, tracking := range trackings {
		dto := s.convertTrackingToDTO(&tracking, tracking.VesselID)
		if dto != nil {
			result = append(result, *dto)
		}
	}

	return result, nil
}

// Note: Geo calculation utilities (HaversineDistance, CalculateBearing, CalculateSpeed, RoundToTwoDecimals)
// are defined in repository.go and exported for reuse across the tracking package

// isGPSSpike checks if a GPS coordinate pair is unreasonably far from the
// previous known position given the time elapsed. Returns true when the jump
// exceeds a realistic maritime speed (~30 knots / 55 km/h threshold).
//
// The threshold scales with time so that:
//   - A 3000 km jump in 1 minute → spike (GPS glitch)
//   - A 500 km jump after 7 days offline → accepted (vessel sailed normally)
//   - A 50 km jump in 5 minutes → spike (max realistic ~5 km in 5 min)
func isGPSSpike(prevLat, prevLon, newLat, newLon float64, timeDiffHours float64) bool {
	if timeDiffHours <= 0 {
		timeDiffHours = 2.0 / 3600 // floor at 2 seconds to avoid div-by-zero
	}

	// Maximum realistic average speed: 30 knots ≈ 55 km/h ≈ 30 NM/h
	maxReasonableNM := 30 * timeDiffHours

	// Floor: 10 NM (~18 km) — any jump this large in under 20 min is a spike
	if maxReasonableNM < 10 {
		maxReasonableNM = 10
	}

	// Ceiling: 2000 NM — prevents unbounded threshold for multi-week gaps
	if maxReasonableNM > 2000 {
		maxReasonableNM = 2000
	}

	distanceNM := HaversineDistance(prevLat, prevLon, newLat, newLon)
	return distanceNM > maxReasonableNM
}

// scanVessel is the core per-vessel spike scanning logic, extracted for reuse.
// Returns the number of spikes found (and nullified) for this vessel.
func (s *service) scanVessel(vesselID uint) int {
	total, err := s.repo.CountTrackingWithCoords(vesselID)
	if err != nil || total == 0 {
		return 0
	}

	log.Printf("[GPS-CLEANUP] Vessel %d: scanning %d records...", vesselID, total)

	const chunkSize = 1000
	var candID uint
	var candLat, candLon *float64
	var candTime *time.Time
	var prevLat, prevLon *float64
	var prevTime *time.Time
	var spikeIDs []uint
	vesselSpikes := 0
	var cursor time.Time

	for {
		chunk, err := s.repo.GetTrackingChunkAfter(vesselID, cursor, chunkSize)
		if err != nil {
			log.Printf("[GPS-CLEANUP] Error fetching chunk for vessel %d: %v", vesselID, err)
			break
		}
		if len(chunk) == 0 {
			break
		}

		for _, rec := range chunk {
			// Flush accumulated spike IDs early to avoid TOAST decompression limit
			if len(spikeIDs) >= 10 {
				n, failed, err := s.repo.BatchNullifyTrackingCoords(spikeIDs)
				if err != nil {
					log.Printf("[GPS-CLEANUP] Error nullifying spikes for vessel %d: %v", vesselID, err)
				}
				vesselSpikes += n
				spikeIDs = failed
			}

			if rec.Lat == nil || rec.Lon == nil {
				continue
			}

			// Check 1: Invalid coordinate range
			if *rec.Lon > 180 || *rec.Lon < -180 || *rec.Lat > 90 || *rec.Lat < -90 {
				log.Printf("[GPS-CLEANUP] Vessel %d: invalid coords at %s (%.4f, %.4f)", vesselID, rec.Time.Format("15:04:05"), *rec.Lat, *rec.Lon)
				spikeIDs = append(spikeIDs, rec.ID)
				continue
			}

			// Check 2: Stored speed is unrealistic (data from a previous wrong calculation)
			if rec.Speed != nil && *rec.Speed > 50 {
				log.Printf("[GPS-CLEANUP] Vessel %d: unrealistic speed %.1f knots at %s", vesselID, *rec.Speed, rec.Time.Format("15:04:05"))
				spikeIDs = append(spikeIDs, rec.ID)
				continue
			}

			// ── Phase 1: Anchor establishment (find 2 close records before detecting spikes) ──
			if prevLat == nil {
				if candLat == nil {
					candID, candLat, candLon, candTime = rec.ID, rec.Lat, rec.Lon, &rec.Time
					continue
				}
				// Compare candidate with this record — if they're close, anchor is valid
				dist := HaversineDistance(*candLat, *candLon, *rec.Lat, *rec.Lon)
				if dist < 10 {
					// Cluster found — both are valid, use candidate as first prev
					prevLat, prevLon, prevTime = candLat, candLon, candTime
				} else {
					// Candidate was a spike — nullify it and shift to this record
					log.Printf("[GPS-CLEANUP] Vessel %d: anchor spike at %s (%.1f NM from next)", vesselID, candTime.Format("15:04:05"), dist)
					spikeIDs = append(spikeIDs, candID)
					candID, candLat, candLon, candTime = rec.ID, rec.Lat, rec.Lon, &rec.Time
					continue
				}
			}

			// ── Phase 2: Spike detection (anchor established) ──
			timeDiffHours := math.Abs(rec.Time.Sub(*prevTime).Hours())
			distNM := HaversineDistance(*prevLat, *prevLon, *rec.Lat, *rec.Lon)

			// Check 3: GPS jitter — small jump with unrealistic speed
			if distNM < 1 && timeDiffHours > 0 {
				equivSpeed := distNM / timeDiffHours
				if equivSpeed > 10 {
					log.Printf("[GPS-CLEANUP] Vessel %d: GPS jitter at %s (%.1f NM, %.1f knots)", vesselID, rec.Time.Format("15:04:05"), distNM, equivSpeed)
					spikeIDs = append(spikeIDs, rec.ID)
					continue
				}
			}

			// Check 4: Big spike — long-distance jump in short time
			if isGPSSpike(*prevLat, *prevLon, *rec.Lat, *rec.Lon, timeDiffHours) {
				spikeIDs = append(spikeIDs, rec.ID)
				continue
			}

			prevLat = rec.Lat
			prevLon = rec.Lon
			prevTime = &rec.Time
		}

		cursor = chunk[len(chunk)-1].Time
	}

	if len(spikeIDs) > 0 {
		n, failed, err := s.repo.BatchNullifyTrackingCoords(spikeIDs)
		if err != nil {
			log.Printf("[GPS-CLEANUP] Error nullifying spikes for vessel %d: %v", vesselID, err)
		}
		vesselSpikes += n
		if len(failed) > 0 {
			log.Printf("[GPS-CLEANUP] Vessel %d: %d IDs permanently failed TOAST decompression", vesselID, len(failed))
		}
	}

	log.Printf("[GPS-CLEANUP] Vessel %d: done (%d spikes cleaned)", vesselID, vesselSpikes)

	if vesselSpikes > 0 {
		s.repo.VacuumTrackings()
	}

	return vesselSpikes
}

// CleanupHistoricalSpikes scans existing tracking records for GPS spikes
// and nullifies coordinates on any record that jumps unrealistically far
// from its predecessor. Runs in a background goroutine at startup and
// repeats automatically every 6 hours.
func (s *service) CleanupHistoricalSpikes() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[GPS-CLEANUP] FATAL: cleanup goroutine panicked: %v", r)
			}
		}()
		s.runSpikeCleanup()
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			s.runSpikeCleanup()
		}
	}()
}

func (s *service) runSpikeCleanup() {
	log.Println("[GPS-CLEANUP] Scanning historical tracking data for spike anomalies...")

	if err := s.repo.DecompressTrackingsChunks(); err != nil {
		log.Printf("[GPS-CLEANUP] Error decompressing chunks: %v", err)
	}

	vesselIDs, err := s.repo.GetVesselIDsWithCoordsOrderedByLatest()
	if err != nil {
		log.Printf("[GPS-CLEANUP] Failed to get vessel IDs: %v", err)
		return
	}

	if len(vesselIDs) == 0 {
		log.Println("[GPS-CLEANUP] No vessels with GPS data found, skipping.")
		return
	}

	const maxConcurrent = 5
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	grandTotalSpikes := 0

	for _, vesselID := range vesselIDs {
		wg.Add(1)
		sem <- struct{}{} // acquire — blocks if maxConcurrent already running

		go func(vid uint) {
			defer wg.Done()
			defer func() { <-sem }() // release
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[GPS-CLEANUP] Panic scanning vessel %d: %v", vid, r)
				}
			}()

			spikes := s.scanVessel(vid)

			mu.Lock()
			grandTotalSpikes += spikes
			mu.Unlock()
		}(vesselID)
	}

	wg.Wait()
	log.Printf("[GPS-CLEANUP] Complete: %d total spike records cleaned across %d vessels", grandTotalSpikes, len(vesselIDs))
}
