package materialhull

import (
	"errors"
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"gorm.io/gorm"
)

type Service interface {
	CreateMaterialHull(req CreateMaterialHullRequest, userID uint, ipAddress, userAgent string) (*MaterialHullResponse, error)
	GetAllMaterialHulls() ([]MaterialHullResponse, error)
	GetMaterialHullByID(id uint) (*MaterialHullResponse, error)
	UpdateMaterialHull(id uint, req UpdateMaterialHullRequest, userID uint, ipAddress, userAgent string) (*MaterialHullResponse, error)
	DeleteMaterialHull(id uint, userID uint, ipAddress, userAgent string) error
}

type service struct {
	repo         Repository
	auditService audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{repo: repo, auditService: auditService}
}

func buildAuditData(mh *model.MMaterialHull) MaterialHullAuditData {
	return MaterialHullAuditData{
		ID:   mh.ID,
		Name: mh.Name,
	}
}

func (s *service) CreateMaterialHull(req CreateMaterialHullRequest, userID uint, ipAddress, userAgent string) (*MaterialHullResponse, error) {
	existingMaterialHull, _ := s.repo.GetByName(req.Name)
	if existingMaterialHull != nil {
		return nil, utils.ErrDuplicateRecord
	}

	materialHull := &model.MMaterialHull{
		Name: req.Name,
	}

	if s.auditService != nil {
		err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Create(materialHull); err != nil {
				return err
			}

			auditData := buildAuditData(materialHull)
			if err := txAudit.Log(userID, "m_material_hull", materialHull.ID, "CREATE", nil, auditData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log CREATE for material_hull %d: %v", materialHull.ID, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Create(materialHull); err != nil {
			return nil, err
		}
	}

	return mapToResponse(materialHull), nil
}

func (s *service) GetAllMaterialHulls() ([]MaterialHullResponse, error) {
	materialHulls, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []MaterialHullResponse
	for _, materialHull := range materialHulls {
		responses = append(responses, *mapToResponse(&materialHull))
	}

	return responses, nil
}

func (s *service) GetMaterialHullByID(id uint) (*MaterialHullResponse, error) {
	materialHull, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}

	return mapToResponse(materialHull), nil
}

func (s *service) UpdateMaterialHull(id uint, req UpdateMaterialHullRequest, userID uint, ipAddress, userAgent string) (*MaterialHullResponse, error) {
	materialHull, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}

	existingMaterialHull, _ := s.repo.GetByName(req.Name)
	if existingMaterialHull != nil && existingMaterialHull.ID != id {
		return nil, utils.ErrDuplicateRecord
	}

	oldData := buildAuditData(materialHull)

	materialHull.Name = req.Name

	if s.auditService != nil {
		err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Update(materialHull); err != nil {
				return err
			}

			newData := buildAuditData(materialHull)
			if err := txAudit.Log(userID, "m_material_hull", materialHull.ID, "UPDATE", oldData, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log UPDATE for material_hull %d: %v", materialHull.ID, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Update(materialHull); err != nil {
			return nil, err
		}
	}

	return mapToResponse(materialHull), nil
}

func (s *service) DeleteMaterialHull(id uint, userID uint, ipAddress, userAgent string) error {
	materialHull, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrRecordNotFound
		}
		return err
	}

	oldData := buildAuditData(materialHull)

	if s.auditService != nil {
		return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Delete(materialHull.ID); err != nil {
				return err
			}

			if err := txAudit.Log(userID, "m_material_hull", materialHull.ID, "DELETE", oldData, nil, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log DELETE for material_hull %d: %v", materialHull.ID, err)
			}

			return nil
		})
	}

	return s.repo.Delete(materialHull.ID)
}

func mapToResponse(materialHull *model.MMaterialHull) *MaterialHullResponse {
	return &MaterialHullResponse{
		ID:        materialHull.ID,
		Name:      materialHull.Name,
		CreatedAt: materialHull.CreatedAt,
		UpdatedAt: materialHull.UpdatedAt,
	}
}
