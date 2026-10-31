package mcu

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/tracking"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

func formatMissingIDsError(entityType string, missingIDs []uint) string {
	if len(missingIDs) == 0 {
		return fmt.Sprintf("One or more %s IDs not found", entityType)
	}

	idsStr := make([]string, len(missingIDs))
	for i, id := range missingIDs {
		idsStr[i] = fmt.Sprintf("%d", id)
	}

	if len(missingIDs) == 1 {
		return fmt.Sprintf("%s ID not found: %s", strings.Title(entityType), idsStr[0])
	}
	return fmt.Sprintf("%s IDs not found: %s", strings.Title(entityType), strings.Join(idsStr, ", "))
}

type Service interface {
	CreateMCU(req CreateMCURequest, userID uint, ipAddress, userAgent string) (*MCUResponse, error)
	GetMCUByID(id uint) (*MCUResponse, error)
	GetAllMCUs() ([]MCUResponse, error)
	UpdateMCU(id uint, req UpdateMCURequest, userID uint, ipAddress, userAgent string) (*MCUResponse, error)
	DeleteMCU(id uint, userID uint, ipAddress, userAgent string) error

	// Bulk sensor topic conversion
	PreviewConvertSensors(mcuID uint, req ConvertSensorRequest) (*PreviewConvertResponse, error)
	ApplyConvertSensors(mcuID uint, req ConvertSensorRequest, userID uint, ipAddress, userAgent string) (*ConvertSensorResponse, error)
}

type service struct {
	repo         Repository
	auditService audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{repo: repo, auditService: auditService}
}

func buildAuditData(mcu *model.MMCU) MCUAuditData {
	data := MCUAuditData{
		ID:       mcu.ID,
		VesselID: mcu.VesselID,
		Name:     mcu.Name,
		Location: mcu.Location,
	}
	for _, s := range mcu.Sensors {
		data.SensorIDs = append(data.SensorIDs, s.ID)
	}
	return data
}

func (s *service) CreateMCU(req CreateMCURequest, userID uint, ipAddress, userAgent string) (*MCUResponse, error) {
	// Validate sensor IDs first before creating mcu
	if len(req.SensorIDs) > 0 {
		_, missingIDs, err := s.repo.ValidateSensorIDs(req.SensorIDs)
		if err != nil {
			return nil, err
		}
		if len(missingIDs) > 0 {
			return nil, errors.New(formatMissingIDsError("sensor", missingIDs))
		}
	}

	mcu := &model.MMCU{
		Name:     req.Name,
		Location: req.Location,
	}

	err := s.repo.Create(mcu)
	if err != nil {
		return nil, err
	}

	// Assign sensors ke mcu ini
	if len(req.SensorIDs) > 0 {
		_, err := s.repo.AssignSensorsToMCU(mcu.ID, req.SensorIDs)
		if err != nil {
			return nil, err
		}
	}

	// Get mcu with preloaded sensors
	mcu, err = s.repo.GetByID(mcu.ID)
	if err != nil {
		return nil, err
	}

	// Audit log for create
	if s.auditService != nil {
		auditData := buildAuditData(mcu)
		if err := s.auditService.Log(userID, "m_mcu", mcu.ID, "CREATE", nil, auditData, ipAddress, userAgent); err != nil {
			log.Printf("[AUDIT ERROR] Failed to log CREATE for mcu %d: %v", mcu.ID, err)
		}
	}

	return s.mapToResponse(mcu), nil
}

func (s *service) GetMCUByID(id uint) (*MCUResponse, error) {
	mcu, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("MCU not added yet")
	}

	return s.mapToResponse(mcu), nil
}

func (s *service) GetAllMCUs() ([]MCUResponse, error) {
	mcus, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	// Return empty array instead of error when no data exists
	if len(mcus) == 0 {
		return []MCUResponse{}, nil
	}

	var response []MCUResponse
	for _, mcu := range mcus {
		response = append(response, *s.mapToResponse(&mcu))
	}

	return response, nil
}

func (s *service) UpdateMCU(id uint, req UpdateMCURequest, userID uint, ipAddress, userAgent string) (*MCUResponse, error) {
	mcu, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("MCU not found")
	}

	oldData := buildAuditData(mcu)

	// Validate sensor IDs first if provided
	if len(req.SensorIDs) > 0 {
		_, missingIDs, err := s.repo.ValidateSensorIDs(req.SensorIDs)
		if err != nil {
			return nil, err
		}
		if len(missingIDs) > 0 {
			return nil, errors.New(formatMissingIDsError("sensor", missingIDs))
		}
	}

	if req.VesselID != nil {
		// Invalidate realtime cache for old vessel if vessel assignment is changing
		if mcu.VesselID != nil && *mcu.VesselID != *req.VesselID {
			tracking.InvalidateRealtimeCache(*mcu.VesselID)
			log.Printf("[MCU] Vessel assignment changing: MCU %d from vessel %d to vessel %d", id, *mcu.VesselID, *req.VesselID)
		}
		// Also invalidate cache for new vessel to ensure fresh state
		tracking.InvalidateRealtimeCache(*req.VesselID)
		mcu.VesselID = req.VesselID
	}
	if req.Name != "" {
		mcu.Name = req.Name
	}
	if req.Location != "" {
		mcu.Location = req.Location
	}

	err = s.repo.Update(mcu)
	if err != nil {
		return nil, err
	}

	// Reassign sensors jika ada
	if len(req.SensorIDs) > 0 {
		_, err := s.repo.AssignSensorsToMCU(mcu.ID, req.SensorIDs)
		if err != nil {
			return nil, err
		}
	}

	// Get updated mcu with preloaded sensors
	mcu, err = s.repo.GetByID(mcu.ID)
	if err != nil {
		return nil, err
	}

	// Audit log for update
	if s.auditService != nil {
		newData := buildAuditData(mcu)
		if err := s.auditService.Log(userID, "m_mcu", mcu.ID, "UPDATE", oldData, newData, ipAddress, userAgent); err != nil {
			log.Printf("[AUDIT ERROR] Failed to log UPDATE for mcu %d: %v", mcu.ID, err)
		}
	}

	return s.mapToResponse(mcu), nil
}

// convertOldToNewTopic converts a 5-part old-format topic to 7-part new format.
// Old: hcml/vs003/bd/gps/1 → New: hcml/flt001/sqd001/vs003/brd/gps/1
func convertOldToNewTopic(oldTopic string, fleetCode, squadCode string) (string, string, bool) {
	parts := strings.Split(oldTopic, "/")
	if len(parts) != 5 {
		return "", "", false
	}

	// Map old location codes to new
	location := parts[2]
	switch location {
	case "bd":
		location = "brd"
	case "er":
		location = "eng"
	}

	newTopic := fmt.Sprintf("%s/%s/%s/%s/%s/%s/%s",
		parts[0], fleetCode, squadCode, parts[1], location, parts[3], parts[4])

	// compute new topic key: company:vessel_code:location:sensor_group:sensor_number
	newKey := fmt.Sprintf("%s:%s:%s:%s:%s",
		parts[0], parts[1], location, parts[3], parts[4])

	return newTopic, newKey, true
}

// isOldFormatTopic checks if a topic uses the old 5-part format
func isOldFormatTopic(topic string) bool {
	parts := strings.Split(topic, "/")
	return len(parts) == 5
}

// isNewFormatTopic checks if a topic uses the new 7-part format
func isNewFormatTopic(topic string) bool {
	parts := strings.Split(topic, "/")
	return len(parts) == 7
}

// PreviewConvertSensors previews bulk sensor topic conversion without applying
func (s *service) PreviewConvertSensors(mcuID uint, req ConvertSensorRequest) (*PreviewConvertResponse, error) {
	mcu, err := s.repo.GetByID(mcuID)
	if err != nil {
		return nil, errors.New("MCU not found")
	}

	// Get all sensors for this MCU
	var sensors []model.MSensor
	if len(req.SensorIDs) > 0 {
		for _, sid := range req.SensorIDs {
			for _, ms := range mcu.Sensors {
				if ms.ID == sid {
					sensors = append(sensors, ms)
					break
				}
			}
		}
	} else {
		sensors = mcu.Sensors
	}

	if len(sensors) == 0 {
		return &PreviewConvertResponse{Items: []ConvertSensorItem{}, Convertible: 0, Conflicts: 0, AlreadyNew: 0}, nil
	}

	// Get vessel fleet/squad info
	var fleetCode, squadCode string
	if mcu.VesselID != nil {
		vessel, err := s.repo.GetVesselByID(*mcu.VesselID)
		if err == nil && vessel != nil {
			if vessel.FleetID != nil {
				if vessel.Fleet.Code != "" {
					fleetCode = vessel.Fleet.Code
				}
			}
			if vessel.SquadID != nil {
				if vessel.Squad.Code != "" {
					squadCode = vessel.Squad.Code
				}
			}
		}
	}
	if fleetCode == "" {
		fleetCode = "flt001" // default fallback
	}
	if squadCode == "" {
		squadCode = "sqd001" // default fallback
	}

	var items []ConvertSensorItem
	convertible := 0
	conflicts := 0
	alreadyNew := 0

	for _, sensor := range sensors {
		item := ConvertSensorItem{
			SensorID:    sensor.ID,
			Name:        sensor.Name,
			OldTopic:    sensor.Topic,
			IsNewFormat: isNewFormatTopic(sensor.Topic),
		}

		if isNewFormatTopic(sensor.Topic) {
			alreadyNew++
			item.NewTopic = sensor.Topic
			items = append(items, item)
			continue
		}

		if !isOldFormatTopic(sensor.Topic) {
			// Unrecognized format, skip
			item.NewTopic = sensor.Topic
			item.ConflictMsg = "unrecognized topic format"
			items = append(items, item)
			continue
		}

		newTopic, _, ok := convertOldToNewTopic(sensor.Topic, fleetCode, squadCode)
		if !ok {
			item.ConflictMsg = "failed to convert topic"
			items = append(items, item)
			continue
		}

		item.NewTopic = newTopic

		// Check if new topic already exists (conflict)
		existing, err := s.repo.GetSensorByTopic(newTopic)
		if err == nil && existing != nil && existing.ID != sensor.ID {
			item.Conflict = true
			item.ConflictWith = &existing.ID
			item.ConflictMsg = fmt.Sprintf("Topic already used by sensor #%d", existing.ID)
			conflicts++
		} else {
			convertible++
		}

		items = append(items, item)
	}

	return &PreviewConvertResponse{
		Items:       items,
		Convertible: convertible,
		Conflicts:   conflicts,
		AlreadyNew:  alreadyNew,
	}, nil
}

// ApplyConvertSensors applies bulk sensor topic conversion
func (s *service) ApplyConvertSensors(mcuID uint, req ConvertSensorRequest, userID uint, ipAddress, userAgent string) (*ConvertSensorResponse, error) {
	// First preview to get conversion plan
	preview, err := s.PreviewConvertSensors(mcuID, req)
	if err != nil {
		return nil, err
	}

	var updates []SensorTopicUpdate
	converted := 0
	skipped := 0
	conflicts := 0

	for _, item := range preview.Items {
		if item.Conflict {
			conflicts++
			skipped++
			continue
		}
		if item.IsNewFormat {
			continue // skip already-new sensors
		}
		if item.NewTopic == "" || item.OldTopic == item.NewTopic {
			skipped++
			continue
		}

		newKey := computeTopicKey(item.NewTopic)
		updates = append(updates, SensorTopicUpdate{
			ID:       item.SensorID,
			NewTopic: item.NewTopic,
			NewKey:   newKey,
		})
		converted++
	}

	if len(updates) > 0 {
		if err := s.repo.BatchUpdateSensorTopic(updates); err != nil {
			return nil, err
		}
	}

	// Audit log for each converted sensor
	if s.auditService != nil && len(updates) > 0 {
		for _, u := range updates {
			oldTopic := ""
			for _, item := range preview.Items {
				if item.SensorID == u.ID {
					oldTopic = item.OldTopic
					break
				}
			}
			auditOld := map[string]interface{}{"id": u.ID, "topic": oldTopic}
			auditNew := map[string]interface{}{"id": u.ID, "topic": u.NewTopic}
			if err := s.auditService.Log(userID, "m_sensor", u.ID, "TOPIC_CONVERT", auditOld, auditNew, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log TOPIC_CONVERT for sensor %d: %v", u.ID, err)
			}
		}
	}

	return &ConvertSensorResponse{
		Items:     preview.Items,
		Converted: converted,
		Skipped:   skipped,
		Conflicts: conflicts,
	}, nil
}

// computeTopicKey generates a normalized unique key from a topic string
func computeTopicKey(topic string) string {
	parts := strings.Split(topic, "/")
	var company, vesselCode, location, sensorGroup, sensorNumber string
	switch len(parts) {
	case 5:
		company, vesselCode, location, sensorGroup, sensorNumber = parts[0], parts[1], parts[2], parts[3], parts[4]
	case 7:
		company, vesselCode, location, sensorGroup, sensorNumber = parts[0], parts[3], parts[4], parts[5], parts[6]
	default:
		return strings.ReplaceAll(topic, "/", ":")
	}
	return fmt.Sprintf("%s:%s:%s:%s:%s", company, vesselCode, location, sensorGroup, sensorNumber)
}

func (s *service) DeleteMCU(id uint, userID uint, ipAddress, userAgent string) error {
	mcu, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("MCU not found")
	}

	oldData := buildAuditData(mcu)

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	if s.auditService != nil {
		if err := s.auditService.Log(userID, "m_mcu", mcu.ID, "DELETE", oldData, nil, ipAddress, userAgent); err != nil {
			log.Printf("[AUDIT ERROR] Failed to log DELETE for mcu %d: %v", mcu.ID, err)
		}
	}

	return nil
}

// Helper function to map model to response
func (s *service) mapToResponse(mcu *model.MMCU) *MCUResponse {
	sensors := []SensorInfo{}
	for _, sensor := range mcu.Sensors {
		sensors = append(sensors, SensorInfo{
			ID:            sensor.ID,
			Topic:         sensor.Topic,
			Name:          sensor.Name,
			IsVisible:     sensor.IsVisible,
			EngineName:    sensor.EngineName,
			FlowDirection: sensor.FlowDirection,
			KFactor:       sensor.KFactor,
		})
	}

	return &MCUResponse{
		ID:        mcu.ID,
		VesselID:  mcu.VesselID,
		Name:      mcu.Name,
		Location:  mcu.Location,
		Sensors:   sensors,
		CreatedAt: mcu.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: mcu.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
