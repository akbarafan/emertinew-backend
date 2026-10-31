package model

import "time"

// VesselGeofenceAssignment assigns a vessel to follow a specific geofence route
type VesselGeofenceAssignment struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	VesselID    uint      `gorm:"column:vessel_id;not null;index;uniqueIndex:idx_vessel_geofence_unique" json:"vesselId"`
	GeofenceID  uint      `gorm:"column:geofence_id;not null;index;uniqueIndex:idx_vessel_geofence_unique" json:"geofenceId"`
	Status      string    `gorm:"column:status;type:varchar(20);not null;default:'assigned'" json:"status"`
	StartedAt   *time.Time `gorm:"column:started_at" json:"startedAt"`
	CompletedAt *time.Time `gorm:"column:completed_at" json:"completedAt"`

	BaseModel
}

// VesselCpLog records when a vessel reaches a checkpoint on a geofence route
type VesselCpLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	AssignmentID uint      `gorm:"column:assignment_id;not null;index" json:"assignmentId"`
	VesselID     uint      `gorm:"column:vessel_id;not null;index" json:"vesselId"`
	GeofenceID   uint      `gorm:"column:geofence_id;not null;index" json:"geofenceId"`
	PointIndex   int       `gorm:"column:point_index;not null" json:"pointIndex"`
	Latitude     float64   `gorm:"column:latitude;not null" json:"latitude"`
	Longitude    float64   `gorm:"column:longitude;not null" json:"longitude"`
	DistanceMeters float64 `gorm:"column:distance_meters;not null" json:"distanceMeters"`
	ArrivedAt    time.Time `gorm:"column:arrived_at;not null" json:"arrivedAt"`
	IsSkipped    bool      `gorm:"column:is_skipped;not null;default:false" json:"isSkipped"`
	OrderReached int       `gorm:"column:order_reached;not null;default:0" json:"orderReached"`

	BaseModel
}
