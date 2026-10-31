package manifest

import (
	"fmt"
	"strings"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

type Service interface {
	CreateManifest(req CreateManifestRequest) (*ManifestResponse, error)
	GetManifestByID(id uint) (*ManifestResponse, error)
	GetAllManifests() ([]ManifestResponse, error)
	GetManifestsByVesselID(vesselID uint) ([]ManifestResponse, error)
	GetManifestsByStatus(status string) ([]ManifestResponse, error)
	UpdateManifest(id uint, req UpdateManifestRequest) (*ManifestResponse, error)
	DeleteManifest(id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateManifest(req CreateManifestRequest) (*ManifestResponse, error) {
	// Validate vessel exists
	vesselExists, err := s.repo.VesselExists(req.VesselID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate vessel: %w", err)
	}
	if !vesselExists {
		return nil, fmt.Errorf("vessel with ID %d not found", req.VesselID)
	}

	// Validate captain if provided
	if req.CaptainID != nil {
		crew, err := s.repo.GetCrewByID(*req.CaptainID)
		if err != nil {
			return nil, fmt.Errorf("captain with ID %d not found", *req.CaptainID)
		}
		// Check if crew belongs to the vessel
		if crew.VesselID == nil || *crew.VesselID != req.VesselID {
			return nil, fmt.Errorf("crew with ID %d does not belong to vessel with ID %d", *req.CaptainID, req.VesselID)
		}
		// Check if crew has Captain job position
		if !strings.EqualFold(crew.JobPosition, "captain") {
			return nil, fmt.Errorf("crew with ID %d is not a captain (job position: %s)", *req.CaptainID, crew.JobPosition)
		}
	}

	// Generate manifest name
	manifestName := s.generateManifestName()

	manifest := &model.Manifest{
		Name:          manifestName,
		VesselID:      req.VesselID,
		CaptainID:     req.CaptainID,
		GrossTonnage:  req.GrossTonnage,
		Date:          req.Date,
		ArrivalTime:   req.ArrivalTime,
		DepartureTime: req.DepartureTime,
		Status:        "draft", // Always start as draft
	}

	if err := s.repo.Create(manifest); err != nil {
		return nil, err
	}

	// Reload to get vessel and captain info
	manifest, _ = s.repo.GetByID(manifest.ID)

	return s.buildResponse(manifest), nil
}

func (s *service) GetManifestByID(id uint) (*ManifestResponse, error) {
	manifest, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.buildResponse(manifest), nil
}

func (s *service) GetAllManifests() ([]ManifestResponse, error) {
	manifests, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var response []ManifestResponse
	for _, manifest := range manifests {
		response = append(response, *s.buildResponse(&manifest))
	}

	return response, nil
}

func (s *service) GetManifestsByVesselID(vesselID uint) ([]ManifestResponse, error) {
	manifests, err := s.repo.GetByVesselID(vesselID)
	if err != nil {
		return nil, err
	}

	var response []ManifestResponse
	for _, manifest := range manifests {
		response = append(response, *s.buildResponse(&manifest))
	}

	return response, nil
}

func (s *service) GetManifestsByStatus(status string) ([]ManifestResponse, error) {
	manifests, err := s.repo.GetByStatus(status)
	if err != nil {
		return nil, err
	}

	var response []ManifestResponse
	for _, manifest := range manifests {
		response = append(response, *s.buildResponse(&manifest))
	}

	return response, nil
}

// validStatuses contains allowed manifest statuses
var validStatuses = map[string]bool{
	"draft":    true,
	"review":   true,
	"approved": true,
}

func (s *service) UpdateManifest(id uint, req UpdateManifestRequest) (*ManifestResponse, error) {
	manifest, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Validate status if provided
	if req.Status != nil {
		if !validStatuses[*req.Status] {
			return nil, fmt.Errorf("invalid status '%s', must be one of: draft, review, approved", *req.Status)
		}
	}

	// If status is approved, no updates allowed
	if manifest.Status == "approved" {
		return nil, fmt.Errorf("cannot update manifest with status: %s", manifest.Status)
	}

	// If status is review, only status field can be updated
	if manifest.Status == "review" {
		if req.VesselID != nil || req.CaptainID != nil || req.GrossTonnage != nil ||
			req.Date != nil || req.ArrivalTime != nil || req.DepartureTime != nil {
			return nil, fmt.Errorf("only status can be updated when manifest is in review")
		}
		if req.Status != nil {
			manifest.Status = *req.Status
		}
		if err := s.repo.Update(manifest); err != nil {
			return nil, err
		}
		manifest, _ = s.repo.GetByID(manifest.ID)
		return s.buildResponse(manifest), nil
	}

	// Status is draft - allow full updates
	// Determine the target vessel ID (new one if provided, otherwise current)
	targetVesselID := manifest.VesselID
	if req.VesselID != nil {
		targetVesselID = *req.VesselID
	}

	// Validate vessel if being changed
	if req.VesselID != nil && *req.VesselID != manifest.VesselID {
		vesselExists, err := s.repo.VesselExists(*req.VesselID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate vessel: %w", err)
		}
		if !vesselExists {
			return nil, fmt.Errorf("vessel with ID %d not found", *req.VesselID)
		}
		manifest.VesselID = *req.VesselID
	}

	// Validate captain if provided
	if req.CaptainID != nil {
		crew, err := s.repo.GetCrewByID(*req.CaptainID)
		if err != nil {
			return nil, fmt.Errorf("captain with ID %d not found", *req.CaptainID)
		}
		// Check if crew belongs to the vessel
		if crew.VesselID == nil || *crew.VesselID != targetVesselID {
			return nil, fmt.Errorf("crew with ID %d does not belong to vessel with ID %d", *req.CaptainID, targetVesselID)
		}
		// Check if crew has Captain job position
		if !strings.EqualFold(crew.JobPosition, "captain") {
			return nil, fmt.Errorf("crew with ID %d is not a captain (job position: %s)", *req.CaptainID, crew.JobPosition)
		}
		manifest.CaptainID = req.CaptainID
	}

	// Update other fields if provided
	if req.GrossTonnage != nil {
		manifest.GrossTonnage = req.GrossTonnage
	}
	if req.Date != nil {
		manifest.Date = req.Date
	}
	if req.ArrivalTime != nil {
		manifest.ArrivalTime = req.ArrivalTime
	}
	if req.DepartureTime != nil {
		manifest.DepartureTime = req.DepartureTime
	}
	if req.Status != nil {
		manifest.Status = *req.Status
	}

	if err := s.repo.Update(manifest); err != nil {
		return nil, err
	}

	// Reload to get updated vessel and captain info
	manifest, _ = s.repo.GetByID(manifest.ID)

	return s.buildResponse(manifest), nil
}

func (s *service) DeleteManifest(id uint) error {
	manifest, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Only allow deletion if status is draft
	if manifest.Status != "draft" {
		return fmt.Errorf("can only delete draft manifests")
	}

	return s.repo.Delete(id)
}

func (s *service) generateManifestName() string {
	now := time.Now()
	datePrefix := now.Format("20060102")

	// Get count of manifests created today
	count, _ := s.repo.CountByDatePrefix(datePrefix)

	// Generate name: MNF-YYYYMMDD-XXX
	return fmt.Sprintf("MNF-%s-%03d", datePrefix, count+1)
}

func (s *service) buildResponse(manifest *model.Manifest) *ManifestResponse {
	// Convert times to local timezone (Asia/Jakarta)
	var date *string
	var arrivalTime, departureTime *time.Time
	if manifest.Date != nil {
		d := manifest.Date.In(time.Local).Format("2006-01-02")
		date = &d
	}
	if manifest.ArrivalTime != nil {
		t := manifest.ArrivalTime.In(time.Local)
		arrivalTime = &t
	}
	if manifest.DepartureTime != nil {
		t := manifest.DepartureTime.In(time.Local)
		departureTime = &t
	}

	response := &ManifestResponse{
		ID:            manifest.ID,
		Name:          manifest.Name,
		GrossTonnage:  manifest.GrossTonnage,
		Date:          date,
		ArrivalTime:   arrivalTime,
		DepartureTime: departureTime,
		Status:        manifest.Status,
		CreatedAt:     manifest.CreatedAt.In(time.Local),
		UpdatedAt:     manifest.UpdatedAt.In(time.Local),
	}

	if manifest.Vessel.ID > 0 {
		response.Vessel = &VesselInfo{
			ID:   manifest.Vessel.ID,
			Name: manifest.Vessel.Name,
		}
	}

	if manifest.CaptainID != nil && manifest.Captain.ID > 0 {
		response.Captain = &CaptainInfo{
			ID:          manifest.Captain.ID,
			Name:        manifest.Captain.Name,
			Department:  manifest.Captain.Department,
			JobPosition: manifest.Captain.JobPosition,
		}
	}

	return response
}
