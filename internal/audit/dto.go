package audit

import (
	"encoding/json"
	"time"
)

// AuditLogResponse represents the response for audit log
type AuditLogResponse struct {
	ID        uint            `json:"id"`
	UserID    uint            `json:"user_id"`
	UserName  string          `json:"user_name,omitempty"`
	TableName string          `json:"table_name"`
	RecordID  uint            `json:"record_id"`
	Action    string          `json:"action"`
	OldValue  json.RawMessage `json:"old_value,omitempty"`
	NewValue  json.RawMessage `json:"new_value,omitempty"`
	IPAddress string          `json:"ip_address,omitempty"`
	UserAgent string          `json:"user_agent,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// AuditHistoryRequest represents the request for getting audit history
type AuditHistoryRequest struct {
	TableName string `form:"table_name"`
	RecordID  *uint  `form:"record_id"`
	Action    string `form:"action"`
	Page      int    `form:"page"`
	Limit     int    `form:"limit"`
}

// UserActivityRequest represents the request for getting user activity
type UserActivityRequest struct {
	UserID uint `form:"user_id"`
	Limit  int  `form:"limit"`
}
