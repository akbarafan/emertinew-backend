package geo

import (
	"errors"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	// Assignment
	CreateAssignment(a *model.VesselGeofenceAssignment) error
	GetAssignmentByID(id uint) (*model.VesselGeofenceAssignment, error)
	GetAssignmentByVesselAndGeofence(vesselID, geofenceID uint) (*model.VesselGeofenceAssignment, error)
	ListAssignments(vesselID, geofenceID *uint, status string, page, limit int) ([]model.VesselGeofenceAssignment, int64, error)
	UpdateAssignment(a *model.VesselGeofenceAssignment) error
	DeleteAssignment(id uint) error
	GetActiveAssignments() ([]model.VesselGeofenceAssignment, error)

	// CP Log
	CreateCpLog(l *model.VesselCpLog) error
	ListCpLogs(assignmentID, vesselID, geofenceID *uint, page, limit int) ([]model.VesselCpLog, int64, error)
	GetCpLogsByAssignment(assignmentID uint) ([]model.VesselCpLog, error)
	GetLatestCpLogByAssignment(assignmentID uint) (*model.VesselCpLog, error)
	CountReachedCP(assignmentID uint) (int64, error)

	// Violation Log (existing GeofenceLog)
	ListViolationLogs(vesselID, geofenceID *uint, isOutside *bool, page, limit int) ([]model.GeofenceLog, int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ── Assignment ─────────────────────────────────────────────────────

func (r *repository) CreateAssignment(a *model.VesselGeofenceAssignment) error {
	return r.db.Create(a).Error
}

func (r *repository) GetAssignmentByID(id uint) (*model.VesselGeofenceAssignment, error) {
	var a model.VesselGeofenceAssignment
	err := r.db.First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *repository) GetAssignmentByVesselAndGeofence(vesselID, geofenceID uint) (*model.VesselGeofenceAssignment, error) {
	var a model.VesselGeofenceAssignment
	err := r.db.Unscoped().Where("vessel_id = ? AND geofence_id = ?", vesselID, geofenceID).First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *repository) ListAssignments(vesselID, geofenceID *uint, status string, page, limit int) ([]model.VesselGeofenceAssignment, int64, error) {
	var assignments []model.VesselGeofenceAssignment
	var total int64

	query := r.db.Model(&model.VesselGeofenceAssignment{})

	if vesselID != nil {
		query = query.Where("vessel_id = ?", *vesselID)
	}
	if geofenceID != nil {
		query = query.Where("geofence_id = ?", *geofenceID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	err := query.Order("id DESC").Find(&assignments).Error
	return assignments, total, err
}

func (r *repository) UpdateAssignment(a *model.VesselGeofenceAssignment) error {
	return r.db.Unscoped().Save(a).Error
}

func (r *repository) DeleteAssignment(id uint) error {
	r.db.Where("assignment_id = ?", id).Delete(&model.VesselCpLog{})
	return r.db.Delete(&model.VesselGeofenceAssignment{}, id).Error
}

func (r *repository) GetActiveAssignments() ([]model.VesselGeofenceAssignment, error) {
	var assignments []model.VesselGeofenceAssignment
	err := r.db.Where("status IN ?", []string{"assigned", "in_progress"}).
		Order("id ASC").Find(&assignments).Error
	return assignments, err
}

// ── CP Log ─────────────────────────────────────────────────────────

func (r *repository) CreateCpLog(l *model.VesselCpLog) error {
	return r.db.Create(l).Error
}

func (r *repository) ListCpLogs(assignmentID, vesselID, geofenceID *uint, page, limit int) ([]model.VesselCpLog, int64, error) {
	var logs []model.VesselCpLog
	var total int64

	query := r.db.Model(&model.VesselCpLog{})

	if assignmentID != nil {
		query = query.Where("assignment_id = ?", *assignmentID)
	}
	if vesselID != nil {
		query = query.Where("vessel_id = ?", *vesselID)
	}
	if geofenceID != nil {
		query = query.Where("geofence_id = ?", *geofenceID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	err := query.Order("arrived_at DESC").Find(&logs).Error
	return logs, total, err
}

func (r *repository) GetCpLogsByAssignment(assignmentID uint) ([]model.VesselCpLog, error) {
	var logs []model.VesselCpLog
	err := r.db.Where("assignment_id = ?", assignmentID).
		Order("arrived_at ASC").Find(&logs).Error
	return logs, err
}

func (r *repository) GetLatestCpLogByAssignment(assignmentID uint) (*model.VesselCpLog, error) {
	var log model.VesselCpLog
	err := r.db.Where("assignment_id = ?", assignmentID).
		Order("arrived_at DESC").First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *repository) CountReachedCP(assignmentID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.VesselCpLog{}).
		Where("assignment_id = ?", assignmentID).
		Count(&count).Error
	return count, err
}

// ── Violation Log ──────────────────────────────────────────────────

func (r *repository) ListViolationLogs(vesselID, geofenceID *uint, isOutside *bool, page, limit int) ([]model.GeofenceLog, int64, error) {
	var logs []model.GeofenceLog
	var total int64

	query := r.db.Model(&model.GeofenceLog{})

	if vesselID != nil {
		query = query.Where("vessel_id = ?", *vesselID)
	}
	if geofenceID != nil {
		query = query.Where("geofence_id = ?", *geofenceID)
	}
	if isOutside != nil {
		query = query.Where("is_outside_geofence = ?", *isOutside)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}
