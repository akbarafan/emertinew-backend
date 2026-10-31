package operationstatus

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
	return &Handler{service}
}

// CreateOperationStatus godoc
// @Summary Create operation status
// @Description Create a new operation status
// @Tags Master - Operation Status
// @Accept json
// @Produce json
// @Param request body CreateOperationStatusRequest true "Operation status data"
// @Success 201 {object} OperationStatusResponse
// @Failure 400 {object} base.ErrorResponse
// @Failure 409 {object} base.ErrorResponse
// @Security BearerAuth
// @Router /api/master/operation-statuses [post]
func (h *Handler) CreateOperationStatus(c *gin.Context) {
	var req CreateOperationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	operationStatus, err := h.service.CreateOperationStatus(req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Operation status already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create operation status", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Operation status created successfully", operationStatus)
}

// GetAllOperationStatuses godoc
// @Summary Get all operation statuses
// @Description Get all operation statuses
// @Tags Master - Operation Status
// @Produce json
// @Success 200 {array} OperationStatusResponse
// @Failure 500 {object} base.ErrorResponse
// @Security BearerAuth
// @Router /api/master/operation-statuses [get]
func (h *Handler) GetAllOperationStatuses(c *gin.Context) {
	operationStatuses, err := h.service.GetAllOperationStatuses()
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to get operation statuses", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Operation statuses fetched successfully", operationStatuses)
}

// GetOperationStatusByID godoc
// @Summary Get operation status by ID
// @Description Get operation status by ID
// @Tags Master - Operation Status
// @Produce json
// @Param id path int true "Operation Status ID"
// @Success 200 {object} OperationStatusResponse
// @Failure 404 {object} base.ErrorResponse
// @Security BearerAuth
// @Router /api/master/operation-statuses/{id} [get]
func (h *Handler) GetOperationStatusByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid operation status ID", err)
		return
	}

	operationStatus, err := h.service.GetOperationStatusByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Operation status not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Operation status fetched successfully", operationStatus)
}

// UpdateOperationStatus godoc
// @Summary Update operation status
// @Description Update operation status by ID
// @Tags Master - Operation Status
// @Accept json
// @Produce json
// @Param id path int true "Operation Status ID"
// @Param request body UpdateOperationStatusRequest true "Operation status data"
// @Success 200 {object} OperationStatusResponse
// @Failure 400 {object} base.ErrorResponse
// @Failure 404 {object} base.ErrorResponse
// @Failure 409 {object} base.ErrorResponse
// @Security BearerAuth
// @Router /api/master/operation-statuses/{id} [put]
func (h *Handler) UpdateOperationStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid operation status ID", err)
		return
	}

	var req UpdateOperationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	operationStatus, err := h.service.UpdateOperationStatus(uint(id), req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Operation status not found", err)
			return
		}
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Operation status already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update operation status", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Operation status updated successfully", operationStatus)
}

// DeleteOperationStatus godoc
// @Summary Delete operation status
// @Description Delete operation status by ID
// @Tags Master - Operation Status
// @Produce json
// @Param id path int true "Operation Status ID"
// @Success 200 {object} base.MessageResponse
// @Failure 404 {object} base.ErrorResponse
// @Security BearerAuth
// @Router /api/master/operation-statuses/{id} [delete]
func (h *Handler) DeleteOperationStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid operation status ID", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.service.DeleteOperationStatus(uint(id), userID, ipAddress, userAgent); err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Operation status not found", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to delete operation status", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Operation status deleted successfully", nil)
}
