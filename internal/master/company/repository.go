package company

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"gorm.io/gorm"
)

type Repository interface {
	Create(company *model.MCompany) error
	GetByID(id uint) (*model.MCompany, error)
	GetByCode(code string) (*model.MCompany, error)
	GetByName(name string) (*model.MCompany, error)
	GetAll() ([]model.MCompany, error)
	Update(company *model.MCompany) error
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

func (r *repository) Create(company *model.MCompany) error {
	return r.db.Create(company).Error
}

func (r *repository) GetByID(id uint) (*model.MCompany, error) {
	var company model.MCompany
	err := r.db.First(&company, id).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *repository) GetByCode(code string) (*model.MCompany, error) {
	var company model.MCompany
	err := r.db.Where("code = ?", code).First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *repository) GetByName(name string) (*model.MCompany, error) {
	var company model.MCompany
	err := r.db.Where("name = ?", name).First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *repository) GetAll() ([]model.MCompany, error) {
	var companies []model.MCompany
	err := r.db.Order("id ASC").Find(&companies).Error
	return companies, err
}

func (r *repository) Update(company *model.MCompany) error {
	return r.db.Save(company).Error
}

func (r *repository) Delete(id uint) error {
	var company model.MCompany
	if err := r.db.First(&company, id).Error; err != nil {
		return err
	}

	company.Code = company.Code + utils.GenerateDeletedSuffix()
	if err := r.db.Save(&company).Error; err != nil {
		return err
	}

	return r.db.Delete(&company).Error
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}
