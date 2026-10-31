package fleet

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

func (h *Handler) CreateFleet(c *gin.Context) {
	var req CreateFleetRequest
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

	response, err := h.service.CreateFleet(req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Fleet already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create fleet", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Fleet created successfully", response)
}

func (h *Handler) GetFleetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid fleet ID", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	response, err := h.service.GetFleetByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Fleet not found", err)
		return
	}

	if (callerRole == "company_admin" || callerRole == "operator") && callerCompID != nil {
		if response.CompanyID != nil && *response.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Access denied to fleet of another company", nil)
			return
		}
	}

	base.SuccessResponse(c, http.StatusOK, "Fleet fetched successfully", response)
}

func (h *Handler) GetAllFleets(c *gin.Context) {
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

	response, err := h.service.GetAllFleets(targetCompID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch fleets", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Fleets fetched successfully", response)
}

func (h *Handler) UpdateFleet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid fleet ID", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if callerRole == "company_admin" && callerCompID != nil {
		existing, err := h.service.GetFleetByID(uint(id))
		if err != nil {
			base.ErrorResponse(c, http.StatusNotFound, "Fleet not found", err)
			return
		}
		if existing.CompanyID != nil && *existing.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Cannot modify fleet of another company", nil)
			return
		}
	}

	var req UpdateFleetRequest
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

	response, err := h.service.UpdateFleet(uint(id), req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Fleet not found", err)
			return
		}
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Fleet already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update fleet", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Fleet updated successfully", response)
}

func (h *Handler) DeleteFleet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid fleet ID", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if callerRole == "company_admin" && callerCompID != nil {
		existing, err := h.service.GetFleetByID(uint(id))
		if err != nil {
			base.ErrorResponse(c, http.StatusNotFound, "Fleet not found", err)
			return
		}
		if existing.CompanyID != nil && *existing.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Cannot delete fleet of another company", nil)
			return
		}
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.service.DeleteFleet(uint(id), userID, ipAddress, userAgent); err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Fleet not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Fleet deleted successfully", nil)
}
