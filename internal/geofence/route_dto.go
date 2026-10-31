package geofence

import (
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

// WaypointDTO data transfer object untuk waypoint rute
type WaypointDTO struct {
	ID             *uint   `json:"id,omitempty"`
	Sequence       int     `json:"sequence"`
	Name           string  `json:"name"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	WaypointType   string  `json:"waypointType"` // port_call, waypoint, checkpoint
	CorridorWidthM float64 `json:"corridorWidthM"`
	RadiusM        float64 `json:"radiusM"`
	Notes          string  `json:"notes,omitempty"`
}

// CreateRouteRequest DTO untuk create rute geofence
type CreateRouteRequest struct {
	CompanyID              *uint         `json:"company_id"`
	VoyageID               *uint         `json:"voyage_id"`
	VesselID               *uint         `json:"vessel_id"`
	Name                   string        `json:"name" binding:"required"`
	Description            string        `json:"description"`
	TotalDistanceNM        float64       `json:"totalDistanceNm"`
	EstimatedDurationHours *float64      `json:"estimatedDurationHours"`
	DefaultCorridorWidthM  float64       `json:"defaultCorridorWidthM"`
	Status                 string        `json:"status"`
	IsActive               *bool         `json:"isActive"`
	AlertOnDeviation       *bool         `json:"alertOnDeviation"`
	Waypoints              []WaypointDTO `json:"waypoints" binding:"required,min=2"`
}

// UpdateRouteRequest DTO untuk update rute geofence
type UpdateRouteRequest struct {
	CompanyID              *uint         `json:"company_id"`
	VoyageID               *uint         `json:"voyage_id"`
	VesselID               *uint         `json:"vessel_id"`
	Name                   *string       `json:"name"`
	Description            *string       `json:"description"`
	TotalDistanceNM        *float64      `json:"totalDistanceNm"`
	EstimatedDurationHours *float64      `json:"estimatedDurationHours"`
	DefaultCorridorWidthM  *float64      `json:"defaultCorridorWidthM"`
	Status                 *string       `json:"status"`
	IsActive               *bool         `json:"isActive"`
	AlertOnDeviation       *bool         `json:"alertOnDeviation"`
	Waypoints              []WaypointDTO `json:"waypoints"`
}

// GenerateRouteFromVoyageRequest DTO untuk auto-generate rute dari voyage clearance
type GenerateRouteFromVoyageRequest struct {
	VoyageID              uint    `json:"voyage_id" binding:"required"`
	Name                  string  `json:"name"`
	DefaultCorridorWidthM float64 `json:"defaultCorridorWidthM"`
	WaypointIntervalNM    float64 `json:"waypointIntervalNm"`
}

// RouteResponse DTO untuk response data rute
type RouteResponse struct {
	ID                     uint                           `json:"id"`
	CompanyID              *uint                          `json:"company_id,omitempty"`
	CompanyName            string                         `json:"company_name,omitempty"`
	VoyageID               *uint                          `json:"voyage_id,omitempty"`
	VoyageNumber           string                         `json:"voyage_number,omitempty"`
	VesselID               *uint                          `json:"vessel_id,omitempty"`
	VesselName             string                         `json:"vessel_name,omitempty"`
	Name                   string                         `json:"name"`
	Description            string                         `json:"description"`
	TotalDistanceNM        float64                        `json:"totalDistanceNm"`
	EstimatedDurationHours *float64                       `json:"estimatedDurationHours,omitempty"`
	DefaultCorridorWidthM  float64                        `json:"defaultCorridorWidthM"`
	Status                 string                         `json:"status"`
	IsActive               bool                           `json:"isActive"`
	AlertOnDeviation       bool                           `json:"alertOnDeviation"`
	Waypoints              []model.MGeofenceRouteWaypoint `json:"waypoints"`
	CreatedAt              time.Time                      `json:"created_at"`
	UpdatedAt              time.Time                      `json:"updated_at"`
}
