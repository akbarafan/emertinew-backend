package classification

import (
	"errors"
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"gorm.io/gorm"
)

type Service interface {
	CreateClassification(req CreateClassificationRequest, userID uint, ipAddress, userAgent string) (*ClassificationResponse, error)
	GetAllClassifications() ([]ClassificationResponse, error)
	GetClassificationByID(id uint) (*ClassificationResponse, error)
	UpdateClassification(id uint, req UpdateClassificationRequest, userID uint, ipAddress, userAgent string) (*ClassificationResponse, error)
	DeleteClassification(id uint, userID uint, ipAddress, userAgent string) error
}

type service struct {
	repo         Repository
	auditService audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{repo: repo, auditService: auditService}
}

func buildAuditData(c *model.MClassification) ClassificationAuditData {
	return ClassificationAuditData{
		ID:   c.ID,
		Name: c.Name,
	}
}

func (s *service) CreateClassification(req CreateClassificationRequest, userID uint, ipAddress, userAgent string) (*ClassificationResponse, error) {
	existingClassification, _ := s.repo.GetByName(req.Name)
	if existingClassification != nil {
		return nil, utils.ErrDuplicateRecord
	}

	classification := &model.MClassification{
		Name: req.Name,
	}

	if s.auditService != nil {
		err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Create(classification); err != nil {
				return err
			}

			auditData := buildAuditData(classification)
			if err := txAudit.Log(userID, "m_classification", classification.ID, "CREATE", nil, auditData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log CREATE for classification %d: %v", classification.ID, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Create(classification); err != nil {
			return nil, err
		}
	}

	return mapToResponse(classification), nil
}

func (s *service) GetAllClassifications() ([]ClassificationResponse, error) {
	classifications, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []ClassificationResponse
	for _, classification := range classifications {
		responses = append(responses, *mapToResponse(&classification))
	}

	return responses, nil
}

func (s *service) GetClassificationByID(id uint) (*ClassificationResponse, error) {
	classification, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}

	return mapToResponse(classification), nil
}

func (s *service) UpdateClassification(id uint, req UpdateClassificationRequest, userID uint, ipAddress, userAgent string) (*ClassificationResponse, error) {
	classification, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}

	existingClassification, _ := s.repo.GetByName(req.Name)
	if existingClassification != nil && existingClassification.ID != id {
		return nil, utils.ErrDuplicateRecord
	}

	oldData := buildAuditData(classification)

	classification.Name = req.Name

	if s.auditService != nil {
		err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Update(classification); err != nil {
				return err
			}

			newData := buildAuditData(classification)
			if err := txAudit.Log(userID, "m_classification", classification.ID, "UPDATE", oldData, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log UPDATE for classification %d: %v", classification.ID, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Update(classification); err != nil {
			return nil, err
		}
	}

	return mapToResponse(classification), nil
}

func (s *service) DeleteClassification(id uint, userID uint, ipAddress, userAgent string) error {
	classification, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrRecordNotFound
		}
		return err
	}

	oldData := buildAuditData(classification)

	if s.auditService != nil {
		return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Delete(classification.ID); err != nil {
				return err
			}

			if err := txAudit.Log(userID, "m_classification", classification.ID, "DELETE", oldData, nil, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log DELETE for classification %d: %v", classification.ID, err)
			}

			return nil
		})
	}

	return s.repo.Delete(classification.ID)
}

func mapToResponse(classification *model.MClassification) *ClassificationResponse {
	return &ClassificationResponse{
		ID:        classification.ID,
		Name:      classification.Name,
		CreatedAt: classification.CreatedAt,
		UpdatedAt: classification.UpdatedAt,
	}
}
