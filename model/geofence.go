package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Point represents a geographical coordinate
type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Points adalah array of Point untuk disimpan di PostgreSQL sebagai JSONB
type Points []Point

// Scan implements sql.Scanner interface
func (p *Points) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSONB value")
	}

	return json.Unmarshal(bytes, p)
}

// Value implements driver.Valuer interface
func (p Points) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

// Geofence model untuk menyimpan zona geografis
type MGeofence struct {
	ID            uint      `gorm:"primaryKey"`
	CompanyID     *uint     `gorm:"index" json:"company_id"`
	Company       *MCompany `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"company,omitempty"`
	Name          string    `gorm:"type:varchar(255);not null;unique" json:"name"`
	Description   string    `gorm:"type:text" json:"description"`
	ZoneType      string    `gorm:"column:zone_type;type:varchar(50);not null;default:'work_zone'" json:"zoneType"`
	Points        Points    `gorm:"type:jsonb;not null" json:"points"` // Array of coordinates
	IsActive      bool      `gorm:"column:is_active;default:true" json:"isActive"`
	RadiusMiles   float64   `gorm:"column:radius_miles;not null" json:"radiusMiles"` // Radius dalam miles dari garis geofence
	MaxSpeedKnots *float64  `gorm:"column:max_speed_knots" json:"maxSpeedKnots,omitempty"`
	AlertOnExit   bool      `gorm:"column:alert_on_exit;default:true" json:"alertOnExit"`
	AlertOnEntry  bool      `gorm:"column:alert_on_entry;default:false" json:"alertOnEntry"`

	BaseModel
}

func (MGeofence) TableName() string {
	return "m_geofences"
}

// GeofenceLog model untuk menyimpan log tracking vessel vs geofence
type GeofenceLog struct {
	ID                uint       `gorm:"primaryKey"`
	VesselID          uint       `gorm:"column:vessel_id;not null;index" json:"vesselId"`
	Vessel            *MVessel   `gorm:"foreignKey:VesselID" json:"vessel,omitempty"`
	GeofenceID        uint       `gorm:"column:geofence_id;not null;index" json:"geofenceId"`
	Geofence          *MGeofence `gorm:"foreignKey:GeofenceID" json:"geofence,omitempty"`
	EventType         string     `gorm:"column:event_type;type:varchar(50);not null;default:'outside'" json:"eventType"`
	Longitude         float64    `gorm:"column:longitude;not null" json:"longitude"`
	Latitude          float64    `gorm:"column:latitude;not null" json:"latitude"`
	Speed             float64    `gorm:"column:speed;not null;default:0" json:"speed"`
	IsOutsideGeofence bool       `gorm:"column:is_outside_geofence;not null" json:"isOutsideGeofence"`
	Distance          float64    `gorm:"column:distance;not null" json:"distance"` // Distance in meters (positive if outside, negative if inside)
	Message           string     `gorm:"column:message;type:text" json:"message,omitempty"`

	BaseModel
}

func (GeofenceLog) TableName() string {
	return "geofence_logs"
}

// MGeofenceRoute model untuk menyimpan rute geofence koridor pelayaran
type MGeofenceRoute struct {
	ID                     uint                     `gorm:"primaryKey" json:"id"`
	CompanyID              *uint                    `gorm:"index" json:"company_id"`
	Company                *MCompany                `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"company,omitempty"`
	VoyageID               *uint                    `gorm:"index" json:"voyage_id,omitempty"`
	Voyage                 *Voyage                  `gorm:"foreignKey:VoyageID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"voyage,omitempty"`
	VesselID               *uint                    `gorm:"index" json:"vessel_id,omitempty"`
	Vessel                 *MVessel                 `gorm:"foreignKey:VesselID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"vessel,omitempty"`
	Name                   string                   `gorm:"type:varchar(255);not null" json:"name"`
	Description            string                   `gorm:"type:text" json:"description"`
	TotalDistanceNM        float64                  `gorm:"column:total_distance_nm;default:0" json:"totalDistanceNm"`
	EstimatedDurationHours *float64                 `gorm:"column:estimated_duration_hours" json:"estimatedDurationHours,omitempty"`
	DefaultCorridorWidthM  float64                  `gorm:"column:default_corridor_width_m;default:926" json:"defaultCorridorWidthM"` // in meters (0.5 NM = 926m)
	Status                 string                   `gorm:"type:varchar(50);not null;default:'draft'" json:"status"`                   // draft, active, archived
	IsActive               bool                     `gorm:"column:is_active;default:true" json:"isActive"`
	AlertOnDeviation       bool                     `gorm:"column:alert_on_deviation;default:true" json:"alertOnDeviation"`
	Waypoints              []MGeofenceRouteWaypoint `gorm:"foreignKey:RouteID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"waypoints"`

	BaseModel
}

func (MGeofenceRoute) TableName() string {
	return "m_geofence_routes"
}

// MGeofenceRouteWaypoint model untuk menyimpan waypoint titik dan toleransi koridor per segmen
type MGeofenceRouteWaypoint struct {
	ID             uint    `gorm:"primaryKey" json:"id"`
	RouteID        uint    `gorm:"not null;index" json:"routeId"`
	Sequence       int     `gorm:"not null" json:"sequence"` // 1, 2, 3...
	Name           string  `gorm:"type:varchar(255)" json:"name"`
	Latitude       float64 `gorm:"type:decimal(10,6);not null" json:"latitude"`
	Longitude      float64 `gorm:"type:decimal(10,6);not null" json:"longitude"`
	WaypointType   string  `gorm:"type:varchar(50);not null;default:'waypoint'" json:"waypointType"` // 'port_call', 'waypoint', 'checkpoint'
	CorridorWidthM float64 `gorm:"column:corridor_width_m;default:926" json:"corridorWidthM"`         // toleransi koridor ke waypoint selanjutnya
	RadiusM        float64 `gorm:"column:radius_m;default:185" json:"radiusM"`                         // trigger radius saat kapal mendekat
	Notes          string  `gorm:"type:text" json:"notes,omitempty"`

	BaseModel
}

func (MGeofenceRouteWaypoint) TableName() string {
	return "m_geofence_route_waypoints"
}
