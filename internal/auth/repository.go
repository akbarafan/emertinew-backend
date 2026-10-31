package auth

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	FindByUsername(username string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Company").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Company").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
