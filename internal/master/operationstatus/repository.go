package operationstatus

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"gorm.io/gorm"
)

type Repository interface {
	Create(operationStatus *model.MOperationStatus) error
	GetByID(id uint) (*model.MOperationStatus, error)
	GetByName(name string) (*model.MOperationStatus, error)
	GetAll() ([]model.MOperationStatus, error)
	Update(operationStatus *model.MOperationStatus) error
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

func (r *repository) Create(operationStatus *model.MOperationStatus) error {
	return r.db.Create(operationStatus).Error
}

func (r *repository) GetByID(id uint) (*model.MOperationStatus, error) {
	var operationStatus model.MOperationStatus
	err := r.db.First(&operationStatus, id).Error
	if err != nil {
		return nil, err
	}
	return &operationStatus, nil
}

func (r *repository) GetByName(name string) (*model.MOperationStatus, error) {
	var operationStatus model.MOperationStatus
	err := r.db.Where("name = ?", name).First(&operationStatus).Error
	if err != nil {
		return nil, err
	}
	return &operationStatus, nil
}

func (r *repository) GetAll() ([]model.MOperationStatus, error) {
	var operationStatuses []model.MOperationStatus
	err := r.db.Order("id ASC").Find(&operationStatuses).Error
	return operationStatuses, err
}

func (r *repository) Update(operationStatus *model.MOperationStatus) error {
	return r.db.Save(operationStatus).Error
}

func (r *repository) Delete(id uint) error {
	// Get the record first
	var operationStatus model.MOperationStatus
	if err := r.db.First(&operationStatus, id).Error; err != nil {
		return err
	}

	// Add suffix to unique field before soft delete to release constraint
	operationStatus.Name = operationStatus.Name + utils.GenerateDeletedSuffix()
	if err := r.db.Save(&operationStatus).Error; err != nil {
		return err
	}

	return r.db.Delete(&operationStatus).Error
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}
