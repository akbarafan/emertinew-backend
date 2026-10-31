package ticket

import "time"

type CreateTicketRequest struct {
	VesselID      *uint  `json:"vesselId"`
	CompanyID     *uint  `json:"companyId"`
	ReporterName  string `json:"reporterName" binding:"required"`
	ReporterPhone string `json:"reporterPhone"`
	ReporterEmail string `json:"reporterEmail"`
	Category      string `json:"category" binding:"required"` // sensor, gateway, voyage, anomaly, account, other
	Priority      string `json:"priority" binding:"required"` // critical, high, medium, low
	Subject       string `json:"subject" binding:"required"`
	Description   string `json:"description" binding:"required"`
}

type AcknowledgeTicketRequest struct {
	AdminName string `json:"adminName"`
	Note      string `json:"note"`
}

type ReplyTicketRequest struct {
	Message    string `json:"message" binding:"required"`
	SenderName string `json:"senderName"`
	SenderRole string `json:"senderRole"` // admin or user
}

type ResolveTicketRequest struct {
	ResolutionNotes string `json:"resolutionNotes" binding:"required"`
	AdminName       string `json:"adminName"`
}

type TicketFilterQuery struct {
	Status         string `form:"status"`
	Priority       string `form:"priority"`
	VesselID       *uint  `form:"vesselId"`
	CompanyID      *uint  `form:"companyId"`
	ReporterUserID *uint  `form:"reporterUserId"`
	Search         string `form:"search"`
}

type TicketReplyResponse struct {
	ID         uint      `json:"id"`
	TicketID   uint      `json:"ticketId"`
	UserID     *uint     `json:"userId,omitempty"`
	SenderName string    `json:"senderName"`
	SenderRole string    `json:"senderRole"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"createdAt"`
}

type TicketResponse struct {
	ID               uint                  `json:"id"`
	TicketNumber     string                `json:"ticketNumber"`
	VesselID         *uint                 `json:"vesselId,omitempty"`
	VesselName       string                `json:"vesselName,omitempty"`
	CompanyID        *uint                 `json:"companyId,omitempty"`
	CompanyName      string                `json:"companyName,omitempty"`
	CompanyCode      string                `json:"companyCode,omitempty"`
	ReporterUserID   *uint                 `json:"reporterUserId,omitempty"`
	ReporterName     string                `json:"reporterName"`
	ReporterPhone    string                `json:"reporterPhone,omitempty"`
	ReporterEmail    string                `json:"reporterEmail,omitempty"`
	Category         string                `json:"category"`
	Priority         string                `json:"priority"`
	Subject          string                `json:"subject"`
	Description      string                `json:"description"`
	Status           string                `json:"status"`
	AcknowledgedBy   string                `json:"acknowledgedBy,omitempty"`
	AcknowledgedAt   *time.Time            `json:"acknowledgedAt,omitempty"`
	SlaTargetAt      time.Time             `json:"slaTargetAt"`
	ResolvedAt       *time.Time            `json:"resolvedAt,omitempty"`
	ClosedAt         *time.Time            `json:"closedAt,omitempty"`
	ResolutionNotes  string                `json:"resolutionNotes,omitempty"`
	TelegramNotified bool                  `json:"telegramNotified"`
	RepliesCount     int                   `json:"repliesCount"`
	Replies          []TicketReplyResponse `json:"replies,omitempty"`
	CreatedAt        time.Time             `json:"createdAt"`
	UpdatedAt        time.Time             `json:"updatedAt"`
}

type TicketMetricsResponse struct {
	TotalTickets int64 `json:"totalTickets"`
	OpenTickets  int64 `json:"openTickets"`
	CriticalP1   int64 `json:"criticalP1"`
	InProgress   int64 `json:"inProgress"`
	Resolved     int64 `json:"resolved"`
	SlaBreached  int64 `json:"slaBreached"`
}
