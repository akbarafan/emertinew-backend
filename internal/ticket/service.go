package ticket

import (
	"errors"
	"time"

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
	GetMetrics(companyID *uint, reporterUserID *uint) (*TicketMetricsResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateTicket(req CreateTicketRequest, userID *uint, userCompanyID *uint) (*TicketResponse, error) {
	ticketNumber, err := s.repo.GetNextTicketNumber()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Compute SLA target deadline based on priority (Critical is max 3 hours)
	var slaDuration time.Duration
	switch req.Priority {
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
	}

	if err := s.repo.Create(t); err != nil {
		return nil, err
	}

	// Fetch created ticket with associations
	created, err := s.repo.GetByID(t.ID)
	if err != nil {
		return toResponse(t), nil
	}

	return toResponse(created), nil
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

	updated, err := s.repo.GetByID(ticket.ID)
	if err != nil {
		return toResponse(ticket), nil
	}
	return toResponse(updated), nil
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
		TicketID:   ticket.ID,
		UserID:     userID,
		SenderName: senderName,
		SenderRole: senderRole,
		Message:    req.Message,
		CreatedAt:  now,
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

	return &TicketReplyResponse{
		ID:         reply.ID,
		TicketID:   reply.TicketID,
		UserID:     reply.UserID,
		SenderName: reply.SenderName,
		SenderRole: reply.SenderRole,
		Message:    reply.Message,
		CreatedAt:  reply.CreatedAt,
	}, nil
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

	updated, err := s.repo.GetByID(ticket.ID)
	if err != nil {
		return toResponse(ticket), nil
	}
	return toResponse(updated), nil
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
			ID:         r.ID,
			TicketID:   r.TicketID,
			UserID:     r.UserID,
			SenderName: r.SenderName,
			SenderRole: r.SenderRole,
			Message:    r.Message,
			CreatedAt:  r.CreatedAt,
		}
	}

	return &TicketResponse{
		ID:               t.ID,
		TicketNumber:     t.TicketNumber,
		VesselID:         t.VesselID,
		VesselName:       vesselName,
		CompanyID:        compID,
		CompanyName:      companyName,
		CompanyCode:      companyCode,
		ReporterUserID:   t.ReporterUserID,
		ReporterName:     t.ReporterName,
		ReporterPhone:    t.ReporterPhone,
		ReporterEmail:    t.ReporterEmail,
		Category:         t.Category,
		Priority:         t.Priority,
		Subject:          t.Subject,
		Description:      t.Description,
		Status:           t.Status,
		AcknowledgedBy:   t.AcknowledgedBy,
		AcknowledgedAt:   t.AcknowledgedAt,
		SlaTargetAt:      t.SlaTargetAt,
		ResolvedAt:       t.ResolvedAt,
		ClosedAt:         t.ClosedAt,
		ResolutionNotes:  t.ResolutionNotes,
		TelegramNotified: t.TelegramNotified,
		RepliesCount:     len(t.Replies),
		Replies:          repliesRes,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
	}
}
