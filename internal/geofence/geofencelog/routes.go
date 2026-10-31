package geofencelog

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, geofenceRepo GeofenceRepository) {
	// Routes are registered in parent geofence/routes.go to match old structure:
	// POST /geofences/check -> CheckGeofence
	// POST /geofence-logs -> CreateGeofenceLog
	// GET /geofences/:id/logs -> GetLogsByGeofenceID
	// GET /vessels/:id/geofence-logs -> GetLogsByVesselID

	// This function is kept for backward compatibility
}
