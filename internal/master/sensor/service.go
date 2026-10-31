package sensor

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/datatypes"
)

type MQTTNotifier interface {
	NotifySensorChange() error
	ReloadSensors() error
}

type Service interface {
	// Sensor Data (TimescaleDB hypertable)
	CreateData(sensorID uint, message map[string]interface{}) error
	CreateLogWithRaw(sensorID uint, message map[string]interface{}, rawMessage string, statusCode int, statusMsg string) error // Kept for backward compatibility
	GetAllData(limit int, sensorID uint) ([]SensorDataResponse, error)
	GetDataBySensorID(sensorID uint) ([]SensorDataResponse, error)
	GetLatestDataBySensorID(sensorID uint) (*model.SensorData, error)

	// Sensor Master CRUD
	CreateSensor(req CreateSensorRequest, userID uint, ipAddress, userAgent string) (*SensorResponse, error)
	GetSensorByID(id uint) (*SensorResponse, error)
	GetSensorByTopic(topic string) (*SensorResponse, error)
	GetAllSensors() ([]SensorResponse, error)
	UpdateSensor(id uint, req UpdateSensorRequest, userID uint, ipAddress, userAgent string) (*SensorResponse, error)
	DeleteSensor(id uint, userID uint, ipAddress, userAgent string) error
	ToggleVisibility(id uint, req ToggleVisibilityRequest) (*SensorResponse, error)
	GetVesselSensorVisibility(vesselID uint) (*VesselSensorVisibilityResponse, error)

	// Flow Mapping
	GetEngineMapping(vesselID uint) (*VesselEngineMappingResponse, error)

	// MQTT Integration
	AddMQTTNotifier(notifier MQTTNotifier)
}

type service struct {
	repo          Repository
	mqttNotifiers []MQTTNotifier
	auditService  audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{
		repo:          repo,
		mqttNotifiers: nil,
		auditService:  auditService,
	}
}

// computeTopicKey generates a normalized unique key from a topic string.
// Format: company:vessel_code:location:sensor_group:sensor_number
// For 5-part topics: key = topic (with / replaced by :)
// For 7-part topics: flight/squad are stripped
// For unrecognized formats: falls back to the raw topic (with / replaced by :) for compatibility
func computeTopicKey(topic string) string {
	parts := parseTopicParts(topic)
	if parts != nil {
		return strings.Join([]string{
			parts.Company,
			parts.VesselCode,
			parts.Location,
			parts.SensorGroup,
			parts.SensorNumber,
		}, ":")
	}
	// Fallback for unrecognized format: just normalize separators
	return strings.ReplaceAll(topic, "/", ":")
}

func (s *service) mapToResponse(sensor *model.MSensor) *SensorResponse {
	if sensor == nil {
		return nil
	}
	return &SensorResponse{
		ID:            sensor.ID,
		MCUID:         sensor.MCUID,
		Topic:         sensor.Topic,
		Name:          sensor.Name,
		KFactor:       sensor.KFactor,
		IsVisible:     sensor.IsVisible,
		EngineName:    sensor.EngineName,
		FlowDirection: sensor.FlowDirection,
	}
}

func (s *service) buildAuditData(sensor *model.MSensor) SensorAuditData {
	return SensorAuditData{
		ID:            sensor.ID,
		MCUID:         sensor.MCUID,
		Topic:         sensor.Topic,
		Name:          sensor.Name,
		KFactor:       sensor.KFactor,
		EngineName:    sensor.EngineName,
		FlowDirection: sensor.FlowDirection,
	}
}

func (s *service) AddMQTTNotifier(notifier MQTTNotifier) {
	s.mqttNotifiers = append(s.mqttNotifiers, notifier)
	log.Printf("[SENSOR] MQTT notifier registered (total %d)", len(s.mqttNotifiers))
}

// CreateData saves sensor data to sensor_data table (new TimescaleDB structure)
func (s *service) CreateData(sensorID uint, message map[string]interface{}) error {
	data := &model.SensorData{
		SensorID: sensorID,
		Value:    datatypes.JSON(mustMarshal(message)),
		Time:     time.Now(),
	}
	return s.repo.CreateData(data)
}

// CreateLogWithRaw kept for backward compatibility with MQTT handler
func (s *service) CreateLogWithRaw(sensorID uint, message map[string]interface{}, rawMessage string, statusCode int, statusMsg string) error {
	// Ignore rawMessage and statusCode - just save to sensor_data
	if sensorID == 0 {
		// If no sensor ID, skip saving (error case)
		log.Printf("[SENSOR] Skipping save - no valid sensor ID")
		return nil
	}
	return s.CreateData(sensorID, message)
}

func (s *service) GetAllData(limit int, sensorID uint) ([]SensorDataResponse, error) {
	data, err := s.repo.GetAll(limit, sensorID)
	if err != nil {
		return nil, err
	}

	var response []SensorDataResponse
	for _, d := range data {
		response = append(response, s.buildDataResponse(&d))
	}

	return response, nil
}

func (s *service) GetDataBySensorID(sensorID uint) ([]SensorDataResponse, error) {
	data, err := s.repo.GetBySensorID(sensorID)
	if err != nil {
		return nil, err
	}

	var response []SensorDataResponse
	for _, d := range data {
		response = append(response, s.buildDataResponse(&d))
	}

	return response, nil
}

func (s *service) GetLatestDataBySensorID(sensorID uint) (*model.SensorData, error) {
	return s.repo.GetLatestBySensorID(sensorID)
}

// buildDataResponse builds sensor data response from Value JSON
func (s *service) buildDataResponse(data *model.SensorData) SensorDataResponse {
	var value map[string]interface{}
	mustUnmarshal(data.Value, &value)

	// Get sensor and vessel info
	var vesselID *uint
	var vesselName string
	var topic string
	var sensorName string

	// Get sensor by ID to fetch related data
	sensor, err := s.repo.GetSensorByID(data.SensorID)
	if err == nil && sensor != nil {
		topic = sensor.Topic
		sensorName = sensor.Name

		// Get MCU info if sensor has MCU
		if sensor.MCUID != nil {
			mcu, err := s.repo.GetMCUByID(*sensor.MCUID)
			if err == nil && mcu != nil && mcu.VesselID != nil {
				vesselID = mcu.VesselID

				// Get vessel info
				vessel, err := s.repo.GetVesselByID(*mcu.VesselID)
				if err == nil && vessel != nil {
					vesselName = vessel.Name
				}
			}
		}
	}

	return SensorDataResponse{
		Time:       data.Time,
		SensorID:   data.SensorID,
		Topic:      topic,
		SensorName: sensorName,
		Value:      value,
		TrackingID: data.TrackingID,
		VesselID:   vesselID,
		VesselName: vesselName,
	}
}

// Sensor Master CRUD implementations
func (s *service) CreateSensor(req CreateSensorRequest, userID uint, ipAddress, userAgent string) (*SensorResponse, error) {
	// Check if topic already exists (topic is unique)
	_, err := s.repo.GetSensorByTopic(req.Topic)
	if err == nil {
		return nil, errors.New("Sensor with this topic already exists")
	}

	if err := validateSensorConfig(req.Topic, req.KFactor, req.EngineName, req.FlowDirection); err != nil {
		return nil, err
	}

	sensor := &model.MSensor{
		MCUID:         req.MCUID,
		Topic:         req.Topic,
		TopicKey:      computeTopicKey(req.Topic),
		Name:          req.Name,
		KFactor:       req.KFactor,
		EngineName:    normaliseOptionalString(req.EngineName),
		FlowDirection: normaliseOptionalString(req.FlowDirection),
	}

	if req.IsVisible != nil {
		sensor.IsVisible = *req.IsVisible
	}

	err = s.repo.CreateSensor(sensor)
	if err != nil {
		return nil, err
	}

	// Audit log
	if s.auditService != nil {
		newData := s.buildAuditData(sensor)
		if err := s.auditService.Log(userID, "m_sensor", sensor.ID, "CREATE", nil, newData, ipAddress, userAgent); err != nil {
			log.Printf("[AUDIT ERROR] Failed to log sensor create: %v", err)
		}
	}

	// Notify MQTT service about sensor change
	for _, n := range s.mqttNotifiers {
		if err := n.NotifySensorChange(); err != nil {
			log.Printf("[SENSOR] Failed to notify MQTT: %v", err)
		}
	}

	return s.mapToResponse(sensor), nil
}

func (s *service) GetSensorByID(id uint) (*SensorResponse, error) {
	sensor, err := s.repo.GetSensorByID(id)
	if err != nil {
		return nil, errors.New("Sensor not added yet")
	}

	return s.mapToResponse(sensor), nil
}

func (s *service) GetSensorByTopic(topic string) (*SensorResponse, error) {
	sensor, err := s.repo.GetSensorByTopic(topic)
	if err != nil {
		return nil, errors.New("Sensor with topic not found")
	}

	return &SensorResponse{
		ID:      sensor.ID,
		MCUID:   sensor.MCUID,
		Topic:   sensor.Topic,
		Name:    sensor.Name,
		KFactor: sensor.KFactor,
	}, nil
}

func (s *service) GetAllSensors() ([]SensorResponse, error) {
	sensors, err := s.repo.GetAllSensors()
	if err != nil {
		return nil, err
	}

	// Return empty array instead of error when no data exists
	if len(sensors) == 0 {
		return []SensorResponse{}, nil
	}

	var response []SensorResponse
	for _, sensor := range sensors {
		resp := s.mapToResponse(&sensor)
		if resp != nil {
			response = append(response, *resp)
		}
	}

	return response, nil
}

func (s *service) UpdateSensor(id uint, req UpdateSensorRequest, userID uint, ipAddress, userAgent string) (*SensorResponse, error) {
	sensor, err := s.repo.GetSensorByID(id)
	if err != nil {
		return nil, errors.New("Sensor not found")
	}

	// Capture old data for audit
	oldData := s.buildAuditData(sensor)

	if req.MCUID != nil {
		sensor.MCUID = req.MCUID
	}
	if req.Topic != "" {
		sensor.Topic = req.Topic
		sensor.TopicKey = computeTopicKey(req.Topic)
	}
	if req.Name != "" {
		sensor.Name = req.Name
	}
	if req.KFactor != nil {
		sensor.KFactor = req.KFactor
	}
	if req.IsVisible != nil {
		sensor.IsVisible = *req.IsVisible
	}
	// `Present` distinguishes "field not sent" from "field explicitly nulled".
	// Without it, picking "Single" in MCU Management could never clear a
	// direction that had been set before: JSON null unmarshals into a nil
	// pointer, which was indistinguishable from an absent field.
	if req.EngineName.Present {
		sensor.EngineName = normaliseOptionalString(req.EngineName.Value)
	}
	if req.FlowDirection.Present {
		sensor.FlowDirection = normaliseOptionalString(req.FlowDirection.Value)
	}

	if err := validateSensorConfig(sensor.Topic, sensor.KFactor, sensor.EngineName, sensor.FlowDirection); err != nil {
		return nil, err
	}

	err = s.repo.UpdateSensor(sensor)
	if err != nil {
		return nil, err
	}

	// Audit log
	if s.auditService != nil {
		newData := s.buildAuditData(sensor)
		if err := s.auditService.Log(userID, "m_sensor", sensor.ID, "UPDATE", oldData, newData, ipAddress, userAgent); err != nil {
			log.Printf("[AUDIT ERROR] Failed to log sensor update: %v", err)
		}
	}

	// Notify MQTT service about sensor change
	for _, n := range s.mqttNotifiers {
		if err := n.NotifySensorChange(); err != nil {
			log.Printf("[SENSOR] Failed to notify MQTT: %v", err)
		}
	}

	return s.mapToResponse(sensor), nil
}

func (s *service) ToggleVisibility(id uint, req ToggleVisibilityRequest) (*SensorResponse, error) {
	sensor, err := s.repo.GetSensorByID(id)
	if err != nil {
		return nil, errors.New("Sensor not found")
	}

	sensor.IsVisible = req.IsVisible
	if err := s.repo.UpdateSensor(sensor); err != nil {
		return nil, err
	}

	// Notify MQTT service about sensor change
	for _, n := range s.mqttNotifiers {
		if err := n.NotifySensorChange(); err != nil {
			log.Printf("[SENSOR] Failed to notify MQTT: %v", err)
		}
	}

	return s.mapToResponse(sensor), nil
}

// fieldName maps a sensor topic's group+number+location to a visibility key.
// - rpm, flow, door: uses group + parsed number (default 1)
// - ac, dc, bat, temp, humidity: uses group + normalized location (top/bot)
// - dc/3 is special: maps to bat_top/bat_bot (Battery panel)
// - Other groups (accel, gps): uses group name only
func fieldName(sensorGroup, sensorNumber, location string) string {
	switch sensorGroup {
	case "rpm", "flow", "door":
		num := 1
		if n, err := strconv.Atoi(sensorNumber); err == nil {
			num = n
		}
		return sensorGroup + strconv.Itoa(num)
	case "ac", "dc", "bat", "temp", "humidity":
		loc := strings.ToLower(location)
		side := "bot"
		if loc == "bd" || loc == "brd" || loc == "mdb" {
			side = "top"
		}
		if sensorGroup == "dc" && sensorNumber == "3" {
			return "bat_" + side
		}
		return sensorGroup + "_" + side
	default:
		return sensorGroup
	}
}

func (s *service) GetVesselSensorVisibility(vesselID uint) (*VesselSensorVisibilityResponse, error) {
	sensors, err := s.repo.GetSensorsByVesselID(vesselID)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]bool)
	for _, sensor := range sensors {
		parts := parseTopicParts(sensor.Topic)
		if parts == nil {
			continue
		}
		fn := fieldName(parts.SensorGroup, parts.SensorNumber, parts.Location)
		if sensor.IsVisible {
			fields[fn] = true
		} else if _, exists := fields[fn]; !exists {
			fields[fn] = false
		}
	}

	return &VesselSensorVisibilityResponse{
		VesselID: vesselID,
		Fields:   fields,
	}, nil
}

// GetEngineMapping groups the sensors of a vessel by the engine they were
// labelled with in MCU Management.
//
// Both the flow meters and the tachometers can carry a label, so the result
// tells the client which slot is a fuel meter and which is a tachometer. The
// channel index cannot be used for that: the flow channels are ordered by the
// meters of the tanks while the RPM channels are ordered by the tachometer
// pins, and the two orders are independent of each other.
//
// `visibility_fields` keeps listing the flow slots only, so the charts of the
// daily report behave exactly as before.
func (s *service) GetEngineMapping(vesselID uint) (*VesselEngineMappingResponse, error) {
	sensors, err := s.repo.GetSensorsByVesselID(vesselID)
	if err != nil {
		return nil, err
	}

	// Group the labelled sensors by engine name.
	engineMap := make(map[string][]model.MSensor)
	for _, sensor := range sensors {
		parts := parseTopicParts(sensor.Topic)
		if parts == nil {
			continue
		}
		if parts.SensorGroup != "flow" && parts.SensorGroup != "rpm" {
			continue
		}
		if sensor.EngineName == nil || *sensor.EngineName == "" {
			continue // unconfigured sensor - skip
		}
		engineName := *sensor.EngineName
		engineMap[engineName] = append(engineMap[engineName], sensor)
	}

	var engines []EngineMapping
	for name, engineSensors := range engineMap {
		var mapped []EngineSensorMapping
		visibilitySet := make(map[string]bool)
		for _, s := range engineSensors {
			parts := parseTopicParts(s.Topic)
			if parts == nil {
				continue
			}
			slot := fieldName(parts.SensorGroup, parts.SensorNumber, parts.Location)
			dir := ""
			if parts.SensorGroup == "flow" && s.FlowDirection != nil {
				dir = *s.FlowDirection
			}
			if parts.SensorGroup == "flow" {
				visibilitySet[slot] = true
			}
			mapped = append(mapped, EngineSensorMapping{
				SensorID:   s.ID,
				Topic:      s.Topic,
				Slot:       slot,
				Kind:       parts.SensorGroup,
				Direction:  dir,
				SensorName: s.Name,
			})
		}
		var visFields []string
		for k := range visibilitySet {
			visFields = append(visFields, k)
		}
		engines = append(engines, EngineMapping{
			Name:             name,
			Sensors:          mapped,
			VisibilityFields: visFields,
		})
	}

	// The engines were collected in a map, so their order was random and the
	// cards on screen jumped around on every reload. They are ordered by the
	// channel they start on, which is the physical order of the meters, and
	// then by name so the result is stable.
	sort.SliceStable(engines, func(i, j int) bool {
		li, lj := lowestChannelIndex(engines[i].Sensors), lowestChannelIndex(engines[j].Sensors)
		if li != lj {
			return li < lj
		}
		return engines[i].Name < engines[j].Name
	})

	if engines == nil {
		engines = []EngineMapping{}
	}

	return &VesselEngineMappingResponse{
		VesselID: vesselID,
		Engines:  engines,
	}, nil
}

// lowestChannelIndex returns the numeric part of the lowest flow slot an engine
// uses ("flow3" -> 3). Slots that are not a numbered flow channel sort last.
func lowestChannelIndex(sensors []EngineSensorMapping) int {
	lowest := math.MaxInt
	for _, s := range sensors {
		if len(s.Slot) < 5 || s.Slot[:4] != "flow" {
			continue
		}
		index, err := strconv.Atoi(s.Slot[4:])
		if err != nil {
			continue
		}
		if index < lowest {
			lowest = index
		}
	}
	return lowest
}

func (s *service) DeleteSensor(id uint, userID uint, ipAddress, userAgent string) error {
	sensor, err := s.repo.GetSensorByID(id)
	if err != nil {
		return errors.New("Sensor not found")
	}

	// Capture old data for audit
	oldData := s.buildAuditData(sensor)

	// Just soft delete - topic suffix is handled by repository
	err = s.repo.DeleteSensor(id)
	if err != nil {
		return err
	}

	// Audit log
	if s.auditService != nil {
		if err := s.auditService.Log(userID, "m_sensor", id, "DELETE", oldData, nil, ipAddress, userAgent); err != nil {
			log.Printf("[AUDIT ERROR] Failed to log sensor delete: %v", err)
		}
	}

	// Notify MQTT service about sensor change
	for _, n := range s.mqttNotifiers {
		if err := n.NotifySensorChange(); err != nil {
			log.Printf("[SENSOR] Failed to notify MQTT: %v", err)
		}
	}

	return nil
}

// Helper functions for JSON marshaling
func mustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

func mustUnmarshal(data []byte, v interface{}) {
	_ = json.Unmarshal(data, v)
}
