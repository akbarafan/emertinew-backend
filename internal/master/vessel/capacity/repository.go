package vesselcapacity

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(capacity *model.VesselCapacity) error
	GetByID(id uint) (*model.VesselCapacity, error)
	GetByVesselID(vesselID uint) (*model.VesselCapacity, error)
	Update(capacity *model.VesselCapacity) error
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

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}

func (r *repository) Create(capacity *model.VesselCapacity) error {
	return r.db.Create(capacity).Error
}

func (r *repository) GetByID(id uint) (*model.VesselCapacity, error) {
	var capacity model.VesselCapacity
	err := r.db.First(&capacity, id).Error
	if err != nil {
		return nil, err
	}
	return &capacity, nil
}

func (r *repository) GetByVesselID(vesselID uint) (*model.VesselCapacity, error) {
	var capacity model.VesselCapacity
	err := r.db.Where("vessel_id = ?", vesselID).First(&capacity).Error
	if err != nil {
		return nil, err
	}
	return &capacity, nil
}

func (r *repository) Update(capacity *model.VesselCapacity) error {
	return r.db.Save(capacity).Error
}

func (r *repository) Delete(id uint) error {
	// Get the record first
	var capacity model.VesselCapacity
	if err := r.db.First(&capacity, id).Error; err != nil {
		return err
	}

	// Soft delete - partial unique index will ignore deleted records
	return r.db.Delete(&capacity).Error
}

func (r *repository) VesselExists(vesselID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.MVessel{}).Where("id = ?", vesselID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
