package ticket

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

func (h *Handler) GetAllTicketsHandler(c *gin.Context) {
	var filter TicketFilterQuery
	if err := c.ShouldBindQuery(&filter); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	userID, companyID, role := extractUserAndCompany(c)

	// RBAC Scoping: Non-superadmin is strictly scoped to ONLY tickets they submitted
	if role != "superadmin" {
		filter.ReporterUserID = userID
		if companyID != nil && *companyID > 0 {
			filter.CompanyID = companyID
		}
	}

	tickets, err := h.service.GetAllTickets(filter)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve tickets", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Tickets retrieved successfully", tickets)
}

func (h *Handler) GetTicketByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid ticket ID", err)
		return
	}

	ticket, err := h.service.GetTicketByID(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusNotFound, "Ticket not found", err)
		return
	}

	// RBAC Security check: If non-superadmin, check that ticket belongs to this user
	userID, companyID, role := extractUserAndCompany(c)
	if role != "superadmin" {
		if ticket.ReporterUserID != nil && userID != nil && *ticket.ReporterUserID != *userID {
			base.ErrorResponse(c, http.StatusForbidden, "Access denied: you can only view tickets you submitted", nil)
			return
		}
		if ticket.CompanyID != nil && companyID != nil && *ticket.CompanyID != *companyID {
			base.ErrorResponse(c, http.StatusForbidden, "Access denied to this ticket", nil)
			return
		}
	}

	base.SuccessResponse(c, http.StatusOK, "Ticket retrieved successfully", ticket)
}

func (h *Handler) CreateTicketHandler(c *gin.Context) {
	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	userID, companyID, _ := extractUserAndCompany(c)

	res, err := h.service.CreateTicket(req, userID, companyID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to create support ticket", err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Support ticket created successfully", res)
}

func (h *Handler) AcknowledgeTicketHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid ticket ID", err)
		return
	}

	var req AcknowledgeTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Optional body
		req = AcknowledgeTicketRequest{}
	}

	if req.AdminName == "" {
		if uName, exists := c.Get("username"); exists {
			if s, ok := uName.(string); ok && s != "" {
				req.AdminName = s
			}
		}
	}

	res, err := h.service.AcknowledgeTicket(uint(id), req)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to acknowledge ticket", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Ticket acknowledged successfully", res)
}

func (h *Handler) ReplyTicketHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid ticket ID", err)
		return
	}

	var req ReplyTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Message is required", err)
		return
	}

	userID, _, role := extractUserAndCompany(c)

	// RBAC Security check: If non-superadmin, check that ticket belongs to this user
	if role != "superadmin" {
		ticket, err := h.service.GetTicketByID(uint(id))
		if err != nil {
			base.ErrorResponse(c, http.StatusNotFound, "Ticket not found", err)
			return
		}
		if ticket.ReporterUserID != nil && userID != nil && *ticket.ReporterUserID != *userID {
			base.ErrorResponse(c, http.StatusForbidden, "Access denied: you can only reply to your own tickets", nil)
			return
		}
	}

	if req.SenderRole == "" {
		if role == "superadmin" || role == "admin" {
			req.SenderRole = "admin"
		} else {
			req.SenderRole = "user"
		}
	}

	if req.SenderName == "" {
		if uName, exists := c.Get("username"); exists {
			if s, ok := uName.(string); ok && s != "" {
				req.SenderName = s
			}
		}
	}

	res, err := h.service.ReplyTicket(uint(id), req, userID)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	base.SuccessResponse(c, http.StatusCreated, "Reply added successfully", res)
}

func (h *Handler) ResolveTicketHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid ticket ID", err)
		return
	}

	var req ResolveTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Resolution notes are required", err)
		return
	}

	if req.AdminName == "" {
		if uName, exists := c.Get("username"); exists {
			if s, ok := uName.(string); ok && s != "" {
				req.AdminName = s
			}
		}
	}

	res, err := h.service.ResolveTicket(uint(id), req)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to resolve ticket", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Ticket resolved successfully", res)
}

func (h *Handler) CloseTicketHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Invalid ticket ID", err)
		return
	}

	res, err := h.service.CloseTicket(uint(id))
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Ticket closed successfully", res)
}

func (h *Handler) GetMetricsHandler(c *gin.Context) {
	userID, companyID, role := extractUserAndCompany(c)

	var targetCompanyID *uint
	var targetUserID *uint

	if role == "superadmin" {
		if q := c.Query("companyId"); q != "" {
			if parsed, err := strconv.ParseUint(q, 10, 32); err == nil && parsed > 0 {
				uid := uint(parsed)
				targetCompanyID = &uid
			}
		}
		if qUser := c.Query("reporterUserId"); qUser != "" {
			if parsed, err := strconv.ParseUint(qUser, 10, 32); err == nil && parsed > 0 {
				uid := uint(parsed)
				targetUserID = &uid
			}
		}
	} else {
		targetUserID = userID
		if companyID != nil && *companyID > 0 {
			targetCompanyID = companyID
		}
	}

	metrics, err := h.service.GetMetrics(targetCompanyID, targetUserID)
	if err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve ticket metrics", err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Ticket metrics retrieved successfully", metrics)
}
