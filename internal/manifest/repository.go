package manifest

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(manifest *model.Manifest) error
	GetByID(id uint) (*model.Manifest, error)
	GetAll() ([]model.Manifest, error)
	GetByVesselID(vesselID uint) ([]model.Manifest, error)
	GetByStatus(status string) ([]model.Manifest, error)
	Update(manifest *model.Manifest) error
	Delete(id uint) error
	CountByDatePrefix(datePrefix string) (int64, error)
	VesselExists(vesselID uint) (bool, error)
	GetCrewByID(crewID uint) (*model.MCrew, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(manifest *model.Manifest) error {
	return r.db.Create(manifest).Error
}

func (r *repository) GetByID(id uint) (*model.Manifest, error) {
	var manifest model.Manifest
	err := r.db.Preload("Vessel").Preload("Captain").First(&manifest, id).Error
	if err != nil {
		return nil, err
	}
	return &manifest, nil
}

func (r *repository) GetAll() ([]model.Manifest, error) {
	var manifests []model.Manifest
	err := r.db.Preload("Vessel").Preload("Captain").Order("id ASC").Find(&manifests).Error
	return manifests, err
}

func (r *repository) GetByVesselID(vesselID uint) ([]model.Manifest, error) {
	var manifests []model.Manifest
	err := r.db.Preload("Vessel").Preload("Captain").Where("vessel_id = ?", vesselID).Order("id ASC").Find(&manifests).Error
	return manifests, err
}

func (r *repository) GetByStatus(status string) ([]model.Manifest, error) {
	var manifests []model.Manifest
	err := r.db.Preload("Vessel").Preload("Captain").Where("status = ?", status).Order("id ASC").Find(&manifests).Error
	return manifests, err
}

func (r *repository) Update(manifest *model.Manifest) error {
	return r.db.Omit("Vessel", "Captain").Save(manifest).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&model.Manifest{}, id).Error
}

func (r *repository) CountByDatePrefix(datePrefix string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Manifest{}).Where("name LIKE ?", "MNF-"+datePrefix+"%").Count(&count).Error
	return count, err
}

func (r *repository) VesselExists(vesselID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.MVessel{}).Where("id = ?", vesselID).Count(&count).Error
	return count > 0, err
}

func (r *repository) GetCrewByID(crewID uint) (*model.MCrew, error) {
	var crew model.MCrew
	err := r.db.First(&crew, crewID).Error
	if err != nil {
		return nil, err
	}
	return &crew, nil
}
