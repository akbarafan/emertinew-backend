# Contoh Create Data - Sensor & Device

## 1. Create Sensor Master Data

### Endpoint
```
POST /api/sensor/master
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json
```

### Request Body
```json
{
  "code": "VS0011",
  "topic": "sensor/rpm/njk",
  "name": "Vessel 1 RPM Sensor 1"
}
```

### cURL Example
```bash
curl -X POST http://localhost:8080/api/sensor/master \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "VS0011",
    "topic": "sensor/rpm/njk",
    "name": "Vessel 1 RPM Sensor 1"
  }'
```

### Response (201 Created)
```json
{
  "status": "success",
  "status_code": 201,
  "message": "Sensor created successfully",
  "data": {
    "id": 1,
    "code": "VS0011",
    "topic": "sensor/rpm/njk",
    "name": "Vessel 1 RPM Sensor 1"
  }
}
```

### Contoh Multiple Sensors (Vessel 1 & 2)
```bash
# Vessel 1 - Sensor 1
curl -X POST http://localhost:8080/api/sensor/master \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code": "VS0011", "topic": "sensor/rpm/njk", "name": "Vessel 1 RPM Sensor 1"}'

# Vessel 1 - Sensor 2
curl -X POST http://localhost:8080/api/sensor/master \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code": "VS0012", "topic": "sensor/rpm/njk", "name": "Vessel 1 RPM Sensor 2"}'

# Vessel 1 - Sensor 3
curl -X POST http://localhost:8080/api/sensor/master \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code": "VS0013", "topic": "sensor/rpm/njk", "name": "Vessel 1 RPM Sensor 3"}'

# Vessel 1 - Sensor 4
curl -X POST http://localhost:8080/api/sensor/master \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code": "VS0014", "topic": "sensor/rpm/njk", "name": "Vessel 1 RPM Sensor 4"}'

# Vessel 2 - Sensor 1
curl -X POST http://localhost:8080/api/sensor/master \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code": "VS0021", "topic": "sensor/rpm/njk", "name": "Vessel 2 RPM Sensor 1"}'

# Vessel 2 - Sensor 2
curl -X POST http://localhost:8080/api/sensor/master \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code": "VS0022", "topic": "sensor/rpm/njk", "name": "Vessel 2 RPM Sensor 2"}'

# Vessel 2 - Sensor 3
curl -X POST http://localhost:8080/api/sensor/master \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code": "VS0023", "topic": "sensor/rpm/njk", "name": "Vessel 2 RPM Sensor 3"}'

# Vessel 2 - Sensor 4
curl -X POST http://localhost:8080/api/sensor/master \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code": "VS0024", "topic": "sensor/rpm/njk", "name": "Vessel 2 RPM Sensor 4"}'
```

---

## 2. Create Device

### Endpoint
```
POST /api/device
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json
```

### Request Body
```json
{
  "name": "Device Top Vessel 1",
  "sensor_ids": [1, 2, 3, 4],
  "type": "top"
}
```

### Field Descriptions
- `name` (required): Nama device
- `sensor_ids` (required): Array of sensor IDs (dari tabel sensors)
- `type` (required): "top" atau "bottom"

### cURL Example
```bash
curl -X POST http://localhost:8080/api/device \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Device Top Vessel 1",
    "sensor_ids": [1, 2, 3, 4],
    "type": "top"
  }'
```

### Response (201 Created)
```json
{
  "status": "success",
  "status_code": 201,
  "message": "Device created successfully",
  "data": {
    "id": 1,
    "name": "Device Top Vessel 1",
    "sensor_ids": [1, 2, 3, 4],
    "type": "top",
    "created_at": "2025-12-23 10:00:00",
    "updated_at": "2025-12-23 10:00:00"
  }
}
```

### Contoh Multiple Devices
```bash
# Device 1 - Top
curl -X POST http://localhost:8080/api/device \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Device Top Vessel 1",
    "sensor_ids": [1, 2, 3, 4],
    "type": "top"
  }'

# Device 2 - Bottom
curl -X POST http://localhost:8080/api/device \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Device Bottom Vessel 1",
    "sensor_ids": [5, 6, 7, 8],
    "type": "bottom"
  }'
```

---

## 3. Error Examples

### Error: Invalid Type
```json
{
  "name": "Device Test",
  "sensor_ids": [1, 2],
  "type": "middle"
}
```

**Response (400 Bad Request):**
```json
{
  "status": "error",
  "status_code": 400,
  "message": "Invalid request",
  "error": "Key: 'CreateDeviceRequest.Type' Error:Field validation for 'Type' failed on the 'oneof' tag"
}
```

### Error: Missing Required Field
```json
{
  "name": "Device Test",
  "sensor_ids": [1, 2]
}
```

**Response (400 Bad Request):**
```json
{
  "status": "error",
  "status_code": 400,
  "message": "Invalid request",
  "error": "Key: 'CreateDeviceRequest.Type' Error:Field validation for 'Type' failed on the 'required' tag"
}
```

### Error: Duplicate Sensor Code
```json
{
  "code": "VS0011",
  "topic": "sensor/rpm/njk",
  "name": "Duplicate Sensor"
}
```

**Response (400 Bad Request):**
```json
{
  "status": "error",
  "status_code": 400,
  "message": "Failed to create sensor",
  "error": "Sensor code already exists"
}
```

---

## 4. Using Postman

### Setup
1. Create new request
2. Method: POST
3. URL: `http://localhost:8080/api/device`
4. Headers:
   - `Authorization`: `Bearer YOUR_TOKEN`
   - `Content-Type`: `application/json`
5. Body → raw → JSON:
```json
{
  "name": "Device Top Vessel 1",
  "sensor_ids": [1, 2, 3, 4],
  "type": "top"
}
```

---

## 5. Complete Flow Example

```bash
# Step 1: Login to get token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your_password"
  }'

# Response akan berisi token
# Simpan token dari response

# Step 2: Create Sensors
TOKEN="your_token_here"

curl -X POST http://localhost:8080/api/sensor/master \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code": "VS0011", "topic": "sensor/rpm/njk", "name": "Vessel 1 RPM Sensor 1"}'

curl -X POST http://localhost:8080/api/sensor/master \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code": "VS0012", "topic": "sensor/rpm/njk", "name": "Vessel 1 RPM Sensor 2"}'

# Step 3: Create Device
curl -X POST http://localhost:8080/api/device \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Device Top Vessel 1",
    "sensor_ids": [1, 2],
    "type": "top"
  }'

# Step 4: Verify
curl -X GET http://localhost:8080/api/device \
  -H "Authorization: Bearer $TOKEN"
```

---

## Notes

- ✅ `sensor_ids` harus berisi ID yang sudah ada di tabel sensors
- ✅ `type` hanya menerima nilai "top" atau "bottom"
- ✅ `code` sensor harus unique
- ✅ Format code sensor: {vessel}{sensor_id} (contoh: VS0011)
- ✅ Token didapat dari endpoint login
