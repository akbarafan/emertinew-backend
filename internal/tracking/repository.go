package tracking

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	// Tracking
	CreateTracking(tracking *model.Tracking) error

	// TimescaleDB optimized queries untuk API
	GetTrackingWithTimescale(vesselID uint, startTime, endTime time.Time, sort string, limit int, hasCoords bool) ([]model.Tracking, error)
	GetAllTrackingWithTimescale(startTime, endTime time.Time, sort string, limit int, hasCoords bool) ([]model.Tracking, error)

	// Historical spike cleanup
	GetDistinctVesselIDsWithCoords() ([]uint, error)
	GetVesselIDsWithCoordsOrderedByLatest() ([]uint, error)
	GetTrackingChunk(vesselID uint, offset, limit int) ([]model.Tracking, error)
	GetTrackingChunkAfter(vesselID uint, afterTime time.Time, limit int) ([]model.Tracking, error)
	CountTrackingWithCoords(vesselID uint) (int64, error)
	NullifyTrackingCoords(ids []uint) error
	BatchNullifyTrackingCoords(ids []uint) (int, []uint, error)
	VacuumTrackings() error
	DecompressTrackingsChunks() error

	// Sensor Data
	CreateSensorData(data *model.SensorData) error
	CreateSensorDataWithTracking(data *model.SensorData, vesselID uint) (*model.Tracking, error)

	// Helper methods
	GetAllVessels() ([]model.MVessel, error)
	GetSensorByID(sensorID uint) (*model.MSensor, error)
	GetVesselByID(vesselID uint) (*model.MVessel, error)
	GetVesselIDBySensorID(sensorID uint) (uint, error)
	GetMCUByID(mcuID uint) (*model.MMCU, error)
	GetVesselByName(name string) (*model.MVessel, error)
	GetVesselByCode(code string) (*model.MVessel, error)
	GetCompanyByCode(code string) (*model.MCompany, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Tracking implementations
func (r *repository) CreateTracking(tracking *model.Tracking) error {
	return r.db.Create(tracking).Error
}

// CreateSensorData saves sensor data without tracking
func (r *repository) CreateSensorData(data *model.SensorData) error {
	return r.db.Create(data).Error
}

// CreateSensorDataWithTracking saves sensor data only.
// Tracking creation is handled by the scheduler every minute.
// This keeps the data ingestion fast and simple.
func (r *repository) CreateSensorDataWithTracking(data *model.SensorData, vesselID uint) (*model.Tracking, error) {
	// Just save sensor_data, tracking will be created by scheduler
	if err := r.db.Create(data).Error; err != nil {
		return nil, err
	}
	return nil, nil
}

// ========================================
// Geo Calculation Utilities (Exported)
// ========================================

// HaversineDistance calculates distance between two lat/lon points in nautical miles
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusNM = 3440.065 // Earth radius in nautical miles

	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusNM * c
}

// CalculateBearing calculates bearing/heading from point A to point B (0-360 degrees)
func CalculateBearing(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	// Calculate bearing
	y := math.Sin(deltaLon) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) -
		math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(deltaLon)

	bearing := math.Atan2(y, x) * 180 / math.Pi

	// Normalize to 0-360
	bearing = math.Mod(bearing+360, 360)

	return bearing
}

// CalculateSpeed calculates speed in knots from two positions and time difference
// Returns 0 if time difference is invalid (<=0 or >=1 hour)
func CalculateSpeed(lat1, lon1, lat2, lon2 float64, timeDiffHours float64) float64 {
	if timeDiffHours <= 0 || timeDiffHours >= 1 {
		return 0
	}
	distanceNM := HaversineDistance(lat1, lon1, lat2, lon2)
	return RoundToTwoDecimals(distanceNM / timeDiffHours)
}

// RoundToTwoDecimals rounds a float64 to 2 decimal places
func RoundToTwoDecimals(val float64) float64 {
	return math.Round(val*100) / 100
}

// Helper methods
func (r *repository) GetAllVessels() ([]model.MVessel, error) {
	var vessels []model.MVessel
	err := r.db.Where("is_active = ?", true).Find(&vessels).Error
	return vessels, err
}

func (r *repository) GetSensorByID(sensorID uint) (*model.MSensor, error) {
	var sensor model.MSensor
	err := r.db.First(&sensor, sensorID).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &sensor, err
}

func (r *repository) GetVesselByID(vesselID uint) (*model.MVessel, error) {
	var vessel model.MVessel
	err := r.db.First(&vessel, vesselID).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &vessel, err
}

func (r *repository) GetVesselIDBySensorID(sensorID uint) (uint, error) {
	// Use raw SQL JOIN for reliable resolution through nullable foreign keys
	// Chain: m_sensors.mcu_id → m_mcu.id → m_mcu.vessel_id
	var result struct {
		VesselID uint
	}
	err := r.db.Raw(`
		SELECT m.vessel_id
		FROM m_sensors s
		JOIN m_mcu m ON s.mcu_id = m.id
		WHERE s.id = ? AND m.vessel_id IS NOT NULL
	`, sensorID).Scan(&result).Error
	if err != nil {
		return 0, err
	}
	if result.VesselID == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	return result.VesselID, nil
}

func (r *repository) GetMCUByID(mcuID uint) (*model.MMCU, error) {
	var mcu model.MMCU
	err := r.db.Preload("Vessel").First(&mcu, mcuID).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &mcu, err
}

func (r *repository) GetVesselByName(name string) (*model.MVessel, error) {
	var vessel model.MVessel
	err := r.db.Where("name = ?", name).First(&vessel).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &vessel, err
}

// GetVesselByCode finds vessel by code (case-insensitive for flexibility)
func (r *repository) GetVesselByCode(code string) (*model.MVessel, error) {
	var vessel model.MVessel
	err := r.db.Where("LOWER(code) = LOWER(?)", code).First(&vessel).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &vessel, err
}

// GetCompanyByCode finds company by code (case-insensitive for flexibility)
func (r *repository) GetCompanyByCode(code string) (*model.MCompany, error) {
	var company model.MCompany
	err := r.db.Where("LOWER(code) = LOWER(?)", code).First(&company).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &company, err
}

// ========================================
// TimescaleDB Optimized Queries
// ========================================

// GetTrackingWithTimescale gets tracking data for a specific vessel using TimescaleDB optimizations
// Returns only completed minutes (excludes current minute to avoid partial data)
func (r *repository) GetTrackingWithTimescale(vesselID uint, startTime, endTime time.Time, sort string, limit int, hasCoords bool) ([]model.Tracking, error) {
	var trackings []model.Tracking

	// Validate sort parameter (default: ASC for chronological order)
	orderDirection := "ASC"
	if sort == "desc" {
		orderDirection = "DESC"
	}

	// Current minute boundary - exclude current minute to get complete data only
	currentMinute := time.Now().Truncate(time.Minute)

	// Build WHERE clauses
	conditions := `vessel_id = ?
		AND time >= ?
		AND time <= ?
		AND time < ?`
	if hasCoords {
		conditions += ` AND lat IS NOT NULL AND lon IS NOT NULL`
	}

	// Build query - only add LIMIT if limit > 0
	var query string
	var err error

	if limit > 0 {
		query = `
			SELECT * FROM trackings
			WHERE ` + conditions + `
			ORDER BY time ` + orderDirection + `
			LIMIT ?
		`
		err = r.db.Raw(query, vesselID, startTime, endTime, currentMinute, limit).Scan(&trackings).Error
	} else {
		query = `
			SELECT * FROM trackings
			WHERE ` + conditions + `
			ORDER BY time ` + orderDirection + `
		`
		err = r.db.Raw(query, vesselID, startTime, endTime, currentMinute).Scan(&trackings).Error
	}

	return trackings, err
}

// GetAllTrackingWithTimescale gets tracking data for all vessels using TimescaleDB optimizations
// Returns only completed minutes (excludes current minute to avoid partial data)
func (r *repository) GetAllTrackingWithTimescale(startTime, endTime time.Time, sort string, limit int, hasCoords bool) ([]model.Tracking, error) {
	var trackings []model.Tracking

	// Validate sort parameter (default: ASC for chronological order)
	orderDirection := "ASC"
	if sort == "desc" {
		orderDirection = "DESC"
	}

	// Current minute boundary - exclude current minute to get complete data only
	currentMinute := time.Now().Truncate(time.Minute)

	// Build WHERE clauses
	conditions := `time >= ?
		AND time <= ?
		AND time < ?`
	if hasCoords {
		conditions += ` AND lat IS NOT NULL AND lon IS NOT NULL`
	}

	// Build query - only add LIMIT if limit > 0
	var query string
	var err error

	if limit > 0 {
		query = `
			SELECT * FROM trackings
			WHERE ` + conditions + `
			ORDER BY time ` + orderDirection + `
			LIMIT ?
		`
		err = r.db.Raw(query, startTime, endTime, currentMinute, limit).Scan(&trackings).Error
	} else {
		query = `
			SELECT * FROM trackings
			WHERE ` + conditions + `
			ORDER BY time ` + orderDirection + `
		`
		err = r.db.Raw(query, startTime, endTime, currentMinute).Scan(&trackings).Error
	}

	return trackings, err
}

// ── Historical Spike Cleanup ──────────────────────────────────────────

func (r *repository) GetDistinctVesselIDsWithCoords() ([]uint, error) {
	var ids []uint
	err := r.db.Model(&model.Tracking{}).
		Select("DISTINCT vessel_id").
		Where("lat IS NOT NULL AND lon IS NOT NULL").
		Pluck("vessel_id", &ids).Error
	return ids, err
}

func (r *repository) GetVesselIDsWithCoordsOrderedByLatest() ([]uint, error) {
	var ids []uint
	err := r.db.Model(&model.Tracking{}).
		Select("vessel_id").
		Where("lat IS NOT NULL AND lon IS NOT NULL").
		Group("vessel_id").
		Order("MAX(time) DESC").
		Pluck("vessel_id", &ids).Error
	return ids, err
}

func (r *repository) CountTrackingWithCoords(vesselID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Tracking{}).
		Where("vessel_id = ? AND lat IS NOT NULL AND lon IS NOT NULL", vesselID).
		Count(&count).Error
	return count, err
}

func (r *repository) GetTrackingChunk(vesselID uint, offset, limit int) ([]model.Tracking, error) {
	var trackings []model.Tracking
	err := r.db.Where("vessel_id = ? AND lat IS NOT NULL AND lon IS NOT NULL", vesselID).
		Order("time ASC").
		Offset(offset).Limit(limit).
		Find(&trackings).Error
	return trackings, err
}

func (r *repository) GetTrackingChunkAfter(vesselID uint, afterTime time.Time, limit int) ([]model.Tracking, error) {
	var trackings []model.Tracking
	err := r.db.Where("vessel_id = ? AND lat IS NOT NULL AND lon IS NOT NULL AND time > ?", vesselID, afterTime).
		Order("time ASC").
		Limit(limit).
		Find(&trackings).Error
	return trackings, err
}

func (r *repository) NullifyTrackingCoords(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Model(&model.Tracking{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"lat":     nil,
			"lon":     nil,
			"speed":   nil,
			"heading": nil,
		}).Error
}

func (r *repository) BatchNullifyTrackingCoords(ids []uint) (int, []uint, error) {
	if len(ids) == 0 {
		return 0, nil, nil
	}
	totalNullified := 0
	var failed []uint
	for _, id := range ids {
		// DELETE + INSERT via CTE avoids TimescaleDB decompression cost on UPDATE
		err := r.db.Exec(`
			WITH deleted AS (
				DELETE FROM trackings WHERE id = ? RETURNING *
			)
			INSERT INTO trackings (
				id, time, vessel_id, lat, lon, speed, heading,
				rpm1, rpm2, rpm3, rpm4,
				flow1, flow2, flow3, flow4,
				top_ac_volt, top_ac_amp, top_ac_freq, top_ac_power,
				top_dc_volt, top_dc_amp, top_bat_volt, top_bat_amp,
				bot_ac_volt, bot_ac_amp, bot_ac_freq, bot_ac_power,
				bot_dc_volt, bot_dc_amp, bot_bat_volt, bot_bat_amp,
				top_door1, top_door2, bot_door1, bot_door2,
				top_humidity, bot_humidity, top_temp, bot_temp,
				accel_x, accel_y, accel_z
			)
			SELECT
				id, time, vessel_id,
				NULL, NULL, NULL, NULL,
				rpm1, rpm2, rpm3, rpm4,
				flow1, flow2, flow3, flow4,
				top_ac_volt, top_ac_amp, top_ac_freq, top_ac_power,
				top_dc_volt, top_dc_amp, top_bat_volt, top_bat_amp,
				bot_ac_volt, bot_ac_amp, bot_ac_freq, bot_ac_power,
				bot_dc_volt, bot_dc_amp, bot_bat_volt, bot_bat_amp,
				top_door1, top_door2, bot_door1, bot_door2,
				top_humidity, bot_humidity, top_temp, bot_temp,
				accel_x, accel_y, accel_z
			FROM deleted
		`, id).Error
		if err != nil {
			log.Printf("[GPS-CLEANUP] Error nullifying id=%d: %v", id, err)
			failed = append(failed, id)
			continue
		}
		totalNullified++
	}
	if len(failed) > 0 {
		return totalNullified, failed, fmt.Errorf("TOAST errors: %d IDs failed", len(failed))
	}
	return totalNullified, nil, nil
}

func (r *repository) VacuumTrackings() error {
	return r.db.Exec("VACUUM ANALYZE trackings").Error
}

func (r *repository) DecompressTrackingsChunks() error {
	return r.db.Exec(`
		SELECT decompress_chunk(chunk_schema || '.' || chunk_name)
		FROM timescaledb_information.chunks
		WHERE hypertable_name = 'trackings' AND is_compressed = true
	`).Error
}
