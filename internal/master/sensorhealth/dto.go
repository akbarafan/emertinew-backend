package sensorhealth

type SensorHealthRequest struct {
	Month    string `form:"month" binding:"required"` // YYYY-MM
	VesselID *uint  `form:"vessel_id"`
}

type VesselHealthResponse struct {
	VesselID            uint               `json:"vessel_id"`
	VesselName          string             `json:"vessel_name"`
	TotalPercentage     float64            `json:"total_percentage"`
	TotalActive         int                `json:"total_active"`
	TotalSlots          int                `json:"total_slots"`
	InternetDownSlots   int                `json:"internet_down_slots"`
	SensorOffSlots      int                `json:"sensor_off_slots"`
	InternetPercentage  float64            `json:"internet_percentage"`
	SensorOffPercentage float64            `json:"sensor_off_percentage"`
	Sensors             []SensorHealthResp `json:"sensors"`
}

type SensorHealthResp struct {
	SensorID            uint            `json:"sensor_id"`
	Topic               string          `json:"topic"`
	Name                string          `json:"name"`
	Category            string          `json:"category"`
	TotalPercentage     float64         `json:"total_percentage"`
	TotalActive         int             `json:"total_active"`
	TotalSlots          int             `json:"total_slots"`
	InternetDownSlots   int             `json:"internet_down_slots"`
	SensorOffSlots      int             `json:"sensor_off_slots"`
	InternetPercentage  float64         `json:"internet_percentage"`
	SensorOffPercentage float64         `json:"sensor_off_percentage"`
	Daily               []DayHealthResp `json:"daily"`
}

type DayHealthResp struct {
	Date        string  `json:"date"` // YYYY-MM-DD
	ActiveSlots int     `json:"active_slots"`
	TotalSlots  int     `json:"total_slots"`
	Percentage  float64 `json:"percentage"`
	SlotStatus  []int   `json:"slot_status"` // 0=future/no_data, 1=active, 2=internet_down, 3=sensor_off
}
