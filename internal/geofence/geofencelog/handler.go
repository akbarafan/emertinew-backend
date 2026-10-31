package geofencelog

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/base"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CheckGeofence godoc
// @Summary Check if vessel is inside or outside geofence (without logging)
// @Description Check vessel position against geofence without creating log entry
// @Tags Geofence Logs
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CheckGeofenceRequest true "Check geofence request"
// @Success 200 {object} map[string]interface{} "success: true, data: CheckGeofenceResponse"
// @Failure 400 {object} map[string]string "error: Invalid request"
// @Router /geofences/check [post]
func (h *Handler) CheckGeofence(c *gin.Context) {
	var req CheckGeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := h.service.CheckGeofence(req)
	if err != nil {
		if errors.Is(err, ErrGeofenceNotFound) {
			base.ErrorResponse(c, http.StatusNotFound, "Geofence tidak ditemukan", err)
			return
		}
		if errors.Is(err, ErrGeofenceInactive) {
			base.ErrorResponse(c, http.StatusBadRequest, "Geofence tidak aktif", err)
			return
		}
		base.ErrorResponse(c, http.StatusInternalServerError, "Gagal melakukan pengecekan geofence", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Berhasil melakukan pengecekan geofence", result)
}

// CreateGeofenceLog godoc
// @Summary Create geofence log entry
// @Description Create log entry with data from frontend (isOutsideGeofence and distance)
// @Tags Geofence Logs
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateGeofenceLogRequest true "Create log request"
// @Success 201 {object} map[string]interface{} "success: true, data: GeofenceLogResponse"
// @Failure 400 {object} map[string]string "error: Invalid request"
// @Router /geofence-logs [post]
func (h *Handler) CreateGeofenceLog(c *gin.Context) {
	var req CreateGeofenceLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := h.service.CreateGeofenceLog(req)
	if err != nil {
		if errors.Is(err, ErrGeofenceNotFound) {
			base.ErrorResponse(c, http.StatusNotFound, "Geofence tidak ditemukan", err)
			return
		}
		if errors.Is(err, ErrGeofenceInactive) {
			base.ErrorResponse(c, http.StatusBadRequest, "Geofence tidak aktif", err)
			return
		}
		base.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat log geofence", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Berhasil membuat log geofence", result)
}

// GetLogsByVesselID godoc
// @Summary Get geofence logs by vessel ID
// @Description Get all geofence check logs for a specific vessel
// @Tags Geofence Logs
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "success: true, data: logs"
// @Failure 400 {object} map[string]string "error: Invalid vessel ID"
// @Router /vessels/{id}/geofence-logs [get]
func (h *Handler) GetLogsByVesselID(c *gin.Context) {
	vesselID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "vessel_id tidak valid", err)
		return
	}

	page, limit, _ := base.GetPaginationParams(c)

	logs, total, err := h.service.GetLogsByVesselID(uint(vesselID), page, limit)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data log", err)
		return
	}

	if page > 0 && limit > 0 {
		base.PaginatedResponse(c, http.StatusOK, "Berhasil mengambil data log", logs, page, limit, total)
	} else {
		base.SuccessResponse(c, http.StatusOK, "Berhasil mengambil data log", logs)
	}
}

// GetLogsByGeofenceID godoc
// @Summary Get geofence logs by geofence ID
// @Description Get all vessel check logs for a specific geofence
// @Tags Geofence Logs
// @Produce json
// @Security Bearer
// @Param id path int true "Geofence ID"
// @Param vesselId query int false "Filter by Vessel ID"
// @Param isOutsideGeofence query bool false "Filter by outside status (true/false)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "success: true, data: logs"
// @Failure 400 {object} map[string]string "error: Invalid geofence ID"
// @Router /geofences/{id}/logs [get]
func (h *Handler) GetLogsByGeofenceID(c *gin.Context) {
	geofenceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "geofence_id tidak valid", err)
		return
	}

	// Parse filters
	var vesselID *uint
	if vesselIDStr := c.Query("vesselId"); vesselIDStr != "" {
		if vid, err := strconv.ParseUint(vesselIDStr, 10, 32); err == nil {
			vidUint := uint(vid)
			vesselID = &vidUint
		}
	}

	var isOutside *bool
	if isOutsideStr := c.Query("isOutsideGeofence"); isOutsideStr != "" {
		if isOut, err := strconv.ParseBool(isOutsideStr); err == nil {
			isOutside = &isOut
		}
	}

	page, limit, _ := base.GetPaginationParams(c)

	logs, total, err := h.service.GetLogsByGeofenceID(uint(geofenceID), vesselID, isOutside, page, limit)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data log", err)
		return
	}

	if page > 0 && limit > 0 {
		base.PaginatedResponse(c, http.StatusOK, "Berhasil mengambil data log", logs, page, limit, total)
	} else {
		base.SuccessResponse(c, http.StatusOK, "Berhasil mengambil data log", logs)
	}
}

// ListAllLogs godoc
// @Summary List all geofence logs with company, vessel, and event filters
// @Tags Geofence Logs
// @Produce json
// @Security Bearer
// @Router /geofence-logs [get]
func (h *Handler) ListAllLogs(c *gin.Context) {
	page, limit, _ := base.GetPaginationParams(c)

	role := c.GetString("role")
	var filterCompID *uint
	if cidVal, exists := c.Get("company_id"); exists {
		switch v := cidVal.(type) {
		case uint:
			if v > 0 { filterCompID = &v }
		case int:
			if v > 0 { u := uint(v); filterCompID = &u }
		}
	}

	// Superadmin / Admin can filter by company
	if role == "superadmin" || role == "admin" || role == "tech_admin" {
		if cidStr := c.Query("company_id"); cidStr != "" && cidStr != "all" {
			if cid, err := strconv.ParseUint(cidStr, 10, 32); err == nil {
				u := uint(cid)
				filterCompID = &u
			}
		} else {
			filterCompID = nil
		}
	}

	var vesselID *uint
	if vidStr := c.Query("vessel_id"); vidStr != "" && vidStr != "all" {
		if vid, err := strconv.ParseUint(vidStr, 10, 32); err == nil {
			u := uint(vid)
			vesselID = &u
		}
	}

	var geofenceID *uint
	if gidStr := c.Query("geofence_id"); gidStr != "" && gidStr != "all" {
		if gid, err := strconv.ParseUint(gidStr, 10, 32); err == nil {
			u := uint(gid)
			geofenceID = &u
		}
	}

	eventType := c.Query("event_type")

	logs, total, err := h.service.ListLogs(filterCompID, vesselID, geofenceID, eventType, page, limit)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data log geofence", err)
		return
	}

	if page > 0 && limit > 0 {
		base.PaginatedResponse(c, http.StatusOK, "Berhasil mengambil data log geofence", logs, page, limit, total)
	} else {
		base.SuccessResponse(c, http.StatusOK, "Berhasil mengambil data log geofence", logs)
	}
}
