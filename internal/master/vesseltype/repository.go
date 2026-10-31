package vesseltype

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"gorm.io/gorm"
)

type Repository interface {
	Create(vesselType *model.MVesselType) error
	GetByID(id uint) (*model.MVesselType, error)
	GetByName(name string) (*model.MVesselType, error)
	GetAll() ([]model.MVesselType, error)
	Update(vesselType *model.MVesselType) error
	Delete(id uint) error
	GetDB() *gorm.DB
	WithTx(tx *gorm.DB) Repository
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) Create(vesselType *model.MVesselType) error {
	return r.db.Create(vesselType).Error
}

func (r *repository) GetByID(id uint) (*model.MVesselType, error) {
	var vesselType model.MVesselType
	err := r.db.First(&vesselType, id).Error
	if err != nil {
		return nil, err
	}
	return &vesselType, nil
}

func (r *repository) GetByName(name string) (*model.MVesselType, error) {
	var vesselType model.MVesselType
	err := r.db.Where("name = ?", name).First(&vesselType).Error
	if err != nil {
		return nil, err
	}
	return &vesselType, nil
}

func (r *repository) GetAll() ([]model.MVesselType, error) {
	var vesselTypes []model.MVesselType
	err := r.db.Order("id ASC").Find(&vesselTypes).Error
	return vesselTypes, err
}

func (r *repository) Update(vesselType *model.MVesselType) error {
	return r.db.Save(vesselType).Error
}

func (r *repository) Delete(id uint) error {
	// Get the record first
	var vesselType model.MVesselType
	if err := r.db.First(&vesselType, id).Error; err != nil {
		return err
	}

	// Add suffix to unique field before soft delete to release constraint
	vesselType.Name = vesselType.Name + utils.GenerateDeletedSuffix()
	if err := r.db.Save(&vesselType).Error; err != nil {
		return err
	}

	return r.db.Delete(&vesselType).Error
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}
