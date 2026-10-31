package model

import "time"

type Manifest struct {
	ID            uint       `gorm:"primaryKey"`
	Name          string     `gorm:"type:varchar(255);not null"` // Auto-generated
	VesselID      uint       `gorm:"not null;index"`
	Vessel        MVessel    `gorm:"foreignKey:VesselID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CaptainID     *uint      `gorm:"index"`
	Captain       MCrew      `gorm:"foreignKey:CaptainID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	GrossTonnage  *float64   `gorm:"type:decimal(10,2)"`
	Date          *time.Time `gorm:"type:date"`
	ArrivalTime   *time.Time `gorm:"type:timestamptz"`
	DepartureTime *time.Time `gorm:"type:timestamptz"`
	Status        string     `gorm:"type:varchar(50);not null;default:'draft'"` // draft, review, approved

	BaseModel
}

func (Manifest) TableName() string {
	return "r_manifest"
}
