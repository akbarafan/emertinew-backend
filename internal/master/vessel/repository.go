package vessel

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(vessel *model.MVessel) error
	GetByID(id uint) (*model.MVessel, error)
	GetByName(name string) (*model.MVessel, error)
	GetByCodeCI(code string) (*model.MVessel, error)
	GetAll() ([]model.MVessel, error)
	Update(vessel *model.MVessel) error
	Delete(id uint) error
	Restore(id uint) error
	GetDeleted() ([]model.MVessel, error)
	AssignMCUsToVessel(vesselID uint, mcuIDs []uint) ([]uint, error)
	ValidateMCUIDs(mcuIDs []uint) ([]uint, []uint, error)
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

func (r *repository) Create(vessel *model.MVessel) error {
	return r.db.Create(vessel).Error
}

func (r *repository) GetByCodeCI(code string) (*model.MVessel, error) {
	var vessel model.MVessel
	err := r.db.Where("UPPER(code) = UPPER(?)", code).First(&vessel).Error
	if err != nil {
		return nil, err
	}
	return &vessel, nil
}

func (r *repository) GetByID(id uint) (*model.MVessel, error) {
	var vessel model.MVessel
	err := r.db.Preload("MCUs").
		Preload("Company").
		Preload("VesselType").
		Preload("Classification").
		Preload("MaterialHull").
		Preload("OperationStatus").
		Preload("Fleet").
		Preload("Squad").
		Preload("FuelType").
		Preload("Dimensions").
		Preload("Capacity").
		First(&vessel, id).Error
	if err != nil {
		return nil, err
	}
	return &vessel, nil
}

func (r *repository) GetByName(name string) (*model.MVessel, error) {
	var vessel model.MVessel
	err := r.db.Preload("MCUs").
		Preload("Company").
		Preload("VesselType").
		Preload("Classification").
		Preload("MaterialHull").
		Preload("OperationStatus").
		Preload("Fleet").
		Preload("Squad").
		Preload("FuelType").
		Where("name = ?", name).First(&vessel).Error
	if err != nil {
		return nil, err
	}
	return &vessel, nil
}

func (r *repository) GetAll() ([]model.MVessel, error) {
	var vessels []model.MVessel
	err := r.db.Preload("MCUs").
		Preload("Company").
		Preload("VesselType").
		Preload("Classification").
		Preload("MaterialHull").
		Preload("OperationStatus").
		Preload("Fleet").
		Preload("Squad").
		Preload("FuelType").
		Order("id ASC").Find(&vessels).Error
	return vessels, err
}

func (r *repository) Update(vessel *model.MVessel) error {
	return r.db.Model(&model.MVessel{}).Where("id = ?", vessel.ID).Select(
		"Code", "Name", "IsActive",
		"CompanyID", "VesselTypeID", "ClassID", "MaterialID", "OperationID",
		"FleetID", "SquadID",
		"CallSign", "IMONumber", "MMSINumber", "Flag", "PortOfRegistry",
		"ProductionYear", "Builder", "HullNumber", "ClassNotation", "Nickname",
		"TrackingSource", "UpdatedBy",
	).Updates(vessel).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.MVessel{}).Error
}

func (r *repository) Restore(id uint) error {
	return r.db.Unscoped().Model(&model.MVessel{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *repository) GetDeleted() ([]model.MVessel, error) {
	var vessels []model.MVessel
	err := r.db.Unscoped().
		Preload("MCUs").
		Preload("Company").
		Preload("VesselType").
		Preload("Classification").
		Preload("MaterialHull").
		Preload("OperationStatus").
		Preload("Fleet").
		Preload("Squad").
		Preload("FuelType").
		Where("deleted_at IS NOT NULL").
		Order("deleted_at DESC").
		Find(&vessels).Error
	return vessels, err
}

func (r *repository) ValidateMCUIDs(mcuIDs []uint) ([]uint, []uint, error) {
	var existingIDs []uint
	err := r.db.Model(&model.MMCU{}).Where("id IN ?", mcuIDs).Pluck("id", &existingIDs).Error
	if err != nil {
		return nil, nil, err
	}

	// Find missing IDs
	missingIDs := []uint{}
	for _, id := range mcuIDs {
		found := false
		for _, existingID := range existingIDs {
			if id == existingID {
				found = true
				break
			}
		}
		if !found {
			missingIDs = append(missingIDs, id)
		}
	}

	return existingIDs, missingIDs, nil
}

func (r *repository) AssignMCUsToVessel(vesselID uint, mcuIDs []uint) ([]uint, error) {
	// Validate that all mcu IDs exist
	if len(mcuIDs) > 0 {
		_, missingIDs, err := r.ValidateMCUIDs(mcuIDs)
		if err != nil {
			return nil, err
		}
		if len(missingIDs) > 0 {
			return missingIDs, gorm.ErrRecordNotFound
		}
	}

	// Unassign all mcus from this vessel first
	err := r.db.Model(&model.MMCU{}).Where("vessel_id = ?", vesselID).Update("vessel_id", nil).Error
	if err != nil {
		return nil, err
	}

	// Assign new mcus to this vessel
	if len(mcuIDs) > 0 {
		return nil, r.db.Model(&model.MMCU{}).Where("id IN ?", mcuIDs).Update("vessel_id", vesselID).Error
	}

	return nil, nil
}
