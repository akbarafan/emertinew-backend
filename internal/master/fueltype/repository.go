package fueltype

import (
	"errors"
	"strings"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"
	"gorm.io/gorm"
)

type Repository interface {
	Create(fuelType *model.MFuelType) error
	GetByID(id uint) (*model.MFuelType, error)
	GetByCode(code string) (*model.MFuelType, error)
	GetAll() ([]model.MFuelType, error)
	Update(fuelType *model.MFuelType) error
	Delete(id uint) error
	GetDB() *gorm.DB
	WithTx(tx *gorm.DB) Repository
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(fuelType *model.MFuelType) error {
	err := r.db.Create(fuelType).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "SQLSTATE 23505") {
			return utils.ErrDuplicateRecord
		}
	}
	return err
}

func (r *repository) GetByID(id uint) (*model.MFuelType, error) {
	var fuelType model.MFuelType
	if err := r.db.First(&fuelType, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}
	return &fuelType, nil
}

func (r *repository) GetByCode(code string) (*model.MFuelType, error) {
	var fuelType model.MFuelType
	if err := r.db.Where("code = ?", code).First(&fuelType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}
	return &fuelType, nil
}

func (r *repository) GetAll() ([]model.MFuelType, error) {
	var fuelTypes []model.MFuelType
	err := r.db.Order("id ASC").Find(&fuelTypes).Error
	return fuelTypes, err
}

func (r *repository) Update(fuelType *model.MFuelType) error {
	err := r.db.Save(fuelType).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "SQLSTATE 23505") {
			return utils.ErrDuplicateRecord
		}
	}
	return err
}

func (r *repository) Delete(id uint) error {
	var fuelType model.MFuelType
	if err := r.db.First(&fuelType, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrRecordNotFound
		}
		return err
	}

	// Add suffix to unique field before soft delete to release constraint
	fuelType.Code = fuelType.Code + utils.GenerateDeletedSuffix()
	if err := r.db.Save(&fuelType).Error; err != nil {
		return err
	}

	return r.db.Delete(&fuelType).Error
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}
