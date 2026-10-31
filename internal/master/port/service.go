package port

import (
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Service interface {
	CreatePort(req CreatePortRequest, userID uint, ipAddress, userAgent string) (*PortResponse, error)
	GetPortByID(id uint) (*PortResponse, error)
	GetAllPorts() ([]PortResponse, error)
	UpdatePort(id uint, req UpdatePortRequest, userID uint, ipAddress, userAgent string) (*PortResponse, error)
	DeletePort(id uint, userID uint, ipAddress, userAgent string) error
}

type service struct {
	repo         Repository
	auditService audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{repo: repo, auditService: auditService}
}

func (s *service) buildAuditData(port *model.MPort) PortAuditData {
	return PortAuditData{
		ID:   port.ID,
		Name: port.Name,
		Code: port.Code,
		Long: port.Long,
		Lat:  port.Lat,
	}
}

func (s *service) CreatePort(req CreatePortRequest, userID uint, ipAddress, userAgent string) (*PortResponse, error) {
	port := &model.MPort{
		Name: req.Name,
		Code: req.Code,
		Long: req.Long,
		Lat:  req.Lat,
	}

	err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Create(port); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			newData := s.buildAuditData(port)
			if err := txAudit.Log(userID, "m_port", port.ID, "CREATE", nil, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log port create: %v", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.buildResponse(port), nil
}

func (s *service) GetPortByID(id uint) (*PortResponse, error) {
	port, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.buildResponse(port), nil
}

func (s *service) GetAllPorts() ([]PortResponse, error) {
	ports, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var response []PortResponse
	for _, port := range ports {
		response = append(response, *s.buildResponse(&port))
	}

	return response, nil
}

func (s *service) UpdatePort(id uint, req UpdatePortRequest, userID uint, ipAddress, userAgent string) (*PortResponse, error) {
	port, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	oldData := s.buildAuditData(port)

	if req.Name != "" {
		port.Name = req.Name
	}
	if req.Code != "" {
		port.Code = req.Code
	}
	if req.Long != 0 {
		port.Long = req.Long
	}
	if req.Lat != 0 {
		port.Lat = req.Lat
	}

	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Update(port); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			newData := s.buildAuditData(port)
			if err := txAudit.Log(userID, "m_port", port.ID, "UPDATE", oldData, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log port update: %v", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.buildResponse(port), nil
}

func (s *service) DeletePort(id uint, userID uint, ipAddress, userAgent string) error {
	port, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	oldData := s.buildAuditData(port)

	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Delete(id); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			if err := txAudit.Log(userID, "m_port", id, "DELETE", oldData, nil, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log port delete: %v", err)
			}
		}

		return nil
	})
	return err
}

func (s *service) buildResponse(port *model.MPort) *PortResponse {
	return &PortResponse{
		ID:        port.ID,
		Name:      port.Name,
		Code:      port.Code,
		Long:      port.Long,
		Lat:       port.Lat,
		CreatedAt: port.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: port.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
