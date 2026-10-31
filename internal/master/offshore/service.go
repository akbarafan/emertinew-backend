package offshore

import (
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Service interface {
	CreateOffshore(req CreateOffshoreRequest, userID uint, ipAddress, userAgent string) (*OffshoreResponse, error)
	GetOffshoreByID(id uint) (*OffshoreResponse, error)
	GetAllOffshores(companyID *uint) ([]OffshoreResponse, error)
	UpdateOffshore(id uint, req UpdateOffshoreRequest, userID uint, ipAddress, userAgent string) (*OffshoreResponse, error)
	DeleteOffshore(id uint, userID uint, ipAddress, userAgent string) error
}

type service struct {
	repo         Repository
	auditService audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{repo: repo, auditService: auditService}
}

func (s *service) buildAuditData(offshore *model.MOffshore) OffshoreAuditData {
	return OffshoreAuditData{
		ID:   offshore.ID,
		Name: offshore.Name,
		Code: offshore.Code,
		Type: offshore.Type,
		Long: offshore.Long,
		Lat:  offshore.Lat,
	}
}

func (s *service) CreateOffshore(req CreateOffshoreRequest, userID uint, ipAddress, userAgent string) (*OffshoreResponse, error) {
	offshore := &model.MOffshore{
		CompanyID: req.CompanyID,
		Name:      req.Name,
		Code:      req.Code,
		Type:      req.Type,
		Long:      req.Long,
		Lat:       req.Lat,
	}

	err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Create(offshore); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			newData := s.buildAuditData(offshore)
			if err := txAudit.Log(userID, "m_offshore", offshore.ID, "CREATE", nil, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log offshore create: %v", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.buildResponse(offshore), nil
}

func (s *service) GetOffshoreByID(id uint) (*OffshoreResponse, error) {
	offshore, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.buildResponse(offshore), nil
}

func (s *service) GetAllOffshores(companyID *uint) ([]OffshoreResponse, error) {
	offshores, err := s.repo.GetAll(companyID)
	if err != nil {
		return nil, err
	}

	var response []OffshoreResponse
	for _, offshore := range offshores {
		response = append(response, *s.buildResponse(&offshore))
	}

	return response, nil
}

func (s *service) UpdateOffshore(id uint, req UpdateOffshoreRequest, userID uint, ipAddress, userAgent string) (*OffshoreResponse, error) {
	offshore, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	oldData := s.buildAuditData(offshore)

	if req.CompanyID != nil {
		offshore.CompanyID = req.CompanyID
	}
	if req.Name != "" {
		offshore.Name = req.Name
	}
	if req.Code != "" {
		offshore.Code = req.Code
	}
	if req.Type != "" {
		offshore.Type = req.Type
	}
	if req.Long != 0 {
		offshore.Long = req.Long
	}
	if req.Lat != 0 {
		offshore.Lat = req.Lat
	}

	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Update(offshore); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			newData := s.buildAuditData(offshore)
			if err := txAudit.Log(userID, "m_offshore", offshore.ID, "UPDATE", oldData, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log offshore update: %v", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.buildResponse(offshore), nil
}

func (s *service) DeleteOffshore(id uint, userID uint, ipAddress, userAgent string) error {
	offshore, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	oldData := s.buildAuditData(offshore)

	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Delete(id); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			if err := txAudit.Log(userID, "m_offshore", id, "DELETE", oldData, nil, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log offshore delete: %v", err)
			}
		}

		return nil
	})
	return err
}

func (s *service) buildResponse(offshore *model.MOffshore) *OffshoreResponse {
	return &OffshoreResponse{
		ID:        offshore.ID,
		CompanyID: offshore.CompanyID,
		Name:      offshore.Name,
		Code:      offshore.Code,
		Type:      offshore.Type,
		Long:      offshore.Long,
		Lat:       offshore.Lat,
		CreatedAt: offshore.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: offshore.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
