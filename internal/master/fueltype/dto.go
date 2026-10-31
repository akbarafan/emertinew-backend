package fueltype

// CreateFuelTypeRequest untuk request payload create fuel type.
//
// The parameters are the raw ones from the IMO references, so a figure can be
// traced back to the standard it came from. See `model.MFuelType` for the
// formulas the per-litre factors are derived from.
type CreateFuelTypeRequest struct {
	Code string `json:"code" binding:"required" example:"MGO"`
	Name string `json:"name" binding:"required" example:"Marine Gas Oil (MDO/MGO)"`

	// Carbon factors, tank-to-wake. IMO DCS / MEPC.364(79).
	CfCO2KgPerKg *float64 `json:"cf_co2_kg_per_kg" example:"3.206"`
	CfCH4KgPerKg *float64 `json:"cf_ch4_kg_per_kg" example:"0.00005"`
	CfN2OKgPerKg *float64 `json:"cf_n2o_kg_per_kg" example:"0.00018"`
	// Unburned methane escaping the engine, in mass % of the fuel.
	MethaneSlipPc *float64 `json:"methane_slip_percent" example:"3.5"`

	// Well-to-tank intensity and heating value, MEPC.391(81) Appendix 2.
	WttIntensityGCO2eqPerMJ *float64 `json:"wtt_intensity_gco2eq_per_mj" example:"17.7"`
	LhvMJPerKg              *float64 `json:"lhv_mj_per_kg" example:"42.7"`

	// Density, kg per litre. Fourth IMO GHG Study 2020, Table 10.
	DensityKgPerLitre *float64 `json:"density_kg_per_litre" example:"0.895"`

	Reference string `json:"reference" example:"MEPC.364(79) Annex 9"`
	Method    string `json:"method" example:"IMO DCS / IMO LCA 2024"`
}

// UpdateFuelTypeRequest untuk request payload update fuel type
type UpdateFuelTypeRequest struct {
	Name string `json:"name" example:"Marine Gas Oil (MDO/MGO)"`

	CfCO2KgPerKg            *float64 `json:"cf_co2_kg_per_kg" example:"3.206"`
	CfCH4KgPerKg            *float64 `json:"cf_ch4_kg_per_kg" example:"0.00005"`
	CfN2OKgPerKg            *float64 `json:"cf_n2o_kg_per_kg" example:"0.00018"`
	MethaneSlipPc           *float64 `json:"methane_slip_percent" example:"3.5"`
	WttIntensityGCO2eqPerMJ *float64 `json:"wtt_intensity_gco2eq_per_mj" example:"17.7"`
	LhvMJPerKg              *float64 `json:"lhv_mj_per_kg" example:"42.7"`
	DensityKgPerLitre       *float64 `json:"density_kg_per_litre" example:"0.895"`

	Reference string `json:"reference" example:"MEPC.364(79) Annex 9"`
	Method    string `json:"method" example:"IMO DCS / IMO LCA 2024"`
	IsActive  *bool  `json:"is_active" example:"true"`
}

// FuelTypeResponse untuk response data fuel type.
//
// The derived per-litre factors are included so a client never has to repeat
// the formula — and therefore never has to disagree with the backend about
// it. A factor that cannot be derived is reported as null.
type FuelTypeResponse struct {
	ID   uint   `json:"id" example:"1"`
	Code string `json:"code" example:"MGO"`
	Name string `json:"name" example:"Marine Gas Oil (MDO/MGO)"`

	CfCO2KgPerKg            *float64 `json:"cf_co2_kg_per_kg" example:"3.206"`
	CfCH4KgPerKg            *float64 `json:"cf_ch4_kg_per_kg" example:"0.00005"`
	CfN2OKgPerKg            *float64 `json:"cf_n2o_kg_per_kg" example:"0.00018"`
	MethaneSlipPc           *float64 `json:"methane_slip_percent" example:"3.5"`
	WttIntensityGCO2eqPerMJ *float64 `json:"wtt_intensity_gco2eq_per_mj" example:"17.7"`
	LhvMJPerKg              *float64 `json:"lhv_mj_per_kg" example:"42.7"`
	DensityKgPerLitre       *float64 `json:"density_kg_per_litre" example:"0.895"`

	// Derived, per litre of fuel.
	TtWCO2KgPerLitre  *float64 `json:"ttw_co2_kg_per_litre" example:"2.869"`
	TtWCO2eKgPerLitre *float64 `json:"ttw_co2e_kg_per_litre" example:"2.894"`
	WtWCO2eKgPerLitre *float64 `json:"wtw_co2e_kg_per_litre" example:"3.59"`

	Reference string `json:"reference" example:"MEPC.364(79) Annex 9"`
	Method    string `json:"method" example:"IMO DCS / IMO LCA 2024"`
	IsActive  bool   `json:"is_active" example:"true"`

	CreatedAt string `json:"created_at" example:"2024-01-01T10:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2024-01-01T10:00:00Z"`
}

// FuelTypeAuditData untuk audit log fuel type
type FuelTypeAuditData struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`

	CfCO2KgPerKg            *float64 `json:"cf_co2_kg_per_kg,omitempty"`
	CfCH4KgPerKg            *float64 `json:"cf_ch4_kg_per_kg,omitempty"`
	CfN2OKgPerKg            *float64 `json:"cf_n2o_kg_per_kg,omitempty"`
	MethaneSlipPc           *float64 `json:"methane_slip_percent,omitempty"`
	WttIntensityGCO2eqPerMJ *float64 `json:"wtt_intensity_gco2eq_per_mj,omitempty"`
	LhvMJPerKg              *float64 `json:"lhv_mj_per_kg,omitempty"`
	DensityKgPerLitre       *float64 `json:"density_kg_per_litre,omitempty"`

	Reference string `json:"reference,omitempty"`
	Method    string `json:"method,omitempty"`
	IsActive  bool   `json:"is_active"`
}
