package auth

// LoginRequest untuk request payload login
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse untuk response setelah login berhasil
type LoginResponse struct {
	Token       string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	UserID      uint   `json:"user_id" example:"1"`
	Username    string `json:"username" example:"admin"`
	Role        string `json:"role" example:"superadmin"`
	CompanyID   *uint  `json:"company_id,omitempty" example:"1"`
	CompanyName string `json:"company_name,omitempty" example:"PT Pelayaran Bahtera"`
	CompanyCode string `json:"company_code,omitempty" example:"hcml"`
}

// AuthAuditData untuk data audit event autentikasi
type AuthAuditData struct {
	Username string `json:"username"`
	Role     string `json:"role,omitempty"`
	Reason   string `json:"reason,omitempty"`
}
