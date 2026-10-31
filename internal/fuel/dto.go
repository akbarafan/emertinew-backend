package fuel

import "strings"

// FuelQueryParams untuk query params API fuel
type FuelQueryParams struct {
	Date string `form:"date" example:"2026-04-17"` // Format: YYYY-MM-DD
}

// EngineFuelDTO per-engine fuel breakdown.
//
// Direction is stored per sensor in `m_sensors.flow_direction` and is read from
// the engine's point of view (marine standard):
//
//	"in"  — fuel entering the engine from the tank (supply / feed line)
//	"out" — fuel exiting the engine back to the tank (return line)
//	""    — "single": an undirected meter; its reading is the consumption
//
// Consumption therefore is `in - out` (fuel supplied minus fuel returned).
type EngineFuelDTO struct {
	EngineName string  `json:"engine_name" example:"Port Engine"`
	FlowIn     float64 `json:"flow_in" example:"250.3"`
	FlowOut    float64 `json:"flow_out" example:"50.1"`
	FlowSingle float64 `json:"flow_single" example:"0"`
	// Consumed is the fuel actually burned by this engine, in litres.
	// -1 when the engine has no meter that can measure consumption.
	Consumed float64 `json:"consumed" example:"200.2"`
	// StockChange is the change in the tank: fuel returned minus fuel burned.
	// Negative while the vessel is burning fuel it bunkered earlier.
	StockChange float64 `json:"stock_change" example:"-200.2"`
	// Deprecated: kept as an alias of Consumed for older clients. The old
	// meaning (`in - out`, clamped) is what this field used to carry.
	NetTotal float64 `json:"net_total" example:"200.2"`
}

// FuelTotalDTO untuk response total fuel per vessel pada hari tertentu
type FuelTotalDTO struct {
	VesselID   uint    `json:"vessel_id" example:"1"`
	VesselName string  `json:"vessel_name" example:"KM Nusantara 01"`
	Date       string  `json:"date" example:"2026-04-17"`
	Flow1Total float64 `json:"flow1_total" example:"120.5"`
	Flow2Total float64 `json:"flow2_total" example:"115.3"`
	Flow3Total float64 `json:"flow3_total" example:"118.7"`
	Flow4Total float64 `json:"flow4_total" example:"122.1"`
	// FuelTotal is the raw metered volume across every channel, in litres.
	FuelTotal float64 `json:"fuel_total" example:"476.6"`
	// Consumed is the fuel burned, in litres. -1 when no engine can measure it.
	Consumed float64 `json:"consumed" example:"306.5"`
	// FuelReturned is the volume that flowed back into the tanks, in litres.
	FuelReturned float64 `json:"fuel_returned" example:"50.1"`
	// StockChange is fuel returned minus fuel burned. Negative when burning.
	StockChange float64 `json:"stock_change" example:"-256.4"`
	// Deprecated: alias of Consumed, kept for older clients.
	NetTotal float64         `json:"net_total" example:"306.5"`
	Engines  []EngineFuelDTO `json:"engines,omitempty"`
}

// extractFlowSlot extracts the flow slot name (flow1-flow4) from a sensor topic.
// Topic format: hcml/vs001/er/flow/{N} (5-part) or hcml/flt001/sqd001/vs001/brd/flow/{N} (7-part)
func extractFlowSlot(topic string) string {
	parts := strings.Split(strings.ToLower(topic), "/")
	var group, num string
	if len(parts) == 5 {
		group = parts[3]
		num = parts[4]
	} else if len(parts) == 7 {
		group = parts[5]
		num = parts[6]
	} else {
		return ""
	}
	if group != "flow" && group != "fuel" {
		return ""
	}
	n := 1
	if len(num) > 0 {
		n = int(num[0] - '0')
	}
	if n < 1 || n > 4 {
		return ""
	}
	return "flow" + string(rune('0'+n))
}
