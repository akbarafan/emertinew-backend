package operationstatus

import "time"

type CreateOperationStatusRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateOperationStatusRequest struct {
	Name string `json:"name" binding:"required"`
}

type OperationStatusResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OperationStatusAuditData represents audit trail data for operation status changes
type OperationStatusAuditData struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
