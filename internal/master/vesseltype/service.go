package vesseltype

import (
	"errors"
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"gorm.io/gorm"
)

type Service interface {
	CreateVesselType(req CreateVesselTypeRequest, userID uint, ipAddress, userAgent string) (*VesselTypeResponse, error)
	GetAllVesselTypes() ([]VesselTypeResponse, error)
	GetVesselTypeByID(id uint) (*VesselTypeResponse, error)
	UpdateVesselType(id uint, req UpdateVesselTypeRequest, userID uint, ipAddress, userAgent string) (*VesselTypeResponse, error)
	DeleteVesselType(id uint, userID uint, ipAddress, userAgent string) error
}

type service struct {
	repo         Repository
	auditService audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{repo: repo, auditService: auditService}
}

func buildAuditData(vt *model.MVesselType) VesselTypeAuditData {
	return VesselTypeAuditData{
		ID:   vt.ID,
		Name: vt.Name,
	}
}

func (s *service) CreateVesselType(req CreateVesselTypeRequest, userID uint, ipAddress, userAgent string) (*VesselTypeResponse, error) {
	existingVesselType, _ := s.repo.GetByName(req.Name)
	if existingVesselType != nil {
		return nil, utils.ErrDuplicateRecord
	}

	vesselType := &model.MVesselType{
		Name: req.Name,
	}

	if s.auditService != nil {
		err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Create(vesselType); err != nil {
				return err
			}

			auditData := buildAuditData(vesselType)
			if err := txAudit.Log(userID, "m_vessel_type", vesselType.ID, "CREATE", nil, auditData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log CREATE for vessel_type %d: %v", vesselType.ID, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Create(vesselType); err != nil {
			return nil, err
		}
	}

	return mapToResponse(vesselType), nil
}

func (s *service) GetAllVesselTypes() ([]VesselTypeResponse, error) {
	vesselTypes, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []VesselTypeResponse
	for _, vesselType := range vesselTypes {
		responses = append(responses, *mapToResponse(&vesselType))
	}

	return responses, nil
}

func (s *service) GetVesselTypeByID(id uint) (*VesselTypeResponse, error) {
	vesselType, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}

	return mapToResponse(vesselType), nil
}

func (s *service) UpdateVesselType(id uint, req UpdateVesselTypeRequest, userID uint, ipAddress, userAgent string) (*VesselTypeResponse, error) {
	vesselType, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}

	existingVesselType, _ := s.repo.GetByName(req.Name)
	if existingVesselType != nil && existingVesselType.ID != id {
		return nil, utils.ErrDuplicateRecord
	}

	oldData := buildAuditData(vesselType)

	vesselType.Name = req.Name

	if s.auditService != nil {
		err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Update(vesselType); err != nil {
				return err
			}

			newData := buildAuditData(vesselType)
			if err := txAudit.Log(userID, "m_vessel_type", vesselType.ID, "UPDATE", oldData, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log UPDATE for vessel_type %d: %v", vesselType.ID, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Update(vesselType); err != nil {
			return nil, err
		}
	}

	return mapToResponse(vesselType), nil
}

func (s *service) DeleteVesselType(id uint, userID uint, ipAddress, userAgent string) error {
	vesselType, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrRecordNotFound
		}
		return err
	}

	oldData := buildAuditData(vesselType)

	if s.auditService != nil {
		return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Delete(vesselType.ID); err != nil {
				return err
			}

			if err := txAudit.Log(userID, "m_vessel_type", vesselType.ID, "DELETE", oldData, nil, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log DELETE for vessel_type %d: %v", vesselType.ID, err)
			}

			return nil
		})
	}

	return s.repo.Delete(vesselType.ID)
}

func mapToResponse(vesselType *model.MVesselType) *VesselTypeResponse {
	return &VesselTypeResponse{
		ID:        vesselType.ID,
		Name:      vesselType.Name,
		CreatedAt: vesselType.CreatedAt,
		UpdatedAt: vesselType.UpdatedAt,
	}
}
