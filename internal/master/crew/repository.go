package crew

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(crew *model.MCrew) error
	GetByID(id uint) (*model.MCrew, error)
	GetAll() ([]model.MCrew, error)
	GetByVesselID(vesselID uint) ([]model.MCrew, error)
	Update(crew *model.MCrew) error
	Delete(id uint) error
	VesselExists(vesselID uint) (bool, error)
	GetDB() *gorm.DB
	WithTx(tx *gorm.DB) Repository
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(crew *model.MCrew) error {
	return r.db.Create(crew).Error
}

func (r *repository) GetByID(id uint) (*model.MCrew, error) {
	var crew model.MCrew
	err := r.db.Preload("Vessel").First(&crew, id).Error
	if err != nil {
		return nil, err
	}
	return &crew, nil
}

func (r *repository) GetAll() ([]model.MCrew, error) {
	var crews []model.MCrew
	err := r.db.Preload("Vessel").Order("id ASC").Find(&crews).Error
	return crews, err
}

func (r *repository) GetByVesselID(vesselID uint) ([]model.MCrew, error) {
	var crews []model.MCrew
	err := r.db.Preload("Vessel").Where("vessel_id = ?", vesselID).Order("id ASC").Find(&crews).Error
	return crews, err
}

func (r *repository) Update(crew *model.MCrew) error {
	return r.db.Omit("Vessel").Save(crew).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&model.MCrew{}, id).Error
}

func (r *repository) VesselExists(vesselID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.MVessel{}).Where("id = ?", vesselID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}
