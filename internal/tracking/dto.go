package tracking

import (
	"time"
)

// VesselTrackingDTO untuk response tracking per vessel
type VesselTrackingDTO struct {
	ID         uint     `json:"id,omitempty" example:"1"`
	VesselID   uint     `json:"vessel_id" example:"1"`
	VesselName string   `json:"vessel_name" example:"KM Nusantara 01"`
	Lat        *float64 `json:"lat" example:"-7.2575"`
	Lon        *float64 `json:"lon" example:"112.7521"`
	Speed      *float64 `json:"speed" example:"12.5"`  // knots, calculated
	Heading    *float64 `json:"heading" example:"180"` // degrees (0-360), calculated
	RPM1       *int     `json:"rpm1" example:"800"`    // RPM value 1
	RPM2       *int     `json:"rpm2" example:"850"`    // RPM value 2
	RPM3       *int     `json:"rpm3" example:"900"`    // RPM value 3
	RPM4       *int     `json:"rpm4" example:"920"`    // RPM value 4
	Flow1      *float64 `json:"flow1" example:"45.5"`  // Flow value 1
	Flow2      *float64 `json:"flow2" example:"46.2"`  // Flow value 2
	Flow3      *float64 `json:"flow3" example:"44.8"`  // Flow value 3
	Flow4      *float64 `json:"flow4" example:"48.1"`  // Flow value 4
	// Top MCU Electrical
	TopACVolt  *float64 `json:"top_ac_volt" example:"220.0"`   // Top AC Voltage
	TopACAmp   *float64 `json:"top_ac_amp" example:"10.5"`     // Top AC Ampere
	TopACFreq  *float64 `json:"top_ac_freq" example:"50.0"`    // Top AC Frequency
	TopACPower *float64 `json:"top_ac_power" example:"2200.0"` // Top AC Power
	TopDCVolt  *float64 `json:"top_dc_volt" example:"24.0"`    // Top DC Voltage
	TopDCAmp   *float64 `json:"top_dc_amp" example:"15.5"`     // Top DC Ampere
	TopBatVolt *float64 `json:"top_bat_volt" example:"12.6"`   // Top Battery Voltage
	TopBatAmp  *float64 `json:"top_bat_amp" example:"8.5"`     // Top Battery Ampere
	// Bottom MCU Electrical
	BotACVolt  *float64 `json:"bot_ac_volt" example:"220.0"`   // Bottom AC Voltage
	BotACAmp   *float64 `json:"bot_ac_amp" example:"10.5"`     // Bottom AC Ampere
	BotACFreq  *float64 `json:"bot_ac_freq" example:"50.0"`    // Bottom AC Frequency
	BotACPower *float64 `json:"bot_ac_power" example:"2200.0"` // Bottom AC Power
	BotDCVolt  *float64 `json:"bot_dc_volt" example:"24.0"`    // Bottom DC Voltage
	BotDCAmp   *float64 `json:"bot_dc_amp" example:"15.5"`     // Bottom DC Ampere
	BotBatVolt *float64 `json:"bot_bat_volt" example:"12.6"`   // Bottom Battery Voltage
	BotBatAmp  *float64 `json:"bot_bat_amp" example:"8.5"`     // Bottom Battery Ampere
	// Door Status
	TopDoor1 *bool `json:"top_door1" example:"false"` // Top Door 1 Status
	TopDoor2 *bool `json:"top_door2" example:"false"` // Top Door 2 Status
	BotDoor1 *bool `json:"bot_door1" example:"false"` // Bottom Door 1 Status
	BotDoor2 *bool `json:"bot_door2" example:"false"` // Bottom Door 2 Status
	// Environment
	TopHumidity *float64 `json:"top_humidity" example:"65.5"` // Top Humidity
	BotHumidity *float64 `json:"bot_humidity" example:"70.2"` // Bottom Humidity
	TopTemp     *float64 `json:"top_temp" example:"28.5"`     // Top Temperature
	BotTemp     *float64 `json:"bot_temp" example:"30.2"`     // Bottom Temperature
	// Accelerometer
	AccelX    *float64  `json:"accel_x" example:"0.1"` // Accelerometer X
	AccelY    *float64  `json:"accel_y" example:"0.2"` // Accelerometer Y
	AccelZ    *float64  `json:"accel_z" example:"9.8"` // Accelerometer Z
	// BMKG Maritime Weather Data
	WaveHeight   *float64 `json:"wave_height,omitempty" example:"1.2"`
	WaveCat      *string  `json:"wave_cat,omitempty" example:"Sedang"`
	CurrentSpeed *float64 `json:"current_speed,omitempty" example:"1.5"`
	CurrentTo    *string  `json:"current_to,omitempty" example:"Barat (270°)"`
	WindSpeed    *float64 `json:"wind_speed,omitempty" example:"15.0"`
	WindGust     *float64 `json:"wind_gust,omitempty" example:"22.0"`
	WindFrom     *string  `json:"wind_from,omitempty" example:"Timur (090°)"`
	Weather      *string  `json:"weather,omitempty" example:"Cerah Berawan"`
	Timestamp    time.Time `json:"timestamp" example:"2024-01-01T10:00:00Z"`
}

// TrackingQueryParams untuk query params API tracking
type TrackingQueryParams struct {
	StartTime string `form:"start_time" example:"2024-01-01T00:00:00Z"` // Format: RFC3339
	EndTime   string `form:"end_time" example:"2024-01-31T23:59:59Z"`   // Format: RFC3339
	Sort      string `form:"sort" example:"desc"`                       // asc or desc (default: desc)
	Limit     int    `form:"limit" example:"100"`                       // Limit number of records (default: 100)
	HasCoords *bool  `form:"has_coords" example:"true"`                 // Filter hanya record dengan lat/lon tidak null
}

// TrackingBroadcast untuk broadcast filtering di WebSocket
type TrackingBroadcast struct {
	VesselID uint
}

// WebSocketFilter represents the filter criteria for WebSocket clients
type WebSocketFilter struct {
	VesselID *uint
}

// RealtimeTrackingState menyimpan state tracking realtime per vessel
// Data diakumulasi dari berbagai sensor (GPS, RPM, Flow, dll)
type RealtimeTrackingState struct {
	VesselID   uint
	VesselName string
	Lat        *float64
	Lon        *float64
	Speed      *float64
	Heading    *float64
	RPM1       *int
	RPM2       *int
	RPM3       *int
	RPM4       *int
	Flow1      *float64
	Flow2      *float64
	Flow3      *float64
	Flow4      *float64
	// Top MCU Electrical
	TopACVolt  *float64
	TopACAmp   *float64
	TopACFreq  *float64
	TopACPower *float64
	TopDCVolt  *float64
	TopDCAmp   *float64
	TopBatVolt *float64
	TopBatAmp  *float64
	// Bottom MCU Electrical
	BotACVolt  *float64
	BotACAmp   *float64
	BotACFreq  *float64
	BotACPower *float64
	BotDCVolt  *float64
	BotDCAmp   *float64
	BotBatVolt *float64
	BotBatAmp  *float64
	// Door Status
	TopDoor1 *bool
	TopDoor2 *bool
	BotDoor1 *bool
	BotDoor2 *bool
	// Environment
	TopHumidity *float64
	BotHumidity *float64
	TopTemp     *float64
	BotTemp     *float64
	// Accelerometer
	AccelX     *float64
	AccelY     *float64
	AccelZ     *float64
	// BMKG Maritime Weather Data
	WaveHeight   *float64
	WaveCat      *string
	CurrentSpeed *float64
	CurrentTo    *string
	WindSpeed    *float64
	WindGust     *float64
	WindFrom     *string
	Weather      *string
	LastUpdate   time.Time
	// Previous position for speed/heading calculation
	PrevLat  *float64
	PrevLon  *float64
	PrevTime time.Time
	// Spike breakthrough — accept position after N consecutive same coordinates
	SpikeCandidateLat   *float64
	SpikeCandidateLon   *float64
	SpikeCandidateCount int
}
