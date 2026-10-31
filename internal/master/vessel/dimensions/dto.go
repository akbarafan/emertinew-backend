package vesseldimensions

import "github.com/shopspring/decimal"

// CreateVesselDimensionsRequest untuk request payload create vessel dimensions
type CreateVesselDimensionsRequest struct {
	LOA          *float64         `json:"loa" example:"150.5"`
	LBP          *float64         `json:"lbp" example:"140.0"`
	Beam         *float64         `json:"beam" example:"25.5"`
	Depth        *float64         `json:"depth" example:"12.0"`
	Draft        *float64         `json:"draft" example:"8.5"`
	AirDraft     *float64         `json:"air_draft" example:"50.0"`
	GT           *decimal.Decimal `json:"gt" example:"5000.00"`
	NT           *decimal.Decimal `json:"nt" example:"3000.00"`
	DWT          *decimal.Decimal `json:"dwt" example:"10000.00"`
	Displacement *decimal.Decimal `json:"displacement" example:"12000.00"`
}

// UpdateVesselDimensionsRequest untuk request payload update vessel dimensions
type UpdateVesselDimensionsRequest struct {
	LOA          *float64         `json:"loa" example:"155.0"`
	LBP          *float64         `json:"lbp" example:"145.0"`
	Beam         *float64         `json:"beam" example:"26.0"`
	Depth        *float64         `json:"depth" example:"13.0"`
	Draft        *float64         `json:"draft" example:"9.0"`
	AirDraft     *float64         `json:"air_draft" example:"52.0"`
	GT           *decimal.Decimal `json:"gt" example:"5200.00"`
	NT           *decimal.Decimal `json:"nt" example:"3100.00"`
	DWT          *decimal.Decimal `json:"dwt" example:"10500.00"`
	Displacement *decimal.Decimal `json:"displacement" example:"12500.00"`
}

// VesselDimensionsResponse untuk response data vessel dimensions
type VesselDimensionsResponse struct {
	ID           uint     `json:"id" example:"1"`
	VesselID     uint     `json:"vessel_id" example:"1"`
	LOA          *float64 `json:"loa,omitempty" example:"150.5"`
	LBP          *float64 `json:"lbp,omitempty" example:"140.0"`
	Beam         *float64 `json:"beam,omitempty" example:"25.5"`
	Depth        *float64 `json:"depth,omitempty" example:"12.0"`
	Draft        *float64 `json:"draft,omitempty" example:"8.5"`
	AirDraft     *float64 `json:"air_draft,omitempty" example:"50.0"`
	GT           *string  `json:"gt,omitempty" example:"5000.00"`
	NT           *string  `json:"nt,omitempty" example:"3000.00"`
	DWT          *string  `json:"dwt,omitempty" example:"10000.00"`
	Displacement *string  `json:"displacement,omitempty" example:"12000.00"`
	CreatedAt    string   `json:"created_at" example:"2024-01-01T10:00:00Z"`
	UpdatedAt    string   `json:"updated_at" example:"2024-01-01T10:00:00Z"`
}

// VesselDimensionsAuditData for audit logging
type VesselDimensionsAuditData struct {
	ID           uint     `json:"id"`
	VesselID     uint     `json:"vessel_id"`
	LOA          *float64 `json:"loa,omitempty"`
	LBP          *float64 `json:"lbp,omitempty"`
	Beam         *float64 `json:"beam,omitempty"`
	Depth        *float64 `json:"depth,omitempty"`
	Draft        *float64 `json:"draft,omitempty"`
	AirDraft     *float64 `json:"air_draft,omitempty"`
	GT           *string  `json:"gt,omitempty"`
	NT           *string  `json:"nt,omitempty"`
	DWT          *string  `json:"dwt,omitempty"`
	Displacement *string  `json:"displacement,omitempty"`
}
