package alert

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/base"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// SyncEventsHandler records the alert transitions observed in one poll.
//
// @Summary      Sync alert events
// @Description  Upserts newly observed conditions and closes any open alert the client stopped reporting
// @Tags         9. Alerts
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        events body UpsertEventsRequest true "Currently observed alert conditions"
// @Success      200 {object} map[string]interface{} "success: true"
// @Failure      400 {object} map[string]string "error: invalid payload"
// @Router       /alerts/events [post]
func (h *Handler) SyncEventsHandler(c *gin.Context) {
	var body UpsertEventsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Derive the key set from the payload rather than trusting a second list,
	// so a client cannot accidentally claim a key is still present while also
	// omitting its row.
	keys := make(map[string]bool, len(body.Events))
	for _, e := range body.Events {
		keys[e.EventKey] = true
	}

	result, err := h.service.SyncEvents(body.Events, keys, body.Complete)
	if err != nil {
		badRequest(err, c)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Alert events synced", result)
}

// ListEventsHandler returns a page of alert history.
//
// @Summary      List alert history
// @Tags         9. Alerts
// @Produce      json
// @Security     Bearer
// @Param        vessel_id query int    false "Filter by vessel"
// @Param        type      query string false "Filter by alert type"
// @Param        severity  query string false "Filter by severity (info|warning|error)"
// @Param        from      query string false "Start date YYYY-MM-DD (inclusive, WIB)"
// @Param        to        query string false "End date YYYY-MM-DD (inclusive, WIB)"
// @Param        open_only query bool   false "Only unresolved alerts"
// @Param        page      query int    false "Page number (default 1)"
// @Param        limit     query int    false "Page size (default 50, max 200)"
// @Success      200 {object} map[string]interface{} "success: true"
// @Failure      400 {object} map[string]string "error: invalid query"
// @Router       /alerts/events [get]
func (h *Handler) ListEventsHandler(c *gin.Context) {
	var query ListEventsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	result, err := h.service.ListEvents(query)
	if err != nil {
		badRequest(err, c)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Alert history fetched", result)
}

// AckEventHandler acknowledges one alert without resolving it.
//
// @Summary      Acknowledge an alert
// @Tags         9. Alerts
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path int  true "Alert event ID"
// @Param        note body AckEventRequest false "Optional note"
// @Success      200 {object} map[string]interface{} "success: true"
// @Failure      404 {object} map[string]string "error: not found"
// @Router       /alerts/events/{id}/ack [patch]
func (h *Handler) AckEventHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid alert ID", err)
		return
	}

	var body AckEventRequest
	// An empty body is valid: acknowledging needs no payload.
	_ = c.ShouldBindJSON(&body)

	result, err := h.service.AckEvent(uint(id), body.Note)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			base.ErrorResponse(c, http.StatusNotFound, "Alert event not found", err)
			return
		}
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to acknowledge alert", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Alert acknowledged", result)
}

func badRequest(err error, c *gin.Context) {
	switch {
	case errors.Is(err, ErrMissingEventKey),
		errors.Is(err, ErrDuplicateEventKey),
		errors.Is(err, ErrMissingDetectedAt),
		errors.Is(err, ErrDetectedAtInFuture),
		errors.Is(err, ErrInvalidSeverity),
		errors.Is(err, ErrInvalidDate):
		base.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
	default:
		base.ErrorResponse(c, http.StatusInternalServerError, "Alert sync failed", err)
	}
}
