package company

import (
	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/middleware"
	"github.com/Alugara-Inovasi-Utama/emerti/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	// Initialize audit service
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)

	// Initialize company module with audit
	repo := NewRepository(db)
	service := NewService(repo, auditService)
	handler := NewHandler(service)

	companies := r.Group("/company")
	{
		// Only superadmin can create or delete companies
		companies.POST("", middleware.RequireSuperadmin(), handler.CreateCompany)
		companies.DELETE("/:id", middleware.RequireSuperadmin(), handler.DeleteCompany)

		// Authenticated users can read companies
		companies.GET("", middleware.RequireGuest(), handler.GetAllCompanies)
		companies.GET("/:id", middleware.RequireGuest(), handler.GetCompanyByID)

		// Superadmin and company_admin can update company profile
		companies.PATCH("/:id", middleware.RequireRole(model.RoleSuperadmin, model.RoleCompanyAdmin), handler.UpdateCompany)
	}
}
