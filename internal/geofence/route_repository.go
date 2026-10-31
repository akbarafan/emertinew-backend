package geofence

import (
	"errors"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type RouteRepository interface {
	Create(route *model.MGeofenceRoute) error
	FindByID(id uint) (*model.MGeofenceRoute, error)
	FindAll(page, limit int, companyID *uint, voyageID *uint, vesselID *uint) ([]model.MGeofenceRoute, int64, error)
	Update(route *model.MGeofenceRoute, waypoints []model.MGeofenceRouteWaypoint) error
	Delete(id uint) error
	GetVoyageByID(id uint) (*model.Voyage, error)
}

type routeRepository struct {
	db *gorm.DB
}

func NewRouteRepository(db *gorm.DB) RouteRepository {
	return &routeRepository{db: db}
}

func (r *routeRepository) Create(route *model.MGeofenceRoute) error {
	return r.db.Create(route).Error
}

func (r *routeRepository) FindByID(id uint) (*model.MGeofenceRoute, error) {
	var route model.MGeofenceRoute
	err := r.db.Preload("Company").
		Preload("Voyage").
		Preload("Vessel").
		Preload("Waypoints", func(db *gorm.DB) *gorm.DB {
			return db.Order("m_geofence_route_waypoints.sequence ASC")
		}).
		First(&route, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("rute geofence tidak ditemukan")
		}
		return nil, err
	}
	return &route, nil
}

func (r *routeRepository) FindAll(page, limit int, companyID *uint, voyageID *uint, vesselID *uint) ([]model.MGeofenceRoute, int64, error) {
	var routes []model.MGeofenceRoute
	var total int64

	query := r.db.Model(&model.MGeofenceRoute{}).
		Preload("Company").
		Preload("Voyage").
		Preload("Vessel").
		Preload("Waypoints", func(db *gorm.DB) *gorm.DB {
			return db.Order("m_geofence_route_waypoints.sequence ASC")
		})

	if companyID != nil {
		query = query.Where("company_id = ? OR company_id IS NULL", *companyID)
	}
	if voyageID != nil {
		query = query.Where("voyage_id = ?", *voyageID)
	}
	if vesselID != nil {
		query = query.Where("vessel_id = ?", *vesselID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("created_at DESC")

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	if err := query.Find(&routes).Error; err != nil {
		return nil, 0, err
	}

	return routes, total, nil
}

func (r *routeRepository) Update(route *model.MGeofenceRoute, waypoints []model.MGeofenceRouteWaypoint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update parent route
		if err := tx.Save(route).Error; err != nil {
			return err
		}

		// If waypoints are provided, replace them
		if waypoints != nil {
			if err := tx.Where("route_id = ?", route.ID).Delete(&model.MGeofenceRouteWaypoint{}).Error; err != nil {
				return err
			}
			for i := range waypoints {
				waypoints[i].RouteID = route.ID
				if err := tx.Create(&waypoints[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *routeRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("route_id = ?", id).Delete(&model.MGeofenceRouteWaypoint{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.MGeofenceRoute{}, id).Error
	})
}

func (r *routeRepository) GetVoyageByID(id uint) (*model.Voyage, error) {
	var v model.Voyage
	if err := r.db.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}
