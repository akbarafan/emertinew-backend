package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// MCompany represents a company in the system
// Used for multi-company support and topic routing
type MCompany struct {
	ID      uint      `gorm:"primaryKey"`
	Code    string    `gorm:"type:varchar(50);uniqueIndex;not null"` // e.g., "hcml" - used in MQTT topic
	Name    string    `gorm:"type:varchar(255);not null"`
	Vessels []MVessel `gorm:"foreignKey:CompanyID"`

	BaseModel
}

func (MCompany) TableName() string {
	return "m_company"
}

type MClassification struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(255);not null;uniqueIndex"`

	BaseModel
}

func (MClassification) TableName() string {
	return "m_classification"
}

type MVesselType struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(255);not null;uniqueIndex"`

	BaseModel
}

func (MVesselType) TableName() string {
	return "m_vessel_type"
}

type MOperationStatus struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(255);not null;uniqueIndex"`

	BaseModel
}

func (MOperationStatus) TableName() string {
	return "m_operation_status"
}

type MMaterialHull struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(255);not null;uniqueIndex"`

	BaseModel
}

func (MMaterialHull) TableName() string {
	return "m_material_hull"
}

// MFleet represents a fleet/group of vessels
type MFleet struct {
	ID        uint     `gorm:"primaryKey"`
	CompanyID *uint    `gorm:"index"`
	Company   MCompany `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Code      string   `gorm:"type:varchar(50);uniqueIndex;not null"`
	Name      string   `gorm:"type:varchar(255);not null"`

	BaseModel
}

func (MFleet) TableName() string {
	return "m_fleets"
}

// MSquad represents a squadron within a fleet
type MSquad struct {
	ID        uint     `gorm:"primaryKey"`
	CompanyID *uint    `gorm:"index"`
	Company   MCompany `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	FleetID   *uint    `gorm:"index"`
	Fleet     MFleet   `gorm:"foreignKey:FleetID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Code      string   `gorm:"type:varchar(50);uniqueIndex;not null"`
	Name      string   `gorm:"type:varchar(255);not null"`

	BaseModel
}

func (MSquad) TableName() string {
	return "m_squads"
}

// MVessel represents a vessel/ship in the system
// Note: Vessel.Code is used for display and initial topic generation only.
// The actual data routing is determined by MCU.VesselID relationship.
type MVessel struct {
	ID              uint             `gorm:"primaryKey"`
	Code            string           `gorm:"type:varchar(50);uniqueIndex;not null"` // Display identifier, e.g., "VS001"
	Name            string           `gorm:"type:varchar(255);not null"`
	IsActive        bool             `gorm:"default:true"`
	CompanyID       *uint            `gorm:"index"`
	Company         MCompany         `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	VesselTypeID    *uint            `gorm:"index"`
	VesselType      MVesselType      `gorm:"foreignKey:VesselTypeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	ClassID         *uint            `gorm:"index"`
	Classification  MClassification  `gorm:"foreignKey:ClassID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	MaterialID      *uint            `gorm:"index"`
	MaterialHull    MMaterialHull    `gorm:"foreignKey:MaterialID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	OperationID     *uint            `gorm:"index"`
	OperationStatus MOperationStatus `gorm:"foreignKey:OperationID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	CallSign        *string          `gorm:"type:varchar(50)"`
	IMONumber       *string          `gorm:"type:varchar(50)"`
	MMSINumber      *string          `gorm:"type:varchar(50)"`
	Flag            *string          `gorm:"type:varchar(100)"`
	PortOfRegistry  *string          `gorm:"type:varchar(255)"`
	ProductionYear  *int             `gorm:"type:int"`
	Builder         *string          `gorm:"type:varchar(255)"`
	HullNumber      *string          `gorm:"type:varchar(100)"`
	ClassNotation   *string          `gorm:"type:varchar(255)"`
	Nickname        *string          `gorm:"type:varchar(255)"`
	FleetID         *uint            `gorm:"index"`
	Fleet           MFleet           `gorm:"foreignKey:FleetID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	SquadID         *uint            `gorm:"index"`
	Squad           MSquad           `gorm:"foreignKey:SquadID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	TrackingSource  *string          `gorm:"type:varchar(10)"`
	// The fuel the vessel runs on. NULL when it has not been declared, in which
	// case no emission figure can be derived for it.
	FuelTypeID *uint     `gorm:"index"`
	FuelType   MFuelType `gorm:"foreignKey:FuelTypeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	MCUs       []MMCU    `gorm:"foreignKey:VesselID"`

	Dimensions VesselDimensions `gorm:"foreignKey:VesselID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Capacity   VesselCapacity   `gorm:"foreignKey:VesselID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	BaseModel
}

type VesselDimensions struct {
	ID           uint             `gorm:"primaryKey;autoIncrement"`
	VesselID     uint             `gorm:"index"` // Changed from uniqueIndex - unique constraint handled by partial index in migration
	LOA          *float64         `gorm:"type:decimal(10,2)"`
	LBP          *float64         `gorm:"type:decimal(10,2)"`
	Beam         *float64         `gorm:"type:decimal(10,2)"`
	Depth        *float64         `gorm:"type:decimal(10,2)"`
	Draft        *float64         `gorm:"type:decimal(10,2)"`
	AirDraft     *float64         `gorm:"type:decimal(10,2)"`
	GT           *decimal.Decimal `gorm:"type:decimal(10,2)"`
	NT           *decimal.Decimal `gorm:"type:decimal(10,2)"`
	DWT          *decimal.Decimal `gorm:"type:decimal(10,2)"`
	Displacement *decimal.Decimal `gorm:"type:decimal(10,2)"`

	BaseModel
}

type VesselCapacity struct {
	ID              uint             `gorm:"primaryKey;autoIncrement"`
	VesselID        uint             `gorm:"index"` // Changed from uniqueIndex - unique constraint handled by partial index in migration
	LoadCapacity    *decimal.Decimal `gorm:"type:decimal(10,2)"`
	LoadUnit        *string          `gorm:"type:varchar(50)"`
	DeckArea        *decimal.Decimal `gorm:"type:decimal(10,2)"`
	CargoHoldVolume *decimal.Decimal `gorm:"type:decimal(10,2)"`
	HasCrane        *bool
	SWLCrane        *decimal.Decimal `gorm:"type:decimal(10,2)"`
	BollardPull     *decimal.Decimal `gorm:"type:decimal(10,2)"` // khusus tugboat

	BaseModel
}

// MMCU represents a Micro Controller Unit (panel) on a vessel
// Note: VesselID is mutable - when an MCU is moved to another vessel,
// simply update VesselID and all sensor data will automatically route to the new vessel.
// This allows hardware swaps without reflashing firmware.
type MMCU struct {
	ID       uint      `gorm:"primaryKey"`
	VesselID *uint     `gorm:"index"` // Mutable: logical vessel assignment (can be changed when MCU moves)
	Vessel   MVessel   `gorm:"foreignKey:VesselID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Name     string    `gorm:"type:varchar(255);not null"`
	Location string    `gorm:"type:varchar(50);not null"` // bd (bridge deck) or er (engine room)
	Sensors  []MSensor `gorm:"foreignKey:MCUID"`

	BaseModel
}

func (MMCU) TableName() string {
	return "m_mcu"
}

// MSensor represents a sensor attached to an MCU
// Note: Topic is immutable (hardcoded in firmware) and acts as physical device identifier.
// Data routing to vessel is determined by: Sensor.MCUID → MCU.VesselID → Vessel
// When MCU moves to another vessel, sensor data automatically routes to new vessel.
// TopicKey is a normalized unique key: company:vessel_code:location:sensor_group:sensor_number
// Used to prevent duplicate sensors across old (5-part) and new (7-part) topic formats.
// IsVisible controls whether this sensor appears in the daily report charts.
type MSensor struct {
	ID            uint     `gorm:"primaryKey"`
	MCUID         *uint    `gorm:"column:mcu_id"`
	MCU           MMCU     `gorm:"foreignKey:MCUID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Topic         string   `gorm:"type:varchar(255);uniqueIndex;not null"`         // Immutable: physical device identifier, e.g., hcml/vs001/top/gps
	TopicKey      string   `gorm:"type:varchar(255);uniqueIndex;column:topic_key"` // Normalized: hcml:vs001:bd:gps:1 (NOT NULL enforced by migration after backfill)
	Name          string   `gorm:"type:varchar(255)"`
	KFactor       *float64 `gorm:"column:k_factor;type:decimal(15,6);default:0"`        // Calibration factor (pulse/L). Volume = pulse / k_factor. 0 = skip.
	IsVisible     bool     `gorm:"column:is_visible;default:true"`                      // Show in daily report charts
	EngineName    *string  `gorm:"column:engine_name;type:varchar(100);default:null"`   // Engine this sensor belongs to (flow meters and tachometers)
	FlowDirection *string  `gorm:"column:flow_direction;type:varchar(10);default:null"` // "in", "out", or null for single

	BaseModel
}

type MOffshore struct {
	ID        uint     `gorm:"primaryKey"`
	CompanyID *uint    `gorm:"index" json:"company_id"`
	Company   MCompany `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"company,omitempty"`
	Name      string   `gorm:"type:varchar(255);not null" json:"name"`
	Code      string   `gorm:"type:varchar(100);uniqueIndex" json:"code"`
	Type      string   `gorm:"type:varchar(50);not null;default:field" json:"type"`
	Long      float64  `gorm:"type:decimal(10,6)" json:"long"`
	Lat       float64  `gorm:"type:decimal(10,6)" json:"lat"`

	BaseModel
}

func (MOffshore) TableName() string {
	return "m_offshore"
}

// GWP100 global warming potentials used by the 2024 IMO LCA Guidelines
// (MEPC.391(81), paragraph 2.4): CO2 = 1, CH4 = 28, N2O = 265.
const (
	GWP100Methane = 28.0
	GWP100Nitrous = 265.0
)

// MFuelType is a fuel together with the emission parameters of the IMO
// references.
//
// The raw parameters are stored, never a pre-computed factor, so a figure can
// always be traced back to the standard it came from:
//
//	cf_co2_kg_per_kg             IMO DCS carbon factor (MEPC.364(79))
//	cf_ch4_kg_per_kg, cf_n2o_... non-CO2 greenhouse gases (MEPC.391(81) App. 2)
//	methane_slip_percent          unburned methane escaping the engine
//	wtt_intensity_gco2eq_per_mj   well-to-tank intensity (MEPC.391(81) App. 2)
//	lhv_mj_per_kg                 lower heating value (MEPC.364(79))
//	density_kg_per_litre          density (Fourth IMO GHG Study 2020, Table 10)
//
// The per-litre factors are derived from them, so the two can never drift
// apart:
//
//	TtW CO2  [kg/L] = cf_co2                                        x density
//	TtW CO2e [kg/L] = (cf_co2 + 28*cf_ch4 + 265*cf_n2o + slip)      x density
//	WtW CO2e [kg/L] = (wtt_intensity*lhv + TtW CO2e)                x density
//
// A missing density or well-to-tank intensity means the figure is not
// derivable from an IMO reference. That is reported as unavailable rather
// than guessed.
type MFuelType struct {
	ID   uint   `gorm:"primaryKey"`
	Code string `gorm:"type:varchar(50);not null;uniqueIndex"`
	Name string `gorm:"type:varchar(255);not null"`

	CfCO2KgPerKg  *float64 `gorm:"column:cf_co2_kg_per_kg;type:numeric(12,6)"`
	CfCH4KgPerKg  *float64 `gorm:"column:cf_ch4_kg_per_kg;type:numeric(12,8)"`
	CfN2OKgPerKg  *float64 `gorm:"column:cf_n2o_kg_per_kg;type:numeric(12,8)"`
	MethaneSlipPc *float64 `gorm:"column:methane_slip_percent;type:numeric(6,3)"`

	WttIntensityGCO2eqPerMJ *float64 `gorm:"column:wtt_intensity_gco2eq_per_mj;type:numeric(12,4)"`
	LhvMJPerKg              *float64 `gorm:"column:lhv_mj_per_kg;type:numeric(10,3)"`
	DensityKgPerLitre       *float64 `gorm:"column:density_kg_per_litre;type:numeric(8,4)"`

	Reference string `gorm:"type:varchar(255)"`
	Method    string `gorm:"type:varchar(100)"`

	IsActive bool `gorm:"default:true"`

	BaseModel
}

func (MFuelType) TableName() string {
	return "m_fuel_types"
}

// TtWCO2KgPerKg returns the tank-to-wake CO2 factor in kg per kg of fuel.
// The second return value is false when the factor cannot be derived.
func (f MFuelType) TtWCO2KgPerKg() (float64, bool) {
	if f.CfCO2KgPerKg == nil {
		return 0, false
	}
	return *f.CfCO2KgPerKg, true
}

// TtWCO2eKgPerKg returns the tank-to-wake CO2-equivalent factor in kg per kg
// of fuel, including the greenhouse gases and the methane that escapes the
// engine unburned.
func (f MFuelType) TtWCO2eKgPerKg() (float64, bool) {
	if f.CfCO2KgPerKg == nil {
		return 0, false
	}
	co2 := *f.CfCO2KgPerKg
	if f.CfCH4KgPerKg != nil {
		co2 += GWP100Methane * *f.CfCH4KgPerKg
	}
	if f.CfN2OKgPerKg != nil {
		co2 += GWP100Nitrous * *f.CfN2OKgPerKg
	}
	if f.MethaneSlipPc != nil {
		// The slipped methane leaves the engine unburned and is released as CH4.
		co2 += GWP100Methane * (*f.MethaneSlipPc / 100)
	}
	return co2, true
}

// TtWCO2KgPerLitre returns the tank-to-wake CO2 factor per litre of fuel.
func (f MFuelType) TtWCO2KgPerLitre() (float64, bool) {
	perKg, ok := f.TtWCO2KgPerKg()
	if !ok || f.DensityKgPerLitre == nil {
		return 0, false
	}
	return perKg * *f.DensityKgPerLitre, true
}

// TtWCO2eKgPerLitre returns the tank-to-wake CO2-equivalent factor per litre.
func (f MFuelType) TtWCO2eKgPerLitre() (float64, bool) {
	perKg, ok := f.TtWCO2eKgPerKg()
	if !ok || f.DensityKgPerLitre == nil {
		return 0, false
	}
	return perKg * *f.DensityKgPerLitre, true
}

// WtWCO2eKgPerKg returns the well-to-wake CO2-equivalent factor in kg per kg
// of fuel: the well-to-tank part plus the tank-to-wake part.
//
// The IMO has not published a well-to-tank default for every fuel, so this
// returns false for those, and the application reports the well-to-wake figure
// as unavailable rather than inventing one.
func (f MFuelType) WtWCO2eKgPerKg() (float64, bool) {
	ttw, ok := f.TtWCO2eKgPerKg()
	if !ok || f.WttIntensityGCO2eqPerMJ == nil || f.LhvMJPerKg == nil {
		return 0, false
	}
	// gCO2eq/MJ x MJ/g = gCO2eq/g fuel, which is 1/1000 of kg CO2e per kg.
	wttKgPerKg := *f.WttIntensityGCO2eqPerMJ * *f.LhvMJPerKg / 1000
	return wttKgPerKg + ttw, true
}

// WtWCO2eKgPerLitre returns the well-to-wake CO2-equivalent factor per litre.
func (f MFuelType) WtWCO2eKgPerLitre() (float64, bool) {
	perKg, ok := f.WtWCO2eKgPerKg()
	if !ok || f.DensityKgPerLitre == nil {
		return 0, false
	}
	return perKg * *f.DensityKgPerLitre, true
}

type MPort struct {
	ID   uint    `gorm:"primaryKey"`
	Name string  `gorm:"type:varchar(255);not null"`
	Code string  `gorm:"type:varchar(100);uniqueIndex"`
	Long float64 `gorm:"type:decimal(10,6)"`
	Lat  float64 `gorm:"type:decimal(10,6)"`

	BaseModel
}

func (MPort) TableName() string {
	return "m_port"
}

type MCrew struct {
	ID                   uint       `gorm:"primaryKey"`
	Name                 string     `gorm:"type:varchar(255);not null"`
	Department           string     `gorm:"type:varchar(50);not null"` // Deck or Engine
	JobPosition          string     `gorm:"type:varchar(255)"`
	VesselID             *uint      `gorm:"index"`
	Vessel               MVessel    `gorm:"foreignKey:VesselID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Nationality          string     `gorm:"type:varchar(100);default:'Indonesian'"`
	PassportNumber       string     `gorm:"type:varchar(100)"`
	PassportExpiryDate   *time.Time `gorm:"type:date"`
	SeamanBookNumber     string     `gorm:"type:varchar(100)"`
	SeamanBookExpiryDate *time.Time `gorm:"type:date"`
	CocGradeNumber       string     `gorm:"type:varchar(255)"`
	BirthPlace           string     `gorm:"type:varchar(100)"`
	BirthDate            *time.Time `gorm:"type:date"`
	Gender               string     `gorm:"type:varchar(10);default:'M'"`
	SignOnDate           *time.Time `gorm:"type:date"`

	BaseModel
}

func (MCrew) TableName() string {
	return "m_crew"
}
