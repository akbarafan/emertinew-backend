package geo

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

// ── Assignment CRUD ────────────────────────────────────────────────

func (h *Handler) CreateAssignment(c *gin.Context) {
	var req CreateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Request tidak valid", err)
		return
	}

	assignment, err := h.service.CreateAssignment(req)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Gagal membuat penugasan", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Penugasan berhasil dibuat", assignment)
}

func (h *Handler) GetAssignment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	detail, err := h.service.GetAssignment(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Penugasan tidak ditemukan", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Detail penugasan", detail)
}

func (h *Handler) ListAssignments(c *gin.Context) {
	page, limit, _ := base.GetPaginationParams(c)

	var vesselID, geofenceID *uint
	if vid := c.Query("vessel_id"); vid != "" {
		if v, err := strconv.ParseUint(vid, 10, 32); err == nil {
			id := uint(v)
			vesselID = &id
		}
	}
	if gid := c.Query("geofence_id"); gid != "" {
		if v, err := strconv.ParseUint(gid, 10, 32); err == nil {
			id := uint(v)
			geofenceID = &id
		}
	}
	status := c.Query("status")

	assignments, total, err := h.service.ListAssignments(vesselID, geofenceID, status, page, limit)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data penugasan", err)
		return
	}

	if page > 0 && limit > 0 {
		base.PaginatedResponse(c, http.StatusOK, "Data penugasan berhasil diambil", assignments, page, limit, total)
	} else {
		base.SuccessResponse(c, http.StatusOK, "Data penugasan berhasil diambil", assignments)
	}
}

func (h *Handler) UpdateAssignment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	var req UpdateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Request tidak valid", err)
		return
	}

	assignment, err := h.service.UpdateAssignment(uint(id), req)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Gagal update penugasan", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Penugasan berhasil diupdate", assignment)
}

func (h *Handler) DeleteAssignment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid", err)
		return
	}

	if err := h.service.DeleteAssignment(uint(id)); err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Gagal hapus penugasan", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Penugasan berhasil dihapus", nil)
}

// ── CP Logs ────────────────────────────────────────────────────────

func (h *Handler) ListCpLogs(c *gin.Context) {
	page, limit, _ := base.GetPaginationParams(c)

	var assignmentID, vesselID, geofenceID *uint
	if aid := c.Query("assignment_id"); aid != "" {
		if v, err := strconv.ParseUint(aid, 10, 32); err == nil {
			id := uint(v)
			assignmentID = &id
		}
	}
	if vid := c.Query("vessel_id"); vid != "" {
		if v, err := strconv.ParseUint(vid, 10, 32); err == nil {
			id := uint(v)
			vesselID = &id
		}
	}
	if gid := c.Query("geofence_id"); gid != "" {
		if v, err := strconv.ParseUint(gid, 10, 32); err == nil {
			id := uint(v)
			geofenceID = &id
		}
	}

	logs, total, err := h.service.ListCpLogs(assignmentID, vesselID, geofenceID, page, limit)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil log CP", err)
		return
	}

	if page > 0 && limit > 0 {
		base.PaginatedResponse(c, http.StatusOK, "Log CP berhasil diambil", logs, page, limit, total)
	} else {
		base.SuccessResponse(c, http.StatusOK, "Log CP berhasil diambil", logs)
	}
}

// ── Violation Logs ─────────────────────────────────────────────────

func (h *Handler) ListViolationLogs(c *gin.Context) {
	page, limit, _ := base.GetPaginationParams(c)

	var vesselID, geofenceID *uint
	if vid := c.Query("vessel_id"); vid != "" {
		if v, err := strconv.ParseUint(vid, 10, 32); err == nil {
			id := uint(v)
			vesselID = &id
		}
	}
	if gid := c.Query("geofence_id"); gid != "" {
		if v, err := strconv.ParseUint(gid, 10, 32); err == nil {
			id := uint(v)
			geofenceID = &id
		}
	}

	var isOutside *bool
	if io := c.Query("is_outside"); io != "" {
		b := io == "true"
		isOutside = &b
	}

	logs, total, err := h.service.ListViolationLogs(vesselID, geofenceID, isOutside, page, limit)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil log violation", err)
		return
	}

	if page > 0 && limit > 0 {
		base.PaginatedResponse(c, http.StatusOK, "Log violation berhasil diambil", logs, page, limit, total)
	} else {
		base.SuccessResponse(c, http.StatusOK, "Log violation berhasil diambil", logs)
	}
}
