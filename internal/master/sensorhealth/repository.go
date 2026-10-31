package sensorhealth

import (
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	GetActiveSlots(sensorIDs []uint, start, end time.Time) ([]ActiveSlotRow, error)
	GetVesselConnectivity(vesselIDs []uint, start, end time.Time) ([]VesselSlotRow, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

type ActiveSlotRow struct {
	SensorID uint
	Day      time.Time
	SlotIdx  int // 0-11, each representing a 2-hour block
}

type VesselSlotRow struct {
	VesselID uint
	Day      time.Time
	SlotIdx  int
}

func (r *repository) GetActiveSlots(sensorIDs []uint, start, end time.Time) ([]ActiveSlotRow, error) {
	var rows []ActiveSlotRow
	err := r.db.Table("sensor_data").
		Select("sensor_id, date_trunc('day', time AT TIME ZONE 'UTC') AT TIME ZONE 'UTC' as day, floor(extract(hour from time) / 2)::int as slot_idx").
		Where("sensor_id IN ? AND time >= ? AND time < ?", sensorIDs, start, end).
		Group("sensor_id, day, slot_idx").
		Order("sensor_id, day, slot_idx").
		Scan(&rows).Error
	return rows, err
}

func (r *repository) GetVesselConnectivity(vesselIDs []uint, start, end time.Time) ([]VesselSlotRow, error) {
	var rows []VesselSlotRow
	err := r.db.Table("sensor_data sd").
		Select("m.vessel_id as vessel_id, date_trunc('day', sd.time AT TIME ZONE 'UTC') AT TIME ZONE 'UTC' as day, floor(extract(hour from sd.time) / 2)::int as slot_idx").
		Joins("JOIN m_sensors s ON s.id = sd.sensor_id").
		Joins("JOIN m_mcu m ON m.id = s.mcu_id").
		Where("m.vessel_id IN ? AND sd.time >= ? AND sd.time < ?", vesselIDs, start, end).
		Group("m.vessel_id, day, slot_idx").
		Order("m.vessel_id, day, slot_idx").
		Scan(&rows).Error
	return rows, err
}
