package fuel

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	fuel := r.Group("/fuel")
	{
		fuel.GET("", handler.GetFuelTotalHandler)
		// The vessel route is registered first on purpose. Gin moves the
		// children of a node under a wildcard that is added to it, so a static
		// route registered before the wildcard ends up being served by the
		// wildcard handler.
		fuel.GET("/:vessel_id", handler.GetFuelTotalByVesselHandler)
		fuel.GET("/window", handler.GetFuelTotalByRangeHandler)
	}
}
