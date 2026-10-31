package classification

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

// CreateClassification godoc
// @Summary Create classification
// @Description Create a new classification
// @Tags Master Data - Classification
// @Accept json
// @Produce json
// @Security Bearer
// @Param classification body CreateClassificationRequest true "Classification data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/classifications [post]
func (h *Handler) CreateClassification(c *gin.Context) {
	var req CreateClassificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.CreateClassification(req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Classification already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create classification", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Classification created successfully", response)
}

// GetClassification godoc
// @Summary Get classification by ID
// @Description Get classification details by ID
// @Tags Master Data - Classification
// @Produce json
// @Security Bearer
// @Param id path int true "Classification ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/classifications/{id} [get]
func (h *Handler) GetClassificationByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid classification ID", err)
		return
	}

	response, err := h.service.GetClassificationByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Classification not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Classification fetched successfully", response)
}

// GetAllClassifications godoc
// @Summary Get all classifications
// @Description Get all classifications
// @Tags Master Data - Classification
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /master/classifications [get]
func (h *Handler) GetAllClassifications(c *gin.Context) {
	response, err := h.service.GetAllClassifications()
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch classifications", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Classifications fetched successfully", response)
}

// UpdateClassification godoc
// @Summary Update classification
// @Description Update classification by ID
// @Tags Master Data - Classification
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Classification ID"
// @Param classification body UpdateClassificationRequest true "Classification data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/classifications/{id} [put]
func (h *Handler) UpdateClassification(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid classification ID", err)
		return
	}

	var req UpdateClassificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.UpdateClassification(uint(id), req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Classification not found", err)
			return
		}
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Classification already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update classification", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Classification updated successfully", response)
}

// DeleteClassification godoc
// @Summary Delete classification
// @Description Delete classification by ID
// @Tags Master Data - Classification
// @Produce json
// @Security Bearer
// @Param id path int true "Classification ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/classifications/{id} [delete]
func (h *Handler) DeleteClassification(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid classification ID", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.service.DeleteClassification(uint(id), userID, ipAddress, userAgent); err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Classification not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Classification deleted successfully", nil)
}
