package offshore

// CreateOffshoreRequest untuk request payload create offshore
type CreateOffshoreRequest struct {
	CompanyID *uint   `json:"company_id"`
	Name      string  `json:"name" binding:"required" example:"BD North"`
	Code      string  `json:"code" binding:"required" example:"BD-NORTH"`
	Type      string  `json:"type" binding:"required" example:"field"`
	Long      float64 `json:"long" binding:"required" example:"113.276267"`
	Lat       float64 `json:"lat" binding:"required" example:"-7.364350"`
}

// UpdateOffshoreRequest untuk request payload update offshore
type UpdateOffshoreRequest struct {
	CompanyID *uint   `json:"company_id"`
	Name      string  `json:"name" example:"BD North Updated"`
	Code      string  `json:"code" example:"BD-NORTH-UPD"`
	Type      string  `json:"type" example:"field"`
	Long      float64 `json:"long" example:"113.276267"`
	Lat       float64 `json:"lat" example:"-7.364350"`
}

// OffshoreResponse untuk response data offshore
type OffshoreResponse struct {
	ID        uint    `json:"id" example:"1"`
	CompanyID *uint   `json:"company_id,omitempty"`
	Name      string  `json:"name" example:"BD North"`
	Code      string  `json:"code" example:"BD-NORTH"`
	Type      string  `json:"type" example:"field"`
	Long      float64 `json:"long" example:"113.276267"`
	Lat       float64 `json:"lat" example:"-7.364350"`
	CreatedAt string  `json:"created_at" example:"2024-01-01T10:00:00Z"`
	UpdatedAt string  `json:"updated_at" example:"2024-01-01T10:00:00Z"`
}

// OffshoreAuditData untuk audit log offshore
type OffshoreAuditData struct {
	ID   uint    `json:"id"`
	Name string  `json:"name"`
	Code string  `json:"code"`
	Type string  `json:"type"`
	Long float64 `json:"long"`
	Lat  float64 `json:"lat"`
}
