package sensorhealth

import (
	"math"
	"strings"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Service interface {
	GetSensorHealth(month string, vesselID *uint) ([]VesselHealthResponse, error)
}

type service struct {
	db   *gorm.DB
	repo Repository
}

func NewService(db *gorm.DB, repo Repository) Service {
	return &service{db: db, repo: repo}
}

func (s *service) GetSensorHealth(month string, vesselID *uint) ([]VesselHealthResponse, error) {
	monthStart, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	monthEnd := monthStart.AddDate(0, 1, 0)
	totalDaysInMonth := int(monthEnd.Sub(monthStart).Hours() / 24)

	// For current month, only count days up to today; for past months, full month
	daysInMonth := totalDaysInMonth
	if now.Year() == monthStart.Year() && now.Month() == monthStart.Month() {
		daysInMonth = today.Day()
		if daysInMonth > totalDaysInMonth {
			daysInMonth = totalDaysInMonth
		}
	}

	// Data query range: monthStart to endDate (start of next day after last counted day)
	endDate := monthStart.AddDate(0, 0, daysInMonth)

	// Get vessels
	var vessels []model.MVessel
	query := s.db.Where("is_active = ?", true)
	if vesselID != nil {
		query = query.Where("id = ?", *vesselID)
	}
	if err := query.Preload("MCUs", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Sensors", "is_visible = ?", true)
	}).Find(&vessels).Error; err != nil {
		return nil, err
	}

	var result []VesselHealthResponse

	for _, v := range vessels {
		var sensorIDs []uint
		sensorMap := make(map[uint]struct {
			Topic string
			Name  string
		})
		for _, mcu := range v.MCUs {
			for _, sensor := range mcu.Sensors {
				sensorIDs = append(sensorIDs, sensor.ID)
				sensorMap[sensor.ID] = struct {
					Topic string
					Name  string
				}{Topic: sensor.Topic, Name: sensor.Name}
			}
		}
		if len(sensorIDs) == 0 {
			continue
		}

		// Query active slots (any data = active, even value 0)
		activeRows, err := s.repo.GetActiveSlots(sensorIDs, monthStart, endDate)
		if err != nil {
			return nil, err
		}

		// Query vessel connectivity (slots where ANY sensor on vessel sent data)
		vesselIDs := []uint{v.ID}
		connRows, err := s.repo.GetVesselConnectivity(vesselIDs, monthStart, endDate)
		if err != nil {
			return nil, err
		}

		type dayKey struct {
			Day   time.Time
			Index int
		}

		// Build connectivity map: vessel (day, slot) -> connected
		connectivityMap := make(map[dayKey]bool)
		for _, row := range connRows {
			connectivityMap[dayKey{Day: row.Day.UTC(), Index: row.SlotIdx}] = true
		}

		// Build active map: sensorID -> (day, slot) -> active
		activeMap := make(map[uint]map[dayKey]bool)
		for _, row := range activeRows {
			if activeMap[row.SensorID] == nil {
				activeMap[row.SensorID] = make(map[dayKey]bool)
			}
			activeMap[row.SensorID][dayKey{Day: row.Day.UTC(), Index: row.SlotIdx}] = true
		}

		// Build per-sensor response
		var sensorResp []SensorHealthResp
		totalVesselActive := 0
		totalVesselInternetDown := 0
		totalVesselSensorOff := 0

		for _, sid := range sensorIDs {
			info := sensorMap[sid]
			sensorTotalActive := 0
			sensorTotalInternetDown := 0
			sensorTotalSensorOff := 0
			var daily []DayHealthResp

			for d := 0; d < daysInMonth; d++ {
				day := monthStart.AddDate(0, 0, d)
				slotStatus := make([]int, 12)
				dayActive := 0
				dayInternetDown := 0
				daySensorOff := 0

				for slot := 0; slot < 12; slot++ {
					dk := dayKey{Day: day.UTC(), Index: slot}
					connected := connectivityMap[dk]
					active := activeMap[sid][dk]

					if active {
						slotStatus[slot] = 1
						dayActive++
					} else if connected {
						slotStatus[slot] = 3
						daySensorOff++
					} else {
						slotStatus[slot] = 2
						dayInternetDown++
					}
				}

				sensorTotalActive += dayActive
				sensorTotalInternetDown += dayInternetDown
				sensorTotalSensorOff += daySensorOff

				wibDay := day.In(loc)
				daily = append(daily, DayHealthResp{
					Date:        wibDay.Format("2006-01-02"),
					ActiveSlots: dayActive,
					TotalSlots:  12,
					Percentage:  roundPct(dayActive, 12),
					SlotStatus:  slotStatus,
				})
			}

			totalVesselActive += sensorTotalActive
			totalVesselInternetDown += sensorTotalInternetDown
			totalVesselSensorOff += sensorTotalSensorOff

			totalSlots := daysInMonth * 12
			sensorResp = append(sensorResp, SensorHealthResp{
				SensorID:            sid,
				Topic:               info.Topic,
				Name:                info.Name,
				Category:            extractCategory(info.Topic),
				TotalPercentage:     roundPct(sensorTotalActive, totalSlots),
				TotalActive:         sensorTotalActive,
				TotalSlots:          totalSlots,
				InternetDownSlots:   sensorTotalInternetDown,
				SensorOffSlots:      sensorTotalSensorOff,
				InternetPercentage:  roundPct(sensorTotalInternetDown, totalSlots),
				SensorOffPercentage: roundPct(sensorTotalSensorOff, totalSlots),
				Daily:               daily,
			})
		}

		totalSlots := daysInMonth * 12 * len(sensorIDs)
		result = append(result, VesselHealthResponse{
			VesselID:            v.ID,
			VesselName:          v.Name,
			TotalPercentage:     roundPct(totalVesselActive, totalSlots),
			TotalActive:         totalVesselActive,
			TotalSlots:          totalSlots,
			InternetDownSlots:   totalVesselInternetDown,
			SensorOffSlots:      totalVesselSensorOff,
			InternetPercentage:  roundPct(totalVesselInternetDown, totalSlots),
			SensorOffPercentage: roundPct(totalVesselSensorOff, totalSlots),
			Sensors:             sensorResp,
		})
	}

	return result, nil
}

func extractCategory(topic string) string {
	parts := strings.Split(topic, "/")
	var group string
	switch len(parts) {
	case 5:
		group = parts[3]
	case 7:
		group = parts[5]
	default:
		group = ""
	}

	switch group {
	case "gps":
		return "GPS"
	case "rpm":
		return "RPM"
	case "flow":
		return "Flow"
	case "ac":
		return "AC"
	case "dc", "bat":
		return "DC / Battery"
	case "door":
		return "Door"
	case "temp":
		return "Temperature"
	case "humidity":
		return "Humidity"
	case "accel":
		return "Accelerometer"
	default:
		return group
	}
}

func roundPct(active, total int) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(float64(active)/float64(total)*10000) / 100
}
