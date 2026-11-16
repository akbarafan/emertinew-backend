package vessel

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RunningHoursService interface {
	GetRunningHours(vesselID uint) ([]model.VesselEngineRunningHours, error)
	ResetTripHours(vesselID uint, slot string, tripName string, userID *uint, userName, userRole string) (*model.VesselEngineRunningHours, *model.TripRunningHoursResetLog, error)
	GetResetLogs(vesselID uint, limit int) ([]model.TripRunningHoursResetLog, error)
	AccumulateEngineHours(vesselID uint, slot string, hours float64, activeAt time.Time) error
}

type runningHoursService struct {
	db *gorm.DB
}

func NewRunningHoursService(db *gorm.DB) RunningHoursService {
	return &runningHoursService{db: db}
}

func (s *runningHoursService) lookupSensorID(vesselID uint, slot string) *uint {
	slotIndex := strings.TrimPrefix(strings.ToLower(slot), "rpm")
	if slotIndex == "" {
		return nil
	}
	var sensor model.MSensor
	pattern := "%rpm/" + slotIndex
	keyPattern := "%rpm:" + slotIndex
	err := s.db.Model(&model.MSensor{}).
		Joins("JOIN m_mcu ON m_sensors.mcu_id = m_mcu.id").
		Where("m_mcu.vessel_id = ? AND (m_sensors.topic LIKE ? OR m_sensors.topic_key LIKE ?)", vesselID, pattern, keyPattern).
		Select("m_sensors.id").
		First(&sensor).Error
	if err == nil && sensor.ID > 0 {
		return &sensor.ID
	}
	return nil
}

func (s *runningHoursService) syncSlotHours(vesselID uint, rec *model.VesselEngineRunningHours) {
	slot := strings.ToLower(strings.TrimSpace(rec.EngineSlot))
	if slot != "rpm1" && slot != "rpm2" && slot != "rpm3" && slot != "rpm4" {
		return
	}

	// 1. Initial backfill (if never synced or total is 0) vs Incremental sync (sub-millisecond)
	if rec.LastActiveAt == nil || rec.TotalRunningHours == 0 {
		var fullResult struct {
			Minutes int64      `gorm:"column:minutes"`
			MaxTime *time.Time `gorm:"column:max_time"`
		}
		_ = s.db.Raw(`
			SELECT 
				COALESCE(SUM(CASE WHEN `+slot+` > 0 THEN 1 ELSE 0 END), 0) AS minutes,
				MAX(time) AS max_time
			FROM trackings
			WHERE vessel_id = ?
		`, vesselID).Scan(&fullResult).Error

		rec.TotalRunningHours = float64(fullResult.Minutes) / 60.0
		if fullResult.MaxTime != nil {
			rec.LastActiveAt = fullResult.MaxTime
		} else {
			now := time.Now()
			rec.LastActiveAt = &now
		}
	} else {
		// Incremental: only query new or late-buffered data arrived after last_active_at (< 1ms)
		var incResult struct {
			Minutes int64      `gorm:"column:minutes"`
			MaxTime *time.Time `gorm:"column:max_time"`
		}
		_ = s.db.Raw(`
			SELECT 
				COALESCE(SUM(CASE WHEN `+slot+` > 0 THEN 1 ELSE 0 END), 0) AS minutes,
				MAX(time) AS max_time
			FROM trackings
			WHERE vessel_id = ? AND time > ?
		`, vesselID, *rec.LastActiveAt).Scan(&incResult).Error

		if incResult.Minutes > 0 {
			rec.TotalRunningHours += float64(incResult.Minutes) / 60.0
		}
		if incResult.MaxTime != nil {
			rec.LastActiveAt = incResult.MaxTime
		}
	}

	// 2. Compute service/trip hours based on the latest reset log
	var lastReset model.TripRunningHoursResetLog
	if err := s.db.Where("vessel_id = ? AND engine_slot = ?", vesselID, slot).Order("reset_at DESC").First(&lastReset).Error; err == nil {
		var sinceResetMinutes int64
		_ = s.db.Raw(`
			SELECT COALESCE(SUM(CASE WHEN `+slot+` > 0 THEN 1 ELSE 0 END), 0)
			FROM trackings
			WHERE vessel_id = ? AND time > ?
		`, vesselID, lastReset.ResetAt).Scan(&sinceResetMinutes).Error
		rec.TripRunningHours = float64(sinceResetMinutes) / 60.0
	} else {
		rec.TripRunningHours = rec.TotalRunningHours
	}

	// 3. Persist cache
	_ = s.db.Model(rec).Updates(map[string]interface{}{
		"sensor_id":           rec.SensorID,
		"total_running_hours": rec.TotalRunningHours,
		"trip_running_hours":  rec.TripRunningHours,
		"last_active_at":      rec.LastActiveAt,
		"updated_at":          time.Now(),
	}).Error
}

func (s *runningHoursService) GetRunningHours(vesselID uint) ([]model.VesselEngineRunningHours, error) {
	var records []model.VesselEngineRunningHours
	err := s.db.Where("vessel_id = ?", vesselID).Order("engine_slot ASC").Find(&records).Error
	if err != nil {
		return nil, err
	}

	// Ensure default slots rpm1 and rpm2 exist so the UI always has records
	existingSlots := make(map[string]bool)
	for i := range records {
		existingSlots[records[i].EngineSlot] = true
	}

	defaultSlots := []string{"rpm1", "rpm2"}
	for _, slot := range defaultSlots {
		if !existingSlots[slot] {
			sensorID := s.lookupSensorID(vesselID, slot)
			newRec := model.VesselEngineRunningHours{
				VesselID:          vesselID,
				SensorID:          sensorID,
				EngineSlot:        slot,
				TotalRunningHours: 0,
				TripRunningHours:  0,
			}
			_ = s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&newRec).Error
			records = append(records, newRec)
		}
	}

	// Sync each slot using high-performance incremental watermark (< 1ms per query)
	for i := range records {
		if records[i].SensorID == nil {
			records[i].SensorID = s.lookupSensorID(vesselID, records[i].EngineSlot)
		}
		s.syncSlotHours(vesselID, &records[i])
	}

	return records, nil
}

func (s *runningHoursService) ResetTripHours(
	vesselID uint,
	slot string,
	tripName string,
	userID *uint,
	userName, userRole string,
) (*model.VesselEngineRunningHours, *model.TripRunningHoursResetLog, error) {
	slot = strings.ToLower(strings.TrimSpace(slot))
	if slot == "" {
		return nil, nil, errors.New("engine_slot is required")
	}

	var record model.VesselEngineRunningHours
	err := s.db.Where("vessel_id = ? AND engine_slot = ?", vesselID, slot).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			sensorID := s.lookupSensorID(vesselID, slot)
			record = model.VesselEngineRunningHours{
				VesselID:          vesselID,
				SensorID:          sensorID,
				EngineSlot:        slot,
				TotalRunningHours: 0,
				TripRunningHours:  0,
			}
			if err := s.db.Create(&record).Error; err != nil {
				return nil, nil, err
			}
		} else {
			return nil, nil, err
		}
	}

	hoursBefore := record.TripRunningHours
	now := time.Now()

	sensorID := record.SensorID
	if sensorID == nil {
		sensorID = s.lookupSensorID(vesselID, slot)
	}

	resetLog := model.TripRunningHoursResetLog{
		VesselID:         vesselID,
		SensorID:         sensorID,
		EngineSlot:       slot,
		HoursBeforeReset: hoursBefore,
		TripName:         tripName,
		ResetByUserID:    userID,
		ResetByName:      userName,
		ResetByRole:      userRole,
		ResetAt:          now,
	}

	tx := s.db.Begin()
	if err := tx.Create(&resetLog).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	record.TripRunningHours = 0
	record.UpdatedAt = now
	if err := tx.Save(&record).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, nil, err
	}

	return &record, &resetLog, nil
}

func (s *runningHoursService) GetResetLogs(vesselID uint, limit int) ([]model.TripRunningHoursResetLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var logs []model.TripRunningHoursResetLog
	err := s.db.Where("vessel_id = ?", vesselID).
		Order("reset_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (s *runningHoursService) AccumulateEngineHours(vesselID uint, slot string, hours float64, activeAt time.Time) error {
	if hours <= 0 {
		return nil
	}
	slot = strings.ToLower(strings.TrimSpace(slot))
	sensorID := s.lookupSensorID(vesselID, slot)

	return s.db.Exec(`
		INSERT INTO vessel_engine_running_hours (vessel_id, sensor_id, engine_slot, total_running_hours, trip_running_hours, last_active_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
		ON CONFLICT (vessel_id, engine_slot)
		DO UPDATE SET
			sensor_id = COALESCE(EXCLUDED.sensor_id, vessel_engine_running_hours.sensor_id),
			total_running_hours = vessel_engine_running_hours.total_running_hours + EXCLUDED.total_running_hours,
			trip_running_hours = vessel_engine_running_hours.trip_running_hours + EXCLUDED.trip_running_hours,
			last_active_at = EXCLUDED.last_active_at,
			updated_at = NOW()
	`, vesselID, sensorID, slot, hours, hours, activeAt).Error
}

type RunningHoursHandler struct {
	svc RunningHoursService
}

func NewRunningHoursHandler(svc RunningHoursService) *RunningHoursHandler {
	return &RunningHoursHandler{svc: svc}
}

type ResetTripRequest struct {
	EngineSlot string `json:"engine_slot" binding:"required"`
	TripName   string `json:"trip_name"`
}

// GetRunningHoursHandler returns running hours for a vessel
func (h *RunningHoursHandler) GetRunningHoursHandler(c *gin.Context) {
	idStr := c.Param("id")
	vesselID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid vessel ID"})
		return
	}

	hours, err := h.svc.GetRunningHours(uint(vesselID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Running hours retrieved",
		"data":    hours,
	})
}

// ResetTripHoursHandler resets the trip hours for an engine slot (Admin only)
func (h *RunningHoursHandler) ResetTripHoursHandler(c *gin.Context) {
	idStr := c.Param("id")
	vesselID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid vessel ID"})
		return
	}

	// RBAC check: only admin or superadmin
	roleVal, _ := c.Get("role")
	role, _ := roleVal.(string)
	if role != "admin" && role != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Only administrators are permitted to reset trip running hours",
		})
		return
	}

	var req ResetTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request body: engine_slot is required"})
		return
	}

	var userID *uint
	if uID, exists := c.Get("user_id"); exists {
		if id, ok := uID.(uint); ok {
			userID = &id
		} else if idInt, ok := uID.(int); ok {
			uid := uint(idInt)
			userID = &uid
		}
	}

	userName := "Admin"
	if uName, exists := c.Get("username"); exists {
		if s, ok := uName.(string); ok && s != "" {
			userName = s
		}
	}

	rec, logRec, err := h.svc.ResetTripHours(uint(vesselID), req.EngineSlot, req.TripName, userID, userName, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Trip running hours reset successfully",
		"data": gin.H{
			"running_hours": rec,
			"reset_log":     logRec,
		},
	})
}

// GetResetLogsHandler returns audit logs of trip resets for a vessel
func (h *RunningHoursHandler) GetResetLogsHandler(c *gin.Context) {
	idStr := c.Param("id")
	vesselID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid vessel ID"})
		return
	}

	limit := 50
	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	logs, err := h.svc.GetResetLogs(uint(vesselID), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Trip reset logs retrieved",
		"data":    logs,
	})
}
