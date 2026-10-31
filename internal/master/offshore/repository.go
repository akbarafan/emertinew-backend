package offshore

import (
	"errors"
	"strings"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"
	"gorm.io/gorm"
)

type Repository interface {
	Create(offshore *model.MOffshore) error
	GetByID(id uint) (*model.MOffshore, error)
	GetAll(companyID *uint) ([]model.MOffshore, error)
	Update(offshore *model.MOffshore) error
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

func (r *repository) Create(offshore *model.MOffshore) error {
	err := r.db.Create(offshore).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "SQLSTATE 23505") {
			return utils.ErrDuplicateRecord
		}
	}
	return err
}

func (r *repository) GetByID(id uint) (*model.MOffshore, error) {
	var offshore model.MOffshore
	err := r.db.First(&offshore, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}
	return &offshore, nil
}

func (r *repository) GetAll(companyID *uint) ([]model.MOffshore, error) {
	var offshores []model.MOffshore
	q := r.db.Order("id ASC")
	if companyID != nil {
		q = q.Where("company_id = ?", *companyID)
	}
	err := q.Find(&offshores).Error
	return offshores, err
}

func (r *repository) Update(offshore *model.MOffshore) error {
	err := r.db.Save(offshore).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "SQLSTATE 23505") {
			return utils.ErrDuplicateRecord
		}
	}
	return err
}

func (r *repository) Delete(id uint) error {
	// Get the record first
	var offshore model.MOffshore
	if err := r.db.First(&offshore, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrRecordNotFound
		}
		return err
	}

	// Add suffix to unique field before soft delete to release constraint
	offshore.Code = offshore.Code + utils.GenerateDeletedSuffix()
	if err := r.db.Save(&offshore).Error; err != nil {
		return err
	}

	return r.db.Delete(&offshore).Error
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}
