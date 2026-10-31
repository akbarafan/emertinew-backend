package port

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

// CreatePort godoc
// @Summary Create port
// @Description Create a new port location
// @Tags Master Data - Port
// @Accept json
// @Produce json
// @Security Bearer
// @Param port body CreatePortRequest true "Port data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/port [post]
func (h *Handler) CreatePort(c *gin.Context) {
	var req CreatePortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.CreatePort(req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Port already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to create port", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Port created successfully", response)
}

// GetPort godoc
// @Summary Get port by ID
// @Description Get port details by ID
// @Tags Master Data - Port
// @Produce json
// @Security Bearer
// @Param id path int true "Port ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/port/{id} [get]
func (h *Handler) GetPort(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid port ID", err)
		return
	}

	response, err := h.service.GetPortByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Port not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Port fetched successfully", response)
}

// GetAllPorts godoc
// @Summary Get all ports
// @Description Get all port locations
// @Tags Master Data - Port
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /master/port [get]
func (h *Handler) GetAllPorts(c *gin.Context) {
	response, err := h.service.GetAllPorts()
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch ports", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Ports fetched successfully", response)
}

// UpdatePort godoc
// @Summary Update port
// @Description Update port by ID
// @Tags Master Data - Port
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Port ID"
// @Param port body UpdatePortRequest true "Port data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/port/{id} [patch]
func (h *Handler) UpdatePort(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid port ID", err)
		return
	}

	var req UpdatePortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.UpdatePort(uint(id), req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Port not found", err)
			return
		}
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Port already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to update port", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Port updated successfully", response)
}

// DeletePort godoc
// @Summary Delete port
// @Description Delete port by ID
// @Tags Master Data - Port
// @Produce json
// @Security Bearer
// @Param id path int true "Port ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/port/{id} [delete]
func (h *Handler) DeletePort(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid port ID", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.service.DeletePort(uint(id), userID, ipAddress, userAgent); err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Port not found", err)
			return
		}
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete port", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Port deleted successfully", nil)
}
