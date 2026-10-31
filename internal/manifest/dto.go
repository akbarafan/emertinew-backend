package manifest

import "time"

// CreateManifestRequest untuk request payload create manifest
type CreateManifestRequest struct {
	VesselID      uint       `json:"vessel_id" binding:"required" example:"1"`
	CaptainID     *uint      `json:"captain_id" example:"1"`
	GrossTonnage  *float64   `json:"gross_tonnage" example:"1500.50"`
	Date          *time.Time `json:"date" example:"2024-01-15T00:00:00Z"`
	ArrivalTime   *time.Time `json:"arrival_time" example:"2024-01-15T08:00:00Z"`
	DepartureTime *time.Time `json:"departure_time" example:"2024-01-15T18:00:00Z"`
}

// UpdateManifestRequest untuk request payload update manifest
type UpdateManifestRequest struct {
	VesselID      *uint      `json:"vessel_id" example:"2"`
	CaptainID     *uint      `json:"captain_id" example:"2"`
	GrossTonnage  *float64   `json:"gross_tonnage" example:"2000.75"`
	Date          *time.Time `json:"date" example:"2024-01-16T00:00:00Z"`
	ArrivalTime   *time.Time `json:"arrival_time" example:"2024-01-16T09:00:00Z"`
	DepartureTime *time.Time `json:"departure_time" example:"2024-01-16T19:00:00Z"`
	Status        *string    `json:"status" enums:"draft,review,approved" example:"review"`
}

// VesselInfo untuk info vessel dalam response
type VesselInfo struct {
	ID   uint   `json:"id" example:"1"`
	Name string `json:"name" example:"KM Nusantara 01"`
}

// CaptainInfo untuk info captain dalam response
type CaptainInfo struct {
	ID          uint   `json:"id" example:"1"`
	Name        string `json:"name" example:"John Doe"`
	Department  string `json:"department" example:"Deck"`
	JobPosition string `json:"job_position" example:"Captain"`
}

// ManifestResponse untuk response data manifest
type ManifestResponse struct {
	ID            uint         `json:"id" example:"1"`
	Name          string       `json:"name" example:"MNF-20240115-001"`
	Vessel        *VesselInfo  `json:"vessel,omitempty"`
	Captain       *CaptainInfo `json:"captain,omitempty"`
	GrossTonnage  *float64     `json:"gross_tonnage,omitempty" example:"1500.50"`
	Date          *string      `json:"date,omitempty" example:"2024-01-15"`
	ArrivalTime   *time.Time   `json:"arrival_time,omitempty" example:"2024-01-15T08:00:00+07:00"`
	DepartureTime *time.Time   `json:"departure_time,omitempty" example:"2024-01-15T18:00:00+07:00"`
	Status        string       `json:"status" example:"draft"`
	CreatedAt     time.Time    `json:"created_at" example:"2024-01-01T10:00:00+07:00"`
	UpdatedAt     time.Time    `json:"updated_at" example:"2024-01-01T10:00:00+07:00"`
}
