package company

import "time"

type CreateCompanyRequest struct {
	Code string `json:"code" binding:"required" example:"hcml"`
	Name string `json:"name" binding:"required" example:"HCML Shipping"`
}

type UpdateCompanyRequest struct {
	Code string `json:"code" binding:"required" example:"hcml"`
	Name string `json:"name" binding:"required" example:"HCML Shipping"`
}

type CompanyResponse struct {
	ID        uint      `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CompanyAuditData represents audit trail data for company changes
type CompanyAuditData struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}
