package audit

import (
	"net/http"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/base"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetHistory returns audit history for a specific record
// @Summary      Get audit history
// @Description  Mendapatkan history audit untuk record tertentu
// @Tags         Audit Log
// @Produce      json
// @Security     Bearer
// @Param        table_name query string true "Table name (e.g. auth, m_company, m_vessel)"
// @Param        record_id query int false "Record ID (optional, filter by specific record)"
// @Param        action query string false "Action filter (e.g. LOGIN, LOGIN_FAILED, LOGOUT, CREATE, UPDATE, DELETE)"
// @Param        page query int false "Page number (default: no pagination)"
// @Param        limit query int false "Items per page (default: no pagination, max: 1000)"
// @Success      200 {object} map[string]interface{} "success: true, data: audit logs array"
// @Failure      400 {object} map[string]string "error: Invalid request"
// @Failure      500 {object} map[string]string "error: Failed to fetch audit history"
// @Router       /audit/history [get]
func (h *Handler) GetHistory(c *gin.Context) {
	var req AuditHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// Manual validation
	if req.TableName == "" {
		base.ErrorResponse(c, http.StatusBadRequest, "table_name is required", nil)
		return
	}

	// Pagination params
	page, limit, offset := base.GetPaginationParams(c)
	if limit > 1000 {
		limit = 1000
		offset = (page - 1) * limit
	}

	logs, total, err := h.service.GetHistory(req.TableName, req.RecordID, req.Action, limit, offset)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch audit history", err)
		return
	}

	// Return paginated if pagination params provided, otherwise return all
	if page > 0 && limit > 0 {
		base.PaginatedResponse(c, http.StatusOK, "Audit history fetched successfully", logs, page, limit, total)
	} else {
		base.SuccessResponse(c, http.StatusOK, "Audit history fetched successfully", logs)
	}
}

// GetUserActivity returns audit logs for a specific user
// @Summary      Get user activity
// @Description  Mendapatkan activity logs untuk user tertentu
// @Tags         Audit Log
// @Produce      json
// @Security     Bearer
// @Param        user_id query int true "User ID"
// @Param        limit query int false "Limit (default 100, max 1000)"
// @Success      200 {object} map[string]interface{} "success: true, data: audit logs array"
// @Failure      400 {object} map[string]string "error: Invalid request"
// @Failure      500 {object} map[string]string "error: Failed to fetch user activity"
// @Router       /audit/activity [get]
func (h *Handler) GetUserActivity(c *gin.Context) {
	var req UserActivityRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// Manual validation
	if req.UserID == 0 {
		base.ErrorResponse(c, http.StatusBadRequest, "user_id is required", nil)
		return
	}

	// Default limit to 100 if not specified
	limit := req.Limit
	if limit == 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	logs, err := h.service.GetUserActivity(req.UserID, limit)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch user activity", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "User activity fetched successfully", logs)
}
