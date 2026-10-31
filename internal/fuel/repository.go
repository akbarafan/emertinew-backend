package fuel

import (
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	GetFuelTotalByDate(date time.Time) ([]FuelAggregateResult, error)
	GetFuelTotalByRange(start, end time.Time) ([]FuelAggregateResult, error)
	GetFuelTotalByVesselAndDate(vesselID uint, date time.Time) (*FuelAggregateResult, error)
	GetLifetimeFuelByVessel(vesselID uint) (*FuelAggregateResult, error)
	GetFlowSensorDirections() ([]FlowSensorDirection, error)
}

// FuelAggregateResult untuk hasil query aggregate dari trackings
type FuelAggregateResult struct {
	VesselID   uint    `json:"vessel_id"`
	VesselName string  `json:"vessel_name"`
	Flow1Total float64 `json:"flow1_total"`
	Flow2Total float64 `json:"flow2_total"`
	Flow3Total float64 `json:"flow3_total"`
	Flow4Total float64 `json:"flow4_total"`
}

// Direction values stored in `m_sensors.flow_direction`.
const (
	// DirectionIn is fuel flowing back into the tank (return line).
	DirectionIn = "in"
	// DirectionOut is fuel leaving the tank towards the engine.
	DirectionOut = "out"
	// DirectionSingle is an undirected meter: the reading is the consumption.
	// It is stored as NULL in the database and travels through the code as "".
	DirectionSingle = ""
)

// FlowSensorDirection menyimpan mapping flow sensor slot dan direction per vessel
type FlowSensorDirection struct {
	VesselID  uint
	Slot      string // "flow1"-"flow4"
	Direction string // "in", "out", or "" for a single (undirected) meter
	Engine    string // engine group name
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// SensorRow untuk raw query result dari m_sensors join m_mcu
type sensorRow struct {
	VesselID      uint
	Topic         string
	FlowDirection *string
	// EngineName has to match the column name `engine_name`: GORM fills the
	// field from the column of the same name, and a field called `FlowEngine`
	// stayed empty — every engine then came out as "Unassigned".
	EngineName *string
}

func (r *repository) GetFlowSensorDirections() ([]FlowSensorDirection, error) {
	var rows []sensorRow
	query := `
		SELECT
			m.vessel_id,
			s.topic,
			s.flow_direction,
			s.engine_name
		FROM m_sensors s
		JOIN m_mcu m ON m.id = s.mcu_id
		WHERE s.topic LIKE '%/flow/%' OR s.topic LIKE '%/fuel/%'
	`
	if err := r.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]FlowSensorDirection, 0, len(rows))
	for _, row := range rows {
		slot := extractFlowSlot(row.Topic)
		if slot == "" {
			continue
		}
		// A NULL direction is the operator's explicit "Single" choice: one
		// undirected meter, whose reading is the consumption. It used to be
		// folded into "in" here, which made an unconfigured vessel look like it
		// had only bunker meters and turned the balance into zero.
		dir := DirectionSingle
		if row.FlowDirection != nil {
			switch *row.FlowDirection {
			case DirectionOut:
				dir = DirectionOut
			case DirectionIn:
				dir = DirectionIn
			}
		}
		engine := ""
		if row.EngineName != nil {
			engine = *row.EngineName
		}
		result = append(result, FlowSensorDirection{
			VesselID:  row.VesselID,
			Slot:      slot,
			Direction: dir,
			Engine:    engine,
		})
	}
	return result, nil
}

func (r *repository) GetFuelTotalByDate(date time.Time) ([]FuelAggregateResult, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	return r.GetFuelTotalByRange(startOfDay, endOfDay)
}

// GetFuelTotalByRange aggregates the metered fuel over an arbitrary window.
//
// A calendar day is not a fair comparison while the day is still running: the
// fleet has burned two hours of fuel today and twenty-four yesterday, and the
// difference between those two numbers says more about the clock than about the
// fleet. The window endpoint lets a caller compare the same elapsed period on
// two different days.
func (r *repository) GetFuelTotalByRange(start, end time.Time) ([]FuelAggregateResult, error) {
	var results []FuelAggregateResult

	query := `
		SELECT
			t.vessel_id,
			v.name AS vessel_name,
			COALESCE(SUM(t.flow1), 0) AS flow1_total,
			COALESCE(SUM(t.flow2), 0) AS flow2_total,
			COALESCE(SUM(t.flow3), 0) AS flow3_total,
			COALESCE(SUM(t.flow4), 0) AS flow4_total
		FROM trackings t
		JOIN m_vessels v ON v.id = t.vessel_id
		WHERE t.time >= ? AND t.time < ?
		GROUP BY t.vessel_id, v.name
		ORDER BY t.vessel_id
	`

	err := r.db.Raw(query, start, end).Scan(&results).Error
	return results, err
}

func (r *repository) GetFuelTotalByVesselAndDate(vesselID uint, date time.Time) (*FuelAggregateResult, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var result FuelAggregateResult

	query := `
		SELECT
			t.vessel_id,
			v.name AS vessel_name,
			COALESCE(SUM(t.flow1), 0) AS flow1_total,
			COALESCE(SUM(t.flow2), 0) AS flow2_total,
			COALESCE(SUM(t.flow3), 0) AS flow3_total,
			COALESCE(SUM(t.flow4), 0) AS flow4_total
		FROM trackings t
		JOIN m_vessels v ON v.id = t.vessel_id
		WHERE t.vessel_id = ? AND t.time >= ? AND t.time < ?
		GROUP BY t.vessel_id, v.name
	`

	err := r.db.Raw(query, vesselID, startOfDay, endOfDay).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}


func (r *repository) GetLifetimeFuelByVessel(vesselID uint) (*FuelAggregateResult, error) {
	var result FuelAggregateResult

	query := `
		SELECT
			t.vessel_id,
			v.name AS vessel_name,
			COALESCE(SUM(t.flow1), 0) AS flow1_total,
			COALESCE(SUM(t.flow2), 0) AS flow2_total,
			COALESCE(SUM(t.flow3), 0) AS flow3_total,
			COALESCE(SUM(t.flow4), 0) AS flow4_total
		FROM trackings t
		JOIN m_vessels v ON v.id = t.vessel_id
		WHERE t.vessel_id = ?
		GROUP BY t.vessel_id, v.name
	`

	err := r.db.Raw(query, vesselID).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}
