package vesselcapacity

import (
	"errors"
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Service interface {
	CreateVesselCapacity(vesselID uint, req CreateVesselCapacityRequest, userID uint, ip, ua string) (*VesselCapacityResponse, error)
	GetVesselCapacityByID(id uint) (*VesselCapacityResponse, error)
	GetVesselCapacityByVesselID(vesselID uint) (*VesselCapacityResponse, error)
	UpdateVesselCapacity(id uint, req UpdateVesselCapacityRequest, userID uint, ip, ua string) (*VesselCapacityResponse, error)
	UpdateVesselCapacityByVesselID(vesselID uint, req UpdateVesselCapacityRequest, userID uint, ip, ua string) (*VesselCapacityResponse, error)
	DeleteVesselCapacity(id uint, userID uint, ip, ua string) error
	DeleteVesselCapacityByVesselID(vesselID uint, userID uint, ip, ua string) error
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

func (s *service) CreateVesselCapacity(vesselID uint, req CreateVesselCapacityRequest, userID uint, ip, ua string) (*VesselCapacityResponse, error) {
	// Validate vessel exists
	exists, err := s.repo.VesselExists(vesselID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("vessel not found")
	}

	// Check if capacity already exists for this vessel
	existing, _ := s.repo.GetByVesselID(vesselID)
	if existing != nil {
		return nil, errors.New("vessel capacity already exists for this vessel")
	}

	capacity := &model.VesselCapacity{
		VesselID:        vesselID,
		LoadCapacity:    req.LoadCapacity,
		LoadUnit:        req.LoadUnit,
		DeckArea:        req.DeckArea,
		CargoHoldVolume: req.CargoHoldVolume,
		HasCrane:        req.HasCrane,
		SWLCrane:        req.SWLCrane,
		BollardPull:     req.BollardPull,
	}

	db := s.repo.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Create(capacity); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			auditData := s.buildAuditData(capacity)
			if err := txAudit.Log(userID, "vessel_capacity", capacity.ID, "CREATE", nil, auditData, ip, ua); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log CREATE for vessel_capacity ID %d: %v", capacity.ID, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.mapToResponse(capacity), nil
}

func (s *service) GetVesselCapacityByID(id uint) (*VesselCapacityResponse, error) {
	capacity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("vessel capacity not found")
	}

	return s.mapToResponse(capacity), nil
}

func (s *service) GetVesselCapacityByVesselID(vesselID uint) (*VesselCapacityResponse, error) {
	capacity, err := s.repo.GetByVesselID(vesselID)
	if err != nil {
		return nil, errors.New("vessel capacity not found")
	}

	return s.mapToResponse(capacity), nil
}

func (s *service) UpdateVesselCapacity(id uint, req UpdateVesselCapacityRequest, userID uint, ip, ua string) (*VesselCapacityResponse, error) {
	capacity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("vessel capacity not found")
	}

	oldData := *capacity

	// Update only non-nil fields
	if req.LoadCapacity != nil {
		capacity.LoadCapacity = req.LoadCapacity
	}
	if req.LoadUnit != nil {
		capacity.LoadUnit = req.LoadUnit
	}
	if req.DeckArea != nil {
		capacity.DeckArea = req.DeckArea
	}
	if req.CargoHoldVolume != nil {
		capacity.CargoHoldVolume = req.CargoHoldVolume
	}
	if req.HasCrane != nil {
		capacity.HasCrane = req.HasCrane
	}
	if req.SWLCrane != nil {
		capacity.SWLCrane = req.SWLCrane
	}
	if req.BollardPull != nil {
		capacity.BollardPull = req.BollardPull
	}

	db := s.repo.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Update(capacity); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			oldAudit := s.buildAuditData(&oldData)
			newAudit := s.buildAuditData(capacity)
			if err := txAudit.Log(userID, "vessel_capacity", id, "UPDATE", oldAudit, newAudit, ip, ua); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log UPDATE for vessel_capacity ID %d: %v", id, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.mapToResponse(capacity), nil
}

func (s *service) UpdateVesselCapacityByVesselID(vesselID uint, req UpdateVesselCapacityRequest, userID uint, ip, ua string) (*VesselCapacityResponse, error) {
	capacity, err := s.repo.GetByVesselID(vesselID)
	if err != nil {
		return nil, errors.New("vessel capacity not found")
	}

	oldData := *capacity

	// Update only non-nil fields
	if req.LoadCapacity != nil {
		capacity.LoadCapacity = req.LoadCapacity
	}
	if req.LoadUnit != nil {
		capacity.LoadUnit = req.LoadUnit
	}
	if req.DeckArea != nil {
		capacity.DeckArea = req.DeckArea
	}
	if req.CargoHoldVolume != nil {
		capacity.CargoHoldVolume = req.CargoHoldVolume
	}
	if req.HasCrane != nil {
		capacity.HasCrane = req.HasCrane
	}
	if req.SWLCrane != nil {
		capacity.SWLCrane = req.SWLCrane
	}
	if req.BollardPull != nil {
		capacity.BollardPull = req.BollardPull
	}

	db := s.repo.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Update(capacity); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			oldAudit := s.buildAuditData(&oldData)
			newAudit := s.buildAuditData(capacity)
			if err := txAudit.Log(userID, "vessel_capacity", capacity.ID, "UPDATE", oldAudit, newAudit, ip, ua); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log UPDATE for vessel_capacity ID %d: %v", capacity.ID, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.mapToResponse(capacity), nil
}

func (s *service) DeleteVesselCapacity(id uint, userID uint, ip, ua string) error {
	capacity, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("vessel capacity not found")
	}

	oldData := *capacity

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Delete(id); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			oldAudit := s.buildAuditData(&oldData)
			if err := txAudit.Log(userID, "vessel_capacity", id, "DELETE", oldAudit, nil, ip, ua); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log DELETE for vessel_capacity ID %d: %v", id, err)
			}
		}

		return nil
	})
}

func (s *service) DeleteVesselCapacityByVesselID(vesselID uint, userID uint, ip, ua string) error {
	capacity, err := s.repo.GetByVesselID(vesselID)
	if err != nil {
		return errors.New("vessel capacity not found")
	}

	oldData := *capacity

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Delete(capacity.ID); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			oldAudit := s.buildAuditData(&oldData)
			if err := txAudit.Log(userID, "vessel_capacity", capacity.ID, "DELETE", oldAudit, nil, ip, ua); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log DELETE for vessel_capacity ID %d: %v", capacity.ID, err)
			}
		}

		return nil
	})
}

// buildAuditData builds a VesselCapacityAuditData from a capacity model
func (s *service) buildAuditData(capacity *model.VesselCapacity) VesselCapacityAuditData {
	return VesselCapacityAuditData{
		ID:              capacity.ID,
		VesselID:        capacity.VesselID,
		LoadCapacity:    decimalToString(capacity.LoadCapacity),
		LoadUnit:        capacity.LoadUnit,
		DeckArea:        decimalToString(capacity.DeckArea),
		CargoHoldVolume: decimalToString(capacity.CargoHoldVolume),
		HasCrane:        capacity.HasCrane,
		SWLCrane:        decimalToString(capacity.SWLCrane),
		BollardPull:     decimalToString(capacity.BollardPull),
	}
}

func (s *service) mapToResponse(capacity *model.VesselCapacity) *VesselCapacityResponse {
	return &VesselCapacityResponse{
		ID:              capacity.ID,
		VesselID:        capacity.VesselID,
		LoadCapacity:    decimalToString(capacity.LoadCapacity),
		LoadUnit:        capacity.LoadUnit,
		DeckArea:        decimalToString(capacity.DeckArea),
		CargoHoldVolume: decimalToString(capacity.CargoHoldVolume),
		HasCrane:        capacity.HasCrane,
		SWLCrane:        decimalToString(capacity.SWLCrane),
		BollardPull:     decimalToString(capacity.BollardPull),
		CreatedAt:       capacity.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       capacity.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func decimalToString(d *decimal.Decimal) *string {
	if d == nil {
		return nil
	}
	str := d.StringFixed(2)
	return &str
}
