package squad

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

func (h *Handler) CreateSquad(c *gin.Context) {
	var req CreateSquadRequest
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

	response, err := h.service.CreateSquad(req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Squad already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create squad", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Squad created successfully", response)
}

func (h *Handler) GetSquadByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid squad ID", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	response, err := h.service.GetSquadByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Squad not found", err)
		return
	}

	if (callerRole == "company_admin" || callerRole == "operator") && callerCompID != nil {
		if response.CompanyID != nil && *response.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Access denied to squad of another company", nil)
			return
		}
	}

	base.SuccessResponse(c, http.StatusOK, "Squad fetched successfully", response)
}

func (h *Handler) GetAllSquads(c *gin.Context) {
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

	response, err := h.service.GetAllSquads(targetCompID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch squads", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Squads fetched successfully", response)
}

func (h *Handler) GetSquadsByFleetID(c *gin.Context) {
	fleetID, err := strconv.ParseUint(c.Query("fleet_id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid fleet_id query", err)
		return
	}

	response, err := h.service.GetSquadsByFleetID(uint(fleetID))
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch squads", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Squads fetched successfully", response)
}

func (h *Handler) UpdateSquad(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid squad ID", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if callerRole == "company_admin" && callerCompID != nil {
		existing, err := h.service.GetSquadByID(uint(id))
		if err != nil {
			base.ErrorResponse(c, http.StatusNotFound, "Squad not found", err)
			return
		}
		if existing.CompanyID != nil && *existing.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Cannot modify squad of another company", nil)
			return
		}
	}

	var req UpdateSquadRequest
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

	response, err := h.service.UpdateSquad(uint(id), req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Squad not found", err)
			return
		}
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Squad already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update squad", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Squad updated successfully", response)
}

func (h *Handler) DeleteSquad(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid squad ID", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if callerRole == "company_admin" && callerCompID != nil {
		existing, err := h.service.GetSquadByID(uint(id))
		if err != nil {
			base.ErrorResponse(c, http.StatusNotFound, "Squad not found", err)
			return
		}
		if existing.CompanyID != nil && *existing.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Cannot delete squad of another company", nil)
			return
		}
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.service.DeleteSquad(uint(id), userID, ipAddress, userAgent); err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Squad not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Squad deleted successfully", nil)
}
