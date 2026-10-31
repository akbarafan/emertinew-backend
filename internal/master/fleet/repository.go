package fleet

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"gorm.io/gorm"
)

type Repository interface {
	Create(fleet *model.MFleet) error
	GetByID(id uint) (*model.MFleet, error)
	GetByCode(code string) (*model.MFleet, error)
	GetAll(companyID *uint) ([]model.MFleet, error)
	Update(fleet *model.MFleet) error
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

func (r *repository) Create(fleet *model.MFleet) error {
	return r.db.Create(fleet).Error
}

func (r *repository) GetByID(id uint) (*model.MFleet, error) {
	var fleet model.MFleet
	err := r.db.Preload("Company").First(&fleet, id).Error
	if err != nil {
		return nil, err
	}
	return &fleet, nil
}

func (r *repository) GetByCode(code string) (*model.MFleet, error) {
	var fleet model.MFleet
	err := r.db.Where("code = ?", code).First(&fleet).Error
	if err != nil {
		return nil, err
	}
	return &fleet, nil
}

func (r *repository) GetAll(companyID *uint) ([]model.MFleet, error) {
	var fleets []model.MFleet
	q := r.db.Preload("Company").Order("id ASC")
	if companyID != nil {
		q = q.Where("company_id = ?", *companyID)
	}
	err := q.Find(&fleets).Error
	return fleets, err
}

func (r *repository) Update(fleet *model.MFleet) error {
	return r.db.Save(fleet).Error
}

func (r *repository) Delete(id uint) error {
	var fleet model.MFleet
	if err := r.db.First(&fleet, id).Error; err != nil {
		return err
	}

	fleet.Code = fleet.Code + utils.GenerateDeletedSuffix()
	if err := r.db.Save(&fleet).Error; err != nil {
		return err
	}

	return r.db.Delete(&fleet).Error
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}
