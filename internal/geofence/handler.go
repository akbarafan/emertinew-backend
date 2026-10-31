package geofence

import (
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

func extractScope(c *gin.Context) (string, *uint) {
	role := c.GetString("role")
	var compID *uint
	if cidVal, exists := c.Get("company_id"); exists {
		switch v := cidVal.(type) {
		case uint:
			if v > 0 { compID = &v }
		case int:
			if v > 0 { u := uint(v); compID = &u }
		}
	}
	return role, compID
}

// CreateGeofence godoc
// @Summary Create new geofence
// @Description Create a new geofence with multiple points
// @Tags Geofence
// @Accept json
// @Produce json
// @Security Bearer
// @Param geofence body CreateGeofenceRequest true "Geofence data"
// @Success 201 {object} map[string]interface{} "success: true, data: geofence"
// @Failure 400 {object} map[string]string "error: Invalid request"
// @Router /geofences [post]
func (h *Handler) CreateGeofence(c *gin.Context) {
	var req CreateGeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if callerRole == "company_admin" || callerRole == "operator" {
		if callerCompID == nil {
			base.ErrorResponse(c, http.StatusForbidden, "Company user must belong to a company", nil)
			return
		}
		req.CompanyID = callerCompID
	}

	geofence, err := h.service.CreateGeofence(req)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Gagal membuat geofence", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Geofence berhasil dibuat", geofence)
}

// GetGeofence godoc
// @Summary Get geofence by ID
// @Description Get geofence details by ID
// @Tags Geofence
// @Produce json
// @Security Bearer
// @Param id path int true "Geofence ID"
// @Success 200 {object} map[string]interface{} "success: true, data: geofence"
// @Failure 404 {object} map[string]string "error: Geofence not found"
// @Router /geofences/{id} [get]
func (h *Handler) GetGeofence(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid geofence ID", err)
		return
	}

	geofence, err := h.service.GetGeofenceByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Geofence tidak ditemukan", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if (callerRole == "company_admin" || callerRole == "operator" || callerRole == "viewer") && callerCompID != nil {
		if geofence.CompanyID != nil && *geofence.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusNotFound, "Geofence tidak ditemukan", nil)
			return
		}
	}

	base.SuccessResponse(c, http.StatusOK, "Geofence berhasil diambil", geofence)
}

// GetAllGeofences godoc
// @Summary Get all geofences
// @Description Get all geofences with pagination
// @Tags Geofence
// @Produce json
// @Security Bearer
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} map[string]interface{} "success: true, data: geofences"
// @Router /geofences [get]
func (h *Handler) GetAllGeofences(c *gin.Context) {
	page, limit, _ := base.GetPaginationParams(c)

	callerRole, callerCompID := extractScope(c)
	var filterCompID *uint
	if callerRole == "company_admin" || callerRole == "operator" || callerRole == "viewer" {
		filterCompID = callerCompID
	} else if cidStr := c.Query("company_id"); cidStr != "" {
		if cid, err := strconv.ParseUint(cidStr, 10, 32); err == nil {
			u := uint(cid)
			filterCompID = &u
		}
	}

	geofences, total, err := h.service.GetAllGeofences(page, limit, filterCompID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data geofences", err)
		return
	}

	if page > 0 && limit > 0 {
		base.PaginatedResponse(c, http.StatusOK, "Data geofences berhasil diambil", geofences, page, limit, total)
	} else {
		base.SuccessResponse(c, http.StatusOK, "Data geofences berhasil diambil", geofences)
	}
}

// UpdateGeofence godoc
// @Summary Update geofence
// @Description Partial update geofence by ID
// @Tags Geofence
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Geofence ID"
// @Param geofence body UpdateGeofenceRequest true "Updated geofence data (partial)"
// @Success 200 {object} map[string]interface{} "success: true, data: geofence"
// @Failure 400 {object} map[string]string "error: Failed to update"
// @Router /geofences/{id} [patch]
func (h *Handler) UpdateGeofence(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid geofence ID", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if (callerRole == "company_admin" || callerRole == "operator") && callerCompID != nil {
		existing, err := h.service.GetGeofenceByID(uint(id))
		if err != nil {
			base.ErrorResponse(c, http.StatusNotFound, "Geofence tidak ditemukan", err)
			return
		}
		if existing.CompanyID != nil && *existing.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Cannot modify geofence belonging to another company", nil)
			return
		}
	}

	var req UpdateGeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if (callerRole == "company_admin" || callerRole == "operator") && callerCompID != nil {
		req.CompanyID = callerCompID
	}

	geofence, err := h.service.UpdateGeofence(uint(id), req)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Gagal update geofence", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Geofence berhasil diupdate", geofence)
}

// DeleteGeofence godoc
// @Summary Delete geofence
// @Description Delete geofence by ID
// @Tags Geofence
// @Produce json
// @Security Bearer
// @Param id path int true "Geofence ID"
// @Success 200 {object} map[string]interface{} "success: true"
// @Failure 404 {object} map[string]string "error: Geofence not found"
// @Router /geofences/{id} [delete]
func (h *Handler) DeleteGeofence(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid geofence ID", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if (callerRole == "company_admin" || callerRole == "operator") && callerCompID != nil {
		existing, err := h.service.GetGeofenceByID(uint(id))
		if err != nil {
			base.ErrorResponse(c, http.StatusNotFound, "Geofence tidak ditemukan", err)
			return
		}
		if existing.CompanyID != nil && *existing.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Cannot delete geofence belonging to another company", nil)
			return
		}
	}

	if err := h.service.DeleteGeofence(uint(id)); err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Gagal delete geofence", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Geofence berhasil dihapus", nil)
}
