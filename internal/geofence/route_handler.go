package geofence

import (
	"net/http"
	"strconv"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/base"
	"github.com/gin-gonic/gin"
)

type RouteHandler struct {
	service RouteService
}

func NewRouteHandler(service RouteService) *RouteHandler {
	return &RouteHandler{service: service}
}

// CreateRoute godoc
// @Summary Create new geofence route
// @Description Create a corridor route with multiple numbered waypoints
// @Tags Geofence Routes
// @Accept json
// @Produce json
// @Security Bearer
// @Param route body CreateRouteRequest true "Route data"
// @Success 201 {object} map[string]interface{} "success: true, data: route"
// @Router /geofence-routes [post]
func (h *RouteHandler) CreateRoute(c *gin.Context) {
	var req CreateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if callerRole == "company_admin" || callerRole == "operator" {
		if callerCompID == nil {
			base.ErrorResponse(c, http.StatusForbidden, "Company user must belong to a company", nil)
			return
		}
		req.CompanyID = callerCompID
	}

	route, err := h.service.CreateRoute(req)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Gagal membuat rute geofence", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Rute geofence berhasil dibuat", route)
}

// GetRoute godoc
// @Summary Get geofence route by ID
// @Description Get geofence route details including waypoints
// @Tags Geofence Routes
// @Produce json
// @Security Bearer
// @Param id path int true "Route ID"
// @Success 200 {object} map[string]interface{} "success: true, data: route"
// @Router /geofence-routes/{id} [get]
func (h *RouteHandler) GetRoute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid route ID", err)
		return
	}

	route, err := h.service.GetRouteByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Rute geofence tidak ditemukan", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if (callerRole == "company_admin" || callerRole == "operator" || callerRole == "viewer") && callerCompID != nil {
		if route.CompanyID != nil && *route.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusNotFound, "Rute geofence tidak ditemukan", nil)
			return
		}
	}

	base.SuccessResponse(c, http.StatusOK, "Data rute berhasil diambil", route)
}

// GetAllRoutes godoc
// @Summary Get all geofence routes
// @Description Get geofence routes with optional filtering by voyage or vessel
// @Tags Geofence Routes
// @Produce json
// @Security Bearer
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param voyage_id query int false "Voyage ID filter"
// @Param vessel_id query int false "Vessel ID filter"
// @Success 200 {object} map[string]interface{} "success: true, data: routes"
// @Router /geofence-routes [get]
func (h *RouteHandler) GetAllRoutes(c *gin.Context) {
	page, limit, _ := base.GetPaginationParams(c)

	callerRole, callerCompID := extractScope(c)
	var filterCompID *uint
	if callerRole == "company_admin" || callerRole == "operator" || callerRole == "viewer" {
		filterCompID = callerCompID
	} else if cidStr := c.Query("company_id"); cidStr != "" {
		if cid, err := strconv.ParseUint(cidStr, 10, 32); err == nil {
			u := uint(cid)
			filterCompID = &u
		}
	}

	var voyageID *uint
	if vStr := c.Query("voyage_id"); vStr != "" {
		if v, err := strconv.ParseUint(vStr, 10, 32); err == nil {
			u := uint(v)
			voyageID = &u
		}
	}

	var vesselID *uint
	if vesStr := c.Query("vessel_id"); vesStr != "" {
		if ves, err := strconv.ParseUint(vesStr, 10, 32); err == nil {
			u := uint(ves)
			vesselID = &u
		}
	}

	routes, total, err := h.service.GetAllRoutes(page, limit, filterCompID, voyageID, vesselID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data rute", err)
		return
	}

	if page > 0 && limit > 0 {
		base.PaginatedResponse(c, http.StatusOK, "Data rute geofence berhasil diambil", routes, page, limit, total)
	} else {
		base.SuccessResponse(c, http.StatusOK, "Data rute geofence berhasil diambil", routes)
	}
}

// UpdateRoute godoc
// @Summary Update geofence route
// @Description Update route properties or waypoints
// @Tags Geofence Routes
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Route ID"
// @Param route body UpdateRouteRequest true "Updated route data"
// @Success 200 {object} map[string]interface{} "success: true, data: route"
// @Router /geofence-routes/{id} [put]
func (h *RouteHandler) UpdateRoute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid route ID", err)
		return
	}

	var req UpdateRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if callerRole == "company_admin" || callerRole == "operator" {
		if callerCompID == nil {
			base.ErrorResponse(c, http.StatusForbidden, "Company user must belong to a company", nil)
			return
		}
		// Check ownership
		existing, err := h.service.GetRouteByID(uint(id))
		if err != nil {
			base.ErrorResponse(c, http.StatusNotFound, "Rute tidak ditemukan", err)
			return
		}
		if existing.CompanyID != nil && *existing.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Tidak memiliki izin mengubah rute ini", nil)
			return
		}
		req.CompanyID = callerCompID
	}

	route, err := h.service.UpdateRoute(uint(id), req)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Gagal memperbarui rute", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Rute geofence berhasil diperbarui", route)
}

// DeleteRoute godoc
// @Summary Delete geofence route
// @Description Delete route and its waypoints
// @Tags Geofence Routes
// @Produce json
// @Security Bearer
// @Param id path int true "Route ID"
// @Success 200 {object} map[string]string "message: Rute berhasil dihapus"
// @Router /geofence-routes/{id} [delete]
func (h *RouteHandler) DeleteRoute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid route ID", err)
		return
	}

	callerRole, callerCompID := extractScope(c)
	if callerRole == "company_admin" || callerRole == "operator" {
		if callerCompID == nil {
			base.ErrorResponse(c, http.StatusForbidden, "Company user must belong to a company", nil)
			return
		}
		existing, err := h.service.GetRouteByID(uint(id))
		if err != nil {
			base.ErrorResponse(c, http.StatusNotFound, "Rute tidak ditemukan", err)
			return
		}
		if existing.CompanyID != nil && *existing.CompanyID != *callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Tidak memiliki izin menghapus rute ini", nil)
			return
		}
	}

	if err := h.service.DeleteRoute(uint(id)); err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus rute", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Rute geofence berhasil dihapus", nil)
}

// GenerateFromVoyage godoc
// @Summary Auto-generate route from Voyage clearance
// @Description Generate sequential waypoints using voyage origin and destination coordinates
// @Tags Geofence Routes
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body GenerateRouteFromVoyageRequest true "Voyage ID and settings"
// @Success 201 {object} map[string]interface{} "success: true, data: route"
// @Router /geofence-routes/generate-from-voyage [post]
func (h *RouteHandler) GenerateFromVoyage(c *gin.Context) {
	var req GenerateRouteFromVoyageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	_, callerCompID := extractScope(c)

	route, err := h.service.GenerateFromVoyage(req, callerCompID)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Gagal membuat rute dari voyage: "+err.Error(), err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Rute geofence berhasil dibuat dari voyage", route)
}
