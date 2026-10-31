package vessel

import (
	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	vesselcapacity "github.com/Alugara-Inovasi-Utama/emerti/internal/master/vessel/capacity"
	vesseldimensions "github.com/Alugara-Inovasi-Utama/emerti/internal/master/vessel/dimensions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Initialize audit service
	auditRepo := audit.NewRepository(db)
	auditService := audit.NewService(auditRepo)

	// Initialize vessel module with audit
	repo := NewRepository(db)
	service := NewService(repo, auditService)
	handler := NewHandler(service)

	router.POST("/vessel", handler.CreateVessel)
	router.GET("/vessel", handler.GetAllVessels)
	router.GET("/vessel/deleted", handler.GetDeletedVessels)
	router.GET("/vessel/:id", handler.GetVessel)
	router.PATCH("/vessel/:id", handler.UpdateVessel)
	router.DELETE("/vessel/:id", handler.DeleteVessel)
	router.POST("/vessel/:id/restore", handler.RestoreVessel)

	// Register vessel dimensions routes
	vesseldimensions.RegisterRoutes(router.Group("/vessel"), db)

	// Register vessel capacity routes
	vesselcapacity.RegisterRoutes(router.Group("/vessel"), db)
}
