package crew

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

// CreateCrew godoc
// @Summary Create crew
// @Description Create a new crew member
// @Tags Master Data - Crew
// @Accept json
// @Produce json
// @Security Bearer
// @Param crew body CreateCrewRequest true "Crew data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/crew [post]
func (h *Handler) CreateCrew(c *gin.Context) {
	var req CreateCrewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.CreateCrew(req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Crew already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create crew", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Crew created successfully", response)
}

// GetCrew godoc
// @Summary Get crew by ID
// @Description Get crew details by ID
// @Tags Master Data - Crew
// @Produce json
// @Security Bearer
// @Param id path int true "Crew ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/crew/{id} [get]
func (h *Handler) GetCrew(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid crew ID", err)
		return
	}

	response, err := h.service.GetCrewByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Crew not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Crew fetched successfully", response)
}

// GetAllCrews godoc
// @Summary Get all crews
// @Description Get all crew members. Can filter by vessel_id using query parameter
// @Tags Master Data - Crew
// @Produce json
// @Security Bearer
// @Param vessel_id query int false "Filter by vessel ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /master/crew [get]
func (h *Handler) GetAllCrews(c *gin.Context) {
	vesselIDStr := c.Query("vessel_id")

	if vesselIDStr != "" {
		vesselID, err := strconv.ParseUint(vesselIDStr, 10, 32)
		if err != nil {
			base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
			return
		}

		response, err := h.service.GetCrewsByVesselID(uint(vesselID))
		if err != nil {
			base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch crews", err)
			return
		}

		base.SuccessResponse(c, http.StatusOK, "Crews fetched successfully", response)
		return
	}

	response, err := h.service.GetAllCrews()
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch crews", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Crews fetched successfully", response)
}

// UpdateCrew godoc
// @Summary Update crew
// @Description Update crew by ID
// @Tags Master Data - Crew
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Crew ID"
// @Param crew body UpdateCrewRequest true "Crew data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/crew/{id} [patch]
func (h *Handler) UpdateCrew(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid crew ID", err)
		return
	}

	var req UpdateCrewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.UpdateCrew(uint(id), req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Crew not found", err)
			return
		}
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Crew already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update crew", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Crew updated successfully", response)
}

// DeleteCrew godoc
// @Summary Delete crew
// @Description Delete crew by ID
// @Tags Master Data - Crew
// @Produce json
// @Security Bearer
// @Param id path int true "Crew ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/crew/{id} [delete]
func (h *Handler) DeleteCrew(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid crew ID", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.service.DeleteCrew(uint(id), userID, ipAddress, userAgent); err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Crew not found", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to delete crew", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Crew deleted successfully", nil)
}
