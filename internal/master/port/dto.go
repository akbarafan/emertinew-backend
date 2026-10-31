package port

// CreatePortRequest untuk request payload create port
type CreatePortRequest struct {
	Name string  `json:"name" binding:"required" example:"KALIANGET"`
	Code string  `json:"code" binding:"required" example:"KALIANGET"`
	Long float64 `json:"long" binding:"required" example:"113.945389"`
	Lat  float64 `json:"lat" binding:"required" example:"-7.054889"`
}

// UpdatePortRequest untuk request payload update port
type UpdatePortRequest struct {
	Name string  `json:"name" example:"KALIANGET Updated"`
	Code string  `json:"code" example:"KALIANGET-UPD"`
	Long float64 `json:"long" example:"113.945389"`
	Lat  float64 `json:"lat" example:"-7.054889"`
}

// PortResponse untuk response data port
type PortResponse struct {
	ID        uint    `json:"id" example:"1"`
	Name      string  `json:"name" example:"KALIANGET"`
	Code      string  `json:"code" example:"KALIANGET"`
	Long      float64 `json:"long" example:"113.945389"`
	Lat       float64 `json:"lat" example:"-7.054889"`
	CreatedAt string  `json:"created_at" example:"2024-01-01T10:00:00Z"`
	UpdatedAt string  `json:"updated_at" example:"2024-01-01T10:00:00Z"`
}

// PortAuditData untuk audit log port
type PortAuditData struct {
	ID   uint    `json:"id"`
	Name string  `json:"name"`
	Code string  `json:"code"`
	Long float64 `json:"long"`
	Lat  float64 `json:"lat"`
}
