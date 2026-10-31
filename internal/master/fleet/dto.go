package fleet

import "time"

type CreateFleetRequest struct {
	CompanyID *uint  `json:"company_id" example:"1"`
	Code      string `json:"code" binding:"required" example:"FLT001"`
	Name      string `json:"name" binding:"required" example:"Fleet Alpha"`
}

type UpdateFleetRequest struct {
	CompanyID *uint  `json:"company_id" example:"1"`
	Code      string `json:"code" binding:"required" example:"FLT001"`
	Name      string `json:"name" binding:"required" example:"Fleet Alpha Updated"`
}

type FleetResponse struct {
	ID        uint      `json:"id"`
	CompanyID *uint     `json:"company_id,omitempty"`
	Company   *string   `json:"company,omitempty"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FleetAuditData struct {
	ID        uint   `json:"id"`
	CompanyID *uint  `json:"company_id,omitempty"`
	Code      string `json:"code"`
	Name      string `json:"name"`
}
