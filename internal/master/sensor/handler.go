package sensor

import (
	"net/http"
	"strconv"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/base"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetSensorData mendapatkan semua sensor data
// @Summary      Get all sensor data
// @Description  Mendapatkan semua sensor data dari tabel sensor_data (memerlukan authentication)
// @Tags         Sensor Data
// @Produce      json
// @Security     Bearer
// @Param        limit query int false "Limit jumlah data (default: 100)"
// @Param        sensor_id query int false "Filter by sensor ID"
// @Success      200 {object} map[string]interface{} "success: true, data: sensor data array"
// @Failure      401 {object} map[string]string "error: Unauthorized"
// @Failure      500 {object} map[string]string "error: Failed to fetch sensor data"
// @Router       /sensor-data [get]
func (h *Handler) GetSensorData(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	sensorIDStr := c.Query("sensor_id")
	var sensorID uint
	if sensorIDStr != "" {
		id, err := strconv.ParseUint(sensorIDStr, 10, 32)
		if err == nil {
			sensorID = uint(id)
		}
	}

	data, err := h.service.GetAllData(limit, sensorID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch sensor data", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Sensor data fetched successfully", data)
}

// GetSensorDataBySensorID mendapatkan sensor data berdasarkan sensor ID
// @Summary      Get sensor data by sensor ID
// @Description  Mendapatkan semua sensor data yang terkait dengan sensor ID tertentu (memerlukan authentication)
// @Tags         Sensor Data
// @Produce      json
// @Security     Bearer
// @Param        sensor_id path int true "Sensor ID"
// @Success      200 {object} map[string]interface{} "success: true, data: sensor data array"
// @Failure      400 {object} map[string]string "error: Invalid sensor ID"
// @Failure      401 {object} map[string]string "error: Unauthorized"
// @Failure      404 {object} map[string]string "error: No sensor data found for this sensor ID"
// @Router       /sensor-data/{sensor_id} [get]
func (h *Handler) GetSensorDataBySensorID(c *gin.Context) {
	sensorIDStr := c.Param("sensor_id")
	sensorID, err := strconv.ParseUint(sensorIDStr, 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid sensor ID", err)
		return
	}

	data, err := h.service.GetDataBySensorID(uint(sensorID))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "No sensor data found for this sensor ID", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Sensor data fetched successfully", data)
}

// CreateSensor godoc
// @Summary Create sensor
// @Description Create a new sensor master data
// @Tags Master Data - Sensor
// @Accept json
// @Produce json
// @Security Bearer
// @Param sensor body CreateSensorRequest true "Sensor data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/sensor [post]
func (h *Handler) CreateSensor(c *gin.Context) {
	var req CreateSensorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.CreateSensor(req, userID, ipAddress, userAgent)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create sensor", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Sensor created successfully", response)
}

// GetSensor godoc
// @Summary Get sensor by ID
// @Description Get sensor master data by ID
// @Tags Master Data - Sensor
// @Produce json
// @Security Bearer
// @Param id path int true "Sensor ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/sensor/{id} [get]
func (h *Handler) GetSensor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid sensor ID", err)
		return
	}

	response, err := h.service.GetSensorByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Sensor not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Sensor fetched successfully", response)
}

// GetAllSensors godoc
// @Summary Get all sensors
// @Description Get all sensor master data
// @Tags Master Data - Sensor
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /master/sensor [get]
func (h *Handler) GetAllSensors(c *gin.Context) {
	response, err := h.service.GetAllSensors()
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch sensors", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Sensors fetched successfully", response)
}

// UpdateSensor godoc
// @Summary Update sensor
// @Description Update sensor master data by ID
// @Tags Master Data - Sensor
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Sensor ID"
// @Param sensor body UpdateSensorRequest true "Sensor data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/sensor/{id} [patch]
func (h *Handler) UpdateSensor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid sensor ID", err)
		return
	}

	var req UpdateSensorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.UpdateSensor(uint(id), req, userID, ipAddress, userAgent)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update sensor", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Sensor updated successfully", response)
}

// GetVesselSensorVisibility godoc
// @Summary Get sensor visibility map for a vessel
// @Description Returns which tracking fields (rpm1, flow1, etc.) are visible for the daily report
// @Tags Master Data - Sensor
// @Produce json
// @Security Bearer
// @Param vessel_id path int true "Vessel ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/vessels/{vessel_id}/sensor-visibility [get]
func (h *Handler) GetVesselSensorVisibility(c *gin.Context) {
	vesselID, err := strconv.ParseUint(c.Param("vessel_id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	response, err := h.service.GetVesselSensorVisibility(uint(vesselID))
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to get visibility", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Visibility fetched successfully", response)
}

// ToggleVisibility godoc
// @Summary Toggle sensor visibility
// @Description Toggle whether a sensor appears in daily report charts
// @Tags Master Data - Sensor
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Sensor ID"
// @Param body body ToggleVisibilityRequest true "Visibility flag"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/sensor/{id}/visibility [patch]
func (h *Handler) ToggleVisibility(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid sensor ID", err)
		return
	}

	var req ToggleVisibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	response, err := h.service.ToggleVisibility(uint(id), req)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to toggle visibility", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Visibility updated successfully", response)
}

// GetEngineMapping godoc
// @Summary Get engine mapping for a vessel
// @Description Returns the sensors of a vessel grouped by the engine they were labelled with in MCU Management
// @Tags Master Data - Sensor
// @Produce json
// @Security Bearer
// @Param vessel_id path int true "Vessel ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/vessels/{vessel_id}/engine-mapping [get]
func (h *Handler) GetEngineMapping(c *gin.Context) {
	vesselID, err := strconv.ParseUint(c.Param("vessel_id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	response, err := h.service.GetEngineMapping(uint(vesselID))
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to get engine mapping", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Engine mapping fetched successfully", response)
}

// DeleteSensor godoc
// @Summary Delete sensor
// @Description Soft delete sensor master data by ID
// @Tags Master Data - Sensor
// @Produce json
// @Security Bearer
// @Param id path int true "Sensor ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/sensor/{id} [delete]
func (h *Handler) DeleteSensor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid sensor ID", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	err = h.service.DeleteSensor(uint(id), userID, ipAddress, userAgent)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to delete sensor", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Sensor deleted successfully", nil)
}
