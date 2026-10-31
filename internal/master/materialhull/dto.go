package materialhull

import "time"

type CreateMaterialHullRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateMaterialHullRequest struct {
	Name string `json:"name" binding:"required"`
}

type MaterialHullResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MaterialHullAuditData represents audit trail data for material hull changes
type MaterialHullAuditData struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
