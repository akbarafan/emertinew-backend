package model

// User represents a user in the system
type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"uniqueIndex;type:varchar(100);not null" json:"username"`
	Password string `gorm:"type:varchar(255);not null" json:"-"` // '-' agar tidak muncul di JSON
	Role      string    `gorm:"type:varchar(20);default:'user'" json:"role"`
	CompanyID *uint     `gorm:"index" json:"company_id,omitempty"`
	Company   *MCompany `gorm:"foreignKey:CompanyID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"company,omitempty"`

	BaseModel
}
