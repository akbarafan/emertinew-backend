package geofence

import (
	"errors"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Service interface {
	CreateGeofence(req CreateGeofenceRequest) (*GeofenceResponse, error)
	GetGeofenceByID(id uint) (*GeofenceResponse, error)
	GetAllGeofences(page, limit int, companyID *uint) ([]GeofenceResponse, int64, error)
	UpdateGeofence(id uint, req UpdateGeofenceRequest) (*GeofenceResponse, error)
	DeleteGeofence(id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateGeofence(req CreateGeofenceRequest) (*GeofenceResponse, error) {
	// Validasi minimal 2 titik untuk garis
	if len(req.Points) < 2 {
		return nil, errors.New("geofence harus memiliki minimal 2 titik")
	}

	// Validasi radius
	if req.RadiusMiles <= 0 {
		return nil, errors.New("radius harus lebih besar dari 0")
	}

	// Check apakah nama sudah ada
	existing, _ := s.repo.GetByName(req.Name)
	if existing != nil {
		return nil, errors.New("geofence dengan nama tersebut sudah ada")
	}

	zoneType := req.ZoneType
	if zoneType == "" {
		zoneType = "work_zone"
	}

	alertOnExit := true
	if req.AlertOnExit != nil {
		alertOnExit = *req.AlertOnExit
	}

	alertOnEntry := false
	if req.AlertOnEntry != nil {
		alertOnEntry = *req.AlertOnEntry
	}

	geofence := &model.MGeofence{
		CompanyID:     req.CompanyID,
		Name:          req.Name,
		Description:   req.Description,
		ZoneType:      zoneType,
		Points:        req.Points,
		IsActive:      req.IsActive,
		RadiusMiles:   req.RadiusMiles,
		MaxSpeedKnots: req.MaxSpeedKnots,
		AlertOnExit:   alertOnExit,
		AlertOnEntry:  alertOnEntry,
	}

	if err := s.repo.Create(geofence); err != nil {
		return nil, err
	}

	if loaded, err := s.repo.GetByID(geofence.ID); err == nil && loaded != nil {
		return s.toResponse(loaded), nil
	}
	return s.toResponse(geofence), nil
}

func (s *service) GetGeofenceByID(id uint) (*GeofenceResponse, error) {
	geofence, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("geofence tidak ditemukan")
		}
		return nil, err
	}

	return s.toResponse(geofence), nil
}

func (s *service) GetAllGeofences(page, limit int, companyID *uint) ([]GeofenceResponse, int64, error) {
	geofences, total, err := s.repo.GetAll(page, limit, companyID)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]GeofenceResponse, len(geofences))
	for i, g := range geofences {
		responses[i] = *s.toResponse(&g)
	}

	return responses, total, nil
}

func (s *service) UpdateGeofence(id uint, req UpdateGeofenceRequest) (*GeofenceResponse, error) {
	geofence, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("geofence tidak ditemukan")
		}
		return nil, err
	}

	if req.SetPublic != nil && *req.SetPublic {
		geofence.CompanyID = nil
	} else if req.CompanyID != nil {
		geofence.CompanyID = req.CompanyID
	}

	// Update fields jika ada
	if req.Name != "" {
		if req.Name != geofence.Name {
			existing, _ := s.repo.GetByName(req.Name)
			if existing != nil {
				return nil, errors.New("geofence dengan nama tersebut sudah ada")
			}
		}
		geofence.Name = req.Name
	}

	if req.Description != "" {
		geofence.Description = req.Description
	}

	if req.ZoneType != "" {
		geofence.ZoneType = req.ZoneType
	}

	if req.Points != nil && len(*req.Points) >= 2 {
		geofence.Points = *req.Points
	}

	if req.IsActive != nil {
		geofence.IsActive = *req.IsActive
	}

	if req.RadiusMiles != nil && *req.RadiusMiles > 0 {
		geofence.RadiusMiles = *req.RadiusMiles
	}

	if req.MaxSpeedKnots != nil {
		geofence.MaxSpeedKnots = req.MaxSpeedKnots
	}

	if req.AlertOnExit != nil {
		geofence.AlertOnExit = *req.AlertOnExit
	}

	if req.AlertOnEntry != nil {
		geofence.AlertOnEntry = *req.AlertOnEntry
	}

	if err := s.repo.Update(geofence); err != nil {
		return nil, err
	}

	if updated, err := s.repo.GetByID(geofence.ID); err == nil && updated != nil {
		return s.toResponse(updated), nil
	}
	return s.toResponse(geofence), nil
}

func (s *service) DeleteGeofence(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("geofence tidak ditemukan")
		}
		return err
	}

	return s.repo.Delete(id)
}

func (s *service) toResponse(g *model.MGeofence) *GeofenceResponse {
	resp := &GeofenceResponse{
		ID:            g.ID,
		CompanyID:     g.CompanyID,
		Name:          g.Name,
		Description:   g.Description,
		ZoneType:      g.ZoneType,
		Points:        g.Points,
		IsActive:      g.IsActive,
		RadiusMiles:   g.RadiusMiles,
		MaxSpeedKnots: g.MaxSpeedKnots,
		AlertOnExit:   g.AlertOnExit,
		AlertOnEntry:  g.AlertOnEntry,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
	}
	if g.Company != nil {
		resp.CompanyName = g.Company.Name
	}
	return resp
}
