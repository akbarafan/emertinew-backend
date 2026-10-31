package geo

import "time"

// ── Assignment DTOs ────────────────────────────────────────────────

type CreateAssignmentRequest struct {
	VesselID   uint `json:"vesselId" binding:"required"`
	GeofenceID uint `json:"geofenceId" binding:"required"`
}

type UpdateAssignmentRequest struct {
	Status *string `json:"status" binding:"omitempty,oneof=assigned in_progress completed cancelled"`
}

type AssignmentResponse struct {
	ID           uint       `json:"id"`
	VesselID     uint       `json:"vesselId"`
	VesselName   string     `json:"vesselName,omitempty"`
	GeofenceID   uint       `json:"geofenceId"`
	GeofenceName string     `json:"geofenceName,omitempty"`
	TotalCP      int        `json:"totalCp"`
	ReachedCP    int        `json:"reachedCp"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"startedAt"`
	CompletedAt  *time.Time `json:"completedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// ── CP Log DTOs ────────────────────────────────────────────────────

type CpLogResponse struct {
	ID             uint      `json:"id"`
	AssignmentID   uint      `json:"assignmentId"`
	VesselID       uint      `json:"vesselId"`
	VesselName     string    `json:"vesselName,omitempty"`
	GeofenceID     uint      `json:"geofenceId"`
	GeofenceName   string    `json:"geofenceName,omitempty"`
	PointIndex     int       `json:"pointIndex"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	DistanceMeters float64   `json:"distanceMeters"`
	ArrivedAt      time.Time `json:"arrivedAt"`
	IsSkipped      bool      `json:"isSkipped"`
	OrderReached   int       `json:"orderReached"`
}

// ── Violation Log DTOs ─────────────────────────────────────────────

type ViolationLogResponse struct {
	ID                uint      `json:"id"`
	VesselID          uint      `json:"vesselId"`
	VesselName        string    `json:"vesselName,omitempty"`
	GeofenceID        uint      `json:"geofenceId"`
	GeofenceName      string    `json:"geofenceName,omitempty"`
	Latitude          float64   `json:"latitude"`
	Longitude         float64   `json:"longitude"`
	IsOutsideGeofence bool      `json:"isOutsideGeofence"`
	Distance          float64   `json:"distance"`
	CreatedAt         time.Time `json:"createdAt"`
}

// ── Stats DTOs ─────────────────────────────────────────────────────

type CpTimeStat struct {
	FromPointIndex int     `json:"fromPointIndex"`
	ToPointIndex   int     `json:"toPointIndex"`
	DurationHours  float64 `json:"durationHours"`
}

type AssignmentDetailResponse struct {
	Assignment  AssignmentResponse `json:"assignment"`
	CpLogs      []CpLogResponse    `json:"cpLogs"`
	CpTimeStats []CpTimeStat       `json:"cpTimeStats"`
}
