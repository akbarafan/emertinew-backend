package port

import (
	"errors"
	"strings"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"
	"gorm.io/gorm"
)

type Repository interface {
	Create(port *model.MPort) error
	GetByID(id uint) (*model.MPort, error)
	GetAll() ([]model.MPort, error)
	Update(port *model.MPort) error
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

func (r *repository) Create(port *model.MPort) error {
	err := r.db.Create(port).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "SQLSTATE 23505") {
			return utils.ErrDuplicateRecord
		}
	}
	return err
}

func (r *repository) GetByID(id uint) (*model.MPort, error) {
	var port model.MPort
	err := r.db.First(&port, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}
	return &port, nil
}

func (r *repository) GetAll() ([]model.MPort, error) {
	var ports []model.MPort
	err := r.db.Order("id ASC").Find(&ports).Error
	return ports, err
}

func (r *repository) Update(port *model.MPort) error {
	err := r.db.Save(port).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "SQLSTATE 23505") {
			return utils.ErrDuplicateRecord
		}
	}
	return err
}

func (r *repository) Delete(id uint) error {
	// Get the record first
	var port model.MPort
	if err := r.db.First(&port, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrRecordNotFound
		}
		return err
	}

	// Add suffix to unique field before soft delete to release constraint
	port.Code = port.Code + utils.GenerateDeletedSuffix()
	if err := r.db.Save(&port).Error; err != nil {
		return err
	}

	return r.db.Delete(&port).Error
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}
