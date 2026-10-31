package middleware

import (
	"net/http"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/base"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/gin-gonic/gin"
)

// RequireRole middleware untuk cek role user
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			base.ErrorResponse(c, http.StatusForbidden, "No role information found", nil)
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		base.ErrorResponse(c, http.StatusForbidden, "You don't have permission to access this resource", nil)
		c.Abort()
	}
}

// RequireSuperadmin middleware untuk require superadmin role
func RequireSuperadmin() gin.HandlerFunc {
	return RequireRole(model.RoleSuperadmin)
}

// RequireAdmin middleware untuk require admin role (atau lebih tinggi)
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(model.RoleAdmin, model.RoleSuperadmin)
}

// RequireGuest middleware untuk require user role (atau lebih tinggi)
func RequireGuest() gin.HandlerFunc {
	return RequireRole(
		model.RoleSuperadmin,
		model.RoleTechAdmin,
		model.RoleAdmin,
		model.RoleCompanyAdmin,
		model.RoleOperator,
		model.RoleViewer,
		model.RoleGuest,
	)
}

// RequirePlatformStaff: superadmin, tech_admin, admin (cross-company platform access)
func RequirePlatformStaff() gin.HandlerFunc {
	return RequireRole(model.RoleSuperadmin, model.RoleTechAdmin, model.RoleAdmin)
}

// RequireHardwareAdmin: superadmin, tech_admin (MCU, sensors, hardware calibration)
func RequireHardwareAdmin() gin.HandlerFunc {
	return RequireRole(model.RoleSuperadmin, model.RoleTechAdmin)
}

// RequireUserAdmin: superadmin, admin, company_admin (creating and managing users)
func RequireUserAdmin() gin.HandlerFunc {
	return RequireRole(model.RoleSuperadmin, model.RoleAdmin, model.RoleCompanyAdmin)
}

// RequireVoyageManager: superadmin, admin, company_admin, operator (scheduling, updating, finishing voyages)
func RequireVoyageManager() gin.HandlerFunc {
	return RequireRole(model.RoleSuperadmin, model.RoleAdmin, model.RoleCompanyAdmin, model.RoleOperator)
}

// RequireCompanyManager: superadmin, admin, company_admin (managing fleet, squad, offshore, vessels)
func RequireCompanyManager() gin.HandlerFunc {
	return RequireRole(model.RoleSuperadmin, model.RoleAdmin, model.RoleCompanyAdmin)
}

// RequireGeofenceManager: superadmin, admin, company_admin, operator (managing geofence routes)
func RequireGeofenceManager() gin.HandlerFunc {
	return RequireRole(model.RoleSuperadmin, model.RoleAdmin, model.RoleCompanyAdmin, model.RoleOperator)
}
