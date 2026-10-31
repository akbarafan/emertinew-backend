package geofencelog

import (
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

// CheckGeofenceRequest DTO untuk check vessel position
type CheckGeofenceRequest struct {
	VesselID   uint    `json:"vessel_id" binding:"required" example:"1"`
	GeofenceID uint    `json:"geofence_id" binding:"required" example:"1"`
	Latitude   float64 `json:"latitude" binding:"required" example:"-6.2088"`
	Longitude  float64 `json:"longitude" binding:"required" example:"106.8456"`
}

// CreateGeofenceLogRequest DTO untuk create geofence log dengan data dari frontend
type CreateGeofenceLogRequest struct {
	VesselID          uint    `json:"vessel_id" binding:"required" example:"1"`
	GeofenceID        uint    `json:"geofence_id" binding:"required" example:"1"`
	Latitude          float64 `json:"latitude" binding:"required" example:"-6.2088"`
	Longitude         float64 `json:"longitude" binding:"required" example:"106.8456"`
	IsOutsideGeofence bool    `json:"is_outside_geofence" example:"false"`
	Distance          float64 `json:"distance" example:"0"`
}

// CheckGeofenceResponse DTO untuk response check (tanpa create log)
type CheckGeofenceResponse struct {
	VesselID          uint    `json:"vessel_id" example:"1"`
	GeofenceID        uint    `json:"geofence_id" example:"1"`
	GeofenceName      string  `json:"geofence_name" example:"Zona Pelabuhan A"`
	Latitude          float64 `json:"latitude" example:"-6.2088"`
	Longitude         float64 `json:"longitude" example:"106.8456"`
	IsOutsideGeofence bool    `json:"is_outside_geofence" example:"false"`
	Distance          float64 `json:"distance" example:"0"` // in miles, 0 if inside geofence, distance beyond radius if outside
}

// GeofenceLogResponse DTO untuk response geofence log
type GeofenceLogResponse struct {
	ID                uint      `json:"id" example:"1"`
	VesselID          uint      `json:"vessel_id" example:"1"`
	VesselName        string    `json:"vessel_name,omitempty"`
	GeofenceID        uint      `json:"geofence_id" example:"1"`
	GeofenceName      string    `json:"geofence_name" example:"Zona Pelabuhan A"`
	EventType         string    `json:"event_type" example:"exited"`
	Latitude          float64   `json:"latitude" example:"-6.2088"`
	Longitude         float64   `json:"longitude" example:"106.8456"`
	Speed             float64   `json:"speed" example:"10.5"`
	IsOutsideGeofence bool      `json:"is_outside_geofence" example:"false"`
	Distance          float64   `json:"distance" example:"0"` // in miles
	Message           string    `json:"message,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// GeofenceRepository interface untuk akses geofence data
type GeofenceRepository interface {
	GetByID(id uint) (*model.MGeofence, error)
}
