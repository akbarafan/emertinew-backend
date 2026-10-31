package sensorhealth

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

func (h *Handler) GetSensorHealth(c *gin.Context) {
	var req SensorHealthRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	data, err := h.service.GetSensorHealth(req.Month, req.VesselID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to get sensor health", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Sensor health retrieved", data)
}
