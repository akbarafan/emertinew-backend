package user

// CreateUserRequest untuk create user (superadmin only)
type CreateUserRequest struct {
	Username  string `json:"username" binding:"required" example:"newuser"`
	Password  string `json:"password" binding:"required,min=6" example:"password123"`
	Role      string `json:"role" binding:"required,oneof=superadmin tech_admin admin company_admin operator viewer guest" example:"admin"`
	CompanyID *uint  `json:"company_id,omitempty" example:"1"`
}

// UpdateUserRequest untuk update user (superadmin only)
type UpdateUserRequest struct {
	Username  string `json:"username" example:"updateduser"`
	Password  string `json:"password,omitempty" binding:"omitempty,min=6" example:"newpassword123"`
	Role      string `json:"role" binding:"omitempty,oneof=superadmin tech_admin admin company_admin operator viewer guest" example:"guest"`
	CompanyID *uint  `json:"company_id,omitempty" example:"1"`
}

// UserResponse untuk response data user
type UserResponse struct {
	ID          uint   `json:"id" example:"1"`
	Username    string `json:"username" example:"admin"`
	Role        string `json:"role" example:"superadmin"`
	CompanyID   *uint  `json:"company_id,omitempty" example:"1"`
	CompanyName string `json:"company_name,omitempty" example:"PT Pelayaran Bahtera"`
	CompanyCode string `json:"company_code,omitempty" example:"hcml"`
}
