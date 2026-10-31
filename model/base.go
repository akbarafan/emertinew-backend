package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *uint          `gorm:"index"`
	UpdatedBy *uint          `gorm:"index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type BaseModelHardDel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Token struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"uniqueIndex;type:text" json:"token"`
	UserID    string    `gorm:"index" json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`

	BaseModelHardDel
}
