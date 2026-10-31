package manifest

import (
	"net/http"
	"strconv"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/base"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateManifest godoc
// @Summary Create manifest
// @Description Create a new manifest (status will be draft by default)
// @Tags Master Data - Manifest
// @Accept json
// @Produce json
// @Security Bearer
// @Param manifest body CreateManifestRequest true "Manifest data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/manifest [post]
func (h *Handler) CreateManifest(c *gin.Context) {
	var req CreateManifestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	response, err := h.service.CreateManifest(req)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create manifest", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Manifest created successfully", response)
}

// GetManifest godoc
// @Summary Get manifest by ID
// @Description Get manifest details by ID
// @Tags Master Data - Manifest
// @Produce json
// @Security Bearer
// @Param id path int true "Manifest ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/manifest/{id} [get]
func (h *Handler) GetManifest(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid manifest ID", err)
		return
	}

	response, err := h.service.GetManifestByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Manifest not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Manifest fetched successfully", response)
}

// GetAllManifests godoc
// @Summary Get all manifests
// @Description Get all manifests. Can filter by vessel_id or status using query parameters
// @Tags Master Data - Manifest
// @Produce json
// @Security Bearer
// @Param vessel_id query int false "Filter by vessel ID"
// @Param status query string false "Filter by status (draft, review, approved)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /master/manifest [get]
func (h *Handler) GetAllManifests(c *gin.Context) {
	vesselIDStr := c.Query("vessel_id")
	status := c.Query("status")

	if vesselIDStr != "" {
		vesselID, err := strconv.ParseUint(vesselIDStr, 10, 32)
		if err != nil {
			base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
			return
		}

		response, err := h.service.GetManifestsByVesselID(uint(vesselID))
		if err != nil {
			base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch manifests", err)
			return
		}

		base.SuccessResponse(c, http.StatusOK, "Manifests fetched successfully", response)
		return
	}

	if status != "" {
		response, err := h.service.GetManifestsByStatus(status)
		if err != nil {
			base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch manifests", err)
			return
		}

		base.SuccessResponse(c, http.StatusOK, "Manifests fetched successfully", response)
		return
	}

	response, err := h.service.GetAllManifests()
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch manifests", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Manifests fetched successfully", response)
}

// UpdateManifest godoc
// @Summary Update manifest
// @Description Update manifest by ID (only draft manifests can be updated)
// @Tags Master Data - Manifest
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Manifest ID"
// @Param manifest body UpdateManifestRequest true "Manifest data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/manifest/{id} [patch]
func (h *Handler) UpdateManifest(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid manifest ID", err)
		return
	}

	var req UpdateManifestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	response, err := h.service.UpdateManifest(uint(id), req)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update manifest", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Manifest updated successfully", response)
}

// DeleteManifest godoc
// @Summary Delete manifest
// @Description Delete manifest by ID (only draft manifests can be deleted)
// @Tags Master Data - Manifest
// @Produce json
// @Security Bearer
// @Param id path int true "Manifest ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/manifest/{id} [delete]
func (h *Handler) DeleteManifest(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid manifest ID", err)
		return
	}

	if err := h.service.DeleteManifest(uint(id)); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Manifest deleted successfully", nil)
}
