package voyage

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

func extractUserAndCompany(c *gin.Context) (userID *uint, companyID *uint, role string) {
	if r, exists := c.Get("role"); exists {
		if s, ok := r.(string); ok {
			role = s
		}
	}
	if uID, exists := c.Get("user_id"); exists {
		if id, ok := uID.(uint); ok {
			userID = &id
		} else if idInt, ok := uID.(int); ok {
			uid := uint(idInt)
			userID = &uid
		}
	}
	if cID, exists := c.Get("company_id"); exists {
		if cid, ok := cID.(uint); ok {
			companyID = &cid
		} else if cidInt, ok := cID.(int); ok {
			cid := uint(cidInt)
			companyID = &cid
		}
	}
	return
}

func (h *Handler) GetAllVoyagesHandler(c *gin.Context) {
	status := c.Query("status")
	vesselIDStr := c.Query("vessel_id")
	search := c.Query("search")
	var vesselID uint
	if vesselIDStr != "" {
		if id, err := strconv.ParseUint(vesselIDStr, 10, 32); err == nil {
			vesselID = uint(id)
		}
	}

	_, companyID, role := extractUserAndCompany(c)
	var targetCompanyID *uint
	if role != "superadmin" {
		targetCompanyID = companyID
	} else if cIDStr := c.Query("company_id"); cIDStr != "" {
		if cid, err := strconv.ParseUint(cIDStr, 10, 32); err == nil && cid > 0 {
			cID := uint(cid)
			targetCompanyID = &cID
		}
	}

	res, err := h.service.GetAllVoyages(status, vesselID, search, targetCompanyID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve voyages", err)
		return
	}
	base.SuccessResponse(c, http.StatusOK, "Voyages retrieved successfully", res)
}

func (h *Handler) CompleteVoyageHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid voyage ID", err)
		return
	}

	_, companyID, role := extractUserAndCompany(c)
	var targetCompanyID *uint
	if role != "superadmin" {
		targetCompanyID = companyID
	}

	res, err := h.service.CompleteVoyage(uint(id), targetCompanyID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
		return
	}
	base.SuccessResponse(c, http.StatusOK, "Voyage completed successfully", res)
}

func (h *Handler) GetAllActiveVoyagesHandler(c *gin.Context) {
	_, companyID, role := extractUserAndCompany(c)
	var targetCompanyID *uint
	if role != "superadmin" {
		targetCompanyID = companyID
	} else if cIDStr := c.Query("company_id"); cIDStr != "" {
		if cid, err := strconv.ParseUint(cIDStr, 10, 32); err == nil && cid > 0 {
			cID := uint(cid)
			targetCompanyID = &cID
		}
	}

	res, err := h.service.GetAllActiveVoyages(targetCompanyID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve active voyages", err)
		return
	}
	base.SuccessResponse(c, http.StatusOK, "Active voyages retrieved successfully", res)
}

func (h *Handler) GetActiveVoyageHandler(c *gin.Context) {
	vesselIDStr := c.Param("vessel_id")
	vesselID, err := strconv.ParseUint(vesselIDStr, 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	res, err := h.service.GetActiveVoyage(uint(vesselID))
	if err != nil {
		// Not found is normal when vessel has no active voyage
		base.SuccessResponse(c, http.StatusOK, "No active voyage", nil)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Active voyage retrieved successfully", res)
}

func (h *Handler) GetVoyagesByVesselHandler(c *gin.Context) {
	vesselIDStr := c.Param("vessel_id")
	vesselID, err := strconv.ParseUint(vesselIDStr, 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid vessel ID", err)
		return
	}

	res, err := h.service.GetVoyagesByVessel(uint(vesselID))
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve voyages", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Voyages retrieved successfully", res)
}

func (h *Handler) CreateVoyageHandler(c *gin.Context) {
	var req CreateVoyageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	_, companyID, role := extractUserAndCompany(c)
	var targetCompanyID *uint
	if role != "superadmin" {
		targetCompanyID = companyID
	}

	res, err := h.service.CreateVoyage(req, targetCompanyID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Voyage created successfully", res)
}

func (h *Handler) UpdateVoyageHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid voyage ID", err)
		return
	}

	var req UpdateVoyageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	_, companyID, role := extractUserAndCompany(c)
	var targetCompanyID *uint
	if role != "superadmin" {
		targetCompanyID = companyID
	}

	res, err := h.service.UpdateVoyage(uint(id), req, targetCompanyID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Voyage updated successfully", res)
}

func (h *Handler) DeleteVoyageHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid voyage ID", err)
		return
	}

	_, companyID, role := extractUserAndCompany(c)
	var targetCompanyID *uint
	if role != "superadmin" {
		targetCompanyID = companyID
	}

	if err := h.service.DeleteVoyage(uint(id), targetCompanyID); err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Voyage deleted successfully", nil)
}
