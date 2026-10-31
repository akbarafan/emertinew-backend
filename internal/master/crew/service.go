package crew

import (
	"errors"
	"log"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Service interface {
	CreateCrew(req CreateCrewRequest, userID uint, ipAddress, userAgent string) (*CrewResponse, error)
	GetCrewByID(id uint) (*CrewResponse, error)
	GetAllCrews() ([]CrewResponse, error)
	GetCrewsByVesselID(vesselID uint) ([]CrewResponse, error)
	UpdateCrew(id uint, req UpdateCrewRequest, userID uint, ipAddress, userAgent string) (*CrewResponse, error)
	DeleteCrew(id uint, userID uint, ipAddress, userAgent string) error
}

type service struct {
	repo         Repository
	auditService audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{repo: repo, auditService: auditService}
}

func parseDateString(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	if t, err := time.Parse("2006-01-02", *s); err == nil {
		return &t
	}
	if t, err := time.Parse(time.RFC3339, *s); err == nil {
		return &t
	}
	return nil
}

func formatDateString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	str := t.Format("2006-01-02")
	return &str
}

func (s *service) buildAuditData(crew *model.MCrew) CrewAuditData {
	return CrewAuditData{
		ID:          crew.ID,
		Name:        crew.Name,
		Department:  crew.Department,
		JobPosition: crew.JobPosition,
		VesselID:    crew.VesselID,
	}
}

func (s *service) CreateCrew(req CreateCrewRequest, userID uint, ipAddress, userAgent string) (*CrewResponse, error) {
	// Validate vessel exists if vessel_id is provided
	if req.VesselID != nil {
		exists, err := s.repo.VesselExists(*req.VesselID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.New("vessel not found")
		}
	}

	nat := req.Nationality
	if nat == "" {
		nat = "Indonesian"
	}
	gender := req.Gender
	if gender == "" {
		gender = "M"
	}

	crew := &model.MCrew{
		Name:                 req.Name,
		Department:           req.Department,
		JobPosition:          req.JobPosition,
		VesselID:             req.VesselID,
		Nationality:          nat,
		PassportNumber:       req.PassportNumber,
		PassportExpiryDate:   parseDateString(req.PassportExpiryDate),
		SeamanBookNumber:     req.SeamanBookNumber,
		SeamanBookExpiryDate: parseDateString(req.SeamanBookExpiryDate),
		CocGradeNumber:       req.CocGradeNumber,
		BirthPlace:           req.BirthPlace,
		BirthDate:            parseDateString(req.BirthDate),
		Gender:               gender,
		SignOnDate:           parseDateString(req.SignOnDate),
	}

	err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Create(crew); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			newData := s.buildAuditData(crew)
			if err := txAudit.Log(userID, "m_crew", crew.ID, "CREATE", nil, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log crew create: %v", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Reload to get vessel info
	crew, _ = s.repo.GetByID(crew.ID)

	return s.buildResponse(crew), nil
}

func (s *service) GetCrewByID(id uint) (*CrewResponse, error) {
	crew, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.buildResponse(crew), nil
}

func (s *service) GetAllCrews() ([]CrewResponse, error) {
	crews, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var response []CrewResponse
	for _, crew := range crews {
		response = append(response, *s.buildResponse(&crew))
	}

	return response, nil
}

func (s *service) GetCrewsByVesselID(vesselID uint) ([]CrewResponse, error) {
	crews, err := s.repo.GetByVesselID(vesselID)
	if err != nil {
		return nil, err
	}

	var response []CrewResponse
	for _, crew := range crews {
		response = append(response, *s.buildResponse(&crew))
	}

	return response, nil
}

func (s *service) UpdateCrew(id uint, req UpdateCrewRequest, userID uint, ipAddress, userAgent string) (*CrewResponse, error) {
	crew, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	oldData := s.buildAuditData(crew)

	// Validate vessel exists if vessel_id is provided
	if req.VesselID != nil {
		exists, err := s.repo.VesselExists(*req.VesselID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.New("vessel not found")
		}
	}

	if req.Name != "" {
		crew.Name = req.Name
	}
	if req.Department != "" {
		crew.Department = req.Department
	}
	if req.JobPosition != "" {
		crew.JobPosition = req.JobPosition
	}
	// Allow setting vessel_id
	crew.VesselID = req.VesselID

	if req.Nationality != nil {
		crew.Nationality = *req.Nationality
	}
	if req.PassportNumber != nil {
		crew.PassportNumber = *req.PassportNumber
	}
	if req.PassportExpiryDate != nil {
		crew.PassportExpiryDate = parseDateString(req.PassportExpiryDate)
	}
	if req.SeamanBookNumber != nil {
		crew.SeamanBookNumber = *req.SeamanBookNumber
	}
	if req.SeamanBookExpiryDate != nil {
		crew.SeamanBookExpiryDate = parseDateString(req.SeamanBookExpiryDate)
	}
	if req.CocGradeNumber != nil {
		crew.CocGradeNumber = *req.CocGradeNumber
	}
	if req.BirthPlace != nil {
		crew.BirthPlace = *req.BirthPlace
	}
	if req.BirthDate != nil {
		crew.BirthDate = parseDateString(req.BirthDate)
	}
	if req.Gender != nil {
		crew.Gender = *req.Gender
	}
	if req.SignOnDate != nil {
		crew.SignOnDate = parseDateString(req.SignOnDate)
	}

	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Update(crew); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			newData := s.buildAuditData(crew)
			if err := txAudit.Log(userID, "m_crew", crew.ID, "UPDATE", oldData, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log crew update: %v", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Reload to get vessel info
	crew, _ = s.repo.GetByID(crew.ID)

	return s.buildResponse(crew), nil
}

func (s *service) DeleteCrew(id uint, userID uint, ipAddress, userAgent string) error {
	crew, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	oldData := s.buildAuditData(crew)

	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Delete(id); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			if err := txAudit.Log(userID, "m_crew", id, "DELETE", oldData, nil, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log crew delete: %v", err)
			}
		}

		return nil
	})
	return err
}

func (s *service) buildResponse(crew *model.MCrew) *CrewResponse {
	var daysOnBoard *int
	if crew.SignOnDate != nil {
		diff := int(time.Since(*crew.SignOnDate).Hours() / 24)
		if diff < 0 {
			diff = 0
		}
		daysOnBoard = &diff
	}

	response := &CrewResponse{
		ID:                   crew.ID,
		Name:                 crew.Name,
		Department:           crew.Department,
		JobPosition:          crew.JobPosition,
		Nationality:          crew.Nationality,
		PassportNumber:       crew.PassportNumber,
		PassportExpiryDate:   formatDateString(crew.PassportExpiryDate),
		SeamanBookNumber:     crew.SeamanBookNumber,
		SeamanBookExpiryDate: formatDateString(crew.SeamanBookExpiryDate),
		CocGradeNumber:       crew.CocGradeNumber,
		BirthPlace:           crew.BirthPlace,
		BirthDate:            formatDateString(crew.BirthDate),
		Gender:               crew.Gender,
		SignOnDate:           formatDateString(crew.SignOnDate),
		DaysOnBoard:          daysOnBoard,
		CreatedAt:            crew.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:            crew.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if crew.VesselID != nil && crew.Vessel.ID > 0 {
		response.Vessel = &VesselInfo{
			ID:   crew.Vessel.ID,
			Name: crew.Vessel.Name,
		}
	}

	return response
}
