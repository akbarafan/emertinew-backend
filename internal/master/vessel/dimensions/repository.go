package vesseldimensions

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(dimensions *model.VesselDimensions) error
	GetByID(id uint) (*model.VesselDimensions, error)
	GetByVesselID(vesselID uint) (*model.VesselDimensions, error)
	Update(dimensions *model.VesselDimensions) error
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

func (r *repository) Create(dimensions *model.VesselDimensions) error {
	return r.db.Create(dimensions).Error
}

func (r *repository) GetByID(id uint) (*model.VesselDimensions, error) {
	var dimensions model.VesselDimensions
	err := r.db.First(&dimensions, id).Error
	if err != nil {
		return nil, err
	}
	return &dimensions, nil
}

func (r *repository) GetByVesselID(vesselID uint) (*model.VesselDimensions, error) {
	var dimensions model.VesselDimensions
	err := r.db.Where("vessel_id = ?", vesselID).First(&dimensions).Error
	if err != nil {
		return nil, err
	}
	return &dimensions, nil
}

func (r *repository) Update(dimensions *model.VesselDimensions) error {
	return r.db.Save(dimensions).Error
}

func (r *repository) Delete(id uint) error {
	// Get the record first
	var dimensions model.VesselDimensions
	if err := r.db.First(&dimensions, id).Error; err != nil {
		return err
	}

	// Soft delete - partial unique index will ignore deleted records
	return r.db.Delete(&dimensions).Error
}

func (r *repository) VesselExists(vesselID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.MVessel{}).Where("id = ?", vesselID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
