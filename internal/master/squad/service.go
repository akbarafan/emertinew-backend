package squad

import (
	"errors"
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"gorm.io/gorm"
)

type Service interface {
	CreateSquad(req CreateSquadRequest, userID uint, ipAddress, userAgent string) (*SquadResponse, error)
	GetAllSquads(companyID *uint) ([]SquadResponse, error)
	GetSquadByID(id uint) (*SquadResponse, error)
	GetSquadsByFleetID(fleetID uint) ([]SquadResponse, error)
	UpdateSquad(id uint, req UpdateSquadRequest, userID uint, ipAddress, userAgent string) (*SquadResponse, error)
	DeleteSquad(id uint, userID uint, ipAddress, userAgent string) error
}

type service struct {
	repo         Repository
	auditService audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{repo: repo, auditService: auditService}
}

func buildAuditData(squad *model.MSquad) SquadAuditData {
	return SquadAuditData{
		ID:        squad.ID,
		CompanyID: squad.CompanyID,
		FleetID:   squad.FleetID,
		Code:      squad.Code,
		Name:      squad.Name,
	}
}

func (s *service) CreateSquad(req CreateSquadRequest, userID uint, ipAddress, userAgent string) (*SquadResponse, error) {
	existing, _ := s.repo.GetByCode(req.Code)
	if existing != nil {
		return nil, utils.ErrDuplicateRecord
	}

	squad := &model.MSquad{
		CompanyID: req.CompanyID,
		FleetID:   req.FleetID,
		Code:      req.Code,
		Name:      req.Name,
	}

	if s.auditService != nil {
		err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Create(squad); err != nil {
				return err
			}

			auditData := buildAuditData(squad)
			if err := txAudit.Log(userID, "m_squads", squad.ID, "CREATE", nil, auditData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log CREATE for squad %d: %v", squad.ID, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Create(squad); err != nil {
			return nil, err
		}
	}

	return mapToResponse(squad), nil
}

func (s *service) GetAllSquads(companyID *uint) ([]SquadResponse, error) {
	squads, err := s.repo.GetAll(companyID)
	if err != nil {
		return nil, err
	}

	responses := make([]SquadResponse, 0)
	for _, sq := range squads {
		responses = append(responses, *mapToResponse(&sq))
	}

	return responses, nil
}

func (s *service) GetSquadByID(id uint) (*SquadResponse, error) {
	squad, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}

	return mapToResponse(squad), nil
}

func (s *service) GetSquadsByFleetID(fleetID uint) ([]SquadResponse, error) {
	squads, err := s.repo.GetByFleetID(fleetID)
	if err != nil {
		return nil, err
	}

	responses := make([]SquadResponse, 0)
	for _, sq := range squads {
		responses = append(responses, *mapToResponse(&sq))
	}

	return responses, nil
}

func (s *service) UpdateSquad(id uint, req UpdateSquadRequest, userID uint, ipAddress, userAgent string) (*SquadResponse, error) {
	squad, err := s.repo.GetByID(id)
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

	oldData := buildAuditData(squad)

	squad.CompanyID = req.CompanyID
	squad.FleetID = req.FleetID
	squad.Code = req.Code
	squad.Name = req.Name

	if s.auditService != nil {
		err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Update(squad); err != nil {
				return err
			}

			newData := buildAuditData(squad)
			if err := txAudit.Log(userID, "m_squads", squad.ID, "UPDATE", oldData, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log UPDATE for squad %d: %v", squad.ID, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Update(squad); err != nil {
			return nil, err
		}
	}

	return mapToResponse(squad), nil
}

func (s *service) DeleteSquad(id uint, userID uint, ipAddress, userAgent string) error {
	squad, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrRecordNotFound
		}
		return err
	}

	oldData := buildAuditData(squad)

	if s.auditService != nil {
		return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Delete(squad.ID); err != nil {
				return err
			}

			if err := txAudit.Log(userID, "m_squads", squad.ID, "DELETE", oldData, nil, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log DELETE for squad %d: %v", squad.ID, err)
			}

			return nil
		})
	}

	return s.repo.Delete(squad.ID)
}

func mapToResponse(squad *model.MSquad) *SquadResponse {
	resp := &SquadResponse{
		ID:        squad.ID,
		CompanyID: squad.CompanyID,
		FleetID:   squad.FleetID,
		Code:      squad.Code,
		Name:      squad.Name,
		CreatedAt: squad.CreatedAt,
		UpdatedAt: squad.UpdatedAt,
	}
	if squad.CompanyID != nil && squad.Company.ID > 0 {
		companyName := squad.Company.Name
		resp.Company = &companyName
	}
	if squad.FleetID != nil && squad.Fleet.ID > 0 {
		fleetName := squad.Fleet.Name
		resp.Fleet = &fleetName
	}
	return resp
}
