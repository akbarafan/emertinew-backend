package offshore

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

// CreateOffshore godoc
// @Summary Create offshore
// @Description Create a new offshore location
// @Tags Master Data - Offshore
// @Accept json
// @Produce json
// @Security Bearer
// @Param offshore body CreateOffshoreRequest true "Offshore data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/offshore [post]
func (h *Handler) CreateOffshore(c *gin.Context) {
	var req CreateOffshoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if callerRole == "company_admin" || callerRole == "operator" {
		if callerCompID == nil {
			base.ErrorResponse(c, http.StatusForbidden, "Company admin must belong to a company", nil)
			return
		}
		req.CompanyID = callerCompID
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.CreateOffshore(req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Offshore already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to create offshore", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Offshore created successfully", response)
}

// GetOffshore godoc
// @Summary Get offshore by ID
// @Description Get offshore details by ID
// @Tags Master Data - Offshore
// @Produce json
// @Security Bearer
// @Param id path int true "Offshore ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/offshore/{id} [get]
func (h *Handler) GetOffshore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid offshore ID", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	response, err := h.service.GetOffshoreByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Offshore not found", err)
		return
	}

	if (callerRole == "company_admin" || callerRole == "operator") && callerCompID != nil {
		if response.CompanyID != nil && *response.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Access denied to offshore location of another company", nil)
			return
		}
	}

	base.SuccessResponse(c, http.StatusOK, "Offshore fetched successfully", response)
}

// GetAllOffshores godoc
// @Summary Get all offshores
// @Description Get all offshore locations
// @Tags Master Data - Offshore
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /master/offshore [get]
func (h *Handler) GetAllOffshores(c *gin.Context) {
	callerRole, callerCompID := extractScope(c)
	var targetCompID *uint
	if callerRole == "company_admin" || callerRole == "operator" || callerRole == "viewer" {
		targetCompID = callerCompID
	} else if cQuery := c.Query("company_id"); cQuery != "" {
		if cid, err := strconv.ParseUint(cQuery, 10, 32); err == nil && cid > 0 {
			uCid := uint(cid)
			targetCompID = &uCid
		}
	}

	response, err := h.service.GetAllOffshores(targetCompID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch offshores", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Offshores fetched successfully", response)
}

// UpdateOffshore godoc
// @Summary Update offshore
// @Description Update offshore by ID
// @Tags Master Data - Offshore
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Offshore ID"
// @Param offshore body UpdateOffshoreRequest true "Offshore data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/offshore/{id} [patch]
func (h *Handler) UpdateOffshore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid offshore ID", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if callerRole == "company_admin" && callerCompID != nil {
		existing, err := h.service.GetOffshoreByID(uint(id))
		if err != nil {
			base.ErrorResponse(c, http.StatusNotFound, "Offshore not found", err)
			return
		}
		if existing.CompanyID != nil && *existing.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Cannot modify offshore location of another company", nil)
			return
		}
	}

	var req UpdateOffshoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	if callerRole == "company_admin" && callerCompID != nil {
		req.CompanyID = callerCompID
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.UpdateOffshore(uint(id), req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Offshore not found", err)
			return
		}
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Offshore already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to update offshore", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Offshore updated successfully", response)
}

// DeleteOffshore godoc
// @Summary Delete offshore
// @Description Delete offshore by ID
// @Tags Master Data - Offshore
// @Produce json
// @Security Bearer
// @Param id path int true "Offshore ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/offshore/{id} [delete]
func (h *Handler) DeleteOffshore(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid offshore ID", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if callerRole == "company_admin" && callerCompID != nil {
		existing, err := h.service.GetOffshoreByID(uint(id))
		if err != nil {
			base.ErrorResponse(c, http.StatusNotFound, "Offshore not found", err)
			return
		}
		if existing.CompanyID != nil && *existing.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Cannot delete offshore location of another company", nil)
			return
		}
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.service.DeleteOffshore(uint(id), userID, ipAddress, userAgent); err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Offshore not found", err)
			return
		}
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete offshore", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Offshore deleted successfully", nil)
}
