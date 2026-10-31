package model

import "time"

// Voyage represents a vessel's voyage plan and current operational schedule
type Voyage struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	VoyageNumber    string     `gorm:"type:varchar(100);not null" json:"voyageNumber"` // e.g. "V-2026-0012"
	VesselID        uint       `gorm:"not null;index" json:"vesselId"`
	Vessel          MVessel    `gorm:"foreignKey:VesselID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"vessel,omitempty"`
	Status          string     `gorm:"type:varchar(50);not null;default:'en_route'" json:"status"` // en_route, moored, standby, completed

	// Origin (Port, Offshore, or Custom)
	OriginType      string     `gorm:"type:varchar(50);not null;default:'port'" json:"originType"` // port, offshore, custom
	OriginID        *uint      `json:"originId"`
	OriginName      string     `gorm:"type:varchar(255);not null" json:"originName"`
	OriginCode      string     `gorm:"type:varchar(50)" json:"originCode"`
	OriginLat       *float64   `gorm:"type:decimal(10,6)" json:"originLat"`
	OriginLong      *float64   `gorm:"type:decimal(10,6)" json:"originLong"`
	DepartureTime   *time.Time `gorm:"type:timestamptz" json:"departureTime"`

	// Destination (Port, Offshore, or Custom)
	DestinationType string     `gorm:"type:varchar(50);not null;default:'offshore'" json:"destinationType"` // port, offshore, custom
	DestinationID   *uint      `json:"destinationId"`
	DestinationName string     `gorm:"type:varchar(255);not null" json:"destinationName"`
	DestinationCode string     `gorm:"type:varchar(50)" json:"destinationCode"`
	DestinationLat  *float64   `gorm:"type:decimal(10,6)" json:"destinationLat"`
	DestinationLong *float64   `gorm:"type:decimal(10,6)" json:"destinationLong"`
	ArrivalTime     *time.Time `gorm:"type:timestamptz" json:"arrivalTime"` // ETA

	ActualDeparture *time.Time `gorm:"type:timestamptz" json:"actualDeparture"`
	ActualArrival   *time.Time `gorm:"type:timestamptz" json:"actualArrival"`
	Notes           string     `gorm:"type:text" json:"notes"`

	BaseModel
}

func (Voyage) TableName() string {
	return "t_voyages"
}
