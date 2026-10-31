package vesseldimensions

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

// CreateVesselDimensions godoc
// @Summary Create vessel dimensions
// @Description Create vessel dimensions for a vessel
// @Tags Master Data - Vessel Dimensions
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel ID"
// @Param dimensions body CreateVesselDimensionsRequest true "Vessel Dimensions data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/vessel/{id}/dimensions [post]
func (h *Handler) CreateVesselDimensions(c *gin.Context) {
	vesselID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	var req CreateVesselDimensionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// Extract context
	userID := utils.GetUserIDFromContext(c)
	ip := utils.GetClientIP(c)
	ua := utils.GetUserAgent(c)

	response, err := h.service.CreateVesselDimensions(uint(vesselID), req, userID, ip, ua)
	if err != nil {
		if err.Error() == "vessel dimensions already exist for this vessel" {
			base.ErrorResponse(c, http.StatusConflict, "Failed to create vessel dimensions", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create vessel dimensions", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Vessel dimensions created successfully", response)
}

// GetVesselDimensions godoc
// @Summary Get vessel dimensions by vessel ID
// @Description Get vessel dimensions by vessel ID
// @Tags Master Data - Vessel Dimensions
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/vessel/{id}/dimensions [get]
func (h *Handler) GetVesselDimensions(c *gin.Context) {
	vesselID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	response, err := h.service.GetVesselDimensionsByVesselID(uint(vesselID))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Vessel dimensions not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel dimensions fetched successfully", response)
}

// GetVesselDimensionsByVesselID godoc
// @Summary Get vessel dimensions by vessel ID
// @Description Get vessel dimensions by vessel ID
// @Tags Master Data - Vessel Dimensions
// @Produce json
// @Security Bearer
// @Param vessel_id path int true "Vessel ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/vessel/{vessel_id}/dimensions [get]
func (h *Handler) GetVesselDimensionsByVesselID(c *gin.Context) {
	vesselID, err := strconv.ParseUint(c.Param("vessel_id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	response, err := h.service.GetVesselDimensionsByVesselID(uint(vesselID))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Vessel dimensions not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel dimensions fetched successfully", response)
}

// UpdateVesselDimensions godoc
// @Summary Update vessel dimensions
// @Description Update vessel dimensions by vessel ID
// @Tags Master Data - Vessel Dimensions
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel ID"
// @Param dimensions body UpdateVesselDimensionsRequest true "Vessel Dimensions data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/vessel/{id}/dimensions [patch]
func (h *Handler) UpdateVesselDimensions(c *gin.Context) {
	vesselID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	var req UpdateVesselDimensionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// Extract context
	userID := utils.GetUserIDFromContext(c)
	ip := utils.GetClientIP(c)
	ua := utils.GetUserAgent(c)

	response, err := h.service.UpdateVesselDimensionsByVesselID(uint(vesselID), req, userID, ip, ua)
	if err != nil {
		if err.Error() == "vessel dimensions not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Failed to update vessel dimensions", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update vessel dimensions", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel dimensions updated successfully", response)
}

// DeleteVesselDimensions godoc
// @Summary Delete vessel dimensions
// @Description Delete vessel dimensions by vessel ID
// @Tags Master Data - Vessel Dimensions
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/vessel/{id}/dimensions [delete]
func (h *Handler) DeleteVesselDimensions(c *gin.Context) {
	vesselID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	// Extract context
	userID := utils.GetUserIDFromContext(c)
	ip := utils.GetClientIP(c)
	ua := utils.GetUserAgent(c)

	err = h.service.DeleteVesselDimensionsByVesselID(uint(vesselID), userID, ip, ua)
	if err != nil {
		if err.Error() == "vessel dimensions not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Failed to delete vessel dimensions", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to delete vessel dimensions", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel dimensions deleted successfully", nil)
}
