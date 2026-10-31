package vessel

import (
	"net/http"
	"strconv"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/base"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateVessel godoc
// @Summary Create vessel
// @Description Create a new vessel
// @Tags Master Data - Vessel
// @Accept json
// @Produce json
// @Security Bearer
// @Param vessel body CreateVesselRequest true "Vessel data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/vessel [post]
func (h *Handler) CreateVessel(c *gin.Context) {
	var req CreateVesselRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	callerRole := c.GetString("role")
	var callerCompID *uint
	if cid, exists := c.Get("company_id"); exists {
		switch v := cid.(type) {
		case uint:
			if v > 0 { callerCompID = &v }
		case int:
			if v > 0 { u := uint(v); callerCompID = &u }
		}
	}
	if model.IsCompanyScoped(callerRole) {
		if callerCompID == nil {
			base.ErrorResponse(c, http.StatusForbidden, "Company user must belong to a company", nil)
			return
		}
		req.CompanyID = callerCompID
		req.MCUIDs = nil // Company users cannot bind hardware MCU
	}

	// Extract context
	userID := utils.GetUserIDFromContext(c)
	ip := utils.GetClientIP(c)
	ua := utils.GetUserAgent(c)

	response, err := h.service.CreateVessel(req, userID, ip, ua)
	if err != nil {
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Vessel already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create vessel", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Vessel created successfully", response)
}

// GetVessel godoc
// @Summary Get vessel by ID
// @Description Get vessel details by ID
// @Tags Master Data - Vessel
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/vessel/{id} [get]
func (h *Handler) GetVessel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	response, err := h.service.GetVesselByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Vessel not found", err)
		return
	}

	callerRole := c.GetString("role")
	var callerCompID *uint
	if cid, exists := c.Get("company_id"); exists {
		switch v := cid.(type) {
		case uint:
			if v > 0 { callerCompID = &v }
		case int:
			if v > 0 { u := uint(v); callerCompID = &u }
		}
	}
	if model.IsCompanyScoped(callerRole) && callerCompID != nil {
		if response.Company != nil && response.Company.ID != *callerCompID {
			base.ErrorResponse(c, http.StatusNotFound, "Vessel not found", nil)
			return
		}
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel fetched successfully", response)
}

// GetAllVessels godoc
// @Summary Get all vessels
// @Description Get all vessels
// @Tags Master Data - Vessel
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /master/vessel [get]
func (h *Handler) GetAllVessels(c *gin.Context) {
	response, err := h.service.GetAllVessels()
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch vessels", err)
		return
	}

	callerRole := c.GetString("role")
	var callerCompID *uint
	if cid, exists := c.Get("company_id"); exists {
		switch v := cid.(type) {
		case uint:
			if v > 0 {
				callerCompID = &v
			}
		case int:
			if v > 0 {
				u := uint(v)
				callerCompID = &u
			}
		}
	}

	// Company-scoped users only see vessels belonging to their company
	if model.IsCompanyScoped(callerRole) && callerCompID != nil {
		var filtered []VesselResponse
		for _, v := range response {
			if v.Company != nil && v.Company.ID == *callerCompID {
				filtered = append(filtered, v)
			}
		}
		response = filtered
	} else if cQuery := c.Query("company_id"); cQuery != "" {
		if cid, err := strconv.ParseUint(cQuery, 10, 32); err == nil && cid > 0 {
			var filtered []VesselResponse
			for _, v := range response {
				if v.Company != nil && v.Company.ID == uint(cid) {
					filtered = append(filtered, v)
				}
			}
			response = filtered
		}
	}

	base.SuccessResponse(c, http.StatusOK, "Vessels fetched successfully", response)
}

// UpdateVessel godoc
// @Summary Update vessel
// @Description Update vessel by ID
// @Tags Master Data - Vessel
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel ID"
// @Param vessel body UpdateVesselRequest true "Vessel data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/vessel/{id} [patch]
func (h *Handler) UpdateVessel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	callerRole := c.GetString("role")
	var callerCompID *uint
	if cid, exists := c.Get("company_id"); exists {
		switch v := cid.(type) {
		case uint:
			if v > 0 { callerCompID = &v }
		case int:
			if v > 0 { u := uint(v); callerCompID = &u }
		}
	}
	if model.IsCompanyScoped(callerRole) && callerCompID != nil {
		existing, err := h.service.GetVesselByID(uint(id))
		if err != nil {
			base.ErrorResponse(c, http.StatusNotFound, "Vessel not found", err)
			return
		}
		if existing.Company != nil && existing.Company.ID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Cannot modify vessel belonging to another company", nil)
			return
		}
	}

	var req UpdateVesselRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	if model.IsCompanyScoped(callerRole) && callerCompID != nil {
		req.CompanyID = callerCompID
		req.MCUIDs = nil // Company users cannot alter MCU hardware bindings
	}

	// Extract context
	userID := utils.GetUserIDFromContext(c)
	ip := utils.GetClientIP(c)
	ua := utils.GetUserAgent(c)

	response, err := h.service.UpdateVessel(uint(id), req, userID, ip, ua)
	if err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Vessel not found", err)
			return
		}
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Vessel already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update vessel", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel updated successfully", response)
}

// DeleteVessel godoc
// @Summary Delete vessel
// @Description Soft delete vessel by ID
// @Tags Master Data - Vessel
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/vessel/{id} [delete]
func (h *Handler) DeleteVessel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	callerRole := c.GetString("role")
	var callerCompID *uint
	if cid, exists := c.Get("company_id"); exists {
		switch v := cid.(type) {
		case uint:
			if v > 0 { callerCompID = &v }
		case int:
			if v > 0 { u := uint(v); callerCompID = &u }
		}
	}
	if model.IsCompanyScoped(callerRole) && callerCompID != nil {
		existing, err := h.service.GetVesselByID(uint(id))
		if err != nil {
			base.ErrorResponse(c, http.StatusNotFound, "Vessel not found", err)
			return
		}
		if existing.Company != nil && existing.Company.ID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Cannot delete vessel belonging to another company", nil)
			return
		}
	}

	// Extract context
	userID := utils.GetUserIDFromContext(c)
	ip := utils.GetClientIP(c)
	ua := utils.GetUserAgent(c)

	err = h.service.DeleteVessel(uint(id), userID, ip, ua)
	if err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Vessel not found", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to delete vessel", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel deleted successfully", nil)
}

// RestoreVessel godoc
// @Summary Restore deleted vessel
// @Description Restore a soft-deleted vessel by ID
// @Tags Master Data - Vessel
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/vessel/{id}/restore [post]
func (h *Handler) RestoreVessel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	callerRole := c.GetString("role")
	var callerCompID *uint
	if cid, exists := c.Get("company_id"); exists {
		switch v := cid.(type) {
		case uint:
			if v > 0 { callerCompID = &v }
		case int:
			if v > 0 { u := uint(v); callerCompID = &u }
		}
	}
	if model.IsCompanyScoped(callerRole) && callerCompID != nil {
		existing, err := h.service.GetVesselByID(uint(id))
		if err != nil {
			base.ErrorResponse(c, http.StatusNotFound, "Vessel not found", err)
			return
		}
		if existing.Company != nil && existing.Company.ID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Cannot restore vessel belonging to another company", nil)
			return
		}
	}

	userID := utils.GetUserIDFromContext(c)
	ip := utils.GetClientIP(c)
	ua := utils.GetUserAgent(c)

	err = h.service.RestoreVessel(uint(id), userID, ip, ua)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to restore vessel", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel restored successfully", nil)
}

// GetDeletedVessels godoc
// @Summary Get deleted vessels
// @Description Get all soft-deleted vessels
// @Tags Master Data - Vessel
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /master/vessel/deleted [get]
func (h *Handler) GetDeletedVessels(c *gin.Context) {
	response, err := h.service.GetDeletedVessels()
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch deleted vessels", err)
		return
	}

	callerRole := c.GetString("role")
	var callerCompID *uint
	if cid, exists := c.Get("company_id"); exists {
		switch v := cid.(type) {
		case uint:
			if v > 0 { callerCompID = &v }
		case int:
			if v > 0 { u := uint(v); callerCompID = &u }
		}
	}
	if model.IsCompanyScoped(callerRole) && callerCompID != nil {
		var filtered []VesselResponse
		for _, v := range response {
			if v.Company != nil && v.Company.ID == *callerCompID {
				filtered = append(filtered, v)
			}
		}
		response = filtered
	}

	base.SuccessResponse(c, http.StatusOK, "Deleted vessels fetched successfully", response)
}
