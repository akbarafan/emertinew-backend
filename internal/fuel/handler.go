package fuel

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

// GetFuelTotalHandler mendapatkan total fuel semua vessel pada hari tertentu
// @Summary      Get fuel total all vessels
// @Description  Mendapatkan total flow1, flow2, flow3, flow4, dan fuel_total dari trackings pada hari tertentu untuk semua vessel
// @Tags         8. Fuel
// @Produce      json
// @Security     Bearer
// @Param        date query string true "Date (YYYY-MM-DD, e.g., 2026-04-17)"
// @Success      200 {object} map[string]interface{} "success: true, data: fuel total array"
// @Failure      400 {object} map[string]string "error: date is required"
// @Failure      401 {object} map[string]string "error: Unauthorized"
// @Failure      500 {object} map[string]string "error: Failed to fetch fuel total"
// @Router       /fuel [get]
func (h *Handler) GetFuelTotalHandler(c *gin.Context) {
	var params FuelQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	if params.Date == "" {
		base.ErrorResponse(c, http.StatusBadRequest, "date is required", nil)
		return
	}

	date, err := time.ParseInLocation("2006-01-02", params.Date, base.WIB)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid date format, use YYYY-MM-DD", err)
		return
	}

	results, err := h.service.GetFuelTotal(date)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch fuel total", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Fuel total fetched successfully", results)
}

// GetFuelTotalByRangeHandler mendapatkan total fuel semua vessel pada rentang waktu tertentu
// @Summary      Get fuel total for a time window
// @Description  Mengambil total konsumsi bahan bakar semua vessel dalam rentang waktu, untuk membandingkan dua periode dengan durasi yang sama
// @Tags         8. Fuel
// @Produce      json
// @Security     Bearer
// @Param        start query string true "Start time (RFC3339, e.g., 2026-04-17T00:00:00+07:00)"
// @Param        end   query string true "End time (RFC3339, e.g., 2026-04-17T12:00:00+07:00)"
// @Success      200 {object} map[string]interface{} "success: true, data: fuel total array"
// @Failure      400 {object} map[string]string "error: start and end are required"
// @Failure      401 {object} map[string]string "error: Unauthorized"
// @Failure      500 {object} map[string]string "error: Failed to fetch fuel total"
// @Router       /fuel/window [get]
func (h *Handler) GetFuelTotalByRangeHandler(c *gin.Context) {
	start, err := time.Parse(time.RFC3339, c.Query("start"))
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid start time, use RFC3339", err)
		return
	}
	end, err := time.Parse(time.RFC3339, c.Query("end"))
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid end time, use RFC3339", err)
		return
	}
	if !end.After(start) {
		base.ErrorResponse(c, http.StatusBadRequest, "end must be after start", nil)
		return
	}

	results, err := h.service.GetFuelTotalByRange(start, end)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch fuel total", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Fuel total fetched successfully", results)
}

// GetFuelTotalByVesselHandler mendapatkan total fuel untuk vessel tertentu pada hari tertentu
// @Summary      Get fuel total by vessel
// @Description  Mendapatkan total flow1, flow2, flow3, flow4, dan fuel_total dari trackings pada hari tertentu untuk vessel tertentu
// @Tags         8. Fuel
// @Produce      json
// @Security     Bearer
// @Param        vessel_id path int true "Vessel ID"
// @Param        date query string true "Date (YYYY-MM-DD, e.g., 2026-04-17)"
// @Success      200 {object} map[string]interface{} "success: true, data: fuel total"
// @Failure      400 {object} map[string]string "error: date is required"
// @Failure      401 {object} map[string]string "error: Unauthorized"
// @Failure      500 {object} map[string]string "error: Failed to fetch fuel total"
// @Router       /fuel/{vessel_id} [get]
func (h *Handler) GetFuelTotalByVesselHandler(c *gin.Context) {
	vesselID, err := strconv.ParseUint(c.Param("vessel_id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	if c.Query("lifetime") == "true" || c.Query("all_time") == "true" {
		result, err := h.service.GetLifetimeFuelByVessel(uint(vesselID))
		if err != nil {
			base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch lifetime fuel total", err)
			return
		}
		base.SuccessResponse(c, http.StatusOK, "Lifetime fuel total fetched successfully", result)
		return
	}

	var params FuelQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	if params.Date == "" {
		base.ErrorResponse(c, http.StatusBadRequest, "date is required", nil)
		return
	}

	date, err := time.ParseInLocation("2006-01-02", params.Date, base.WIB)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid date format, use YYYY-MM-DD", err)
		return
	}

	result, err := h.service.GetFuelTotalByVessel(uint(vesselID), date)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch fuel total", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Fuel total fetched successfully", result)
}
