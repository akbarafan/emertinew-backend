package fuel

import (
	"math"
	"sort"
	"strconv"
	"time"
)

// NotMeasured is the sentinel used when a vessel has no meter that can report
// consumption. It is deliberately not 0: reporting 0 would claim the vessel
// burned no fuel, which is a different statement from "we cannot tell".
const NotMeasured = -1.0

// UnassignedEngineName is the bucket used for flow channels that carry no
// `engine_name` label. They are still metered, so they belong in the totals;
// they just cannot be attributed to an engine.
const UnassignedEngineName = "Unassigned"

type Service interface {
	GetFuelTotal(date time.Time) ([]FuelTotalDTO, error)
	GetFuelTotalByVessel(vesselID uint, date time.Time) (*FuelTotalDTO, error)
	GetLifetimeFuelByVessel(vesselID uint) (*FuelTotalDTO, error)
	GetFuelTotalByRange(start, end time.Time) ([]FuelTotalDTO, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// slotIndex returns the numeric part of a flow slot ("flow3" -> 3).
// Slots that are not a numbered flow channel sort last.
func slotIndex(slot string) int {
	if len(slot) < 5 || slot[:4] != "flow" {
		return math.MaxInt
	}
	index, err := strconv.Atoi(slot[4:])
	if err != nil {
		return math.MaxInt
	}
	return index
}

// engineAccum collects the three directions a flow channel can carry.
type engineAccum struct {
	inTotal     float64
	outTotal    float64
	singleTotal float64
	hasIn       bool
	hasOut      bool
	hasSingle   bool
	// lowestSlot is the numeric part of the lowest channel this engine uses
	// ("flow3" -> 3). It gives the engines a stable, physical order.
	lowestSlot int
}

// consumed applies the consumption rule for one engine.
//
//	directed pair (in + out) -> in - out, because `in` is fuel supplied to
//	                            the engine, and `out` is fuel returned to the tank
//	in only                  -> in
//	single only              -> the reading itself
//	out only                 -> not measurable, the engine has no supply meter
func (e *engineAccum) consumed() float64 {
	switch {
	case e.hasIn && e.hasOut:
		v := e.inTotal - e.outTotal
		if v < 0 {
			return 0
		}
		return v
	case e.hasIn:
		return e.inTotal + e.singleTotal
	case e.hasSingle:
		return e.singleTotal
	default:
		return NotMeasured
	}
}

// stockChange is the movement of the tank: fuel returned minus fuel burned.
// It is negative while the vessel burns fuel that was bunkered earlier.
func (e *engineAccum) stockChange(consumed float64) float64 {
	if consumed == NotMeasured {
		return 0
	}
	return e.outTotal - consumed
}

func (s *service) buildEngineBreakdown(vesselID uint, flowTotals map[string]float64, directions []FlowSensorDirection) (float64, float64, float64, []EngineFuelDTO) {
	// Group sensors by engine
	engines := make(map[string]*engineAccum)

	// Sensors without engine group go to the unassigned bucket
	unassigned := &engineAccum{}

	for _, d := range directions {
		if d.VesselID != vesselID {
			continue
		}
		val := flowTotals[d.Slot]
		target := unassigned
		if d.Engine != "" {
			ea, ok := engines[d.Engine]
			if !ok {
				ea = &engineAccum{lowestSlot: math.MaxInt}
				engines[d.Engine] = ea
			}
			target = ea
		}
		if slot := slotIndex(d.Slot); slot < target.lowestSlot {
			target.lowestSlot = slot
		}

		switch d.Direction {
		case DirectionOut:
			target.outTotal += val
			target.hasOut = true
		case DirectionIn:
			target.inTotal += val
			target.hasIn = true
		default: // single / undirected meter
			target.singleTotal += val
			target.hasSingle = true
		}
	}

	engineDTOs := make([]EngineFuelDTO, 0, len(engines)+1)
	type orderedEngine struct {
		dto  EngineFuelDTO
		slot int
	}
	ordered := make([]orderedEngine, 0, len(engines)+1)
	totalConsumed := 0.0
	totalReturned := 0.0
	measuredEngines := 0

	appendEngine := func(name string, ea *engineAccum) {
		consumed := ea.consumed()
		ordered = append(ordered, orderedEngine{
			dto: EngineFuelDTO{
				EngineName:  name,
				FlowIn:      ea.inTotal,
				FlowOut:     ea.outTotal,
				FlowSingle:  ea.singleTotal,
				Consumed:    consumed,
				StockChange: ea.stockChange(consumed),
				NetTotal:    consumed,
			},
			slot: ea.lowestSlot,
		})
		if consumed != NotMeasured {
			totalConsumed += consumed
			totalReturned += ea.outTotal
			measuredEngines++
		}
	}

	for name, ea := range engines {
		appendEngine(name, ea)
	}
	if unassigned.hasIn || unassigned.hasOut || unassigned.hasSingle {
		appendEngine(UnassignedEngineName, unassigned)
	}

	// The engines were collected in a map, so their order was random and the
	// table jumped around on every reload. Sorting by the lowest channel puts
	// them back in the physical order of the meters.
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].slot != ordered[j].slot {
			return ordered[i].slot < ordered[j].slot
		}
		return ordered[i].dto.EngineName < ordered[j].dto.EngineName
	})
	for _, item := range ordered {
		engineDTOs = append(engineDTOs, item.dto)
	}

	if measuredEngines == 0 {
		totalConsumed = NotMeasured
		totalReturned = 0
	}

	stockChange := 0.0
	if totalConsumed != NotMeasured {
		stockChange = totalReturned - totalConsumed
	}

	return totalConsumed, totalReturned, stockChange, engineDTOs
}

func (s *service) enrichWithNetTotal(dtos []FuelTotalDTO) ([]FuelTotalDTO, error) {
	directions, err := s.repo.GetFlowSensorDirections()
	if err != nil {
		return dtos, err
	}

	for i, dto := range dtos {
		flowTotals := map[string]float64{
			"flow1": dto.Flow1Total,
			"flow2": dto.Flow2Total,
			"flow3": dto.Flow3Total,
			"flow4": dto.Flow4Total,
		}
		consumed, returned, stockChange, engines := s.buildEngineBreakdown(dto.VesselID, flowTotals, directions)
		dtos[i].Consumed = consumed
		dtos[i].FuelReturned = returned
		dtos[i].StockChange = stockChange
		dtos[i].NetTotal = consumed
		dtos[i].Engines = engines
	}

	return dtos, nil
}

func (s *service) enrichOneWithNetTotal(dto *FuelTotalDTO) (*FuelTotalDTO, error) {
	directions, err := s.repo.GetFlowSensorDirections()
	if err != nil {
		return dto, err
	}

	flowTotals := map[string]float64{
		"flow1": dto.Flow1Total,
		"flow2": dto.Flow2Total,
		"flow3": dto.Flow3Total,
		"flow4": dto.Flow4Total,
	}
	consumed, returned, stockChange, engines := s.buildEngineBreakdown(dto.VesselID, flowTotals, directions)
	dto.Consumed = consumed
	dto.FuelReturned = returned
	dto.StockChange = stockChange
	dto.NetTotal = consumed
	dto.Engines = engines
	return dto, nil
}

func (s *service) convertToDTO(result *FuelAggregateResult, dateStr string) *FuelTotalDTO {
	return &FuelTotalDTO{
		VesselID:   result.VesselID,
		VesselName: result.VesselName,
		Date:       dateStr,
		Flow1Total: result.Flow1Total,
		Flow2Total: result.Flow2Total,
		Flow3Total: result.Flow3Total,
		Flow4Total: result.Flow4Total,
		FuelTotal:  result.Flow1Total + result.Flow2Total + result.Flow3Total + result.Flow4Total,
		// Defaults for a vessel with no flow sensors at all.
		Consumed:     NotMeasured,
		FuelReturned: 0,
		StockChange:  0,
	}
}

func (s *service) GetFuelTotal(date time.Time) ([]FuelTotalDTO, error) {
	results, err := s.repo.GetFuelTotalByDate(date)
	if err != nil {
		return nil, err
	}

	dateStr := date.Format("2006-01-02")
	dtos := make([]FuelTotalDTO, 0, len(results))
	for _, r := range results {
		dto := s.convertToDTO(&r, dateStr)
		dtos = append(dtos, *dto)
	}

	return s.enrichWithNetTotal(dtos)
}

func (s *service) GetFuelTotalByVessel(vesselID uint, date time.Time) (*FuelTotalDTO, error) {
	result, err := s.repo.GetFuelTotalByVesselAndDate(vesselID, date)
	if err != nil {
		return nil, err
	}

	dateStr := date.Format("2006-01-02")
	dto := s.convertToDTO(result, dateStr)

	return s.enrichOneWithNetTotal(dto)
}

// GetFuelTotalByRange returns the fuel of every vessel over an arbitrary
// window. The date field carries the label of the window rather than a single
// day, so a caller can tell a two-hour window from a full one.
func (s *service) GetFuelTotalByRange(start, end time.Time) ([]FuelTotalDTO, error) {
	results, err := s.repo.GetFuelTotalByRange(start, end)
	if err != nil {
		return nil, err
	}

	label := start.Format("2006-01-02 15:04") + " - " + end.Format("2006-01-02 15:04")
	dtos := make([]FuelTotalDTO, 0, len(results))
	for _, r := range results {
		dto := s.convertToDTO(&r, label)
		dtos = append(dtos, *dto)
	}

	return s.enrichWithNetTotal(dtos)
}


func (s *service) GetLifetimeFuelByVessel(vesselID uint) (*FuelTotalDTO, error) {
	result, err := s.repo.GetLifetimeFuelByVessel(vesselID)
	if err != nil {
		return nil, err
	}
	dto := &FuelTotalDTO{
		VesselID:   result.VesselID,
		VesselName: result.VesselName,
		Flow1Total: result.Flow1Total,
		Flow2Total: result.Flow2Total,
		Flow3Total: result.Flow3Total,
		Flow4Total: result.Flow4Total,
		FuelTotal:  result.Flow1Total + result.Flow2Total + result.Flow3Total + result.Flow4Total,
	}
	return s.enrichOneWithNetTotal(dto)
}
