package squad

import "time"

type CreateSquadRequest struct {
	CompanyID *uint  `json:"company_id" example:"1"`
	FleetID   *uint  `json:"fleet_id" binding:"required" example:"1"`
	Code      string `json:"code" binding:"required" example:"SQD001"`
	Name      string `json:"name" binding:"required" example:"Squad Delta"`
}

type UpdateSquadRequest struct {
	CompanyID *uint  `json:"company_id" example:"1"`
	FleetID   *uint  `json:"fleet_id" binding:"required" example:"1"`
	Code      string `json:"code" binding:"required" example:"SQD001"`
	Name      string `json:"name" binding:"required" example:"Squad Delta Updated"`
}

type SquadResponse struct {
	ID        uint      `json:"id"`
	CompanyID *uint     `json:"company_id,omitempty"`
	Company   *string   `json:"company,omitempty"`
	FleetID   *uint     `json:"fleet_id,omitempty"`
	Fleet     *string   `json:"fleet,omitempty"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SquadAuditData struct {
	ID        uint   `json:"id"`
	CompanyID *uint  `json:"company_id,omitempty"`
	FleetID   *uint  `json:"fleet_id,omitempty"`
	Code      string `json:"code"`
	Name      string `json:"name"`
}
