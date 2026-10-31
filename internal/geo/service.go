package geo

import (
	"errors"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Service interface {
	CreateAssignment(req CreateAssignmentRequest) (*AssignmentResponse, error)
	GetAssignment(id uint) (*AssignmentDetailResponse, error)
	ListAssignments(vesselID, geofenceID *uint, status string, page, limit int) ([]AssignmentResponse, int64, error)
	UpdateAssignment(id uint, req UpdateAssignmentRequest) (*AssignmentResponse, error)
	DeleteAssignment(id uint) error

	ListCpLogs(assignmentID, vesselID, geofenceID *uint, page, limit int) ([]CpLogResponse, int64, error)
	ListViolationLogs(vesselID, geofenceID *uint, isOutside *bool, page, limit int) ([]ViolationLogResponse, int64, error)
}

type service struct {
	repo Repository
	db   *gorm.DB
}

func NewService(repo Repository, db *gorm.DB) Service {
	return &service{repo: repo, db: db}
}

// ── Assignment ─────────────────────────────────────────────────────

func (s *service) CreateAssignment(req CreateAssignmentRequest) (*AssignmentResponse, error) {
	if req.VesselID == 0 || req.GeofenceID == 0 {
		return nil, errors.New("vessel dan geofence harus diisi")
	}

	existing, _ := s.repo.GetAssignmentByVesselAndGeofence(req.VesselID, req.GeofenceID)
	if existing != nil && existing.Status != "cancelled" && existing.Status != "completed" {
		return nil, errors.New("vessel sudah memiliki penugasan aktif untuk geofence ini")
	}

	// If there's a cancelled or completed assignment, reset it instead of creating new
	if existing != nil {
		existing.Status = "assigned"
		existing.StartedAt = nil
		existing.CompletedAt = nil
		existing.DeletedAt = gorm.DeletedAt{} // clear soft-delete
		if err := s.repo.UpdateAssignment(existing); err != nil {
			return nil, err
		}
		return s.toAssignmentResponse(existing), nil
	}

	assignment := &model.VesselGeofenceAssignment{
		VesselID:   req.VesselID,
		GeofenceID: req.GeofenceID,
		Status:     "assigned",
	}

	if err := s.repo.CreateAssignment(assignment); err != nil {
		return nil, err
	}

	return s.toAssignmentResponse(assignment), nil
}

func (s *service) GetAssignment(id uint) (*AssignmentDetailResponse, error) {
	a, err := s.repo.GetAssignmentByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("penugasan tidak ditemukan")
		}
		return nil, err
	}

	resp := s.toAssignmentResponse(a)
	cpLogs, _ := s.repo.GetCpLogsByAssignment(id)

	cpResp := make([]CpLogResponse, len(cpLogs))
	for i, l := range cpLogs {
		cpResp[i] = *s.toCpLogResponse(&l)
	}

	stats := calculateCpTimeStats(cpLogs)

	return &AssignmentDetailResponse{
		Assignment:  *resp,
		CpLogs:      cpResp,
		CpTimeStats: stats,
	}, nil
}

func (s *service) ListAssignments(vesselID, geofenceID *uint, status string, page, limit int) ([]AssignmentResponse, int64, error) {
	assignments, total, err := s.repo.ListAssignments(vesselID, geofenceID, status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]AssignmentResponse, len(assignments))
	for i, a := range assignments {
		responses[i] = *s.toAssignmentResponse(&a)
	}

	return responses, total, nil
}

func (s *service) UpdateAssignment(id uint, req UpdateAssignmentRequest) (*AssignmentResponse, error) {
	a, err := s.repo.GetAssignmentByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("penugasan tidak ditemukan")
		}
		return nil, err
	}

	if req.Status != nil {
		status := *req.Status
		switch status {
		case "assigned":
			a.Status = "assigned"
			a.StartedAt = nil
			a.CompletedAt = nil
		case "cancelled":
			a.Status = "cancelled"
		default:
			return nil, errors.New("status tidak valid")
		}
	}

	if err := s.repo.UpdateAssignment(a); err != nil {
		return nil, err
	}

	return s.toAssignmentResponse(a), nil
}

func (s *service) DeleteAssignment(id uint) error {
	_, err := s.repo.GetAssignmentByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("penugasan tidak ditemukan")
		}
		return err
	}

	return s.repo.DeleteAssignment(id)
}

// ── CP Logs ────────────────────────────────────────────────────────

func (s *service) ListCpLogs(assignmentID, vesselID, geofenceID *uint, page, limit int) ([]CpLogResponse, int64, error) {
	logs, total, err := s.repo.ListCpLogs(assignmentID, vesselID, geofenceID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]CpLogResponse, len(logs))
	for i, l := range logs {
		responses[i] = *s.toCpLogResponse(&l)
	}

	return responses, total, nil
}

// ── Violation Logs ─────────────────────────────────────────────────

func (s *service) ListViolationLogs(vesselID, geofenceID *uint, isOutside *bool, page, limit int) ([]ViolationLogResponse, int64, error) {
	logs, total, err := s.repo.ListViolationLogs(vesselID, geofenceID, isOutside, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]ViolationLogResponse, len(logs))
	for i, l := range logs {
		vesselName := ""
		var ves model.MVessel
		if err := s.db.First(&ves, l.VesselID).Error; err == nil {
			vesselName = ves.Name
		}

		geofenceName := ""
		var gf model.MGeofence
		if err := s.db.First(&gf, l.GeofenceID).Error; err == nil {
			geofenceName = gf.Name
		}

		responses[i] = ViolationLogResponse{
			ID:                l.ID,
			VesselID:          l.VesselID,
			VesselName:        vesselName,
			GeofenceID:        l.GeofenceID,
			GeofenceName:      geofenceName,
			Latitude:          l.Latitude,
			Longitude:         l.Longitude,
			IsOutsideGeofence: l.IsOutsideGeofence,
			Distance:          l.Distance,
			CreatedAt:         l.CreatedAt,
		}
	}

	return responses, total, nil
}

// ── Helpers ────────────────────────────────────────────────────────

func (s *service) toAssignmentResponse(a *model.VesselGeofenceAssignment) *AssignmentResponse {
	reachedCP, _ := s.repo.CountReachedCP(a.ID)
	totalCP := 0

	var geofence model.MGeofence
	if err := s.db.First(&geofence, a.GeofenceID).Error; err == nil {
		totalCP = len(geofence.Points)
	}

	vesselName := ""
	var vessel model.MVessel
	if err := s.db.First(&vessel, a.VesselID).Error; err == nil {
		vesselName = vessel.Name
	}

	geofenceName := ""
	if err := s.db.First(&geofence, a.GeofenceID).Error; err == nil {
		geofenceName = geofence.Name
	}

	return &AssignmentResponse{
		ID:           a.ID,
		VesselID:     a.VesselID,
		VesselName:   vesselName,
		GeofenceID:   a.GeofenceID,
		GeofenceName: geofenceName,
		TotalCP:      totalCP,
		ReachedCP:    int(reachedCP),
		Status:       a.Status,
		StartedAt:    a.StartedAt,
		CompletedAt:  a.CompletedAt,
		CreatedAt:    a.CreatedAt,
	}
}

func (s *service) toCpLogResponse(l *model.VesselCpLog) *CpLogResponse {
	vesselName := ""
	var vessel model.MVessel
	if err := s.db.First(&vessel, l.VesselID).Error; err == nil {
		vesselName = vessel.Name
	}

	geofenceName := ""
	var geofence model.MGeofence
	if err := s.db.First(&geofence, l.GeofenceID).Error; err == nil {
		geofenceName = geofence.Name
	}

	return &CpLogResponse{
		ID:             l.ID,
		AssignmentID:   l.AssignmentID,
		VesselID:       l.VesselID,
		VesselName:     vesselName,
		GeofenceID:     l.GeofenceID,
		GeofenceName:   geofenceName,
		PointIndex:     l.PointIndex,
		Latitude:       l.Latitude,
		Longitude:      l.Longitude,
		DistanceMeters: l.DistanceMeters,
		ArrivedAt:      l.ArrivedAt,
		IsSkipped:      l.IsSkipped,
		OrderReached:   l.OrderReached,
	}
}

func calculateCpTimeStats(logs []model.VesselCpLog) []CpTimeStat {
	if len(logs) < 2 {
		return []CpTimeStat{}
	}

	// Sort by order_reached
	sorted := make([]model.VesselCpLog, len(logs))
	copy(sorted, logs)

	var stats []CpTimeStat
	for i := 1; i < len(sorted); i++ {
		prev := sorted[i-1]
		curr := sorted[i]
		duration := curr.ArrivedAt.Sub(prev.ArrivedAt).Hours()

		stats = append(stats, CpTimeStat{
			FromPointIndex: prev.PointIndex,
			ToPointIndex:   curr.PointIndex,
			DurationHours:  roundTo2(duration),
		})
	}

	return stats
}

func roundTo2(v float64) float64 {
	return float64(int(v*100)) / 100
}
