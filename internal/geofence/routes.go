package geofence

import (
	"github.com/Alugara-Inovasi-Utama/emerti/internal/geofence/geofencelog"
	"github.com/Alugara-Inovasi-Utama/emerti/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Initialize repository, service, handler for geofence CRUD
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Initialize geofence log handler
	logRepo := geofencelog.NewRepository(db)
	logService := geofencelog.NewService(logRepo, repo)
	logHandler := geofencelog.NewHandler(logService)

	// Initialize geofence route handler
	routeRepo := NewRouteRepository(db)
	routeService := NewRouteService(routeRepo)
	routeHandler := NewRouteHandler(routeService)

	geofenceRoutes := router.Group("/geofence-routes")
	{
		geofenceRoutes.POST("/generate-from-voyage", middleware.RequireGeofenceManager(), routeHandler.GenerateFromVoyage)
		geofenceRoutes.POST("", middleware.RequireGeofenceManager(), routeHandler.CreateRoute)
		geofenceRoutes.GET("", middleware.JWTMiddleware(), routeHandler.GetAllRoutes)
		geofenceRoutes.GET("/:id", middleware.JWTMiddleware(), routeHandler.GetRoute)
		geofenceRoutes.PUT("/:id", middleware.RequireGeofenceManager(), routeHandler.UpdateRoute)
		geofenceRoutes.PATCH("/:id", middleware.RequireGeofenceManager(), routeHandler.UpdateRoute)
		geofenceRoutes.DELETE("/:id", middleware.RequireGeofenceManager(), routeHandler.DeleteRoute)
	}

	geofences := router.Group("/geofences")
	{
		// Static routes first to avoid conflicts
		geofences.POST("/check", middleware.JWTMiddleware(), logHandler.CheckGeofence)

		// CRUD routes
		geofences.POST("", middleware.RequireGeofenceManager(), handler.CreateGeofence)
		geofences.GET("", middleware.JWTMiddleware(), handler.GetAllGeofences)

		// Routes with :id parameter
		geofences.GET("/:id", middleware.JWTMiddleware(), handler.GetGeofence)
		geofences.PATCH("/:id", middleware.RequireGeofenceManager(), handler.UpdateGeofence)
		geofences.DELETE("/:id", middleware.RequireGeofenceManager(), handler.DeleteGeofence)

		// Logs route - :id is geofence_id here
		geofences.GET("/:id/logs", middleware.JWTMiddleware(), logHandler.GetLogsByGeofenceID)
	}

	// Geofence log routes
	router.POST("/geofence-logs", middleware.JWTMiddleware(), logHandler.CreateGeofenceLog)
	router.GET("/geofence-logs", middleware.JWTMiddleware(), logHandler.ListAllLogs)

	// Vessel logs route
	vessels := router.Group("/vessels")
	{
		vessels.GET("/:id/geofence-logs", middleware.JWTMiddleware(), logHandler.GetLogsByVesselID)
	}
}
