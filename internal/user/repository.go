package user

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(user *model.User) error
	GetAll(companyID *uint) ([]model.User, error)
	GetByID(id uint) (*model.User, error)
	Update(user *model.User) error
	Delete(id uint) error
	FindByUsername(username string) (*model.User, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *repository) GetAll(companyID *uint) ([]model.User, error) {
	var users []model.User
	q := r.db.Preload("Company").Order("id ASC")
	if companyID != nil && *companyID > 0 {
		q = q.Where("company_id = ?", *companyID)
	}
	err := q.Find(&users).Error
	return users, err
}

func (r *repository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Company").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) Update(user *model.User) error {
	return r.db.Omit("Company").Save(user).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.User{}).Error
}

func (r *repository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Company").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
