package materialhull

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

// CreateMaterialHull godoc
// @Summary Create material hull
// @Description Create a new material hull
// @Tags Master Data - Material Hull
// @Accept json
// @Produce json
// @Security Bearer
// @Param materialhull body CreateMaterialHullRequest true "Material hull data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/material-hulls [post]
func (h *Handler) CreateMaterialHull(c *gin.Context) {
	var req CreateMaterialHullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.CreateMaterialHull(req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Material hull already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create material hull", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Material hull created successfully", response)
}

// GetMaterialHull godoc
// @Summary Get material hull by ID
// @Description Get material hull details by ID
// @Tags Master Data - Material Hull
// @Produce json
// @Security Bearer
// @Param id path int true "Material Hull ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/material-hulls/{id} [get]
func (h *Handler) GetMaterialHullByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid material hull ID", err)
		return
	}

	response, err := h.service.GetMaterialHullByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Material hull not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Material hull fetched successfully", response)
}

// GetAllMaterialHulls godoc
// @Summary Get all material hulls
// @Description Get all material hulls
// @Tags Master Data - Material Hull
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /master/material-hulls [get]
func (h *Handler) GetAllMaterialHulls(c *gin.Context) {
	response, err := h.service.GetAllMaterialHulls()
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch material hulls", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Material hulls fetched successfully", response)
}

// UpdateMaterialHull godoc
// @Summary Update material hull
// @Description Update material hull by ID
// @Tags Master Data - Material Hull
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Material Hull ID"
// @Param materialhull body UpdateMaterialHullRequest true "Material hull data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/material-hulls/{id} [put]
func (h *Handler) UpdateMaterialHull(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid material hull ID", err)
		return
	}

	var req UpdateMaterialHullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.UpdateMaterialHull(uint(id), req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Material hull not found", err)
			return
		}
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Material hull already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update material hull", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Material hull updated successfully", response)
}

// DeleteMaterialHull godoc
// @Summary Delete material hull
// @Description Delete material hull by ID
// @Tags Master Data - Material Hull
// @Produce json
// @Security Bearer
// @Param id path int true "Material Hull ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/material-hulls/{id} [delete]
func (h *Handler) DeleteMaterialHull(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid material hull ID", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.service.DeleteMaterialHull(uint(id), userID, ipAddress, userAgent); err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Material hull not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Material hull deleted successfully", nil)
}
