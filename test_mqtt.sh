#!/bin/bash
# MQTT Testing Script
# Description: Test script for validating sensor log functionality

MQTT_HOST="192.168.18.88"
MQTT_TOPIC="sensor/rpm/njk"

echo "=========================================="
echo "MQTT Sensor Log Testing Script"
echo "=========================================="
echo ""

# Test 1: Valid message with 2 sensors
echo "Test 1: Publishing valid message with 2 sensors from Vessel VS002"
mosquitto_pub -h $MQTT_HOST -t $MQTT_TOPIC -m '{
  "vessel": "VS002",
  "device": "top",
  "timestamp": "2025-12-23 10:00:00",
  "sensors": [
    {"sensor_id": 1, "rpm": 100},
    {"sensor_id": 2, "rpm": 150}
  ]
}'
echo "✅ Test 1 completed - Check logs for: VS0021, VS0022"
echo ""
sleep 2

# Test 2: Valid message with 4 sensors
echo "Test 2: Publishing valid message with 4 sensors from Vessel VS001"
mosquitto_pub -h $MQTT_HOST -t $MQTT_TOPIC -m '{
  "vessel": "VS001",
  "device": "bottom",
  "timestamp": "2025-12-23 10:01:00",
  "sensors": [
    {"sensor_id": 1, "rpm": 200},
    {"sensor_id": 2, "rpm": 220},
    {"sensor_id": 3, "rpm": 210},
    {"sensor_id": 4, "rpm": 230}
  ]
}'
echo "✅ Test 2 completed - Check logs for: VS0011, VS0012, VS0013, VS0014"
echo ""
sleep 2

# Test 3: Invalid sensor_id (should log error)
echo "Test 3: Publishing message with invalid sensor_id (5)"
mosquitto_pub -h $MQTT_HOST -t $MQTT_TOPIC -m '{
  "vessel": "VS002",
  "device": "top",
  "timestamp": "2025-12-23 10:02:00",
  "sensors": [
    {"sensor_id": 5, "rpm": 300}
  ]
}'
echo "❌ Test 3 completed - Should log error: sensor VS0025 not found"
echo ""
sleep 2

# Test 4: Missing vessel field (should log error)
echo "Test 4: Publishing message without vessel field"
mosquitto_pub -h $MQTT_HOST -t $MQTT_TOPIC -m '{
  "device": "top",
  "timestamp": "2025-12-23 10:03:00",
  "sensors": [
    {"sensor_id": 1, "rpm": 100}
  ]
}'
echo "❌ Test 4 completed - Should log error: missing vessel field"
echo ""
sleep 2

# Test 5: Invalid JSON (should log error)
echo "Test 5: Publishing invalid JSON"
mosquitto_pub -h $MQTT_HOST -t $MQTT_TOPIC -m 'invalid json data'
echo "❌ Test 5 completed - Should log error: invalid JSON format"
echo ""
sleep 2

# Test 6: Mixed valid and invalid sensors
echo "Test 6: Publishing message with mixed valid/invalid sensors"
mosquitto_pub -h $MQTT_HOST -t $MQTT_TOPIC -m '{
  "vessel": "VS001",
  "device": "top",
  "timestamp": "2025-12-23 10:04:00",
  "sensors": [
    {"sensor_id": 1, "rpm": 100},
    {"sensor_id": 99, "rpm": 999},
    {"sensor_id": 2, "rpm": 200}
  ]
}'
echo "⚠️  Test 6 completed - Should save VS0011 and VS0012, error for VS00199"
echo ""

echo "=========================================="
echo "Testing completed!"
echo "=========================================="
echo ""
echo "Check the following:"
echo "1. Backend logs for MQTT messages"
echo "2. Database sensor_logs table for entries"
echo "3. Logs with status='success' should have valid sensor_id"
echo "4. Logs with status='error' should have sensor_id=0"
echo ""
echo "Query to check logs:"
echo "SELECT id, sensor_id, message, status, created_at FROM sensor_logs ORDER BY created_at DESC LIMIT 20;"
