package model

import (
	"time"

	"gorm.io/datatypes"
)

// Tracking adalah hypertable untuk data tracking yang dikelompokkan per vessel
// Data dari semua sensor dalam 1 vessel dikumpulkan dalam 1 record per time bucket
// Speed dan Heading dihitung otomatis dari tracking sebelumnya
type Tracking struct {
	ID       uint      `gorm:"primaryKey;autoIncrement"`
	Time     time.Time `gorm:"type:timestamptz;not null;index:idx_tracking_time,priority:1"`
	VesselID uint      `gorm:"index:idx_tracking_vessel;not null"` // No foreign key constraint
	Lat      *float64  `gorm:"type:double precision"`              // Latitude
	Lon      *float64  `gorm:"type:double precision"`              // Longitude
	Speed    *float64  `gorm:"type:double precision"`              // Calculated speed in knots
	Heading  *float64  `gorm:"type:double precision"`              // Calculated heading in degrees (0-360)
	RPM1     *int      `gorm:"type:integer"`                       // RPM value 1
	RPM2     *int      `gorm:"type:integer"`                       // RPM value 2
	RPM3     *int      `gorm:"type:integer"`                       // RPM value 3
	RPM4     *int      `gorm:"type:integer"`                       // RPM value 4
	Flow1    *float64  `gorm:"type:double precision"`              // Litres per minute on channel 1 (sum of the minute's 5s samples)
	Flow2    *float64  `gorm:"type:double precision"`              // Litres per minute on channel 2
	Flow3    *float64  `gorm:"type:double precision"`              // Litres per minute on channel 3
	Flow4    *float64  `gorm:"type:double precision"`              // Litres per minute on channel 4
	// Top MCU Electrical
	TopACVolt  *float64 `gorm:"column:top_ac_volt;type:double precision"`  // Top AC Voltage
	TopACAmp   *float64 `gorm:"column:top_ac_amp;type:double precision"`   // Top AC Ampere
	TopACFreq  *float64 `gorm:"column:top_ac_freq;type:double precision"`  // Top AC Frequency
	TopACPower *float64 `gorm:"column:top_ac_power;type:double precision"` // Top AC Power
	TopDCVolt  *float64 `gorm:"column:top_dc_volt;type:double precision"`  // Top DC Voltage
	TopDCAmp   *float64 `gorm:"column:top_dc_amp;type:double precision"`   // Top DC Ampere
	TopBatVolt *float64 `gorm:"column:top_bat_volt;type:double precision"` // Top Battery Voltage
	TopBatAmp  *float64 `gorm:"column:top_bat_amp;type:double precision"`  // Top Battery Ampere
	// Bottom MCU Electrical
	BotACVolt  *float64 `gorm:"column:bot_ac_volt;type:double precision"`  // Bottom AC Voltage
	BotACAmp   *float64 `gorm:"column:bot_ac_amp;type:double precision"`   // Bottom AC Ampere
	BotACFreq  *float64 `gorm:"column:bot_ac_freq;type:double precision"`  // Bottom AC Frequency
	BotACPower *float64 `gorm:"column:bot_ac_power;type:double precision"` // Bottom AC Power
	BotDCVolt  *float64 `gorm:"column:bot_dc_volt;type:double precision"`  // Bottom DC Voltage
	BotDCAmp   *float64 `gorm:"column:bot_dc_amp;type:double precision"`   // Bottom DC Ampere
	BotBatVolt *float64 `gorm:"column:bot_bat_volt;type:double precision"` // Bottom Battery Voltage
	BotBatAmp  *float64 `gorm:"column:bot_bat_amp;type:double precision"`  // Bottom Battery Ampere
	// Door Status
	TopDoor1 *bool `gorm:"column:top_door1;type:boolean"` // Top Door 1 Status
	TopDoor2 *bool `gorm:"column:top_door2;type:boolean"` // Top Door 2 Status
	BotDoor1 *bool `gorm:"column:bot_door1;type:boolean"` // Bottom Door 1 Status
	BotDoor2 *bool `gorm:"column:bot_door2;type:boolean"` // Bottom Door 2 Status
	// Environment
	TopHumidity *float64 `gorm:"column:top_humidity;type:double precision"` // Top Humidity
	BotHumidity *float64 `gorm:"column:bot_humidity;type:double precision"` // Bottom Humidity
	TopTemp     *float64 `gorm:"column:top_temp;type:double precision"`     // Top Temperature
	BotTemp     *float64 `gorm:"column:bot_temp;type:double precision"`     // Bottom Temperature
	// Accelerometer
	AccelX *float64 `gorm:"column:accel_x;type:double precision"` // Accelerometer X
	AccelY *float64 `gorm:"column:accel_y;type:double precision"` // Accelerometer Y
	AccelZ *float64 `gorm:"column:accel_z;type:double precision"` // Accelerometer Z
}

// SensorData adalah hypertable untuk data sensor mentah dari MQTT
// Setiap record adalah 1 sensor reading, linked ke tracking untuk traceability
type SensorData struct {
	Time       time.Time      `gorm:"type:timestamptz;not null;index:idx_sensor_data_time,priority:1"`
	SensorID   uint           `gorm:"index:idx_sensor_data_sensor;not null"` // No foreign key constraint
	Value      datatypes.JSON `gorm:"type:jsonb;not null"`                   // Raw sensor value
	TrackingID *uint          `gorm:"index:idx_sensor_data_tracking"`        // Link to Tracking.ID for data lineage
}
