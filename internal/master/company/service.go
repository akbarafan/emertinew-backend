package company

import (
	"errors"
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"

	"gorm.io/gorm"
)

type Service interface {
	CreateCompany(req CreateCompanyRequest, userID uint, ipAddress, userAgent string) (*CompanyResponse, error)
	GetAllCompanies() ([]CompanyResponse, error)
	GetCompanyByID(id uint) (*CompanyResponse, error)
	GetCompanyByCode(code string) (*CompanyResponse, error)
	UpdateCompany(id uint, req UpdateCompanyRequest, userID uint, ipAddress, userAgent string) (*CompanyResponse, error)
	DeleteCompany(id uint, userID uint, ipAddress, userAgent string) error
}

type service struct {
	repo         Repository
	auditService audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{repo: repo, auditService: auditService}
}

func buildAuditData(company *model.MCompany) CompanyAuditData {
	return CompanyAuditData{
		ID:   company.ID,
		Code: company.Code,
		Name: company.Name,
	}
}

func (s *service) CreateCompany(req CreateCompanyRequest, userID uint, ipAddress, userAgent string) (*CompanyResponse, error) {
	existingCompany, _ := s.repo.GetByCode(req.Code)
	if existingCompany != nil {
		return nil, utils.ErrDuplicateRecord
	}

	company := &model.MCompany{
		Code: req.Code,
		Name: req.Name,
	}

	if s.auditService != nil {
		err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Create(company); err != nil {
				return err
			}

			auditData := buildAuditData(company)
			if err := txAudit.Log(userID, "m_company", company.ID, "CREATE", nil, auditData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log CREATE for company %d: %v", company.ID, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Create(company); err != nil {
			return nil, err
		}
	}

	return mapToResponse(company), nil
}

func (s *service) GetAllCompanies() ([]CompanyResponse, error) {
	companies, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var responses []CompanyResponse
	for _, company := range companies {
		responses = append(responses, *mapToResponse(&company))
	}

	return responses, nil
}

func (s *service) GetCompanyByID(id uint) (*CompanyResponse, error) {
	company, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}

	return mapToResponse(company), nil
}

func (s *service) GetCompanyByCode(code string) (*CompanyResponse, error) {
	company, err := s.repo.GetByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}

	return mapToResponse(company), nil
}

func (s *service) UpdateCompany(id uint, req UpdateCompanyRequest, userID uint, ipAddress, userAgent string) (*CompanyResponse, error) {
	company, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}

	existingCompany, _ := s.repo.GetByCode(req.Code)
	if existingCompany != nil && existingCompany.ID != id {
		return nil, utils.ErrDuplicateRecord
	}

	oldData := buildAuditData(company)

	company.Code = req.Code
	company.Name = req.Name

	if s.auditService != nil {
		err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Update(company); err != nil {
				return err
			}

			newData := buildAuditData(company)
			if err := txAudit.Log(userID, "m_company", company.ID, "UPDATE", oldData, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log UPDATE for company %d: %v", company.ID, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Update(company); err != nil {
			return nil, err
		}
	}

	return mapToResponse(company), nil
}

func (s *service) DeleteCompany(id uint, userID uint, ipAddress, userAgent string) error {
	company, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrRecordNotFound
		}
		return err
	}

	oldData := buildAuditData(company)

	if s.auditService != nil {
		return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
			txRepo := s.repo.WithTx(tx)
			txAudit := s.auditService.WithTx(tx)

			if err := txRepo.Delete(company.ID); err != nil {
				return err
			}

			if err := txAudit.Log(userID, "m_company", company.ID, "DELETE", oldData, nil, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log DELETE for company %d: %v", company.ID, err)
			}

			return nil
		})
	}

	return s.repo.Delete(company.ID)
}

func mapToResponse(company *model.MCompany) *CompanyResponse {
	return &CompanyResponse{
		ID:        company.ID,
		Code:      company.Code,
		Name:      company.Name,
		CreatedAt: company.CreatedAt,
		UpdatedAt: company.UpdatedAt,
	}
}
