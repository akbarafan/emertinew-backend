package materialhull

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"gorm.io/gorm"
)

type Repository interface {
	Create(materialHull *model.MMaterialHull) error
	GetByID(id uint) (*model.MMaterialHull, error)
	GetByName(name string) (*model.MMaterialHull, error)
	GetAll() ([]model.MMaterialHull, error)
	Update(materialHull *model.MMaterialHull) error
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

func (r *repository) Create(materialHull *model.MMaterialHull) error {
	return r.db.Create(materialHull).Error
}

func (r *repository) GetByID(id uint) (*model.MMaterialHull, error) {
	var materialHull model.MMaterialHull
	err := r.db.First(&materialHull, id).Error
	if err != nil {
		return nil, err
	}
	return &materialHull, nil
}

func (r *repository) GetByName(name string) (*model.MMaterialHull, error) {
	var materialHull model.MMaterialHull
	err := r.db.Where("name = ?", name).First(&materialHull).Error
	if err != nil {
		return nil, err
	}
	return &materialHull, nil
}

func (r *repository) GetAll() ([]model.MMaterialHull, error) {
	var materialHulls []model.MMaterialHull
	err := r.db.Order("id ASC").Find(&materialHulls).Error
	return materialHulls, err
}

func (r *repository) Update(materialHull *model.MMaterialHull) error {
	return r.db.Save(materialHull).Error
}

func (r *repository) Delete(id uint) error {
	// Get the record first
	var materialHull model.MMaterialHull
	if err := r.db.First(&materialHull, id).Error; err != nil {
		return err
	}

	// Add suffix to unique field before soft delete to release constraint
	materialHull.Name = materialHull.Name + utils.GenerateDeletedSuffix()
	if err := r.db.Save(&materialHull).Error; err != nil {
		return err
	}

	return r.db.Delete(&materialHull).Error
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}
