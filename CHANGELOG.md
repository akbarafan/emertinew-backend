# Update Summary - Emerti Backend

## Changes Implemented

### 1. ✅ Removed `mac_address` Column from Device Table

**Files Modified:**
- [model/device.go](model/device.go)
- [internal/device/dto.go](internal/device/dto.go)
- [internal/device/service.go](internal/device/service.go)
- [internal/device/repository.go](internal/device/repository.go)

**Changes:**
- Removed `MacAddress` field from `Device` model
- Removed `MacAddress` from `CreateDeviceRequest`, `UpdateDeviceRequest`, and `DeviceResponse` DTOs
- Removed `GetByMacAddress` method from repository
- Removed mac address validation in service layer
- Updated `BeforeDelete` hook to only handle name field

**Migration Required:**
```sql
ALTER TABLE devices DROP COLUMN mac_address;
```

---

### 2. ✅ Added Empty Data Handlers for All Endpoints

**Files Modified:**
- [internal/device/service.go](internal/device/service.go)
- [internal/sensor/service.go](internal/sensor/service.go)
- [internal/user/service.go](internal/user/service.go)
- [internal/vessel/service.go](internal/vessel/service.go)

**Changes:**
When `GetAll` endpoints are called and no data exists, the API will now return:
```json
{
  "status": "error",
  "status_code": 404,
  "message": "Failed to fetch {endpoint}",
  "error": "{endpoint} not created yet"
}
```

Where `{endpoint}` can be: `device`, `sensor`, `user`, or `vessel`

---

### 3. ✅ Status Code and Status in All Responses

**Already Implemented** in [internal/base/dto.go](internal/base/dto.go)

All API responses now include:
- `status`: "success" or "error"
- `status_code`: HTTP status code (200, 201, 400, 404, 500, etc.)
- `message`: Descriptive message
- `data`: Response data (for success)
- `error`: Error details (for error responses)

**Example Success Response:**
```json
{
  "status": "success",
  "status_code": 200,
  "message": "Devices fetched successfully",
  "data": [...]
}
```

**Example Error Response:**
```json
{
  "status": "error",
  "status_code": 404,
  "message": "Failed to fetch devices",
  "error": "device not created yet"
}
```

---

### 4. ✅ MQTT Sensor Log Validation

**File Modified:**
- [internal/sensor/mqtt.go](internal/sensor/mqtt.go)

**New Validation Logic:**

When MQTT receives a message like:
```bash
mosquitto_pub -h 194.59.165.228 -t sensor/rpm/njk -m '{
  "vessel": "VS002",
  "device": "top",
  "timestamp": "2025-12-11 22:42:36",
  "sensors": [
    {"sensor_id": 1, "rpm": 0},
    {"sensor_id": 2, "rpm": 0},
    {"sensor_id": 3, "rpm": 0},
    {"sensor_id": 4, "rpm": 0}
  ]
}'
```

**The system will:**

1. **Parse the JSON message**
   - Extract `vessel` (e.g., "VS002")
   - Extract `sensors` array

2. **For each sensor in the array:**
   - Get `sensor_id` from the sensor data
   - Generate expected code: `{vessel}{sensor_id}`
   - Examples:
     - vessel="VS002", sensor_id=1 → code="VS0021"
     - vessel="VS002", sensor_id=2 → code="VS0022"
     - vessel="VS001", sensor_id=1 → code="VS0011"
     - vessel="VS001", sensor_id=4 → code="VS0014"

3. **Validate against master sensor table:**
   - Check if a sensor exists in `sensors` table with:
     - `code` = expected code (e.g., "VS0021")
     - `topic` = message topic (e.g., "sensor/rpm/njk")
   
4. **Save sensor log:**
   - ✅ **If found**: Save log with matched sensor_id and status="success"
   - ❌ **If not found**: Save error log with sensor_id=0 and status="error"

**Example Master Sensor Table Setup:**
```sql
-- Vessel 1 (VS001) - 4 RPM sensors
INSERT INTO sensors (code, topic, name) VALUES 
('VS0011', 'sensor/rpm/njk', 'Vessel 1 RPM Sensor 1'),
('VS0012', 'sensor/rpm/njk', 'Vessel 1 RPM Sensor 2'),
('VS0013', 'sensor/rpm/njk', 'Vessel 1 RPM Sensor 3'),
('VS0014', 'sensor/rpm/njk', 'Vessel 1 RPM Sensor 4');

-- Vessel 2 (VS002) - 4 RPM sensors  
INSERT INTO sensors (code, topic, name) VALUES
('VS0021', 'sensor/rpm/njk', 'Vessel 2 RPM Sensor 1'),
('VS0022', 'sensor/rpm/njk', 'Vessel 2 RPM Sensor 2'),
('VS0023', 'sensor/rpm/njk', 'Vessel 2 RPM Sensor 3'),
('VS0024', 'sensor/rpm/njk', 'Vessel 2 RPM Sensor 4');
```

**Log Structure:**

Success log entry in `sensor_logs` table:
```json
{
  "sensor_id": 123,
  "message": {
    "vessel": "VS002",
    "sensor_id": 1,
    "code": "VS0021",
    "data": {
      "sensor_id": 1,
      "rpm": 0
    },
    "device": "top",
    "timestamp": "2025-12-11 22:42:36"
  },
  "status": "success"
}
```

Error log entry when sensor not found:
```json
{
  "sensor_id": 0,
  "message": {
    "vessel": "VS002",
    "sensor_id": 5,
    "expected_code": "VS0025",
    "topic": "sensor/rpm/njk",
    "error": "Sensor code not found in master data",
    "data": {
      "sensor_id": 5,
      "rpm": 100
    }
  },
  "status": "error"
}
```

---

### 5. ✅ Error Handler Utilities

**New File Created:**
- [utils/errors.go](utils/errors.go)

**Features:**
- Common error types (ErrNotFound, ErrAlreadyExists, etc.)
- Error code constants with HTTP status codes
- Helper functions to create standardized errors:
  - `NewNotFoundError(resource)`
  - `NewAlreadyExistsError(resource)`
  - `NewNotCreatedYetError(resource)`
  - `NewValidationError(message)`

---

## Testing the Changes

### Test Empty Data Handler
```bash
# Get all devices when none exist
curl -X GET http://localhost:8080/api/device \
  -H "Authorization: Bearer YOUR_TOKEN"

# Expected Response:
{
  "status": "error",
  "status_code": 404,
  "message": "Failed to fetch devices",
  "error": "device not created yet"
}
```

### Test MQTT Validation
```bash
# Publish test data
mosquitto_pub -h 194.59.165.228 -t sensor/rpm/njk -m '{
  "vessel": "VS002",
  "device": "top",
  "timestamp": "2025-12-23 10:00:00",
  "sensors": [
    {"sensor_id": 1, "rpm": 100},
    {"sensor_id": 2, "rpm": 200}
  ]
}'

# Check logs - should see validation messages in console
# Check database - sensor_logs table should have entries with correct sensor_id
```

### Test Device Without MAC Address
```bash
# Create device (no mac_address field)
curl -X POST http://localhost:8080/api/device \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Device 1",
    "sensor_ids": [1, 2, 3]
  }'
```

---

## Database Migration Required

Run these SQL commands to update your database:

```sql
-- Remove mac_address column from devices table
ALTER TABLE devices DROP COLUMN IF EXISTS mac_address;

-- Verify the change
\d devices;
```

---

## Notes

1. **MQTT Message Format**: The message must include `vessel` and `sensors` fields in JSON format
2. **Code Format**: Sensor codes should follow the pattern: `{vessel_code}{sensor_id}` (e.g., VS0021)
3. **Master Data**: Ensure all sensor codes are properly set up in the `sensors` table before publishing MQTT messages
4. **Error Logs**: Invalid messages are logged with status="error" and sensor_id=0 for tracking

---

## Status Codes Reference

| Status Code | Meaning | Usage |
|-------------|---------|-------|
| 200 | OK | Successful GET, PUT, DELETE |
| 201 | Created | Successful POST |
| 400 | Bad Request | Invalid input data |
| 401 | Unauthorized | Missing/invalid token |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found or not created yet |
| 500 | Internal Server Error | Server error |

---

**All changes implemented successfully! ✅**
