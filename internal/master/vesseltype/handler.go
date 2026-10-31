package vesseltype

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

// CreateVesselType godoc
// @Summary Create vessel type
// @Description Create a new vessel type
// @Tags Master Data - Vessel Type
// @Accept json
// @Produce json
// @Security Bearer
// @Param vesseltype body CreateVesselTypeRequest true "Vessel type data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/vessel-types [post]
func (h *Handler) CreateVesselType(c *gin.Context) {
	var req CreateVesselTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.CreateVesselType(req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Vessel type already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create vessel type", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Vessel type created successfully", response)
}

// GetVesselType godoc
// @Summary Get vessel type by ID
// @Description Get vessel type details by ID
// @Tags Master Data - Vessel Type
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel Type ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/vessel-types/{id} [get]
func (h *Handler) GetVesselTypeByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel type ID", err)
		return
	}

	response, err := h.service.GetVesselTypeByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Vessel type not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel type fetched successfully", response)
}

// GetAllVesselTypes godoc
// @Summary Get all vessel types
// @Description Get all vessel types
// @Tags Master Data - Vessel Type
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /master/vessel-types [get]
func (h *Handler) GetAllVesselTypes(c *gin.Context) {
	response, err := h.service.GetAllVesselTypes()
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch vessel types", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel types fetched successfully", response)
}

// UpdateVesselType godoc
// @Summary Update vessel type
// @Description Update vessel type by ID
// @Tags Master Data - Vessel Type
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel Type ID"
// @Param vesseltype body UpdateVesselTypeRequest true "Vessel type data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/vessel-types/{id} [put]
func (h *Handler) UpdateVesselType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel type ID", err)
		return
	}

	var req UpdateVesselTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.UpdateVesselType(uint(id), req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Vessel type not found", err)
			return
		}
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Vessel type already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update vessel type", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel type updated successfully", response)
}

// DeleteVesselType godoc
// @Summary Delete vessel type
// @Description Delete vessel type by ID
// @Tags Master Data - Vessel Type
// @Produce json
// @Security Bearer
// @Param id path int true "Vessel Type ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/vessel-types/{id} [delete]
func (h *Handler) DeleteVesselType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel type ID", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.service.DeleteVesselType(uint(id), userID, ipAddress, userAgent); err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Vessel type not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Vessel type deleted successfully", nil)
}
