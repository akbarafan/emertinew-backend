package vesseltype

import "time"

type CreateVesselTypeRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateVesselTypeRequest struct {
	Name string `json:"name" binding:"required"`
}

type VesselTypeResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VesselTypeAuditData represents audit trail data for vessel type changes
type VesselTypeAuditData struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
