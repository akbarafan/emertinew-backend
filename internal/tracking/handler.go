package tracking

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/base"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ========================================
// NEW API: /tracking
// ========================================

// GetAllTrackingHandler mendapatkan semua tracking data (per-minute)
// @Summary      Get all tracking
// @Description  Mendapatkan semua tracking data per-minute. Default sort DESC (terbaru dulu). Jika pakai start_time, sort menjadi ASC.
// @Tags         7. Tracking
// @Produce      json
// @Security     Bearer
// @Param        start_time query string false "Start time (RFC3339, e.g., 2024-01-01T00:00:00Z) - jika ada, sort default ASC"
// @Param        end_time query string false "End time (RFC3339) - memerlukan start_time"
// @Param        sort query string false "Sort order: asc or desc (default: desc, atau asc jika pakai start_time)"
// @Param        limit query int false "Limit jumlah data"
// @Success      200 {object} map[string]interface{} "success: true, data: tracking array"
// @Failure      400 {object} map[string]string "error: end_time requires start_time"
// @Failure      401 {object} map[string]string "error: Unauthorized"
// @Failure      500 {object} map[string]string "error: Failed to fetch tracking"
// @Router       /tracking [get]
func (h *Handler) GetAllTrackingHandler(c *gin.Context) {
	var params TrackingQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// Validate: end_time requires start_time
	if params.EndTime != "" && params.StartTime == "" {
		base.ErrorResponse(c, http.StatusBadRequest, "end_time requires start_time", nil)
		return
	}

	startTime, endTime, err := parseTimeRangeRFC3339(params.StartTime, params.EndTime)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Determine sort order:
	// - Default (no filter): DESC (newest first)
	// - With start_time filter: ASC (chronological)
	// - User can override with sort param
	if params.Sort == "" {
		if params.StartTime != "" {
			params.Sort = "asc" // Chronological when filtering by time
		} else {
			params.Sort = "desc" // Newest first by default
		}
	}

	// Default limit to 100 if not specified
	if params.Limit <= 0 {
		params.Limit = 100
	}

	hasCoords := false
	if params.HasCoords != nil {
		hasCoords = *params.HasCoords
	}

	trackings, err := h.service.GetAllTracking(startTime, endTime, params.Sort, params.Limit, hasCoords)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tracking", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Tracking fetched successfully", trackings)
}

// GetTrackingByVesselHandler mendapatkan tracking untuk vessel tertentu
// @Summary      Get tracking by vessel
// @Description  Mendapatkan tracking data per-minute untuk vessel tertentu. Default sort DESC (terbaru dulu). Jika pakai start_time, sort menjadi ASC.
// @Tags         7. Tracking
// @Produce      json
// @Security     Bearer
// @Param        vessel_id path int true "Vessel ID"
// @Param        start_time query string false "Start time (RFC3339, e.g., 2024-01-01T00:00:00Z) - jika ada, sort default ASC"
// @Param        end_time query string false "End time (RFC3339) - memerlukan start_time"
// @Param        sort query string false "Sort order: asc or desc (default: desc, atau asc jika pakai start_time)"
// @Param        limit query int false "Limit jumlah data"
// @Success      200 {object} map[string]interface{} "success: true, data: tracking array"
// @Failure      400 {object} map[string]string "error: end_time requires start_time"
// @Failure      401 {object} map[string]string "error: Unauthorized"
// @Failure      500 {object} map[string]string "error: Failed to fetch tracking"
// @Router       /tracking/{vessel_id} [get]
func (h *Handler) GetTrackingByVesselHandler(c *gin.Context) {
	vesselID, err := strconv.ParseUint(c.Param("vessel_id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	var params TrackingQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// Validate: end_time requires start_time
	if params.EndTime != "" && params.StartTime == "" {
		base.ErrorResponse(c, http.StatusBadRequest, "end_time requires start_time", nil)
		return
	}

	startTime, endTime, err := parseTimeRangeRFC3339(params.StartTime, params.EndTime)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Determine sort order:
	// - Default (no filter): DESC (newest first)
	// - With start_time filter: ASC (chronological)
	// - User can override with sort param
	if params.Sort == "" {
		if params.StartTime != "" {
			params.Sort = "asc" // Chronological when filtering by time
		} else {
			params.Sort = "desc" // Newest first by default
		}
	}

	// Default limit to 100 if not specified
	if params.Limit <= 0 {
		params.Limit = 100
	}

	// Determine hasCoords filter
	hasCoords := false
	if params.HasCoords != nil {
		hasCoords = *params.HasCoords
	}

	trackings, err := h.service.GetTracking(uint(vesselID), startTime, endTime, params.Sort, params.Limit, hasCoords)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tracking", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Tracking fetched successfully", trackings)
}

// parseTimeRangeRFC3339 parses RFC3339 formatted start and end time strings
// - No params: no time filter (get all data up to now)
// - start_time only: from start_time until now
// - start_time + end_time: specific range
func parseTimeRangeRFC3339(startTimeStr, endTimeStr string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	} else {
		// Default: no start filter (from epoch / very old date)
		startTime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	} else {
		// Default: now
		endTime = time.Now()
	}

	return startTime, endTime, nil
}
