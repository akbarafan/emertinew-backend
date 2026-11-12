package ticket

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/base"
	"github.com/Alugara-Inovasi-Utama/emerti/internal/telegram"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
	tg      telegram.Service
}

func NewHandler(service Service, tg telegram.Service) *Handler {
	return &Handler{service: service, tg: tg}
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

type BulkDeleteTicketsRequest struct {
	TicketIDs []uint `json:"ticketIds" binding:"required"`
}

func (h *Handler) DeleteTicketHandler(c *gin.Context) {
	_, _, role := extractUserAndCompany(c)
	if role != "superadmin" {
		base.ErrorResponse(c, http.StatusForbidden, "Hanya Super Admin yang memiliki hak akses menghapus tiket", nil)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "ID tiket tidak valid", err)
		return
	}

	if err := h.service.DeleteTicket(uint(id)); err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus tiket: "+err.Error(), err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, "Tiket berhasil dihapus permanen", nil)
}

func (h *Handler) BulkDeleteTicketsHandler(c *gin.Context) {
	_, _, role := extractUserAndCompany(c)
	if role != "superadmin" {
		base.ErrorResponse(c, http.StatusForbidden, "Hanya Super Admin yang memiliki hak akses menghapus tiket", nil)
		return
	}

	var req BulkDeleteTicketsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "Data tiket yang akan dihapus tidak valid: "+err.Error(), err)
		return
	}

	if len(req.TicketIDs) == 0 {
		base.ErrorResponse(c, http.StatusBadRequest, "Daftar tiket yang akan dihapus tidak boleh kosong", nil)
		return
	}

	if err := h.service.BulkDeleteTickets(req.TicketIDs); err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus tiket: "+err.Error(), err)
		return
	}

	base.SuccessResponse(c, http.StatusOK, fmt.Sprintf("%d tiket berhasil dihapus permanen", len(req.TicketIDs)), nil)
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

func resolveCompanyFolder(c *gin.Context) string {
	// 1. Check if explicitly provided in Form or Query parameter
	companyInput := strings.TrimSpace(c.PostForm("companyCode"))
	if companyInput == "" {
		companyInput = strings.TrimSpace(c.PostForm("company"))
	}
	if companyInput == "" {
		companyInput = strings.TrimSpace(c.PostForm("companyName"))
	}
	if companyInput == "" {
		companyInput = strings.TrimSpace(c.Query("companyCode"))
	}

	// 2. Fall back to JWT context claims
	if companyInput == "" {
		if cName, exists := c.Get("company_name"); exists {
			if s, ok := cName.(string); ok && strings.TrimSpace(s) != "" {
				companyInput = s
			}
		}
	}
	if companyInput == "" {
		if cID, exists := c.Get("company_id"); exists {
			if id, ok := cID.(uint); ok && id > 0 {
				companyInput = fmt.Sprintf("company_%d", id)
			} else if idInt, ok := cID.(int); ok && idInt > 0 {
				companyInput = fmt.Sprintf("company_%d", idInt)
			}
		}
	}

	if companyInput == "" {
		return "global"
	}

	// 3. Sanitize folder name into a clean, safe path (lowercase alphanumeric and underscores)
	folder := strings.ToLower(companyInput)
	var sb strings.Builder
	for _, r := range folder {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			sb.WriteRune(r)
		} else if r == ' ' || r == '/' || r == '\\' || r == '.' {
			sb.WriteRune('_')
		}
	}
	clean := strings.Trim(sb.String(), "_-")
	if clean == "" {
		return "global"
	}
	return clean
}

func (h *Handler) UploadAttachmentHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		base.ErrorResponse(c, http.StatusBadRequest, "No file uploaded", err)
		return
	}

	// 10MB limit check
	if file.Size > 10*1024*1024 {
		base.ErrorResponse(c, http.StatusBadRequest, "File size exceeds 10MB limit", nil)
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".pdf":  true,
		".csv":  true,
		".xlsx": true,
		".docx": true,
		".txt":  true,
	}
	if !allowed[ext] {
		base.ErrorResponse(c, http.StatusBadRequest, "Format file tidak didukung. Ekstensi diizinkan: jpg, png, webp, pdf, csv, xlsx, docx, txt", nil)
		return
	}

	companyFolder := resolveCompanyFolder(c)
	uploadDir := filepath.Join("./assets/uploads/tickets", companyFolder)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to create upload directory", err)
		return
	}

	safeName := strings.ReplaceAll(file.Filename, " ", "_")
	filename := fmt.Sprintf("tkt_%d_%s", time.Now().Unix(), safeName)
	dst := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		base.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file", err)
		return
	}

	fileURL := fmt.Sprintf("/uploads/tickets/%s/%s", companyFolder, filename)

	base.SuccessResponse(c, http.StatusOK, "File uploaded successfully", gin.H{
		"url":           fileURL,
		"filename":      file.Filename,
		"size":          file.Size,
		"companyFolder": companyFolder,
	})
}

func (h *Handler) TelegramWebhookHandler(c *gin.Context) {
	if h.tg == nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "message": "Telegram integration not enabled"})
		return
	}

	var update telegram.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
		return
	}

	if err := h.tg.HandleUpdate(update); err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
