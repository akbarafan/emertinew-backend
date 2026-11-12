package mcu

// CreateMCURequest untuk request payload create mcu
type CreateMCURequest struct {
	VesselID  *uint  `json:"vessel_id" example:"1"`
	SensorIDs []uint `json:"sensor_id" example:"1,2"`
	Name      string `json:"name" binding:"required" example:"GPS Module A1"`
	Location  string `json:"location" binding:"required,oneof=bd er brd eng" example:"bd"`
}

// UpdateMCURequest untuk request payload update mcu
type UpdateMCURequest struct {
	VesselID  *uint  `json:"vessel_id" example:"1"`
	SensorIDs []uint `json:"sensor_id" example:"1,2,3"`
	Name      string `json:"name" example:"GPS Module A1 Updated"`
	Location  string `json:"location" binding:"omitempty,oneof=bd er brd eng" example:"er"`
}

// SensorInfo untuk info sensor dalam response
type SensorInfo struct {
	ID            uint     `json:"id" example:"1"`
	Topic         string   `json:"topic" example:"sensor/gps/001"`
	Name          string   `json:"name" example:"GPS Sensor 1"`
	IsVisible     bool     `json:"is_visible" example:"true"`
	EngineName    *string  `json:"engine_name" example:"Port Engine"`
	FlowDirection *string  `json:"flow_direction" example:"in"`
	KFactor       *float64 `json:"k_factor" example:"36.77"`
}

// MCUResponse untuk response data mcu
type MCUResponse struct {
	ID        uint         `json:"id" example:"1"`
	VesselID  *uint        `json:"vessel_id" example:"1"`
	Name      string       `json:"name" example:"GPS Module A1"`
	Location  string       `json:"location" example:"bd"`
	Sensors   []SensorInfo `json:"sensor"`
	CreatedAt string       `json:"created_at" example:"2024-01-01T10:00:00Z"`
	UpdatedAt string       `json:"updated_at" example:"2024-01-01T10:00:00Z"`
}

// MCUAuditData represents audit trail data for MCU changes
type MCUAuditData struct {
	ID        uint   `json:"id"`
	VesselID  *uint  `json:"vessel_id,omitempty"`
	Name      string `json:"name"`
	Location  string `json:"location"`
	SensorIDs []uint `json:"sensor_ids,omitempty"`
}

// ConvertSensorRequest for bulk sensor topic format conversion
type ConvertSensorRequest struct {
	SensorIDs []uint `json:"sensor_ids"` // empty = all sensors on this MCU
}

// ConvertSensorItem represents a single sensor's conversion info
type ConvertSensorItem struct {
	SensorID     uint   `json:"sensor_id"`
	Name         string `json:"name"`
	OldTopic     string `json:"old_topic"`
	NewTopic     string `json:"new_topic"`
	Conflict     bool   `json:"conflict"`
	ConflictWith *uint  `json:"conflict_with,omitempty"`
	ConflictMsg  string `json:"conflict_msg,omitempty"`
	IsNewFormat  bool   `json:"is_new_format"`
}

// ConvertSensorResponse for bulk convert result
type ConvertSensorResponse struct {
	Items     []ConvertSensorItem `json:"items"`
	Converted int                 `json:"converted"`
	Skipped   int                 `json:"skipped"`
	Conflicts int                 `json:"conflicts"`
}

// PreviewConvertResponse for preview before applying
type PreviewConvertResponse struct {
	Items       []ConvertSensorItem `json:"items"`
	Convertible int                 `json:"convertible"`
	Conflicts   int                 `json:"conflicts"`
	AlreadyNew  int                 `json:"already_new"`
}
