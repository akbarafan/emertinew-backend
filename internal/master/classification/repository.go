package classification

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"gorm.io/gorm"
)

type Repository interface {
	Create(classification *model.MClassification) error
	GetByID(id uint) (*model.MClassification, error)
	GetByName(name string) (*model.MClassification, error)
	GetAll() ([]model.MClassification, error)
	Update(classification *model.MClassification) error
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

func (r *repository) Create(classification *model.MClassification) error {
	return r.db.Create(classification).Error
}

func (r *repository) GetByID(id uint) (*model.MClassification, error) {
	var classification model.MClassification
	err := r.db.First(&classification, id).Error
	if err != nil {
		return nil, err
	}
	return &classification, nil
}

func (r *repository) GetByName(name string) (*model.MClassification, error) {
	var classification model.MClassification
	err := r.db.Where("name = ?", name).First(&classification).Error
	if err != nil {
		return nil, err
	}
	return &classification, nil
}

func (r *repository) GetAll() ([]model.MClassification, error) {
	var classifications []model.MClassification
	err := r.db.Order("id ASC").Find(&classifications).Error
	return classifications, err
}

func (r *repository) Update(classification *model.MClassification) error {
	return r.db.Save(classification).Error
}

func (r *repository) Delete(id uint) error {
	// Get the record first
	var classification model.MClassification
	if err := r.db.First(&classification, id).Error; err != nil {
		return err
	}

	// Add suffix to unique field before soft delete to release constraint
	classification.Name = classification.Name + utils.GenerateDeletedSuffix()
	if err := r.db.Save(&classification).Error; err != nil {
		return err
	}

	return r.db.Delete(&classification).Error
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}
