package operationstatus

import (
	"errors"
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"gorm.io/gorm"
)

type Service interface {
	CreateOperationStatus(req CreateOperationStatusRequest, userID uint, ipAddress, userAgent string) (*OperationStatusResponse, error)
	GetAllOperationStatuses() ([]OperationStatusResponse, error)
	GetOperationStatusByID(id uint) (*OperationStatusResponse, error)
	UpdateOperationStatus(id uint, req UpdateOperationStatusRequest, userID uint, ipAddress, userAgent string) (*OperationStatusResponse, error)
	DeleteOperationStatus(id uint, userID uint, ipAddress, userAgent string) error
}

type service struct {
	repo         Repository
	auditService audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{repo: repo, auditService: auditService}
}

func buildAuditData(os *model.MOperationStatus) OperationStatusAuditData {
	return OperationStatusAuditData{
		ID:   os.ID,
		Name: os.Name,
	}
}

func (s *service) CreateOperationStatus(req CreateOperationStatusRequest, userID uint, ipAddress, userAgent string) (*OperationStatusResponse, error) {
	existingOperationStatus, _ := s.repo.GetByName(req.Name)
	if existingOperationStatus != nil {
		return nil, utils.ErrDuplicateRecord
	}

	operationStatus := &model.MOperationStatus{
		Name: req.Name,
	}

	if s.auditService != nil {
		err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Create(operationStatus); err != nil {
				return err
			}

			auditData := buildAuditData(operationStatus)
			if err := txAudit.Log(userID, "m_operation_status", operationStatus.ID, "CREATE", nil, auditData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log CREATE for operation_status %d: %v", operationStatus.ID, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Create(operationStatus); err != nil {
			return nil, err
		}
	}

	return mapToResponse(operationStatus), nil
}

func (s *service) GetAllOperationStatuses() ([]OperationStatusResponse, error) {
	operationStatuses, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []OperationStatusResponse
	for _, operationStatus := range operationStatuses {
		responses = append(responses, *mapToResponse(&operationStatus))
	}

	return responses, nil
}

func (s *service) GetOperationStatusByID(id uint) (*OperationStatusResponse, error) {
	operationStatus, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}

	return mapToResponse(operationStatus), nil
}

func (s *service) UpdateOperationStatus(id uint, req UpdateOperationStatusRequest, userID uint, ipAddress, userAgent string) (*OperationStatusResponse, error) {
	operationStatus, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}

	existingOperationStatus, _ := s.repo.GetByName(req.Name)
	if existingOperationStatus != nil && existingOperationStatus.ID != id {
		return nil, utils.ErrDuplicateRecord
	}

	oldData := buildAuditData(operationStatus)

	operationStatus.Name = req.Name

	if s.auditService != nil {
		err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Update(operationStatus); err != nil {
				return err
			}

			newData := buildAuditData(operationStatus)
			if err := txAudit.Log(userID, "m_operation_status", operationStatus.ID, "UPDATE", oldData, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log UPDATE for operation_status %d: %v", operationStatus.ID, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Update(operationStatus); err != nil {
			return nil, err
		}
	}

	return mapToResponse(operationStatus), nil
}

func (s *service) DeleteOperationStatus(id uint, userID uint, ipAddress, userAgent string) error {
	operationStatus, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrRecordNotFound
		}
		return err
	}

	oldData := buildAuditData(operationStatus)

	if s.auditService != nil {
		return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Delete(operationStatus.ID); err != nil {
				return err
			}

			if err := txAudit.Log(userID, "m_operation_status", operationStatus.ID, "DELETE", oldData, nil, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log DELETE for operation_status %d: %v", operationStatus.ID, err)
			}

			return nil
		})
	}

	return s.repo.Delete(operationStatus.ID)
}

func mapToResponse(operationStatus *model.MOperationStatus) *OperationStatusResponse {
	return &OperationStatusResponse{
		ID:        operationStatus.ID,
		Name:      operationStatus.Name,
		CreatedAt: operationStatus.CreatedAt,
		UpdatedAt: operationStatus.UpdatedAt,
	}
}
