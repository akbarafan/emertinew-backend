package mcu

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

// CreateMCU godoc
// @Summary Create mcu
// @Description Create a new mcu
// @Tags Master Data - MCU
// @Accept json
// @Produce json
// @Security Bearer
// @Param mcu body CreateMCURequest true "MCU data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/mcu [post]
func (h *Handler) CreateMCU(c *gin.Context) {
	var req CreateMCURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.CreateMCU(req, userID, ipAddress, userAgent)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create mcu", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "MCU created successfully", response)
}

// GetMCU godoc
// @Summary Get mcu by ID
// @Description Get mcu details by ID
// @Tags Master Data - MCU
// @Produce json
// @Security Bearer
// @Param id path int true "MCU ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/mcu/{id} [get]
func (h *Handler) GetMCU(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid mcu ID", err)
		return
	}

	response, err := h.service.GetMCUByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "MCU not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "MCU fetched successfully", response)
}

// GetAllMCUs godoc
// @Summary Get all mcus
// @Description Get all mcus
// @Tags Master Data - MCU
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /master/mcu [get]
func (h *Handler) GetAllMCUs(c *gin.Context) {
	response, err := h.service.GetAllMCUs()
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch mcus", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "MCUs fetched successfully", response)
}

// UpdateMCU godoc
// @Summary Update mcu
// @Description Update mcu by ID
// @Tags Master Data - MCU
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "MCU ID"
// @Param mcu body UpdateMCURequest true "MCU data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/mcu/{id} [patch]
func (h *Handler) UpdateMCU(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid mcu ID", err)
		return
	}

	var req UpdateMCURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.UpdateMCU(uint(id), req, userID, ipAddress, userAgent)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update mcu", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "MCU updated successfully", response)
}

// DeleteMCU godoc
// @Summary Delete mcu
// @Description Soft delete mcu by ID
// @Tags Master Data - MCU
// @Produce json
// @Security Bearer
// @Param id path int true "MCU ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/mcu/{id} [delete]
func (h *Handler) DeleteMCU(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid mcu ID", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	err = h.service.DeleteMCU(uint(id), userID, ipAddress, userAgent)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to delete mcu", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "MCU deleted successfully", nil)
}

// PreviewConvertSensors godoc
// @Summary Preview bulk sensor topic conversion
// @Description Preview converting all (or selected) sensors on an MCU from old 5-part to new 7-part topic format
// @Tags Master Data - MCU
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "MCU ID"
// @Param body body ConvertSensorRequest false "Sensor IDs to preview (empty = all)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/mcu/{id}/sensors/preview-convert [post]
func (h *Handler) PreviewConvertSensors(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid MCU ID", err)
		return
	}

	var req ConvertSensorRequest
	c.ShouldBindJSON(&req) // ignore error, optional body

	response, err := h.service.PreviewConvertSensors(uint(id), req)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to preview conversion", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Preview generated", response)
}

// ApplyConvertSensors godoc
// @Summary Apply bulk sensor topic conversion
// @Description Apply converting sensors on an MCU from old 5-part to new 7-part topic format
// @Tags Master Data - MCU
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "MCU ID"
// @Param body body ConvertSensorRequest false "Sensor IDs to convert (empty = all old-format sensors)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/mcu/{id}/sensors/apply-convert [post]
func (h *Handler) ApplyConvertSensors(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid MCU ID", err)
		return
	}

	var req ConvertSensorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.ApplyConvertSensors(uint(id), req, userID, ipAddress, userAgent)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to apply conversion", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Conversion applied", response)
}
