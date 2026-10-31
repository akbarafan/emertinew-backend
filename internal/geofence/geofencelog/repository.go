package geofencelog

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(log *model.GeofenceLog) error
	GetByID(id uint) (*model.GeofenceLog, error)
	GetByVesselID(vesselID uint, page, limit int) ([]model.GeofenceLog, int64, error)
	GetByGeofenceID(geofenceID uint, vesselID *uint, isOutside *bool, page, limit int) ([]model.GeofenceLog, int64, error)
	GetLatestByVesselAndGeofence(vesselID, geofenceID uint) (*model.GeofenceLog, error)
	ListLogs(companyID *uint, vesselID *uint, geofenceID *uint, eventType string, page, limit int) ([]model.GeofenceLog, int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(log *model.GeofenceLog) error {
	return r.db.Create(log).Error
}

func (r *repository) GetByID(id uint) (*model.GeofenceLog, error) {
	var log model.GeofenceLog
	err := r.db.First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *repository) GetByVesselID(vesselID uint, page, limit int) ([]model.GeofenceLog, int64, error) {
	var logs []model.GeofenceLog
	var total int64

	query := r.db.Model(&model.GeofenceLog{}).Where("vessel_id = ?", vesselID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

func (r *repository) GetByGeofenceID(geofenceID uint, vesselID *uint, isOutside *bool, page, limit int) ([]model.GeofenceLog, int64, error) {
	var logs []model.GeofenceLog
	var total int64

	query := r.db.Model(&model.GeofenceLog{}).Where("geofence_id = ?", geofenceID)

	// Apply filters
	if vesselID != nil {
		query = query.Where("vessel_id = ?", *vesselID)
	}
	if isOutside != nil {
		query = query.Where("is_outside_geofence = ?", *isOutside)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

func (r *repository) GetLatestByVesselAndGeofence(vesselID, geofenceID uint) (*model.GeofenceLog, error) {
	var log model.GeofenceLog
	err := r.db.Where("vessel_id = ? AND geofence_id = ?", vesselID, geofenceID).
		Order("created_at DESC").
		First(&log).Error

	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *repository) ListLogs(companyID *uint, vesselID *uint, geofenceID *uint, eventType string, page, limit int) ([]model.GeofenceLog, int64, error) {
	var logs []model.GeofenceLog
	var total int64

	query := r.db.Model(&model.GeofenceLog{})
	if companyID != nil {
		query = query.Joins("JOIN m_vessels ON m_vessels.id = geofence_logs.vessel_id").
			Where("m_vessels.company_id = ?", *companyID)
	}
	if vesselID != nil && *vesselID > 0 {
		query = query.Where("geofence_logs.vessel_id = ?", *vesselID)
	}
	if geofenceID != nil && *geofenceID > 0 {
		query = query.Where("geofence_logs.geofence_id = ?", *geofenceID)
	}
	if eventType != "" && eventType != "all" {
		query = query.Where("geofence_logs.event_type = ?", eventType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	err := query.Preload("Vessel").Preload("Geofence").Order("geofence_logs.created_at DESC").Find(&logs).Error
	return logs, total, err
}
