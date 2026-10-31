package model

// Role constants
const (
	RoleSuperadmin   = "superadmin"
	RoleTechAdmin    = "tech_admin"
	RoleAdmin        = "admin"
	RoleCompanyAdmin = "company_admin"
	RoleOperator     = "operator"
	RoleViewer       = "viewer"

	// Legacy compatibility
	RoleGuest = "guest"
)

// ValidRoles adalah list semua role yang valid
var ValidRoles = []string{
	RoleSuperadmin,
	RoleTechAdmin,
	RoleAdmin,
	RoleCompanyAdmin,
	RoleOperator,
	RoleViewer,
	RoleGuest,
}

// IsValidRole mengecek apakah role valid
func IsValidRole(role string) bool {
	for _, validRole := range ValidRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// IsSuperadmin mengecek apakah role adalah superadmin
func IsSuperadmin(role string) bool {
	return role == RoleSuperadmin
}

// IsTechAdmin mengecek apakah role adalah tech admin / IoT engineer
func IsTechAdmin(role string) bool {
	return role == RoleTechAdmin
}

// IsAdmin mengecek apakah role adalah admin operasional platform
func IsAdmin(role string) bool {
	return role == RoleAdmin
}

// IsCompanyAdmin mengecek apakah role adalah company admin
func IsCompanyAdmin(role string) bool {
	return role == RoleCompanyAdmin
}

// IsOperator mengecek apakah role adalah operator kapal
func IsOperator(role string) bool {
	return role == RoleOperator
}

// IsViewer mengecek apakah role adalah viewer
func IsViewer(role string) bool {
	return role == RoleViewer || role == RoleGuest
}

// IsPlatformStaff mengecek apakah role berada di level platform Alugara/EMERTI (cross-company)
func IsPlatformStaff(role string) bool {
	return role == RoleSuperadmin || role == RoleTechAdmin || role == RoleAdmin
}

// IsCompanyScoped mengecek apakah role terikat strictly pada satu company_id
func IsCompanyScoped(role string) bool {
	return role == RoleCompanyAdmin || role == RoleOperator || role == RoleViewer || role == RoleGuest
}

// CanManageHardware mengecek apakah role berhak mengelola MCU & kalibrasi sensor
func CanManageHardware(role string) bool {
	return role == RoleSuperadmin || role == RoleTechAdmin
}

// CanManageUsers mengecek apakah role berhak membuat / mengedit user
func CanManageUsers(role string) bool {
	return role == RoleSuperadmin || role == RoleAdmin || role == RoleCompanyAdmin
}

// CanMutateVoyage mengecek apakah role berhak membuat / mengedit / menyelesaikan jadwal pelayaran
func CanMutateVoyage(role string) bool {
	return role == RoleSuperadmin || role == RoleAdmin || role == RoleCompanyAdmin || role == RoleOperator
}

// IsReadOnly mengecek apakah role hanya memiliki hak baca (view-only)
func IsReadOnly(role string) bool {
	return role == RoleViewer || role == RoleGuest
}
