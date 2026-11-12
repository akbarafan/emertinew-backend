package ticket

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/telegram"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

type Service interface {
	CreateTicket(req CreateTicketRequest, userID *uint, userCompanyID *uint) (*TicketResponse, error)
	GetAllTickets(filter TicketFilterQuery) ([]TicketResponse, error)
	GetTicketByID(id uint) (*TicketResponse, error)
	AcknowledgeTicket(id uint, req AcknowledgeTicketRequest) (*TicketResponse, error)
	ReplyTicket(id uint, req ReplyTicketRequest, userID *uint) (*TicketReplyResponse, error)
	ResolveTicket(id uint, req ResolveTicketRequest) (*TicketResponse, error)
	CloseTicket(id uint) (*TicketResponse, error)
	DeleteTicket(id uint) error
	BulkDeleteTickets(ids []uint) error
	GetMetrics(companyID *uint, reporterUserID *uint) (*TicketMetricsResponse, error)

	// Telegram integration methods
	SetTelegramService(tg telegram.Service)
	GetTicketByIDModel(id uint) (*model.SupportTicket, error)
	GetTicketByTelegramMessageID(msgID int) (*model.SupportTicket, error)
	AcknowledgeTicketFromTelegram(ticketID uint, adminName string) (*model.SupportTicket, error)
	ReplyTicketFromTelegram(ticketID uint, replyText string, senderName string, tgMsgID int) (*model.TicketReply, error)
	ResolveTicketFromTelegram(ticketID uint, adminName string, notes string) (*model.SupportTicket, error)
	GetUnhandledTickets() ([]model.SupportTicket, error)
	GetCriticalTickets() ([]model.SupportTicket, error)
	GetActiveTickets() ([]model.SupportTicket, error)
	GetNOCMetrics() (*telegram.TicketMetricsSummary, error)
}

type service struct {
	repo Repository
	tg   telegram.Service
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) SetTelegramService(tg telegram.Service) {
	s.tg = tg
}

func (s *service) CreateTicket(req CreateTicketRequest, userID *uint, userCompanyID *uint) (*TicketResponse, error) {
	ticketNumber, err := s.repo.GetNextTicketNumber()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Compute SLA target deadline based on priority (Critical is max 3 hours)
	var slaDuration time.Duration
	switch strings.ToLower(req.Priority) {
	case "critical":
		slaDuration = 3 * time.Hour
	case "high":
		slaDuration = 8 * time.Hour
	case "medium":
		slaDuration = 24 * time.Hour
	case "low":
		slaDuration = 48 * time.Hour
	default:
		req.Priority = "medium"
		slaDuration = 24 * time.Hour
	}

	slaTargetAt := now.Add(slaDuration)

	// Resolve CompanyID
	var companyID *uint
	if userCompanyID != nil && *userCompanyID > 0 {
		companyID = userCompanyID
	} else if req.CompanyID != nil && *req.CompanyID > 0 {
		companyID = req.CompanyID
	} else if req.VesselID != nil && *req.VesselID > 0 {
		vessel, err := s.repo.GetVesselByID(*req.VesselID)
		if err == nil && vessel != nil && vessel.CompanyID != nil {
			companyID = vessel.CompanyID
		}
	}

	vesselStatus := req.VesselStatus
	if vesselStatus == "" {
		vesselStatus = "unknown"
	}

	t := &model.SupportTicket{
		TicketNumber:   ticketNumber,
		VesselID:       req.VesselID,
		CompanyID:      companyID,
		ReporterUserID: userID,
		ReporterName:   req.ReporterName,
		ReporterPhone:  req.ReporterPhone,
		ReporterEmail:  req.ReporterEmail,
		Category:       req.Category,
		Priority:       req.Priority,
		Subject:        req.Subject,
		Description:    req.Description,
		Status:         "open",
		SlaTargetAt:    slaTargetAt,
		VesselStatus:   vesselStatus,
		PortLocation:   req.PortLocation,
		AttachmentURL:  req.AttachmentURL,
	}

	if err := s.repo.Create(t); err != nil {
		return nil, err
	}

	// Automatic greeting & acknowledgment reply
	greetingMsg := fmt.Sprintf("Laporan kendala berhasil diregistrasi dengan No. Tiket %s. Notifikasi darurat telah diteruskan ke tim siaga NOC & Command Center. Tim teknisi kami akan segera menindaklanjuti sesuai target SLA (%s). Anda dapat memantau atau menambahkan info perkembangan di percakapan ini.", t.TicketNumber, strings.ToUpper(t.Priority))
	greetingReply := &model.TicketReply{
		TicketID:   t.ID,
		SenderName: "EMERTI Virtual Desk",
		SenderRole: "admin",
		Message:    greetingMsg,
		CreatedAt:  now,
	}
	_ = s.repo.AddReply(greetingReply)

	// Fetch created ticket with associations
	created, _ := s.repo.GetByID(t.ID)
	targetTicket := t
	if created != nil {
		targetTicket = created
	}

	// Trigger Telegram notification
	if s.tg != nil && s.tg.IsEnabled() {
		msgID, tgErr := s.tg.NotifyNewTicket(targetTicket)
		if tgErr == nil && msgID > 0 {
			targetTicket.TelegramMessageID = &msgID
			targetTicket.TelegramNotified = true
			_ = s.repo.Update(targetTicket)
			log.Printf("[TICKET] Telegram notification sent for %s (msg_id: %d)", targetTicket.TicketNumber, msgID)
		} else if tgErr != nil {
			log.Printf("[TICKET] Warning: Telegram notification failed for %s: %v", targetTicket.TicketNumber, tgErr)
		}
	} else {
		log.Printf("[TICKET] Telegram notification skipped: Telegram service is not enabled (tg=%v)", s.tg != nil)
	}

	if created != nil {
		return toResponse(created), nil
	}
	return toResponse(t), nil
}

func (s *service) GetAllTickets(filter TicketFilterQuery) ([]TicketResponse, error) {
	tickets, err := s.repo.GetAll(filter)
	if err != nil {
		return nil, err
	}

	res := make([]TicketResponse, len(tickets))
	for i, t := range tickets {
		res[i] = *toResponse(&t)
	}
	return res, nil
}

func (s *service) GetTicketByID(id uint) (*TicketResponse, error) {
	ticket, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return toResponse(ticket), nil
}

func (s *service) GetTicketByIDModel(id uint) (*model.SupportTicket, error) {
	return s.repo.GetByID(id)
}

func (s *service) GetTicketByTelegramMessageID(msgID int) (*model.SupportTicket, error) {
	return s.repo.GetByTelegramMessageID(msgID)
}

func (s *service) AcknowledgeTicket(id uint, req AcknowledgeTicketRequest) (*TicketResponse, error) {
	ticket, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	ticket.Status = "in_progress"
	ticket.AcknowledgedAt = &now
	if req.AdminName != "" {
		ticket.AcknowledgedBy = req.AdminName
	} else {
		ticket.AcknowledgedBy = "Support Admin"
	}

	if err := s.repo.Update(ticket); err != nil {
		return nil, err
	}

	if req.Note != "" {
		reply := &model.TicketReply{
			TicketID:   ticket.ID,
			SenderName: ticket.AcknowledgedBy,
			SenderRole: "admin",
			Message:    req.Note,
			CreatedAt:  now,
		}
		_ = s.repo.AddReply(reply)
	}

	// Notify Telegram of claim
	if s.tg != nil && s.tg.IsEnabled() && ticket.TelegramMessageID != nil {
		_ = s.tg.NotifyTicketClaimed(ticket, ticket.AcknowledgedBy)
	}

	updated, err := s.repo.GetByID(ticket.ID)
	if err != nil {
		return toResponse(ticket), nil
	}
	return toResponse(updated), nil
}

func (s *service) AcknowledgeTicketFromTelegram(ticketID uint, adminName string) (*model.SupportTicket, error) {
	ticket, err := s.repo.GetByID(ticketID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	ticket.Status = "in_progress"
	ticket.AcknowledgedAt = &now
	ticket.AcknowledgedBy = adminName

	if err := s.repo.Update(ticket); err != nil {
		return nil, err
	}

	reply := &model.TicketReply{
		TicketID:   ticket.ID,
		SenderName: adminName,
		SenderRole: "admin",
		Message:    fmt.Sprintf("Tiket telah dikonfirmasi dan sedang ditangani oleh %s.", adminName),
		CreatedAt:  now,
	}
	_ = s.repo.AddReply(reply)

	return s.repo.GetByID(ticket.ID)
}

func (s *service) ReplyTicket(id uint, req ReplyTicketRequest, userID *uint) (*TicketReplyResponse, error) {
	ticket, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if ticket.Status == "closed" {
		return nil, errors.New("cannot reply to a closed ticket")
	}

	now := time.Now()
	senderName := req.SenderName
	if senderName == "" {
		senderName = "Operator"
	}
	senderRole := req.SenderRole
	if senderRole == "" {
		senderRole = "user"
	}

	reply := &model.TicketReply{
		TicketID:      ticket.ID,
		UserID:        userID,
		SenderName:    senderName,
		SenderRole:    senderRole,
		Message:       req.Message,
		AttachmentURL: req.AttachmentURL,
		CreatedAt:     now,
	}

	if err := s.repo.AddReply(reply); err != nil {
		return nil, err
	}

	// If admin replies and ticket was open, auto-transition to in_progress
	if senderRole == "admin" && ticket.Status == "open" {
		ticket.Status = "in_progress"
		ticket.AcknowledgedAt = &now
		ticket.AcknowledgedBy = senderName
		_ = s.repo.Update(ticket)
	}

	// Forward client reply to Telegram group
	if senderRole == "user" && s.tg != nil && s.tg.IsEnabled() && ticket.TelegramMessageID != nil {
		tgMsgID, tgErr := s.tg.NotifyTicketReply(ticket, reply)
		if tgErr == nil && tgMsgID > 0 {
			reply.TelegramMessageID = &tgMsgID
			_ = s.repo.Update(ticket)
		}
	}

	return &TicketReplyResponse{
		ID:            reply.ID,
		TicketID:      reply.TicketID,
		UserID:        reply.UserID,
		SenderName:    reply.SenderName,
		SenderRole:    reply.SenderRole,
		Message:       reply.Message,
		AttachmentURL: reply.AttachmentURL,
		CreatedAt:     reply.CreatedAt,
	}, nil
}

func (s *service) ReplyTicketFromTelegram(ticketID uint, replyText string, senderName string, tgMsgID int) (*model.TicketReply, error) {
	ticket, err := s.repo.GetByID(ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.Status == "closed" {
		return nil, errors.New("cannot reply to a closed ticket")
	}

	now := time.Now()
	reply := &model.TicketReply{
		TicketID:          ticket.ID,
		SenderName:        senderName,
		SenderRole:        "admin",
		Message:           replyText,
		TelegramMessageID: &tgMsgID,
		CreatedAt:         now,
	}

	if err := s.repo.AddReply(reply); err != nil {
		return nil, err
	}

	if ticket.Status == "open" {
		ticket.Status = "in_progress"
		ticket.AcknowledgedAt = &now
		ticket.AcknowledgedBy = senderName
		_ = s.repo.Update(ticket)
	}

	return reply, nil
}

func (s *service) ResolveTicket(id uint, req ResolveTicketRequest) (*TicketResponse, error) {
	ticket, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	ticket.Status = "resolved"
	ticket.ResolvedAt = &now
	ticket.ResolutionNotes = req.ResolutionNotes

	adminName := req.AdminName
	if adminName == "" {
		adminName = "Support Admin"
	}

	if err := s.repo.Update(ticket); err != nil {
		return nil, err
	}

	// Record resolution as official reply
	replyMsg := "Laporan telah diselesaikan: " + req.ResolutionNotes
	reply := &model.TicketReply{
		TicketID:   ticket.ID,
		SenderName: adminName,
		SenderRole: "admin",
		Message:    replyMsg,
		CreatedAt:  now,
	}
	_ = s.repo.AddReply(reply)

	// Notify Telegram of resolution
	if s.tg != nil && s.tg.IsEnabled() && ticket.TelegramMessageID != nil {
		_ = s.tg.NotifyTicketResolved(ticket, adminName, req.ResolutionNotes)
	}

	updated, err := s.repo.GetByID(ticket.ID)
	if err != nil {
		return toResponse(ticket), nil
	}
	return toResponse(updated), nil
}

func (s *service) ResolveTicketFromTelegram(ticketID uint, adminName string, notes string) (*model.SupportTicket, error) {
	ticket, err := s.repo.GetByID(ticketID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	ticket.Status = "resolved"
	ticket.ResolvedAt = &now
	ticket.ResolutionNotes = notes

	if err := s.repo.Update(ticket); err != nil {
		return nil, err
	}

	reply := &model.TicketReply{
		TicketID:   ticket.ID,
		SenderName: adminName,
		SenderRole: "admin",
		Message:    "Laporan telah diselesaikan via Telegram NOC: " + notes,
		CreatedAt:  now,
	}
	_ = s.repo.AddReply(reply)

	return s.repo.GetByID(ticket.ID)
}

func (s *service) CloseTicket(id uint) (*TicketResponse, error) {
	ticket, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if ticket.Status != "resolved" && ticket.Status != "in_progress" {
		return nil, errors.New("only in-progress or resolved tickets can be closed")
	}

	now := time.Now()
	ticket.Status = "closed"
	ticket.ClosedAt = &now

	if err := s.repo.Update(ticket); err != nil {
		return nil, err
	}

	return toResponse(ticket), nil
}

func (s *service) DeleteTicket(id uint) error {
	ticket, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ticket.ID)
}

func (s *service) BulkDeleteTickets(ids []uint) error {
	return s.repo.BulkDelete(ids)
}

func (s *service) GetUnhandledTickets() ([]model.SupportTicket, error) {
	return s.repo.GetAll(TicketFilterQuery{Status: "open"})
}

func (s *service) GetCriticalTickets() ([]model.SupportTicket, error) {
	tickets, err := s.repo.GetAll(TicketFilterQuery{Priority: "critical"})
	if err != nil {
		return nil, err
	}
	var active []model.SupportTicket
	for _, t := range tickets {
		if t.Status != "closed" && t.Status != "resolved" {
			active = append(active, t)
		}
	}
	return active, nil
}

func (s *service) GetActiveTickets() ([]model.SupportTicket, error) {
	tickets, err := s.repo.GetAll(TicketFilterQuery{})
	if err != nil {
		return nil, err
	}
	var active []model.SupportTicket
	for _, t := range tickets {
		if t.Status == "open" || t.Status == "in_progress" {
			active = append(active, t)
		}
	}
	return active, nil
}

func (s *service) GetNOCMetrics() (*telegram.TicketMetricsSummary, error) {
	m, err := s.repo.GetMetrics(nil, nil)
	if err != nil {
		return nil, err
	}
	return &telegram.TicketMetricsSummary{
		TotalTickets: m.TotalTickets,
		OpenTickets:  m.OpenTickets,
		CriticalP1:   m.CriticalP1,
		InProgress:   m.InProgress,
		Resolved:     m.Resolved,
		SlaBreached:  m.SlaBreached,
	}, nil
}

func (s *service) GetMetrics(companyID *uint, reporterUserID *uint) (*TicketMetricsResponse, error) {
	return s.repo.GetMetrics(companyID, reporterUserID)
}

func toResponse(t *model.SupportTicket) *TicketResponse {
	if t == nil {
		return nil
	}

	vesselName := ""
	if t.Vessel != nil {
		vesselName = t.Vessel.Name
	}

	var compID *uint
	companyName := ""
	companyCode := ""
	if t.Company != nil {
		compID = &t.Company.ID
		companyName = t.Company.Name
		companyCode = t.Company.Code
	} else if t.CompanyID != nil {
		compID = t.CompanyID
	}

	if companyName == "" && t.Vessel != nil && t.Vessel.Company.ID > 0 {
		compID = &t.Vessel.Company.ID
		companyName = t.Vessel.Company.Name
		companyCode = t.Vessel.Company.Code
	}

	repliesRes := make([]TicketReplyResponse, len(t.Replies))
	for i, r := range t.Replies {
		repliesRes[i] = TicketReplyResponse{
			ID:            r.ID,
			TicketID:      r.TicketID,
			UserID:        r.UserID,
			SenderName:    r.SenderName,
			SenderRole:    r.SenderRole,
			Message:       r.Message,
			AttachmentURL: r.AttachmentURL,
			CreatedAt:     r.CreatedAt,
		}
	}

	return &TicketResponse{
		ID:                t.ID,
		TicketNumber:      t.TicketNumber,
		VesselID:          t.VesselID,
		VesselName:        vesselName,
		CompanyID:         compID,
		CompanyName:       companyName,
		CompanyCode:       companyCode,
		ReporterUserID:    t.ReporterUserID,
		ReporterName:      t.ReporterName,
		ReporterPhone:     t.ReporterPhone,
		ReporterEmail:     t.ReporterEmail,
		Category:          t.Category,
		Priority:          t.Priority,
		Subject:           t.Subject,
		Description:       t.Description,
		Status:            t.Status,
		AcknowledgedBy:    t.AcknowledgedBy,
		AcknowledgedAt:    t.AcknowledgedAt,
		SlaTargetAt:       t.SlaTargetAt,
		ResolvedAt:        t.ResolvedAt,
		ClosedAt:          t.ClosedAt,
		ResolutionNotes:   t.ResolutionNotes,
		TelegramNotified:  t.TelegramNotified,
		TelegramMessageID: t.TelegramMessageID,
		VesselStatus:      t.VesselStatus,
		PortLocation:      t.PortLocation,
		AttachmentURL:     t.AttachmentURL,
		RepliesCount:      len(t.Replies),
		Replies:           repliesRes,
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
	}
}
