package voyage

import "time"

type CreateVoyageRequest struct {
	VesselID        uint       `json:"vesselId" binding:"required"`
	VoyageNumber    string     `json:"voyageNumber"`
	Status          string     `json:"status"` // en_route, moored, standby
	OriginType      string     `json:"originType"`
	OriginID        *uint      `json:"originId"`
	OriginName      string     `json:"originName" binding:"required"`
	OriginCode      string     `json:"originCode"`
	OriginLat       *float64   `json:"originLat"`
	OriginLong      *float64   `json:"originLong"`
	DepartureTime   *time.Time `json:"departureTime"`
	DestinationType string     `json:"destinationType"`
	DestinationID   *uint      `json:"destinationId"`
	DestinationName string     `json:"destinationName" binding:"required"`
	DestinationCode string     `json:"destinationCode"`
	DestinationLat  *float64   `json:"destinationLat"`
	DestinationLong *float64   `json:"destinationLong"`
	ArrivalTime     *time.Time `json:"arrivalTime"`
	Notes           string     `json:"notes"`
}

type UpdateVoyageRequest struct {
	Status          *string    `json:"status"`
	OriginType      *string    `json:"originType"`
	OriginID        *uint      `json:"originId"`
	OriginName      *string    `json:"originName"`
	OriginCode      *string    `json:"originCode"`
	OriginLat       *float64   `json:"originLat"`
	OriginLong      *float64   `json:"originLong"`
	DepartureTime   *time.Time `json:"departureTime"`
	DestinationType *string    `json:"destinationType"`
	DestinationID   *uint      `json:"destinationId"`
	DestinationName *string    `json:"destinationName"`
	DestinationCode *string    `json:"destinationCode"`
	DestinationLat  *float64   `json:"destinationLat"`
	DestinationLong *float64   `json:"destinationLong"`
	ArrivalTime     *time.Time `json:"arrivalTime"`
	ActualDeparture *time.Time `json:"actualDeparture"`
	ActualArrival   *time.Time `json:"actualArrival"`
	Notes           *string    `json:"notes"`
}

type VoyageResponse struct {
	ID              uint       `json:"id"`
	VoyageNumber    string     `json:"voyageNumber"`
	VesselID        uint       `json:"vesselId"`
	VesselName      string     `json:"vesselName,omitempty"`
	Status          string     `json:"status"`
	OriginType      string     `json:"originType"`
	OriginID        *uint      `json:"originId,omitempty"`
	OriginName      string     `json:"originName"`
	OriginCode      string     `json:"originCode"`
	OriginLat       *float64   `json:"originLat,omitempty"`
	OriginLong      *float64   `json:"originLong,omitempty"`
	DepartureTime   *time.Time `json:"departureTime,omitempty"`
	DestinationType string     `json:"destinationType"`
	DestinationID   *uint      `json:"destinationId,omitempty"`
	DestinationName string     `json:"destinationName"`
	DestinationCode string     `json:"destinationCode"`
	DestinationLat  *float64   `json:"destinationLat,omitempty"`
	DestinationLong *float64   `json:"destinationLong,omitempty"`
	ArrivalTime     *time.Time `json:"arrivalTime,omitempty"`
	ActualDeparture *time.Time `json:"actualDeparture,omitempty"`
	ActualArrival   *time.Time `json:"actualArrival,omitempty"`
	Notes           string     `json:"notes,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}
