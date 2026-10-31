package sensor

import "time"

// CreateSensorRequest untuk request payload create sensor
type CreateSensorRequest struct {
	MCUID         *uint    `json:"mcu_id" example:"1"`
	Topic         string   `json:"topic" binding:"required" example:"hcml/vs001/top/gps"` // Primary identifier
	Name          string   `json:"name" binding:"required" example:"GPS Sensor Vessel 1"`
	KFactor       *float64 `json:"k_factor" example:"200"`            // Calibration factor (pulse/L) for flow sensors. 0 = skip.
	IsVisible     *bool    `json:"is_visible" example:"true"`         // Show in daily report charts
	EngineName    *string  `json:"engine_name" example:"Port Engine"` // Engine this sensor belongs to
	FlowDirection *string  `json:"flow_direction" example:"in"`       // "in", "out", or null for single
}

// UpdateSensorRequest untuk request payload update sensor
type UpdateSensorRequest struct {
	MCUID         *uint          `json:"mcu_id" example:"1"`
	Topic         string         `json:"topic" binding:"required" example:"hcml/vs001/top/gps"`
	Name          string         `json:"name" binding:"required" example:"GPS Sensor Vessel 2"`
	KFactor       *float64       `json:"k_factor" example:"200"`            // Calibration factor (pulse/L) for flow sensors. 0 = skip.
	IsVisible     *bool          `json:"is_visible" example:"true"`         // Show in daily report charts
	EngineName    OptionalString `json:"engine_name" example:"Port Engine"` // Engine this sensor belongs to. null clears it.
	FlowDirection OptionalString `json:"flow_direction" example:"in"`       // "in", "out", or null for single
}

// SensorResponse untuk response data sensor
type SensorResponse struct {
	ID            uint     `json:"id" example:"1"`
	MCUID         *uint    `json:"mcu_id" example:"1"`
	Topic         string   `json:"topic" example:"hcml/vs001/top/gps"` // Primary identifier
	Name          string   `json:"name" example:"GPS Sensor Vessel 1"`
	KFactor       *float64 `json:"k_factor" example:"200"`
	IsVisible     bool     `json:"is_visible" example:"true"`         // Show in daily report charts
	EngineName    *string  `json:"engine_name" example:"Port Engine"` // Engine this sensor belongs to
	FlowDirection *string  `json:"flow_direction" example:"in"`       // "in", "out", or null for single
}

// SensorAuditData untuk audit log sensor
type SensorAuditData struct {
	ID            uint     `json:"id"`
	MCUID         *uint    `json:"mcu_id,omitempty"`
	Topic         string   `json:"topic"`
	Name          string   `json:"name"`
	KFactor       *float64 `json:"k_factor,omitempty"`
	EngineName    *string  `json:"engine_name,omitempty"`
	FlowDirection *string  `json:"flow_direction,omitempty"`
}

// EngineSensorMapping represents a single sensor in an engine group
type EngineSensorMapping struct {
	SensorID   uint   `json:"sensor_id"`
	Topic      string `json:"topic"`
	Slot       string `json:"slot"`      // e.g. "flow1", "rpm2"
	Kind       string `json:"kind"`      // "flow" or "rpm"
	Direction  string `json:"direction"` // "in", "out", or "" — flow meters only
	SensorName string `json:"sensor_name"`
}

// EngineMapping represents the sensors of one engine
type EngineMapping struct {
	Name             string                `json:"name"`
	Sensors          []EngineSensorMapping `json:"sensors"`
	VisibilityFields []string              `json:"visibility_fields"`
}

// VesselEngineMappingResponse is the response for the engine-mapping endpoint
type VesselEngineMappingResponse struct {
	VesselID uint            `json:"vessel_id"`
	Engines  []EngineMapping `json:"engines"`
}

// ToggleVisibilityRequest untuk request toggle visibility
type ToggleVisibilityRequest struct {
	IsVisible bool `json:"is_visible" example:"true"`
}

// VesselSensorVisibilityResponse mapping tracking field → visible status
type VesselSensorVisibilityResponse struct {
	VesselID uint            `json:"vessel_id"`
	Fields   map[string]bool `json:"fields"` // e.g. {"rpm1":true, "rpm2":false}
}

// SensorDataResponse untuk response data dari tabel sensor_data (TimescaleDB)
type SensorDataResponse struct {
	Time       time.Time              `json:"time" example:"2024-01-01T10:00:00Z"`
	SensorID   uint                   `json:"sensor_id" example:"1"`
	Topic      string                 `json:"topic" example:"hcml/vs001/top/gps"`
	SensorName string                 `json:"sensor_name" example:"GPS Sensor 1"`
	Value      map[string]interface{} `json:"value" example:"{\"lat\":-7.2575,\"lon\":112.7521}"`
	TrackingID *uint                  `json:"tracking_id,omitempty" example:"1"`
	VesselID   *uint                  `json:"vessel_id,omitempty" example:"1"`
	VesselName string                 `json:"vessel_name,omitempty" example:"KM Nusantara 01"`
}
