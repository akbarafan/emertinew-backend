package sensor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/config"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/redis/go-redis/v9"
)

const (
	RedisSensorChangeChannel = "sensor:changes"
	RedisTopicPrefix         = "mqtt:topics:"
)

// TrackingService interface untuk integrasi dengan tracking
type TrackingService interface {
	CreateTracking(vesselID uint, sensorCode string, sensorID uint, location string, sensorNumber int, kFactor float64, value map[string]interface{}) error
	CreateTrackingByTopicParts(companyCode, vesselCode, location, sensorGroup string, sensorNumber int, sensorID uint, kFactor float64, value map[string]interface{}) error
}

// mqttStats tracks message counts per vessel for periodic summary logging.
// This prevents log flooding when many sensors send data simultaneously.
type mqttStats struct {
	totalMessages atomic.Int64
	totalErrors   atomic.Int64
	perVessel     sync.Map // map[string]*vesselStats
}

type vesselStats struct {
	Messages atomic.Int64
	Errors   atomic.Int64
}

func (s *mqttStats) record(vesselCode string, isError bool) {
	s.totalMessages.Add(1)
	if isError {
		s.totalErrors.Add(1)
	}
	v, _ := s.perVessel.LoadOrStore(vesselCode, &vesselStats{})
	vs := v.(*vesselStats)
	vs.Messages.Add(1)
	if isError {
		vs.Errors.Add(1)
	}
}

func (s *mqttStats) reset() (total, errors int64, perVessel map[string][2]int64) {
	total = s.totalMessages.Swap(0)
	errors = s.totalErrors.Swap(0)
	perVessel = make(map[string][2]int64)
	s.perVessel.Range(func(key, value interface{}) bool {
		vs := value.(*vesselStats)
		msg := vs.Messages.Swap(0)
		err := vs.Errors.Swap(0)
		if msg > 0 || err > 0 {
			perVessel[key.(string)] = [2]int64{msg, err}
		}
		return true
	})
	return
}

type MQTTService struct {
	client          mqtt.Client
	service         Service
	trackingService TrackingService
	redisClient     *redis.Client
	ctx             context.Context
	cancel          context.CancelFunc

	// Label for log identification (e.g. "OLD", "NEW")
	brokerLabel string

	// Map untuk track topic yang sudah disubscribe
	subscribedTopics sync.Map // map[string]bool

	// Map untuk cache sensor by topic
	sensorsByTopic sync.Map // map[string][]SensorResponse

	// Mutex to prevent concurrent ReloadSensors race condition
	reloadMu sync.Mutex

	// Unique instance ID to detect self-sent Redis messages
	instanceID string

	// Message stats for periodic summary logging
	stats mqttStats
}

// NewMQTTService creates a new MQTT service.
// Use label to differentiate brokers in logs (e.g., "OLD", "NEW").
func NewMQTTService(broker string, port string, service Service, label string) *MQTTService {
	var m *MQTTService
	opts := mqtt.NewClientOptions()

	brokerURL := "tcp://" + broker + ":" + port
	opts.AddBroker(brokerURL)
	opts.SetClientID("emerti-" + strings.ToLower(label) + "-" + time.Now().Format("20060102150405"))
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)

	opts.OnConnect = func(client mqtt.Client) {
		log.Printf("[MQTT-%s] Connected to %s", label, brokerURL)
		// Re-subscribe on reconnect: auto-reconnect triggers OnConnect but
		// paho does NOT restore subscriptions with default CleanSession=true.
		// Without this, the backend silently stops receiving MQTT messages
		// until restart.
		m.subscribedTopics = sync.Map{}
		if err := m.SubscribeAll(); err != nil {
			log.Printf("[MQTT-%s] Failed to subscribe after reconnect: %v", label, err)
		}
	}

	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("[MQTT-%s] Connection lost: %v\n", label, err)
	}

	client := mqtt.NewClient(opts)
	ctx, cancel := context.WithCancel(context.Background())

	m = &MQTTService{
		client:           client,
		service:          service,
		trackingService:  nil,
		redisClient:      config.RedisClient,
		ctx:              ctx,
		cancel:           cancel,
		brokerLabel:      label,
		subscribedTopics: sync.Map{},
		sensorsByTopic:   sync.Map{},
		instanceID:       fmt.Sprintf("%s-%d", label, time.Now().UnixNano()),
	}
	return m
}

// SetTrackingService sets the tracking service for MQTT
func (m *MQTTService) SetTrackingService(trackingService TrackingService) {
	m.trackingService = trackingService
	log.Printf("[MQTT-%s] Tracking service integrated", m.brokerLabel)
}

// IsConnected reports whether the underlying MQTT client has an active broker session.
func (m *MQTTService) IsConnected() bool {
	return m != nil && m.client != nil && m.client.IsConnected()
}

// Label returns the human-readable identifier of this broker client.
func (m *MQTTService) Label() string {
	if m == nil {
		return ""
	}
	return m.brokerLabel
}

func (m *MQTTService) Connect() error {
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	log.Printf("[MQTT-%s] Service connected successfully", m.brokerLabel)

	// Start listening to Redis pubsub for sensor changes
	go m.listenSensorChanges()

	// Start periodic stats logger
	go m.logStatsPeriodically()

	return nil
}

// logStatsPeriodically logs a summary of MQTT message stats every 30 seconds.
// This replaces per-message logging to keep logs readable under high throughput.
func (m *MQTTService) logStatsPeriodically() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			total, errors, perVessel := m.stats.reset()
			if total == 0 {
				continue
			}

			// Build per-vessel summary
			var parts []string
			for vessel, counts := range perVessel {
				if counts[1] > 0 {
					parts = append(parts, fmt.Sprintf("%s:%d(%d err)", vessel, counts[0], counts[1]))
				} else {
					parts = append(parts, fmt.Sprintf("%s:%d", vessel, counts[0]))
				}
			}

			if errors > 0 {
				log.Printf("[MQTT-%s] 30s summary: %d messages, %d errors | %s", m.brokerLabel, total, errors, strings.Join(parts, ", "))
			} else {
				log.Printf("[MQTT-%s] 30s summary: %d messages | %s", m.brokerLabel, total, strings.Join(parts, ", "))
			}

		case <-m.ctx.Done():
			return
		}
	}
}

// listenSensorChanges listens to Redis pubsub for sensor changes from other instances
func (m *MQTTService) listenSensorChanges() {
	pubsub := m.redisClient.Subscribe(m.ctx, RedisSensorChangeChannel)
	defer pubsub.Close()

	log.Printf("[MQTT-%s] Listening to sensor changes via Redis", m.brokerLabel)

	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			// Skip self-sent messages to avoid double reload
			var notification struct {
				Instance string `json:"instance"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &notification); err == nil && notification.Instance == m.instanceID {
				continue
			}

			log.Printf("[MQTT-%s] Received sensor change notification: %s", m.brokerLabel, msg.Payload)

			// Reload all sensors and resubscribe
			if err := m.ReloadSensors(); err != nil {
				log.Printf("[MQTT-%s] Failed to reload sensors: %v", m.brokerLabel, err)
			}

		case <-m.ctx.Done():
			log.Printf("[MQTT-%s] Stopping sensor changes listener", m.brokerLabel)
			return
		}
	}
}

// NotifySensorChange reloads sensors immediately AND broadcasts to Redis for other instances.
func (m *MQTTService) NotifySensorChange() error {
	// 1. Reload locally immediately so this instance picks up changes right away
	if err := m.ReloadSensors(); err != nil {
		log.Printf("[MQTT-%s] NotifySensorChange: reload error: %v", m.brokerLabel, err)
	}

	// 2. Broadcast to Redis so other instances also reload
	if m.redisClient != nil {
		message := map[string]interface{}{
			"event":     "sensor_changed",
			"instance":  m.instanceID,
			"timestamp": time.Now().Unix(),
		}
		data, _ := json.Marshal(message)
		if err := m.redisClient.Publish(m.ctx, RedisSensorChangeChannel, string(data)).Err(); err != nil {
			log.Printf("[MQTT-%s] NotifySensorChange: redis publish error: %v", m.brokerLabel, err)
		}
	}

	return nil
}

// ReloadSensors reloads all sensors and updates subscriptions
func (m *MQTTService) ReloadSensors() error {
	m.reloadMu.Lock()
	defer m.reloadMu.Unlock()

	log.Printf("[MQTT-%s] Reloading sensors and subscriptions", m.brokerLabel)

	// Get all sensors from database
	sensors, err := m.service.GetAllSensors()
	if err != nil {
		return err
	}

	// Group sensors by topic
	topicMap := make(map[string][]SensorResponse)
	for _, sensor := range sensors {
		topicMap[sensor.Topic] = append(topicMap[sensor.Topic], sensor)
	}

	// Update cache
	m.sensorsByTopic = sync.Map{}
	for topic, sensorList := range topicMap {
		m.sensorsByTopic.Store(topic, sensorList)
	}

	// Subscribe to new topics
	for topic := range topicMap {
		if _, exists := m.subscribedTopics.Load(topic); !exists {
			if err := m.Subscribe(topic); err != nil {
				log.Printf("[MQTT-%s] Failed to subscribe topic=%s: %v", m.brokerLabel, topic, err)
			}
		}
	}

	// Unsubscribe from topics that no longer exist
	m.subscribedTopics.Range(func(key, value interface{}) bool {
		topic := key.(string)
		if _, exists := topicMap[topic]; !exists {
			if err := m.Unsubscribe(topic); err != nil {
				log.Printf("[MQTT-%s] Failed to unsubscribe topic=%s: %v", m.brokerLabel, topic, err)
			}
		}
		return true
	})

	log.Printf("[MQTT-%s] Sensors reloaded: topics=%d sensors=%d", m.brokerLabel, len(topicMap), len(sensors))
	return nil
}

func (m *MQTTService) Subscribe(topic string) error {
	token := m.client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		m.handleMessage(msg)
	})

	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	m.subscribedTopics.Store(topic, true)
	log.Printf("[MQTT-%s] Subscribed topic=%s", m.brokerLabel, topic)
	return nil
}

func (m *MQTTService) Unsubscribe(topic string) error {
	token := m.client.Unsubscribe(topic)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	m.subscribedTopics.Delete(topic)
	log.Printf("🔕 Unsubscribed from topic: %s", topic)
	return nil
}

// handleMessage processes incoming MQTT messages
// Supports topic format: <company>/<vessel>/<location>/<sensor_group>/<number>
// Example: hcml/vs001/bd/gps/1, hcml/vs001/er/rpm/1
// Payload format: {"time": "...", "lat": -7.2313, "lon": -9.1231}
func (m *MQTTService) handleMessage(msg mqtt.Message) {
	// Skip retained messages to avoid processing old data on startup
	if msg.Retained() {
		return
	}

	rawMessage := string(msg.Payload())

	// Parse message as JSON
	var message map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &message); err != nil {
		log.Printf("[MQTT-%s] Failed to parse JSON topic=%s payload=%q: %v", m.brokerLabel, msg.Topic(), rawMessage, err)
		errorMsg := map[string]interface{}{
			"raw_message": rawMessage,
			"error":       "Invalid JSON format",
			"topic":       msg.Topic(),
		}
		m.service.CreateLogWithRaw(0, errorMsg, rawMessage, 400, "error")
		m.stats.record("unknown", true)
		return
	}

	// Parse topic to extract parts
	topicParts := parseTopicParts(msg.Topic())
	if topicParts != nil {
		if topicParts.Flight != "" {
			// New 7-part format: hcml/flt001/sqd001/vs001/brd/fuel/01f
			m.processNewTopicFormatNew(message, topicParts, rawMessage, msg.Topic())
		} else {
			// Old 5-part format: hcml/vs001/bd/gps/1
			m.processNewTopicFormat(message, topicParts, rawMessage, msg.Topic())
		}
		return
	}

	// Fallback: Invalid topic format
	log.Printf("[MQTT-%s] Invalid topic format: %s", m.brokerLabel, msg.Topic())
	errorMsg := map[string]interface{}{
		"topic": msg.Topic(),
		"error": "Invalid topic format - expected: company/vessel/location/sensor_group/number or company/flight/squad/vessel/board/sensor_group/number",
	}
	m.service.CreateLogWithRaw(0, errorMsg, rawMessage, 400, "error")
	m.stats.record("unknown", true)
}

// TopicParts represents parsed topic components
// Supports two formats:
//
//	Old (5-part):  <company>/<vessel>/<location>/<sensor_group>/<number>
//	               e.g. hcml/vs001/bd/gps/1
//	New (7-part):  <company>/<flight>/<squad>/<vessel>/<board>/<sensor_group>/<number>
//	               e.g. hcml/flt001/sqd001/vs001/brd/fuel/01f
type TopicParts struct {
	Company      string // e.g., "hcml"
	VesselCode   string // e.g., "vs001"
	Location     string // e.g., "bd"/"er" (old), "brd"/"eng" (new)
	SensorGroup  string // e.g., "gps", "rpm", "fuel"
	SensorNumber string // e.g., "1", "01f", "02r"
	// New format only
	Flight string // e.g., "flt001"
	Squad  string // e.g., "sqd001"
}

// parseTopicParts parses topic into structured parts.
// Returns nil if topic doesn't match expected format.
//
//	Old (5-part): hcml/vs001/bd/gps/1
//	New (7-part): hcml/flt001/sqd001/vs001/brd/fuel/01f
func parseTopicParts(topic string) *TopicParts {
	parts := strings.Split(topic, "/")
	switch len(parts) {
	case 5:
		return &TopicParts{
			Company:      parts[0],
			VesselCode:   parts[1],
			Location:     parts[2],
			SensorGroup:  parts[3],
			SensorNumber: parts[4],
		}
	case 7:
		return &TopicParts{
			Company:      parts[0],
			Flight:       parts[1],
			Squad:        parts[2],
			VesselCode:   parts[3],
			Location:     parts[4],
			SensorGroup:  parts[5],
			SensorNumber: parts[6],
		}
	default:
		return nil
	}
}

// processNewTopicFormat handles messages with topic format
// Topic: <company>/<vessel>/<location>/<sensor_group>/<number>
// Payload: {"time": "...", ...sensor_specific_values}
func (m *MQTTService) processNewTopicFormat(message map[string]interface{}, topicParts *TopicParts, rawMessage, topic string) {
	vesselCode := topicParts.VesselCode

	// Get sensor by topic (must be pre-registered)
	sensor, err := m.service.GetSensorByTopic(topic)
	if err != nil || sensor == nil {
		log.Printf("[MQTT-%s] Sensor not registered topic=%s: %v", m.brokerLabel, topic, err)
		errorMsg := map[string]interface{}{
			"topic": topic,
			"error": "Sensor not registered for this topic",
		}
		m.service.CreateLogWithRaw(0, errorMsg, rawMessage, 404, "error")
		m.stats.record(vesselCode, true)
		return
	}

	// Parse sensor number from topic
	sensorNumber := 1
	if n, err := strconv.Atoi(topicParts.SensorNumber); err == nil {
		sensorNumber = n
	}

	// Create tracking data
	if m.trackingService != nil {
		// Get k_factor from sensor for flow calibration
		var kFactor float64
		if sensor.KFactor != nil {
			kFactor = *sensor.KFactor
		}

		if err := m.trackingService.CreateTrackingByTopicParts(
			topicParts.Company,
			topicParts.VesselCode,
			topicParts.Location,
			topicParts.SensorGroup,
			sensorNumber,
			sensor.ID,
			kFactor,
			message,
		); err != nil {
			log.Printf("[MQTT-%s] Tracking error vessel=%s sensor=%s/%d: %v", m.brokerLabel,
				vesselCode, topicParts.SensorGroup, sensorNumber, err)
			m.stats.record(vesselCode, true)
			return
		}
	}

	m.stats.record(vesselCode, false)
}

// processNewTopicFormatNew handles messages with new 7-part topic format
// Topic: <company>/<flight>/<squad>/<vessel>/<board>/<sensor_group>/<number>
// Example: hcml/flt001/sqd001/vs001/brd/fuel/01f
// Payload: {"time": "...", ...sensor_specific_values}
func (m *MQTTService) processNewTopicFormatNew(message map[string]interface{}, topicParts *TopicParts, rawMessage, topic string) {
	vesselCode := topicParts.VesselCode

	// Get sensor by topic (must be pre-registered)
	sensor, err := m.service.GetSensorByTopic(topic)
	if err != nil || sensor == nil {
		log.Printf("[MQTT-%s] Sensor not registered topic=%s: %v", m.brokerLabel, topic, err)
		errorMsg := map[string]interface{}{
			"topic": topic,
			"error": "Sensor not registered for this topic",
		}
		m.service.CreateLogWithRaw(0, errorMsg, rawMessage, 404, "error")
		m.stats.record(vesselCode, true)
		return
	}

	// Parse sensor number from topic
	sensorNumber := 1
	if n, err := strconv.Atoi(topicParts.SensorNumber); err == nil {
		sensorNumber = n
	}

	// Create tracking data
	if m.trackingService != nil {
		var kFactor float64
		if sensor.KFactor != nil {
			kFactor = *sensor.KFactor
		}

		if err := m.trackingService.CreateTrackingByTopicParts(
			topicParts.Company,
			topicParts.VesselCode,
			topicParts.Location,
			topicParts.SensorGroup,
			sensorNumber,
			sensor.ID,
			kFactor,
			message,
		); err != nil {
			log.Printf("[MQTT-%s] Tracking error vessel=%s sensor=%s/%d: %v", m.brokerLabel,
				vesselCode, topicParts.SensorGroup, sensorNumber, err)
			m.stats.record(vesselCode, true)
			return
		}
	}

	m.stats.record(vesselCode, false)
}

// SubscribeAll subscribes to all topics from sensor master data
func (m *MQTTService) SubscribeAll() error {
	return m.ReloadSensors()
}

func (m *MQTTService) Disconnect() {
	// Cancel context to stop listeners
	m.cancel()

	// Unsubscribe from all topics
	m.subscribedTopics.Range(func(key, value interface{}) bool {
		topic := key.(string)
		m.Unsubscribe(topic)
		return true
	})

	// Disconnect MQTT client
	m.client.Disconnect(250)
	log.Printf("[MQTT-%s] Service disconnected", m.brokerLabel)
}
