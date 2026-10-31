package fueltype

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

// CreateFuelType godoc
// @Summary Create fuel type
// @Description Create a new fuel type with its IMO emission parameters
// @Tags Master Data - Fuel Type
// @Accept json
// @Produce json
// @Security Bearer
// @Param fuel_type body CreateFuelTypeRequest true "Fuel type data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/fuel-type [post]
func (h *Handler) CreateFuelType(c *gin.Context) {
	var req CreateFuelTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.CreateFuelType(req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Fuel type already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create fuel type", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Fuel type created successfully", response)
}

// GetFuelType godoc
// @Summary Get fuel type by ID
// @Description Get fuel type details by ID
// @Tags Master Data - Fuel Type
// @Produce json
// @Security Bearer
// @Param id path int true "Fuel type ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/fuel-type/{id} [get]
func (h *Handler) GetFuelType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid fuel type ID", err)
		return
	}

	response, err := h.service.GetFuelTypeByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Fuel type not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Fuel type fetched successfully", response)
}

// GetAllFuelTypes godoc
// @Summary Get all fuel types
// @Description Get all fuel types with their derived emission factors
// @Tags Master Data - Fuel Type
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /master/fuel-type [get]
func (h *Handler) GetAllFuelTypes(c *gin.Context) {
	response, err := h.service.GetAllFuelTypes()
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch fuel types", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Fuel types fetched successfully", response)
}

// UpdateFuelType godoc
// @Summary Update fuel type
// @Description Update fuel type by ID
// @Tags Master Data - Fuel Type
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Fuel type ID"
// @Param fuel_type body UpdateFuelTypeRequest true "Fuel type data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/fuel-type/{id} [patch]
func (h *Handler) UpdateFuelType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid fuel type ID", err)
		return
	}

	var req UpdateFuelTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.UpdateFuelType(uint(id), req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Fuel type not found", err)
			return
		}
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Fuel type already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update fuel type", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Fuel type updated successfully", response)
}

// DeleteFuelType godoc
// @Summary Delete fuel type
// @Description Delete fuel type by ID
// @Tags Master Data - Fuel Type
// @Produce json
// @Security Bearer
// @Param id path int true "Fuel type ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/fuel-type/{id} [delete]
func (h *Handler) DeleteFuelType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid fuel type ID", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.service.DeleteFuelType(uint(id), userID, ipAddress, userAgent); err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Fuel type not found", err)
			return
		}
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete fuel type", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Fuel type deleted successfully", nil)
}
