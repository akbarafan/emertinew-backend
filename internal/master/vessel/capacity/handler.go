package vesselcapacity

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

// CreateVesselCapacity godoc
// @Summary Create vessel capacity
// @Description Create vessel capacity for a vessel
// @Tags Master Data - Vessel Capacity
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel ID"
// @Param capacity body CreateVesselCapacityRequest true "Vessel Capacity data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/vessel/{id}/capacity [post]
func (h *Handler) CreateVesselCapacity(c *gin.Context) {
	vesselID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	var req CreateVesselCapacityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// Extract context
	userID := utils.GetUserIDFromContext(c)
	ip := utils.GetClientIP(c)
	ua := utils.GetUserAgent(c)

	response, err := h.service.CreateVesselCapacity(uint(vesselID), req, userID, ip, ua)
	if err != nil {
		if err.Error() == "vessel capacity already exists for this vessel" {
			base.ErrorResponse(c, http.StatusConflict, "Failed to create vessel capacity", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create vessel capacity", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Vessel capacity created successfully", response)
}

// GetVesselCapacity godoc
// @Summary Get vessel capacity by vessel ID
// @Description Get vessel capacity by vessel ID
// @Tags Master Data - Vessel Capacity
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/vessel/{id}/capacity [get]
func (h *Handler) GetVesselCapacity(c *gin.Context) {
	vesselID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	response, err := h.service.GetVesselCapacityByVesselID(uint(vesselID))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Vessel capacity not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel capacity fetched successfully", response)
}

// UpdateVesselCapacity godoc
// @Summary Update vessel capacity
// @Description Update vessel capacity by vessel ID
// @Tags Master Data - Vessel Capacity
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel ID"
// @Param capacity body UpdateVesselCapacityRequest true "Vessel Capacity data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/vessel/{id}/capacity [patch]
func (h *Handler) UpdateVesselCapacity(c *gin.Context) {
	vesselID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	var req UpdateVesselCapacityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// Extract context
	userID := utils.GetUserIDFromContext(c)
	ip := utils.GetClientIP(c)
	ua := utils.GetUserAgent(c)

	response, err := h.service.UpdateVesselCapacityByVesselID(uint(vesselID), req, userID, ip, ua)
	if err != nil {
		if err.Error() == "vessel capacity not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Failed to update vessel capacity", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update vessel capacity", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel capacity updated successfully", response)
}

// DeleteVesselCapacity godoc
// @Summary Delete vessel capacity
// @Description Delete vessel capacity by vessel ID
// @Tags Master Data - Vessel Capacity
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/vessel/{id}/capacity [delete]
func (h *Handler) DeleteVesselCapacity(c *gin.Context) {
	vesselID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	// Extract context
	userID := utils.GetUserIDFromContext(c)
	ip := utils.GetClientIP(c)
	ua := utils.GetUserAgent(c)

	err = h.service.DeleteVesselCapacityByVesselID(uint(vesselID), userID, ip, ua)
	if err != nil {
		if err.Error() == "vessel capacity not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Failed to delete vessel capacity", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to delete vessel capacity", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel capacity deleted successfully", nil)
}
