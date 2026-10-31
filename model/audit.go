package model

import "time"

// AuditLog tracks all create, update, and delete operations
type AuditLog struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     *uint     `gorm:"index" json:"user_id"`
	User       User      `gorm:"foreignKey:UserID" json:"-"`
	AuditTable string    `gorm:"column:table_name;type:varchar(100);index;not null" json:"table_name"`
	RecordID   uint      `gorm:"index;not null" json:"record_id"`
	Action     string    `gorm:"type:varchar(20);not null" json:"action"` // CREATE, UPDATE, DELETE, LOGIN, LOGIN_FAILED, LOGOUT
	OldValue   *string   `gorm:"type:jsonb" json:"old_value,omitempty"`
	NewValue   *string   `gorm:"type:jsonb" json:"new_value,omitempty"`
	IPAddress  string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent  string    `gorm:"type:varchar(500)" json:"user_agent"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
