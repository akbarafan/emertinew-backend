package geofence

import (
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

// CreateGeofenceRequest DTO untuk create geofence
type CreateGeofenceRequest struct {
	CompanyID     *uint        `json:"company_id"`
	Name          string       `json:"name" binding:"required" example:"Zona Pelabuhan A"`
	Description   string       `json:"description" example:"Area pelabuhan utama"`
	ZoneType      string       `json:"zone_type" example:"work_zone"`
	Points        model.Points `json:"points" binding:"required,min=2" example:"[{\"latitude\":-6.2088,\"longitude\":106.8456}]"`
	IsActive      bool         `json:"is_active" example:"true"`
	RadiusMiles   float64      `json:"radius_miles" binding:"required,gt=0" example:"0.5"`
	MaxSpeedKnots *float64     `json:"max_speed_knots" example:"12.0"`
	AlertOnExit   *bool        `json:"alert_on_exit"`
	AlertOnEntry  *bool        `json:"alert_on_entry"`
}

// UpdateGeofenceRequest DTO untuk update geofence
type UpdateGeofenceRequest struct {
	CompanyID     *uint         `json:"company_id"`
	SetPublic     *bool         `json:"set_public"`
	Name          string        `json:"name" example:"Zona Pelabuhan B"`
	Description   string        `json:"description" example:"Area pelabuhan tambahan"`
	ZoneType      string        `json:"zone_type" example:"restricted"`
	Points        *model.Points `json:"points" binding:"omitempty,min=2"`
	IsActive      *bool         `json:"is_active"`
	RadiusMiles   *float64      `json:"radius_miles" binding:"omitempty,gt=0"`
	MaxSpeedKnots *float64      `json:"max_speed_knots"`
	AlertOnExit   *bool         `json:"alert_on_exit"`
	AlertOnEntry  *bool         `json:"alert_on_entry"`
}

// GeofenceResponse DTO untuk response
type GeofenceResponse struct {
	ID            uint         `json:"id" example:"1"`
	CompanyID     *uint        `json:"company_id,omitempty"`
	CompanyName   string       `json:"company_name,omitempty"`
	Name          string       `json:"name" example:"Zona Pelabuhan A"`
	Description   string       `json:"description" example:"Area pelabuhan utama"`
	ZoneType      string       `json:"zone_type" example:"work_zone"`
	Points        model.Points `json:"points"`
	IsActive      bool         `json:"is_active" example:"true"`
	RadiusMiles   float64      `json:"radius_miles" example:"0.5"`
	MaxSpeedKnots *float64     `json:"max_speed_knots,omitempty"`
	AlertOnExit   bool         `json:"alert_on_exit"`
	AlertOnEntry  bool         `json:"alert_on_entry"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}
