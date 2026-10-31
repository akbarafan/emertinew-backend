# Emerti Backend API Documentation

## 🚀 Server Info
- **Base URL**: `http://localhost:6969`
- **Port**: 6969

## 👤 Default Admin Account
- **Username**: `admin`
- **Password**: `admin123`
- ⚠️ **Please change password after first login!**

## 📋 API Endpoints

### 1. Public Endpoints (No Auth Required)

#### GET / - Status Check
```bash
curl http://localhost:6969/
```

#### POST /api/auth/login - Login
```bash
curl -X POST http://localhost:6969/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

**PowerShell:**
```powershell
$body = @{ username = "admin"; password = "admin123" } | ConvertTo-Json
Invoke-RestMethod -Uri http://localhost:6969/api/auth/login -Method Post -Body $body -ContentType "application/json"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "role": "admin"
    }
  }
}
```

---

### 2. Protected Endpoints (Require Bearer Token)

**Set token in header:**
```
Authorization: Bearer YOUR_TOKEN_HERE
```

#### GET /api/auth/profile - Get Profile
```bash
curl http://localhost:6969/api/auth/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**PowerShell:**
```powershell
$token = "YOUR_TOKEN_HERE"
$headers = @{ Authorization = "Bearer $token" }
Invoke-RestMethod -Uri http://localhost:6969/api/auth/profile -Headers $headers
```

#### GET /api/sensor - Get All Sensor Logs
```bash
curl http://localhost:6969/api/sensor \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**PowerShell:**
```powershell
Invoke-RestMethod -Uri http://localhost:6969/api/sensor -Headers $headers
```

**Response:**
```json
{
  "success": true,
  "count": 5,
  "data": [
    {
      "id": 5,
      "topic": "sensor/hcsr04",
      "message": "{name: test}",
      "status": "success",
      "created_at": "2025-12-03T19:10:00Z"
    }
  ]
}
```

#### GET /api/sensor/latest - Get Latest Sensor Log
```bash
curl http://localhost:6969/api/sensor/latest \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

### 3. Admin Only Endpoints (Require Admin Role)

#### POST /api/admin/users - Create User
```bash
curl -X POST http://localhost:6969/api/admin/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user1",
    "password": "password123",
    "role": "user"
  }'
```

**PowerShell:**
```powershell
$body = @{ 
  username = "user1"
  password = "password123"
  role = "user"
} | ConvertTo-Json

Invoke-RestMethod -Uri http://localhost:6969/api/admin/users -Method Post -Body $body -ContentType "application/json" -Headers $headers
```

**Response:**
```json
{
  "success": true,
  "message": "User created successfully",
  "data": {
    "id": 2,
    "username": "user1",
    "role": "user"
  }
}
```

#### GET /api/admin/users - Get All Users
```bash
curl http://localhost:6969/api/admin/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**Response:**
```json
{
  "success": true,
  "count": 2,
  "data": [
    {
      "id": 1,
      "username": "admin",
      "role": "admin"
    },
    {
      "id": 2,
      "username": "user1",
      "role": "user"
    }
  ]
}
```

---

## 🧪 Testing Flow

### Step 1: Login
```powershell
$loginBody = @{ username = "admin"; password = "admin123" } | ConvertTo-Json
$loginResponse = Invoke-RestMethod -Uri http://localhost:6969/api/auth/login -Method Post -Body $loginBody -ContentType "application/json"
$token = $loginResponse.data.token
Write-Host "Token: $token"
```

### Step 2: Get Profile
```powershell
$headers = @{ Authorization = "Bearer $token" }
Invoke-RestMethod -Uri http://localhost:6969/api/auth/profile -Headers $headers
```

### Step 3: Get Sensor Data
```powershell
Invoke-RestMethod -Uri http://localhost:6969/api/sensor -Headers $headers
```

### Step 4: Create New User (Admin Only)
```powershell
$newUserBody = @{ 
  username = "testuser"
  password = "test123"
  role = "user"
} | ConvertTo-Json

Invoke-RestMethod -Uri http://localhost:6969/api/admin/users -Method Post -Body $newUserBody -ContentType "application/json" -Headers $headers
```

### Step 5: Get All Users (Admin Only)
```powershell
Invoke-RestMethod -Uri http://localhost:6969/api/admin/users -Headers $headers
```

---

## 📊 Database Tables

### users
```sql
- id (PRIMARY KEY, AUTO INCREMENT)
- username (VARCHAR 100, UNIQUE)
- password (VARCHAR 255, HASHED)
- role (VARCHAR 20, 'admin' or 'user')
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
- deleted_at (TIMESTAMP, nullable)
```

### sensor_logs
```sql
- id (PRIMARY KEY, AUTO INCREMENT)
- topic (VARCHAR 255)
- message (TEXT)
- status (VARCHAR 50, DEFAULT 'success')
- created_at (TIMESTAMP)
```

---

## 🔐 Security Notes

1. **JWT Token** expires in 7 days
2. **Password** is hashed using bcrypt
3. **Admin role** required for user management
4. **Bearer token** required for all protected routes
5. Change default admin password immediately!

---

## 🐛 Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request: username is required"
}
```

### 401 Unauthorized
```json
{
  "error": "invalid credentials"
}
```

### 403 Forbidden
```json
{
  "error": "Admin access required"
}
```

### 404 Not Found
```json
{
  "error": "No sensor logs found"
}
```

---

## 📡 MQTT Testing

Send test sensor data:
```bash
mosquitto_pub -h 194.59.165.228 -t sensor/hcsr04 -m '{"sensor":"hcsr04","value":123}'
```

Data will be saved automatically to database and can be retrieved via `/api/sensor` endpoint (with authentication).
