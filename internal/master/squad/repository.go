package squad

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"gorm.io/gorm"
)

type Repository interface {
	Create(squad *model.MSquad) error
	GetByID(id uint) (*model.MSquad, error)
	GetByCode(code string) (*model.MSquad, error)
	GetByFleetID(fleetID uint) ([]model.MSquad, error)
	GetAll(companyID *uint) ([]model.MSquad, error)
	Update(squad *model.MSquad) error
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

func (r *repository) Create(squad *model.MSquad) error {
	return r.db.Create(squad).Error
}

func (r *repository) GetByID(id uint) (*model.MSquad, error) {
	var squad model.MSquad
	err := r.db.Preload("Company").Preload("Fleet").First(&squad, id).Error
	if err != nil {
		return nil, err
	}
	return &squad, nil
}

func (r *repository) GetByCode(code string) (*model.MSquad, error) {
	var squad model.MSquad
	err := r.db.Where("code = ?", code).First(&squad).Error
	if err != nil {
		return nil, err
	}
	return &squad, nil
}

func (r *repository) GetByFleetID(fleetID uint) ([]model.MSquad, error) {
	var squads []model.MSquad
	err := r.db.Where("fleet_id = ?", fleetID).Order("id ASC").Find(&squads).Error
	return squads, err
}

func (r *repository) GetAll(companyID *uint) ([]model.MSquad, error) {
	var squads []model.MSquad
	q := r.db.Preload("Company").Preload("Fleet").Order("id ASC")
	if companyID != nil {
		q = q.Where("company_id = ?", *companyID)
	}
	err := q.Find(&squads).Error
	return squads, err
}

func (r *repository) Update(squad *model.MSquad) error {
	return r.db.Save(squad).Error
}

func (r *repository) Delete(id uint) error {
	var squad model.MSquad
	if err := r.db.First(&squad, id).Error; err != nil {
		return err
	}

	squad.Code = squad.Code + utils.GenerateDeletedSuffix()
	if err := r.db.Save(&squad).Error; err != nil {
		return err
	}

	return r.db.Delete(&squad).Error
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}
