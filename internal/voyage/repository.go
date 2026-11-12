package voyage

import (
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	GetAll(status string, vesselID uint, search string, companyID *uint) ([]model.Voyage, error)
	GetAllActive(companyID *uint) ([]model.Voyage, error)
	GetActiveByVesselID(vesselID uint) (*model.Voyage, error)
	GetByID(id uint) (*model.Voyage, error)
	GetByVesselID(vesselID uint) ([]model.Voyage, error)
	Create(voyage *model.Voyage) error
	Update(voyage *model.Voyage) error
	CompleteVoyage(id uint) error
	CompletePreviousActive(vesselID uint, excludeID uint) error
	Delete(id uint) error
	GetVesselByID(id uint) (*model.MVessel, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAll(status string, vesselID uint, search string, companyID *uint) ([]model.Voyage, error) {
	var list []model.Voyage
	q := r.db.Preload("Vessel").Order("created_at DESC")
	if companyID != nil && *companyID > 0 {
		q = q.Where("vessel_id IN (SELECT id FROM m_vessels WHERE company_id = ?)", *companyID)
	}
	if status != "" && status != "all" {
		if status == "active" {
			q = q.Where("status != 'completed' AND status != 'draft'")
		} else {
			q = q.Where("status = ?", status)
		}
	}
	if vesselID > 0 {
		q = q.Where("vessel_id = ?", vesselID)
	}
	if search != "" {
		s := "%" + search + "%"
		q = q.Where("voyage_number ILIKE ? OR origin_name ILIKE ? OR destination_name ILIKE ?", s, s, s)
	}
	err := q.Find(&list).Error
	return list, err
}

func (r *repository) GetAllActive(companyID *uint) ([]model.Voyage, error) {
	var list []model.Voyage
	q := r.db.Preload("Vessel").
		Where("status != 'completed' AND status != 'draft'").
		Order("COALESCE(departure_time, created_at) ASC")
	if companyID != nil && *companyID > 0 {
		q = q.Where("vessel_id IN (SELECT id FROM m_vessels WHERE company_id = ?)", *companyID)
	}
	err := q.Find(&list).Error
	return list, err
}

func (r *repository) GetActiveByVesselID(vesselID uint) (*model.Voyage, error) {
	var v model.Voyage
	err := r.db.Preload("Vessel").
		Where("vessel_id = ? AND status != 'completed' AND status != 'draft'", vesselID).
		Order("COALESCE(departure_time, created_at) ASC").
		First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *repository) GetByID(id uint) (*model.Voyage, error) {
	var v model.Voyage
	err := r.db.Preload("Vessel").First(&v, id).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *repository) GetByVesselID(vesselID uint) ([]model.Voyage, error) {
	var list []model.Voyage
	err := r.db.Preload("Vessel").
		Where("vessel_id = ?", vesselID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *repository) Create(voyage *model.Voyage) error {
	return r.db.Create(voyage).Error
}

func (r *repository) Update(voyage *model.Voyage) error {
	return r.db.Omit("Vessel").Save(voyage).Error
}

func (r *repository) CompleteVoyage(id uint) error {
	now := time.Now()
	return r.db.Model(&model.Voyage{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":         "completed",
		"actual_arrival": &now,
	}).Error
}

func (r *repository) CompletePreviousActive(vesselID uint, excludeID uint) error {
	q := r.db.Model(&model.Voyage{}).
		Where("vessel_id = ? AND status != 'completed'", vesselID)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	return q.Update("status", "completed").Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Delete(&model.Voyage{}, id).Error
}

func (r *repository) GetVesselByID(id uint) (*model.MVessel, error) {
	var v model.MVessel
	err := r.db.First(&v, id).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}
