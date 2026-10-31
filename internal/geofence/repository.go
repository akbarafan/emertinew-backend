package geofence

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(geofence *model.MGeofence) error
	GetByID(id uint) (*model.MGeofence, error)
	GetByName(name string) (*model.MGeofence, error)
	GetAll(page, limit int, companyID *uint) ([]model.MGeofence, int64, error)
	GetAllActive(companyID *uint) ([]model.MGeofence, error)
	Update(geofence *model.MGeofence) error
	Delete(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(geofence *model.MGeofence) error {
	return r.db.Create(geofence).Error
}

func (r *repository) GetByID(id uint) (*model.MGeofence, error) {
	var geofence model.MGeofence
	err := r.db.Preload("Company").First(&geofence, id).Error
	if err != nil {
		return nil, err
	}
	return &geofence, nil
}

func (r *repository) GetByName(name string) (*model.MGeofence, error) {
	var geofence model.MGeofence
	err := r.db.Where("name = ?", name).First(&geofence).Error
	if err != nil {
		return nil, err
	}
	return &geofence, nil
}

func (r *repository) GetAll(page, limit int, companyID *uint) ([]model.MGeofence, int64, error) {
	var geofences []model.MGeofence
	var total int64

	query := r.db.Model(&model.MGeofence{})
	if companyID != nil {
		query = query.Where("company_id = ? OR company_id IS NULL", *companyID)
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

	err := query.Preload("Company").Order("id ASC").Find(&geofences).Error
	return geofences, total, err
}

func (r *repository) GetAllActive(companyID *uint) ([]model.MGeofence, error) {
	var geofences []model.MGeofence
	query := r.db.Where("is_active = ?", true)
	if companyID != nil {
		query = query.Where("company_id = ? OR company_id IS NULL", *companyID)
	}
	err := query.Preload("Company").Order("id ASC").Find(&geofences).Error
	return geofences, err
}

func (r *repository) Update(geofence *model.MGeofence) error {
	return r.db.Save(geofence).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&model.MGeofence{}, id).Error
}
