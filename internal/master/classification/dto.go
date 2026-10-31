package classification

import "time"

type CreateClassificationRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateClassificationRequest struct {
	Name string `json:"name" binding:"required"`
}

type ClassificationResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ClassificationAuditData represents audit trail data for classification changes
type ClassificationAuditData struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
