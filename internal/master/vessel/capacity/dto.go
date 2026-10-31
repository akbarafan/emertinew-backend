package vesselcapacity

import "github.com/shopspring/decimal"

// CreateVesselCapacityRequest untuk request payload create vessel capacity
type CreateVesselCapacityRequest struct {
	LoadCapacity    *decimal.Decimal `json:"load_capacity" example:"5000.00"`
	LoadUnit        *string          `json:"load_unit" example:"ton"`
	DeckArea        *decimal.Decimal `json:"deck_area" example:"1500.50"`
	CargoHoldVolume *decimal.Decimal `json:"cargo_hold_volume" example:"8000.00"`
	HasCrane        *bool            `json:"has_crane" example:"true"`
	SWLCrane        *decimal.Decimal `json:"swl_crane" example:"50.00"`
	BollardPull     *decimal.Decimal `json:"bollard_pull" example:"75.00"`
}

// UpdateVesselCapacityRequest untuk request payload update vessel capacity
type UpdateVesselCapacityRequest struct {
	LoadCapacity    *decimal.Decimal `json:"load_capacity" example:"5200.00"`
	LoadUnit        *string          `json:"load_unit" example:"ton"`
	DeckArea        *decimal.Decimal `json:"deck_area" example:"1600.00"`
	CargoHoldVolume *decimal.Decimal `json:"cargo_hold_volume" example:"8500.00"`
	HasCrane        *bool            `json:"has_crane" example:"false"`
	SWLCrane        *decimal.Decimal `json:"swl_crane" example:"55.00"`
	BollardPull     *decimal.Decimal `json:"bollard_pull" example:"80.00"`
}

// VesselCapacityResponse untuk response data vessel capacity
type VesselCapacityResponse struct {
	ID              uint    `json:"id" example:"1"`
	VesselID        uint    `json:"vessel_id" example:"1"`
	LoadCapacity    *string `json:"load_capacity,omitempty" example:"5000.00"`
	LoadUnit        *string `json:"load_unit,omitempty" example:"ton"`
	DeckArea        *string `json:"deck_area,omitempty" example:"1500.50"`
	CargoHoldVolume *string `json:"cargo_hold_volume,omitempty" example:"8000.00"`
	HasCrane        *bool   `json:"has_crane,omitempty" example:"true"`
	SWLCrane        *string `json:"swl_crane,omitempty" example:"50.00"`
	BollardPull     *string `json:"bollard_pull,omitempty" example:"75.00"`
	CreatedAt       string  `json:"created_at" example:"2024-01-01T10:00:00Z"`
	UpdatedAt       string  `json:"updated_at" example:"2024-01-01T10:00:00Z"`
}

// VesselCapacityAuditData for audit logging
type VesselCapacityAuditData struct {
	ID              uint    `json:"id"`
	VesselID        uint    `json:"vessel_id"`
	LoadCapacity    *string `json:"load_capacity,omitempty"`
	LoadUnit        *string `json:"load_unit,omitempty"`
	DeckArea        *string `json:"deck_area,omitempty"`
	CargoHoldVolume *string `json:"cargo_hold_volume,omitempty"`
	HasCrane        *bool   `json:"has_crane,omitempty"`
	SWLCrane        *string `json:"swl_crane,omitempty"`
	BollardPull     *string `json:"bollard_pull,omitempty"`
}
