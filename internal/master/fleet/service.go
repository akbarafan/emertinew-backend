package fleet

import (
	"errors"
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"gorm.io/gorm"
)

type Service interface {
	CreateFleet(req CreateFleetRequest, userID uint, ipAddress, userAgent string) (*FleetResponse, error)
	GetAllFleets(companyID *uint) ([]FleetResponse, error)
	GetFleetByID(id uint) (*FleetResponse, error)
	UpdateFleet(id uint, req UpdateFleetRequest, userID uint, ipAddress, userAgent string) (*FleetResponse, error)
	DeleteFleet(id uint, userID uint, ipAddress, userAgent string) error
}

type service struct {
	repo         Repository
	auditService audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{repo: repo, auditService: auditService}
}

func buildAuditData(fleet *model.MFleet) FleetAuditData {
	return FleetAuditData{
		ID:        fleet.ID,
		CompanyID: fleet.CompanyID,
		Code:      fleet.Code,
		Name:      fleet.Name,
	}
}

func (s *service) CreateFleet(req CreateFleetRequest, userID uint, ipAddress, userAgent string) (*FleetResponse, error) {
	existing, _ := s.repo.GetByCode(req.Code)
	if existing != nil {
		return nil, utils.ErrDuplicateRecord
	}

	fleet := &model.MFleet{
		CompanyID: req.CompanyID,
		Code:      req.Code,
		Name:      req.Name,
	}

	if s.auditService != nil {
		err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Create(fleet); err != nil {
				return err
			}

			auditData := buildAuditData(fleet)
			if err := txAudit.Log(userID, "m_fleets", fleet.ID, "CREATE", nil, auditData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log CREATE for fleet %d: %v", fleet.ID, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Create(fleet); err != nil {
			return nil, err
		}
	}

	return mapToResponse(fleet), nil
}

func (s *service) GetAllFleets(companyID *uint) ([]FleetResponse, error) {
	fleets, err := s.repo.GetAll(companyID)
	if err != nil {
		return nil, err
	}

	responses := make([]FleetResponse, 0)
	for _, f := range fleets {
		responses = append(responses, *mapToResponse(&f))
	}

	return responses, nil
}

func (s *service) GetFleetByID(id uint) (*FleetResponse, error) {
	fleet, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}

	return mapToResponse(fleet), nil
}

func (s *service) UpdateFleet(id uint, req UpdateFleetRequest, userID uint, ipAddress, userAgent string) (*FleetResponse, error) {
	fleet, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}

	existing, _ := s.repo.GetByCode(req.Code)
	if existing != nil && existing.ID != id {
		return nil, utils.ErrDuplicateRecord
	}

	oldData := buildAuditData(fleet)

	fleet.CompanyID = req.CompanyID
	fleet.Code = req.Code
	fleet.Name = req.Name

	if s.auditService != nil {
		err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Update(fleet); err != nil {
				return err
			}

			newData := buildAuditData(fleet)
			if err := txAudit.Log(userID, "m_fleets", fleet.ID, "UPDATE", oldData, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log UPDATE for fleet %d: %v", fleet.ID, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Update(fleet); err != nil {
			return nil, err
		}
	}

	return mapToResponse(fleet), nil
}

func (s *service) DeleteFleet(id uint, userID uint, ipAddress, userAgent string) error {
	fleet, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrRecordNotFound
		}
		return err
	}

	oldData := buildAuditData(fleet)

	if s.auditService != nil {
		return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Delete(fleet.ID); err != nil {
				return err
			}

			if err := txAudit.Log(userID, "m_fleets", fleet.ID, "DELETE", oldData, nil, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log DELETE for fleet %d: %v", fleet.ID, err)
			}

			return nil
		})
	}

	return s.repo.Delete(fleet.ID)
}

func mapToResponse(fleet *model.MFleet) *FleetResponse {
	resp := &FleetResponse{
		ID:        fleet.ID,
		CompanyID: fleet.CompanyID,
		Code:      fleet.Code,
		Name:      fleet.Name,
		CreatedAt: fleet.CreatedAt,
		UpdatedAt: fleet.UpdatedAt,
	}
	if fleet.CompanyID != nil && fleet.Company.ID > 0 {
		companyName := fleet.Company.Name
		resp.Company = &companyName
	}
	return resp
}
