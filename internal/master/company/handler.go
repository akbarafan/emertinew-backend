package company

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

// CreateCompany godoc
// @Summary Create company
// @Description Create a new company
// @Tags Master Data - Company
// @Accept json
// @Produce json
// @Security Bearer
// @Param company body CreateCompanyRequest true "Company data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /master/company [post]
func (h *Handler) CreateCompany(c *gin.Context) {
	var req CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.CreateCompany(req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Company already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to create company", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Company created successfully", response)
}

// GetCompanyByID godoc
// @Summary Get company by ID
// @Description Get company details by ID
// @Tags Master Data - Company
// @Produce json
// @Security Bearer
// @Param id path int true "Company ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /master/company/{id} [get]
func (h *Handler) GetCompanyByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid company ID", err)
		return
	}

	response, err := h.service.GetCompanyByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Company not found", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Company fetched successfully", response)
}

// GetAllCompanies godoc
// @Summary Get all companies
// @Description Get all companies
// @Tags Master Data - Company
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /master/company [get]
func (h *Handler) GetAllCompanies(c *gin.Context) {
	response, err := h.service.GetAllCompanies()
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch companies", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Companies fetched successfully", response)
}

// UpdateCompany godoc
// @Summary Update company
// @Description Update company by ID
// @Tags Master Data - Company
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Company ID"
// @Param company body UpdateCompanyRequest true "Company data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /master/company/{id} [patch]
func (h *Handler) UpdateCompany(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid company ID", err)
		return
	}

	callerRole := c.GetString("role")
	if callerRole == "company_admin" {
		var callerCompID uint
		if cid, exists := c.Get("company_id"); exists {
			switch v := cid.(type) {
			case uint:
				callerCompID = v
			case int:
				callerCompID = uint(v)
			}
		}
		if callerCompID == 0 || uint(id) != callerCompID {
			base.ErrorResponse(c, http.StatusForbidden, "Access denied: can only update your own company profile", nil)
			return
		}
	}

	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	response, err := h.service.UpdateCompany(uint(id), req, userID, ipAddress, userAgent)
	if err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Company not found", err)
			return
		}
		if err.Error() == "duplicate record" {
			base.ErrorResponse(c, http.StatusConflict, "Company code already exists", err)
			return
		}
		base.ErrorResponse(c, http.StatusBadRequest, "Failed to update company", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Company updated successfully", response)
}

// DeleteCompany godoc
// @Summary Delete company
// @Description Delete company by ID
// @Tags Master Data - Company
// @Produce json
// @Security Bearer
// @Param id path int true "Company ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /master/company/{id} [delete]
func (h *Handler) DeleteCompany(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid company ID", err)
		return
	}

	userID := utils.GetUserIDFromContext(c)
	ipAddress := utils.GetClientIP(c)
	userAgent := utils.GetUserAgent(c)

	if err := h.service.DeleteCompany(uint(id), userID, ipAddress, userAgent); err != nil {
		if err.Error() == "record not found" {
			base.ErrorResponse(c, http.StatusNotFound, "Company not found", err)
			return
		}
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete company", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Company deleted successfully", nil)
}
