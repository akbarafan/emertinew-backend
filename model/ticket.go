package model

import "time"

// SupportTicket represents a customer support incident or operational issue logged in the service desk
type SupportTicket struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	TicketNumber   string     `gorm:"type:varchar(50);not null;uniqueIndex" json:"ticketNumber"` // e.g. "TKT-202610-001"
	VesselID       *uint      `gorm:"index" json:"vesselId"`
	Vessel         *MVessel   `gorm:"foreignKey:VesselID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"vessel,omitempty"`
	CompanyID      *uint      `gorm:"index" json:"companyId"`
	Company        *MCompany  `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"company,omitempty"`

	// Reporter info
	ReporterUserID *uint      `gorm:"index" json:"reporterUserId"`
	ReporterName   string     `gorm:"type:varchar(100);not null" json:"reporterName"`
	ReporterPhone  string     `gorm:"type:varchar(50)" json:"reporterPhone"`
	ReporterEmail  string     `gorm:"type:varchar(100)" json:"reporterEmail"`

	// Classification & Content
	Category       string     `gorm:"type:varchar(50);not null" json:"category"`                 // sensor, gateway, voyage, anomaly, account, other
	Priority       string     `gorm:"type:varchar(20);not null;default:'medium'" json:"priority"` // critical, high, medium, low
	Subject        string     `gorm:"type:varchar(255);not null" json:"subject"`
	Description    string     `gorm:"type:text;not null" json:"description"`

	// Status & SLA Tracking
	Status         string     `gorm:"type:varchar(30);not null;default:'open'" json:"status"` // open, in_progress, resolved, closed

	// SLA Acknowledgement (Response SLA)
	AcknowledgedBy string     `gorm:"type:varchar(100)" json:"acknowledgedBy"` // admin responder name
	AcknowledgedAt *time.Time `gorm:"type:timestamptz" json:"acknowledgedAt"`

	// SLA Resolution
	SlaTargetAt    time.Time  `gorm:"type:timestamptz;not null" json:"slaTargetAt"` // deadline based on priority
	ResolvedAt     *time.Time `gorm:"type:timestamptz" json:"resolvedAt"`
	ClosedAt       *time.Time `gorm:"type:timestamptz" json:"closedAt"`

	// Resolution details
	ResolutionNotes string    `gorm:"type:text" json:"resolutionNotes"`

	// Telegram integration tracking
	TelegramNotified  bool    `gorm:"default:false" json:"telegramNotified"`
	TelegramMessageID *int    `gorm:"index" json:"telegramMessageId,omitempty"`
	TelegramChatID    *int64  `json:"telegramChatId,omitempty"`

	// Operational context & attachments
	VesselStatus   string     `gorm:"type:varchar(30);default:'unknown'" json:"vesselStatus"` // sailing, docked, anchorage, unknown
	PortLocation   string     `gorm:"type:varchar(100)" json:"portLocation,omitempty"`
	AttachmentURL  string     `gorm:"type:varchar(255)" json:"attachmentUrl,omitempty"`

	// Conversation replies
	Replies        []TicketReply `gorm:"foreignKey:TicketID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"replies,omitempty"`

	BaseModel
}

func (SupportTicket) TableName() string {
	return "t_support_tickets"
}

// TicketReply represents a reply in the ticket conversation
type TicketReply struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	TicketID          uint      `gorm:"not null;index" json:"ticketId"`
	UserID            *uint     `json:"userId"`
	SenderName        string    `gorm:"type:varchar(100);not null" json:"senderName"`
	SenderRole        string    `gorm:"type:varchar(30);not null;default:'user'" json:"senderRole"` // admin, user
	Message           string    `gorm:"type:text;not null" json:"message"`
	AttachmentURL     string    `gorm:"type:varchar(255)" json:"attachmentUrl,omitempty"`
	TelegramMessageID *int      `json:"telegramMessageId,omitempty"`
	CreatedAt         time.Time `gorm:"type:timestamptz;default:now()" json:"createdAt"`
}

func (TicketReply) TableName() string {
	return "t_ticket_replies"
}
