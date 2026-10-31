package vessel

import (
	vesselcapacity "github.com/Alugara-Inovasi-Utama/emerti/internal/master/vessel/capacity"
	vesseldimensions "github.com/Alugara-Inovasi-Utama/emerti/internal/master/vessel/dimensions"
)

// CreateVesselRequest untuk request payload create vessel
type CreateVesselRequest struct {
	MCUIDs         []uint  `json:"mcu_id" example:"1,2"`
	Code           string  `json:"code" binding:"required" example:"v0001"`
	Name           string  `json:"name" binding:"required" example:"KM Nusantara 01"`
	IsActive       *bool   `json:"is_active" example:"true"`
	CompanyID      *uint   `json:"company_id" example:"1"`
	VesselTypeID   *uint   `json:"vessel_type_id" example:"1"`
	ClassID        *uint   `json:"classification_id" example:"1"`
	MaterialID     *uint   `json:"material_hull_id" example:"1"`
	OperationID    *uint   `json:"operation_status_id" example:"1"`
	FleetID        *uint   `json:"fleet_id" example:"1"`
	SquadID        *uint   `json:"squad_id" example:"1"`
	FuelTypeID     *uint   `json:"fuel_type_id" example:"1"`
	CallSign       *string `json:"call_sign" example:"YCBA"`
	IMONumber      *string `json:"imo_number" example:"1234567"`
	MMSINumber     *string `json:"mmsi_number" example:"525012345"`
	Flag           *string `json:"flag" example:"Indonesia"`
	PortOfRegistry *string `json:"port_of_registry" example:"Jakarta"`
	ProductionYear *int    `json:"production_year" example:"2015"`
	Builder        *string `json:"builder" example:"PT. Dumas Tanjung Perak Shipyard"`
	HullNumber     *string `json:"hull_number" example:"HN-001"`
	ClassNotation  *string `json:"class_notation" example:"BKI Class"`
	Nickname       *string `json:"nickname" example:"Hawk"`
	TrackingSource *string `json:"tracking_source" example:"alugara" enums:"alugara,hexa"`
}

// UpdateVesselRequest untuk request payload update vessel
type UpdateVesselRequest struct {
	MCUIDs         []uint  `json:"mcu_id" example:"1,2,3"`
	Code           string  `json:"code" example:"v0001"`
	Name           string  `json:"name" example:"KM Nusantara 01 Updated"`
	IsActive       *bool   `json:"is_active" example:"false"`
	CompanyID      *uint   `json:"company_id" example:"1"`
	VesselTypeID   *uint   `json:"vessel_type_id" example:"2"`
	ClassID        *uint   `json:"class_id" example:"2"`
	MaterialID     *uint   `json:"material_id" example:"2"`
	OperationID    *uint   `json:"operation_id" example:"2"`
	FleetID        *uint   `json:"fleet_id" example:"1"`
	SquadID        *uint   `json:"squad_id" example:"1"`
	FuelTypeID     *uint   `json:"fuel_type_id" example:"2"`
	CallSign       *string `json:"call_sign" example:"YCBB"`
	IMONumber      *string `json:"imo_number" example:"7654321"`
	MMSINumber     *string `json:"mmsi_number" example:"525054321"`
	Flag           *string `json:"flag" example:"Singapore"`
	PortOfRegistry *string `json:"port_of_registry" example:"Surabaya"`
	ProductionYear *int    `json:"production_year" example:"2020"`
	Builder        *string `json:"builder" example:"PT. PAL Indonesia"`
	HullNumber     *string `json:"hull_number" example:"HN-002"`
	ClassNotation  *string `json:"class_notation" example:"ABS Class"`
	Nickname       *string `json:"nickname" example:"Eagle"`
	TrackingSource *string `json:"tracking_source" example:"hexa" enums:"alugara,hexa"`
}

// MCUInfo untuk info mcu dalam response
type MCUInfo struct {
	ID       uint   `json:"id" example:"1"`
	Name     string `json:"name" example:"GPS Module A1"`
	Location string `json:"location" example:"bd"`
}

// LookupInfo untuk info lookup tables
type LookupInfo struct {
	ID   uint   `json:"id" example:"1"`
	Name string `json:"name" example:"BKI"`
}

// FuelTypeInfo untuk info fuel type dalam response.
//
// The derived per-litre factors travel with it so the client never has to
// repeat the IMO formula, and therefore never has to disagree with the
// backend about it. A factor that cannot be derived stays null.
//
// The raw IMO parameters travel with them as well. The command center shows
// how a factor is built - "2.869 kg CO2/L = 3.206 kg CO2/kg (carbon factor)
// x 0.895 kg/L (density)" - and an operator who cannot see where the number
// comes from cannot check it. The backend still derives the factor; the client
// only displays the derivation.
type FuelTypeInfo struct {
	ID   uint   `json:"id" example:"1"`
	Code string `json:"code" example:"MGO"`
	Name string `json:"name" example:"Marine Gas Oil (MDO/MGO)"`

	// kg CO2 per litre of fuel, tank-to-wake.
	TtWCO2KgPerLitre *float64 `json:"ttw_co2_kg_per_litre" example:"2.869"`
	// kg CO2e per litre of fuel, tank-to-wake.
	TtWCO2eKgPerLitre *float64 `json:"ttw_co2e_kg_per_litre" example:"2.894"`
	// kg CO2e per litre of fuel, well-to-wake. Null when the IMO has no
	// well-to-tank default for this fuel.
	WtWCO2eKgPerLitre *float64 `json:"wtw_co2e_kg_per_litre" example:"3.59"`

	// The raw IMO parameters the factors are derived from.
	// IMO DCS carbon factor, kg CO2 per kg of fuel (MEPC.364(79) Annex 9).
	CfCO2KgPerKg *float64 `json:"cf_co2_kg_per_kg" example:"3.206"`
	// kg CH4 per kg of fuel (MEPC.391(81) Appendix 2).
	CfCH4KgPerKg *float64 `json:"cf_ch4_kg_per_kg" example:"0.00005"`
	// kg N2O per kg of fuel (MEPC.391(81) Appendix 2).
	CfN2OKgPerKg *float64 `json:"cf_n2o_kg_per_kg" example:"0.00018"`
	// Unburned methane escaping the engine, in mass % of the fuel.
	MethaneSlipPercent *float64 `json:"methane_slip_percent" example:"0.5"`
	// Well-to-tank intensity, gCO2eq per MJ (MEPC.391(81) Appendix 2).
	WttIntensityGCO2eqPerMJ *float64 `json:"wtt_intensity_gco2eq_per_mj" example:"17.7"`
	// Lower heating value, MJ per kg (MEPC.364(79) Annex 9).
	LhvMJPerKg *float64 `json:"lhv_mj_per_kg" example:"42.7"`
	// Density, kg per litre (Fourth IMO GHG Study 2020, Table 10).
	DensityKgPerLitre *float64 `json:"density_kg_per_litre" example:"0.895"`

	Reference string `json:"reference,omitempty" example:"MEPC.364(79) Annex 9"`
	Method    string `json:"method,omitempty" example:"IMO DCS / IMO LCA 2024"`
}

// VesselResponse untuk response data vessel
type VesselResponse struct {
	ID              uint                                       `json:"id" example:"1"`
	Code            string                                     `json:"code" example:"v0001"`
	Name            string                                     `json:"name" example:"KM Nusantara 01"`
	IsActive        bool                                       `json:"is_active" example:"true"`
	TrackingSource  *string                                    `json:"tracking_source" example:"alugara" enums:"alugara,hexa"`
	Company         *LookupInfo                                `json:"company"`
	VesselType      *LookupInfo                                `json:"vessel_type"`
	Classification  *LookupInfo                                `json:"classification"`
	MaterialHull    *LookupInfo                                `json:"material_hull"`
	OperationStatus *LookupInfo                                `json:"operation_status"`
	Fleet           *LookupInfo                                `json:"fleet"`
	Squad           *LookupInfo                                `json:"squad"`
	FuelType        *FuelTypeInfo                              `json:"fuel_type"`
	CallSign        *string                                    `json:"call_sign" example:"YCBA"`
	IMONumber       *string                                    `json:"imo_number" example:"1234567"`
	MMSINumber      *string                                    `json:"mmsi_number" example:"525012345"`
	Flag            *string                                    `json:"flag" example:"Indonesia"`
	PortOfRegistry  *string                                    `json:"port_of_registry" example:"Jakarta"`
	ProductionYear  *int                                       `json:"production_year" example:"2015"`
	Builder         *string                                    `json:"builder" example:"PT. Dumas Tanjung Perak Shipyard"`
	HullNumber      *string                                    `json:"hull_number" example:"HN-001"`
	ClassNotation   *string                                    `json:"class_notation" example:"BKI Class"`
	Nickname        *string                                    `json:"nickname" example:"Hawk"`
	MCUs            []MCUInfo                                  `json:"mcu"`
	Dimensions      *vesseldimensions.VesselDimensionsResponse `json:"dimensions"`
	Capacity        *vesselcapacity.VesselCapacityResponse     `json:"capacity"`
	CreatedAt       string                                     `json:"created_at" example:"2024-01-01T10:00:00Z"`
	UpdatedAt       string                                     `json:"updated_at" example:"2024-01-01T10:00:00Z"`
}

// VesselAuditData for clean audit logging (all fields, no nested objects)
type VesselAuditData struct {
	ID             uint    `json:"id"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	IsActive       bool    `json:"is_active"`
	CompanyID      *uint   `json:"company_id,omitempty"`
	VesselTypeID   *uint   `json:"vessel_type_id,omitempty"`
	ClassID        *uint   `json:"classification_id,omitempty"`
	MaterialID     *uint   `json:"material_hull_id,omitempty"`
	OperationID    *uint   `json:"operation_status_id,omitempty"`
	FleetID        *uint   `json:"fleet_id,omitempty"`
	SquadID        *uint   `json:"squad_id,omitempty"`
	FuelTypeID     *uint   `json:"fuel_type_id,omitempty"`
	CallSign       *string `json:"call_sign,omitempty"`
	IMONumber      *string `json:"imo_number,omitempty"`
	MMSINumber     *string `json:"mmsi_number,omitempty"`
	Flag           *string `json:"flag,omitempty"`
	PortOfRegistry *string `json:"port_of_registry,omitempty"`
	ProductionYear *int    `json:"production_year,omitempty"`
	Builder        *string `json:"builder,omitempty"`
	HullNumber     *string `json:"hull_number,omitempty"`
	ClassNotation  *string `json:"class_notation,omitempty"`
	Nickname       *string `json:"nickname,omitempty"`
	TrackingSource *string `json:"tracking_source,omitempty"`
	MCUIDs         []uint  `json:"mcu_ids,omitempty"`
}
