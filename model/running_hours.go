package model

import "time"

// VesselEngineRunningHours stores lifetime and trip running hours per engine slot
type VesselEngineRunningHours struct {
	ID                uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	VesselID          uint       `gorm:"index:idx_vessel_engine_slot,unique;not null" json:"vessel_id"`
	SensorID          *uint      `gorm:"index" json:"sensor_id,omitempty"` // ID of the RPM sensor from m_sensor
	EngineSlot        string     `gorm:"type:varchar(20);index:idx_vessel_engine_slot,unique;not null" json:"engine_slot"` // e.g. "rpm1", "rpm2", "rpm3", "rpm4"
	TotalRunningHours float64    `gorm:"type:double precision;default:0;not null" json:"total_running_hours"`
	TripRunningHours  float64    `gorm:"type:double precision;default:0;not null" json:"trip_running_hours"`
	LastActiveAt      *time.Time `gorm:"type:timestamptz" json:"last_active_at,omitempty"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (VesselEngineRunningHours) TableName() string {
	return "vessel_engine_running_hours"
}

// TripRunningHoursResetLog stores audit log for each trip running hours reset
type TripRunningHoursResetLog struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	VesselID         uint      `gorm:"index;not null" json:"vessel_id"`
	SensorID         *uint     `gorm:"index" json:"sensor_id,omitempty"` // ID of the RPM sensor from m_sensor
	EngineSlot       string    `gorm:"type:varchar(20);not null" json:"engine_slot"` // e.g. "rpm1", "rpm2"
	HoursBeforeReset float64   `gorm:"type:double precision;not null" json:"hours_before_reset"`
	TripName         string    `gorm:"type:varchar(255)" json:"trip_name"`
	ResetByUserID    *uint     `gorm:"index" json:"reset_by_user_id,omitempty"`
	ResetByName      string    `gorm:"type:varchar(100);not null" json:"reset_by_name"`
	ResetByRole      string    `gorm:"type:varchar(50);not null" json:"reset_by_role"`
	ResetAt          time.Time `gorm:"autoCreateTime;index" json:"reset_at"`
}

func (TripRunningHoursResetLog) TableName() string {
	return "trip_running_hours_reset_logs"
}
