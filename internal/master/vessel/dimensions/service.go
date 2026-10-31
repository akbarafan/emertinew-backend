package vesseldimensions

import (
	"errors"
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Service interface {
	CreateVesselDimensions(vesselID uint, req CreateVesselDimensionsRequest, userID uint, ip, ua string) (*VesselDimensionsResponse, error)
	GetVesselDimensionsByID(id uint) (*VesselDimensionsResponse, error)
	GetVesselDimensionsByVesselID(vesselID uint) (*VesselDimensionsResponse, error)
	UpdateVesselDimensions(id uint, req UpdateVesselDimensionsRequest, userID uint, ip, ua string) (*VesselDimensionsResponse, error)
	UpdateVesselDimensionsByVesselID(vesselID uint, req UpdateVesselDimensionsRequest, userID uint, ip, ua string) (*VesselDimensionsResponse, error)
	DeleteVesselDimensions(id uint, userID uint, ip, ua string) error
	DeleteVesselDimensionsByVesselID(vesselID uint, userID uint, ip, ua string) error
}

type service struct {
	repo         Repository
	auditService audit.Service
}

func NewService(repo Repository, auditService audit.Service) Service {
	return &service{
		repo:         repo,
		auditService: auditService,
	}
}

func (s *service) CreateVesselDimensions(vesselID uint, req CreateVesselDimensionsRequest, userID uint, ip, ua string) (*VesselDimensionsResponse, error) {
	// Validate vessel exists
	exists, err := s.repo.VesselExists(vesselID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("vessel not found")
	}

	// Check if dimensions already exist for this vessel
	existing, _ := s.repo.GetByVesselID(vesselID)
	if existing != nil {
		return nil, errors.New("vessel dimensions already exist for this vessel")
	}

	dimensions := &model.VesselDimensions{
		VesselID:     vesselID,
		LOA:          req.LOA,
		LBP:          req.LBP,
		Beam:         req.Beam,
		Depth:        req.Depth,
		Draft:        req.Draft,
		AirDraft:     req.AirDraft,
		GT:           req.GT,
		NT:           req.NT,
		DWT:          req.DWT,
		Displacement: req.Displacement,
	}

	db := s.repo.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Create(dimensions); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			auditData := s.buildAuditData(dimensions)
			if err := txAudit.Log(userID, "vessel_dimensions", dimensions.ID, "CREATE", nil, auditData, ip, ua); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log CREATE for vessel_dimensions ID %d: %v", dimensions.ID, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.mapToResponse(dimensions), nil
}

func (s *service) GetVesselDimensionsByID(id uint) (*VesselDimensionsResponse, error) {
	dimensions, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("vessel dimensions not found")
	}

	return s.mapToResponse(dimensions), nil
}

func (s *service) GetVesselDimensionsByVesselID(vesselID uint) (*VesselDimensionsResponse, error) {
	dimensions, err := s.repo.GetByVesselID(vesselID)
	if err != nil {
		return nil, errors.New("vessel dimensions not found")
	}

	return s.mapToResponse(dimensions), nil
}

func (s *service) UpdateVesselDimensions(id uint, req UpdateVesselDimensionsRequest, userID uint, ip, ua string) (*VesselDimensionsResponse, error) {
	dimensions, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("vessel dimensions not found")
	}

	oldData := *dimensions

	// Update only non-nil fields
	if req.LOA != nil {
		dimensions.LOA = req.LOA
	}
	if req.LBP != nil {
		dimensions.LBP = req.LBP
	}
	if req.Beam != nil {
		dimensions.Beam = req.Beam
	}
	if req.Depth != nil {
		dimensions.Depth = req.Depth
	}
	if req.Draft != nil {
		dimensions.Draft = req.Draft
	}
	if req.AirDraft != nil {
		dimensions.AirDraft = req.AirDraft
	}
	if req.GT != nil {
		dimensions.GT = req.GT
	}
	if req.NT != nil {
		dimensions.NT = req.NT
	}
	if req.DWT != nil {
		dimensions.DWT = req.DWT
	}
	if req.Displacement != nil {
		dimensions.Displacement = req.Displacement
	}

	db := s.repo.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Update(dimensions); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			oldAudit := s.buildAuditData(&oldData)
			newAudit := s.buildAuditData(dimensions)
			if err := txAudit.Log(userID, "vessel_dimensions", id, "UPDATE", oldAudit, newAudit, ip, ua); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log UPDATE for vessel_dimensions ID %d: %v", id, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.mapToResponse(dimensions), nil
}

func (s *service) UpdateVesselDimensionsByVesselID(vesselID uint, req UpdateVesselDimensionsRequest, userID uint, ip, ua string) (*VesselDimensionsResponse, error) {
	dimensions, err := s.repo.GetByVesselID(vesselID)
	if err != nil {
		return nil, errors.New("vessel dimensions not found")
	}

	oldData := *dimensions

	// Update only non-nil fields
	if req.LOA != nil {
		dimensions.LOA = req.LOA
	}
	if req.LBP != nil {
		dimensions.LBP = req.LBP
	}
	if req.Beam != nil {
		dimensions.Beam = req.Beam
	}
	if req.Depth != nil {
		dimensions.Depth = req.Depth
	}
	if req.Draft != nil {
		dimensions.Draft = req.Draft
	}
	if req.AirDraft != nil {
		dimensions.AirDraft = req.AirDraft
	}
	if req.GT != nil {
		dimensions.GT = req.GT
	}
	if req.NT != nil {
		dimensions.NT = req.NT
	}
	if req.DWT != nil {
		dimensions.DWT = req.DWT
	}
	if req.Displacement != nil {
		dimensions.Displacement = req.Displacement
	}

	db := s.repo.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Update(dimensions); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			oldAudit := s.buildAuditData(&oldData)
			newAudit := s.buildAuditData(dimensions)
			if err := txAudit.Log(userID, "vessel_dimensions", dimensions.ID, "UPDATE", oldAudit, newAudit, ip, ua); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log UPDATE for vessel_dimensions ID %d: %v", dimensions.ID, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.mapToResponse(dimensions), nil
}

func (s *service) DeleteVesselDimensions(id uint, userID uint, ip, ua string) error {
	dimensions, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("vessel dimensions not found")
	}

	oldData := *dimensions

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Delete(id); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			oldAudit := s.buildAuditData(&oldData)
			if err := txAudit.Log(userID, "vessel_dimensions", id, "DELETE", oldAudit, nil, ip, ua); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log DELETE for vessel_dimensions ID %d: %v", id, err)
			}
		}

		return nil
	})
}

func (s *service) DeleteVesselDimensionsByVesselID(vesselID uint, userID uint, ip, ua string) error {
	dimensions, err := s.repo.GetByVesselID(vesselID)
	if err != nil {
		return errors.New("vessel dimensions not found")
	}

	oldData := *dimensions

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Delete(dimensions.ID); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			oldAudit := s.buildAuditData(&oldData)
			if err := txAudit.Log(userID, "vessel_dimensions", dimensions.ID, "DELETE", oldAudit, nil, ip, ua); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log DELETE for vessel_dimensions ID %d: %v", dimensions.ID, err)
			}
		}

		return nil
	})
}

// buildAuditData builds a VesselDimensionsAuditData from a dimensions model
func (s *service) buildAuditData(dimensions *model.VesselDimensions) VesselDimensionsAuditData {
	return VesselDimensionsAuditData{
		ID:           dimensions.ID,
		VesselID:     dimensions.VesselID,
		LOA:          dimensions.LOA,
		LBP:          dimensions.LBP,
		Beam:         dimensions.Beam,
		Depth:        dimensions.Depth,
		Draft:        dimensions.Draft,
		AirDraft:     dimensions.AirDraft,
		GT:           decimalToString(dimensions.GT),
		NT:           decimalToString(dimensions.NT),
		DWT:          decimalToString(dimensions.DWT),
		Displacement: decimalToString(dimensions.Displacement),
	}
}

func (s *service) mapToResponse(dimensions *model.VesselDimensions) *VesselDimensionsResponse {
	return &VesselDimensionsResponse{
		ID:           dimensions.ID,
		VesselID:     dimensions.VesselID,
		LOA:          dimensions.LOA,
		LBP:          dimensions.LBP,
		Beam:         dimensions.Beam,
		Depth:        dimensions.Depth,
		Draft:        dimensions.Draft,
		AirDraft:     dimensions.AirDraft,
		GT:           decimalToString(dimensions.GT),
		NT:           decimalToString(dimensions.NT),
		DWT:          decimalToString(dimensions.DWT),
		Displacement: decimalToString(dimensions.Displacement),
		CreatedAt:    dimensions.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    dimensions.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func decimalToString(d *decimal.Decimal) *string {
	if d == nil {
		return nil
	}
	str := d.StringFixed(2)
	return &str
}
