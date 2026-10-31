package fueltype

import (
	"errors"
	"log"
	"strings"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Service interface {
	CreateFuelType(req CreateFuelTypeRequest, userID uint, ipAddress, userAgent string) (*FuelTypeResponse, error)
	GetFuelTypeByID(id uint) (*FuelTypeResponse, error)
	GetAllFuelTypes() ([]FuelTypeResponse, error)
	UpdateFuelType(id uint, req UpdateFuelTypeRequest, userID uint, ipAddress, userAgent string) (*FuelTypeResponse, error)
	DeleteFuelType(id uint, userID uint, ipAddress, userAgent string) error
}

type service struct {
	repo         Repository
	auditService audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{repo: repo, auditService: auditService}
}

// validate checks the parameters that have a physical meaning.
//
// The carbon factor and the heating value are per kilogram and per MJ, so they
// must not be negative. The density is a mass per volume, and a density of
// zero would silently turn every emission figure into zero, so it is rejected
// when it is given at all.
func validate(cfCO2, cfCH4, cfN2O, slip, wtt, lhv, density *float64) error {
	for name, value := range map[string]*float64{
		"cf_co2_kg_per_kg":            cfCO2,
		"cf_ch4_kg_per_kg":            cfCH4,
		"cf_n2o_kg_per_kg":            cfN2O,
		"wtt_intensity_gco2eq_per_mj": wtt,
		"lhv_mj_per_kg":               lhv,
	} {
		if value != nil && *value < 0 {
			return errors.New(name + " must not be negative")
		}
	}
	if slip != nil && (*slip < 0 || *slip > 100) {
		return errors.New("methane_slip_percent must be between 0 and 100")
	}
	if density != nil && *density <= 0 {
		return errors.New("density_kg_per_litre must be greater than 0")
	}
	return nil
}

func (s *service) buildAuditData(fuelType *model.MFuelType) FuelTypeAuditData {
	return FuelTypeAuditData{
		ID:                      fuelType.ID,
		Code:                    fuelType.Code,
		Name:                    fuelType.Name,
		CfCO2KgPerKg:            fuelType.CfCO2KgPerKg,
		CfCH4KgPerKg:            fuelType.CfCH4KgPerKg,
		CfN2OKgPerKg:            fuelType.CfN2OKgPerKg,
		MethaneSlipPc:           fuelType.MethaneSlipPc,
		WttIntensityGCO2eqPerMJ: fuelType.WttIntensityGCO2eqPerMJ,
		LhvMJPerKg:              fuelType.LhvMJPerKg,
		DensityKgPerLitre:       fuelType.DensityKgPerLitre,
		Reference:               fuelType.Reference,
		Method:                  fuelType.Method,
		IsActive:                fuelType.IsActive,
	}
}

func (s *service) CreateFuelType(req CreateFuelTypeRequest, userID uint, ipAddress, userAgent string) (*FuelTypeResponse, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		return nil, errors.New("code is required")
	}
	if err := validate(req.CfCO2KgPerKg, req.CfCH4KgPerKg, req.CfN2OKgPerKg, req.MethaneSlipPc, req.WttIntensityGCO2eqPerMJ, req.LhvMJPerKg, req.DensityKgPerLitre); err != nil {
		return nil, err
	}

	fuelType := &model.MFuelType{
		Code:                    code,
		Name:                    req.Name,
		CfCO2KgPerKg:            req.CfCO2KgPerKg,
		CfCH4KgPerKg:            req.CfCH4KgPerKg,
		CfN2OKgPerKg:            req.CfN2OKgPerKg,
		MethaneSlipPc:           req.MethaneSlipPc,
		WttIntensityGCO2eqPerMJ: req.WttIntensityGCO2eqPerMJ,
		LhvMJPerKg:              req.LhvMJPerKg,
		DensityKgPerLitre:       req.DensityKgPerLitre,
		Reference:               req.Reference,
		Method:                  req.Method,
		IsActive:                true,
	}

	err := s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Create(fuelType); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			newData := s.buildAuditData(fuelType)
			if err := txAudit.Log(userID, "m_fuel_type", fuelType.ID, "CREATE", nil, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log fuel type create: %v", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.buildResponse(fuelType), nil
}

func (s *service) GetFuelTypeByID(id uint) (*FuelTypeResponse, error) {
	fuelType, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.buildResponse(fuelType), nil
}

func (s *service) GetAllFuelTypes() ([]FuelTypeResponse, error) {
	fuelTypes, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	response := make([]FuelTypeResponse, 0, len(fuelTypes))
	for _, fuelType := range fuelTypes {
		response = append(response, *s.buildResponse(&fuelType))
	}
	return response, nil
}

func (s *service) UpdateFuelType(id uint, req UpdateFuelTypeRequest, userID uint, ipAddress, userAgent string) (*FuelTypeResponse, error) {
	fuelType, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	oldData := s.buildAuditData(fuelType)

	if req.Name != "" {
		fuelType.Name = req.Name
	}
	if req.CfCO2KgPerKg != nil {
		fuelType.CfCO2KgPerKg = req.CfCO2KgPerKg
	}
	if req.CfCH4KgPerKg != nil {
		fuelType.CfCH4KgPerKg = req.CfCH4KgPerKg
	}
	if req.CfN2OKgPerKg != nil {
		fuelType.CfN2OKgPerKg = req.CfN2OKgPerKg
	}
	if req.MethaneSlipPc != nil {
		fuelType.MethaneSlipPc = req.MethaneSlipPc
	}
	if req.WttIntensityGCO2eqPerMJ != nil {
		fuelType.WttIntensityGCO2eqPerMJ = req.WttIntensityGCO2eqPerMJ
	}
	if req.LhvMJPerKg != nil {
		fuelType.LhvMJPerKg = req.LhvMJPerKg
	}
	if req.DensityKgPerLitre != nil {
		fuelType.DensityKgPerLitre = req.DensityKgPerLitre
	}
	if req.Reference != "" {
		fuelType.Reference = req.Reference
	}
	if req.Method != "" {
		fuelType.Method = req.Method
	}
	if req.IsActive != nil {
		fuelType.IsActive = *req.IsActive
	}

	if err := validate(fuelType.CfCO2KgPerKg, fuelType.CfCH4KgPerKg, fuelType.CfN2OKgPerKg, fuelType.MethaneSlipPc, fuelType.WttIntensityGCO2eqPerMJ, fuelType.LhvMJPerKg, fuelType.DensityKgPerLitre); err != nil {
		return nil, err
	}

	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Update(fuelType); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			newData := s.buildAuditData(fuelType)
			if err := txAudit.Log(userID, "m_fuel_type", fuelType.ID, "UPDATE", oldData, newData, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log fuel type update: %v", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.buildResponse(fuelType), nil
}

func (s *service) DeleteFuelType(id uint, userID uint, ipAddress, userAgent string) error {
	fuelType, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	oldData := s.buildAuditData(fuelType)

	err = s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Delete(id); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			if err := txAudit.Log(userID, "m_fuel_type", id, "DELETE", oldData, nil, ipAddress, userAgent); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log fuel type delete: %v", err)
			}
		}

		return nil
	})
	return err
}

// buildResponse maps the model to the response, including the derived
// per-litre factors. A factor that cannot be derived stays null, which the UI
// reports as "not available" instead of a zero.
func (s *service) buildResponse(fuelType *model.MFuelType) *FuelTypeResponse {
	response := &FuelTypeResponse{
		ID:                      fuelType.ID,
		Code:                    fuelType.Code,
		Name:                    fuelType.Name,
		CfCO2KgPerKg:            fuelType.CfCO2KgPerKg,
		CfCH4KgPerKg:            fuelType.CfCH4KgPerKg,
		CfN2OKgPerKg:            fuelType.CfN2OKgPerKg,
		MethaneSlipPc:           fuelType.MethaneSlipPc,
		WttIntensityGCO2eqPerMJ: fuelType.WttIntensityGCO2eqPerMJ,
		LhvMJPerKg:              fuelType.LhvMJPerKg,
		DensityKgPerLitre:       fuelType.DensityKgPerLitre,
		Reference:               fuelType.Reference,
		Method:                  fuelType.Method,
		IsActive:                fuelType.IsActive,
		CreatedAt:               fuelType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:               fuelType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if value, ok := fuelType.TtWCO2KgPerLitre(); ok {
		response.TtWCO2KgPerLitre = &value
	}
	if value, ok := fuelType.TtWCO2eKgPerLitre(); ok {
		response.TtWCO2eKgPerLitre = &value
	}
	if value, ok := fuelType.WtWCO2eKgPerLitre(); ok {
		response.WtWCO2eKgPerLitre = &value
	}

	return response
}
