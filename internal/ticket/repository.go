package ticket

import (
	"fmt"
	"strings"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ticket *model.SupportTicket) error
	GetByID(id uint) (*model.SupportTicket, error)
	GetAll(filter TicketFilterQuery) ([]model.SupportTicket, error)
	Update(ticket *model.SupportTicket) error
	AddReply(reply *model.TicketReply) error
	GetReplies(ticketID uint) ([]model.TicketReply, error)
	GetMetrics(companyID *uint, reporterUserID *uint) (*TicketMetricsResponse, error)
	GetNextTicketNumber() (string, error)
	GetVesselByID(id uint) (*model.MVessel, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ticket *model.SupportTicket) error {
	return r.db.Create(ticket).Error
}

func (r *repository) GetByID(id uint) (*model.SupportTicket, error) {
	var ticket model.SupportTicket
	err := r.db.Preload("Company").Preload("Vessel").Preload("Vessel.Company").Preload("Replies", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).First(&ticket, id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *repository) GetAll(filter TicketFilterQuery) ([]model.SupportTicket, error) {
	var tickets []model.SupportTicket
	query := r.db.Model(&model.SupportTicket{}).Preload("Company").Preload("Vessel").Preload("Vessel.Company").Preload("Replies")

	if filter.Status != "" && filter.Status != "all" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Priority != "" && filter.Priority != "all" {
		query = query.Where("priority = ?", filter.Priority)
	}

	if filter.VesselID != nil && *filter.VesselID > 0 {
		query = query.Where("vessel_id = ?", *filter.VesselID)
	}

	if filter.CompanyID != nil && *filter.CompanyID > 0 {
		query = query.Where("company_id = ?", *filter.CompanyID)
	}

	if filter.ReporterUserID != nil && *filter.ReporterUserID > 0 {
		query = query.Where("reporter_user_id = ?", *filter.ReporterUserID)
	}

	if filter.Search != "" {
		s := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(ticket_number) LIKE ? OR LOWER(subject) LIKE ? OR LOWER(reporter_name) LIKE ? OR LOWER(description) LIKE ?", s, s, s, s)
	}

	// Order by urgency: Critical first, then by earliest SLA deadline
	err := query.Order("CASE WHEN priority = 'critical' AND status != 'closed' THEN 1 WHEN priority = 'high' AND status != 'closed' THEN 2 ELSE 3 END ASC, sla_target_at ASC, id DESC").Find(&tickets).Error
	return tickets, err
}

func (r *repository) Update(ticket *model.SupportTicket) error {
	return r.db.Save(ticket).Error
}

func (r *repository) AddReply(reply *model.TicketReply) error {
	return r.db.Create(reply).Error
}

func (r *repository) GetReplies(ticketID uint) ([]model.TicketReply, error) {
	var replies []model.TicketReply
	err := r.db.Where("ticket_id = ?", ticketID).Order("created_at ASC").Find(&replies).Error
	return replies, err
}

func (r *repository) GetMetrics(companyID *uint, reporterUserID *uint) (*TicketMetricsResponse, error) {
	now := time.Now()
	var metrics TicketMetricsResponse

	baseQuery := r.db.Model(&model.SupportTicket{})
	if companyID != nil && *companyID > 0 {
		baseQuery = baseQuery.Where("company_id = ?", *companyID)
	}
	if reporterUserID != nil && *reporterUserID > 0 {
		baseQuery = baseQuery.Where("reporter_user_id = ?", *reporterUserID)
	}

	// Total tickets
	baseQuery.Session(&gorm.Session{}).Count(&metrics.TotalTickets)

	// Open / Unhandled tickets
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "open").Count(&metrics.OpenTickets)

	// Critical P1 active tickets
	baseQuery.Session(&gorm.Session{}).Where("priority = ? AND status IN ('open', 'in_progress')", "critical").Count(&metrics.CriticalP1)

	// In Progress
	baseQuery.Session(&gorm.Session{}).Where("status = ?", "in_progress").Count(&metrics.InProgress)

	// Resolved or Closed
	baseQuery.Session(&gorm.Session{}).Where("status IN ('resolved', 'closed')").Count(&metrics.Resolved)

	// Breached SLA (Open or In Progress past SLA target)
	baseQuery.Session(&gorm.Session{}).Where("status IN ('open', 'in_progress') AND sla_target_at < ?", now).Count(&metrics.SlaBreached)

	return &metrics, nil
}

func (r *repository) GetNextTicketNumber() (string, error) {
	yearMonth := time.Now().Format("200601")
	var count int64
	prefix := "TKT-" + yearMonth + "-"
	r.db.Model(&model.SupportTicket{}).Where("ticket_number LIKE ?", prefix+"%").Count(&count)

	// Formatted like TKT-202610-001
	return fmt.Sprintf("%s%03d", prefix, count+1), nil
}

func (r *repository) GetVesselByID(id uint) (*model.MVessel, error) {
	var vessel model.MVessel
	err := r.db.Preload("Company").First(&vessel, id).Error
	if err != nil {
		return nil, err
	}
	return &vessel, nil
}
