package vessel

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Alugara-Inovasi-Utama/emerti/internal/audit"
	vesselcapacity "github.com/Alugara-Inovasi-Utama/emerti/internal/master/vessel/capacity"
	vesseldimensions "github.com/Alugara-Inovasi-Utama/emerti/internal/master/vessel/dimensions"
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// validTrackingSources contains the allowed values for tracking_source
var validTrackingSources = map[string]bool{
	"alugara": true,
	"hexa":    true,
}

func validateTrackingSource(src *string) error {
	if src != nil && !validTrackingSources[strings.ToLower(*src)] {
		return errors.New("tracking_source must be 'alugara' or 'hexa'")
	}
	return nil
}

func formatMissingIDsError(entityType string, missingIDs []uint) string {
	if len(missingIDs) == 0 {
		return fmt.Sprintf("One or more %s IDs not found", entityType)
	}

	idsStr := make([]string, len(missingIDs))
	for i, id := range missingIDs {
		idsStr[i] = fmt.Sprintf("%d", id)
	}

	if len(missingIDs) == 1 {
		return fmt.Sprintf("%s ID not found: %s", strings.Title(entityType), idsStr[0])
	}
	return fmt.Sprintf("%s IDs not found: %s", strings.Title(entityType), strings.Join(idsStr, ", "))
}

type Service interface {
	CreateVessel(req CreateVesselRequest, userID uint, ip, ua string) (*VesselResponse, error)
	GetVesselByID(id uint) (*VesselResponse, error)
	GetAllVessels() ([]VesselResponse, error)
	UpdateVessel(id uint, req UpdateVesselRequest, userID uint, ip, ua string) (*VesselResponse, error)
	DeleteVessel(id uint, userID uint, ip, ua string) error
	RestoreVessel(id uint, userID uint, ip, ua string) error
	GetDeletedVessels() ([]VesselResponse, error)
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

func (s *service) CreateVessel(req CreateVesselRequest, userID uint, ip, ua string) (*VesselResponse, error) {
	// Validate tracking_source
	if err := validateTrackingSource(req.TrackingSource); err != nil {
		return nil, err
	}

	// Normalize code to uppercase
	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))

	// Check for duplicate code (case-insensitive)
	existing, _ := s.repo.GetByCodeCI(req.Code)
	if existing != nil {
		return nil, errors.New("duplicate record")
	}

	// Validate mcu IDs first before creating vessel
	if len(req.MCUIDs) > 0 {
		_, missingIDs, err := s.repo.ValidateMCUIDs(req.MCUIDs)
		if err != nil {
			return nil, err
		}
		if len(missingIDs) > 0 {
			return nil, errors.New(formatMissingIDsError("mcu", missingIDs))
		}
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	vessel := &model.MVessel{
		Code:           req.Code,
		Name:           req.Name,
		IsActive:       isActive,
		CompanyID:      req.CompanyID,
		VesselTypeID:   req.VesselTypeID,
		ClassID:        req.ClassID,
		MaterialID:     req.MaterialID,
		OperationID:    req.OperationID,
		FleetID:        req.FleetID,
		SquadID:        req.SquadID,
		FuelTypeID:     req.FuelTypeID,
		CallSign:       req.CallSign,
		IMONumber:      req.IMONumber,
		MMSINumber:     req.MMSINumber,
		Flag:           req.Flag,
		PortOfRegistry: req.PortOfRegistry,
		ProductionYear: req.ProductionYear,
		Builder:        req.Builder,
		HullNumber:     req.HullNumber,
		ClassNotation:  req.ClassNotation,
		Nickname:       req.Nickname,
		TrackingSource: req.TrackingSource,
	}

	db := s.repo.GetDB()
	err := db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Create(vessel); err != nil {
			return err
		}

		// Assign mcus ke vessel ini
		if len(req.MCUIDs) > 0 {
			if _, err := txRepo.AssignMCUsToVessel(vessel.ID, req.MCUIDs); err != nil {
				return err
			}
		}

		// Set CreatedBy and UpdatedBy
		vessel.CreatedBy = &userID
		vessel.UpdatedBy = &userID
		if err := txRepo.Update(vessel); err != nil {
			return err
		}

		// Log audit within transaction
		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			auditData := s.buildAuditData(vessel)
			if err := txAudit.Log(userID, "m_vessel", vessel.ID, "CREATE", nil, auditData, ip, ua); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log CREATE for m_vessel ID %d: %v", vessel.ID, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Get vessel with preloaded relations
	vessel, err = s.repo.GetByID(vessel.ID)
	if err != nil {
		return nil, err
	}

	return s.mapToResponse(vessel), nil
}

func (s *service) GetVesselByID(id uint) (*VesselResponse, error) {
	vessel, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("Vessel not found")
	}

	return s.mapToResponse(vessel), nil
}

func (s *service) GetAllVessels() ([]VesselResponse, error) {
	vessels, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	// Return empty array instead of error when no data exists
	if len(vessels) == 0 {
		return []VesselResponse{}, nil
	}

	var response []VesselResponse
	for _, vessel := range vessels {
		response = append(response, *s.mapToResponse(&vessel))
	}

	return response, nil
}

func (s *service) UpdateVessel(id uint, req UpdateVesselRequest, userID uint, ip, ua string) (*VesselResponse, error) {
	vessel, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("Vessel not found")
	}

	// Get old data for audit
	oldData := *vessel

	// Validate tracking_source
	if err := validateTrackingSource(req.TrackingSource); err != nil {
		return nil, err
	}

	// Validate mcu IDs first if provided
	if len(req.MCUIDs) > 0 {
		_, missingIDs, err := s.repo.ValidateMCUIDs(req.MCUIDs)
		if err != nil {
			return nil, err
		}
		if len(missingIDs) > 0 {
			return nil, errors.New(formatMissingIDsError("mcu", missingIDs))
		}
	}

	// Update fields
	if req.Code != "" {
		req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
		// Check for duplicate code (case-insensitive), exclude current vessel
		existing, _ := s.repo.GetByCodeCI(req.Code)
		if existing != nil && existing.ID != id {
			return nil, errors.New("duplicate record")
		}
		vessel.Code = req.Code
	}
	if req.Name != "" {
		vessel.Name = req.Name
	}
	if req.IsActive != nil {
		vessel.IsActive = *req.IsActive
	}
	if req.CompanyID != nil {
		if *req.CompanyID == 0 {
			vessel.CompanyID = nil
		} else {
			vessel.CompanyID = req.CompanyID
		}
	}
	if req.FleetID != nil {
		if *req.FleetID == 0 {
			vessel.FleetID = nil
		} else {
			vessel.FleetID = req.FleetID
		}
	}
	if req.SquadID != nil {
		if *req.SquadID == 0 {
			vessel.SquadID = nil
		} else {
			vessel.SquadID = req.SquadID
		}
	}
	if req.FuelTypeID != nil {
		if *req.FuelTypeID == 0 {
			vessel.FuelTypeID = nil
		} else {
			vessel.FuelTypeID = req.FuelTypeID
		}
	}
	if req.VesselTypeID != nil {
		if *req.VesselTypeID == 0 {
			vessel.VesselTypeID = nil
		} else {
			vessel.VesselTypeID = req.VesselTypeID
		}
	}
	if req.ClassID != nil {
		if *req.ClassID == 0 {
			vessel.ClassID = nil
		} else {
			vessel.ClassID = req.ClassID
		}
	}
	if req.MaterialID != nil {
		if *req.MaterialID == 0 {
			vessel.MaterialID = nil
		} else {
			vessel.MaterialID = req.MaterialID
		}
	}
	if req.OperationID != nil {
		if *req.OperationID == 0 {
			vessel.OperationID = nil
		} else {
			vessel.OperationID = req.OperationID
		}
	}
	if req.CallSign != nil {
		if *req.CallSign == "" {
			vessel.CallSign = nil
		} else {
			vessel.CallSign = req.CallSign
		}
	}
	if req.IMONumber != nil {
		if *req.IMONumber == "" {
			vessel.IMONumber = nil
		} else {
			vessel.IMONumber = req.IMONumber
		}
	}
	if req.MMSINumber != nil {
		if *req.MMSINumber == "" {
			vessel.MMSINumber = nil
		} else {
			vessel.MMSINumber = req.MMSINumber
		}
	}
	if req.Flag != nil {
		if *req.Flag == "" {
			vessel.Flag = nil
		} else {
			vessel.Flag = req.Flag
		}
	}
	if req.PortOfRegistry != nil {
		if *req.PortOfRegistry == "" {
			vessel.PortOfRegistry = nil
		} else {
			vessel.PortOfRegistry = req.PortOfRegistry
		}
	}
	if req.ProductionYear != nil {
		if *req.ProductionYear == 0 {
			vessel.ProductionYear = nil
		} else {
			vessel.ProductionYear = req.ProductionYear
		}
	}
	if req.Builder != nil {
		if *req.Builder == "" {
			vessel.Builder = nil
		} else {
			vessel.Builder = req.Builder
		}
	}
	if req.HullNumber != nil {
		if *req.HullNumber == "" {
			vessel.HullNumber = nil
		} else {
			vessel.HullNumber = req.HullNumber
		}
	}
	if req.ClassNotation != nil {
		if *req.ClassNotation == "" {
			vessel.ClassNotation = nil
		} else {
			vessel.ClassNotation = req.ClassNotation
		}
	}
	if req.Nickname != nil {
		if *req.Nickname == "" {
			vessel.Nickname = nil
		} else {
			vessel.Nickname = req.Nickname
		}
	}
	if req.TrackingSource != nil {
		vessel.TrackingSource = req.TrackingSource
	}

	// Set UpdatedBy
	vessel.UpdatedBy = &userID

	db := s.repo.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Update(vessel); err != nil {
			return err
		}

		// Reassign mcus jika ada
		if len(req.MCUIDs) > 0 {
			if _, err := txRepo.AssignMCUsToVessel(vessel.ID, req.MCUIDs); err != nil {
				return err
			}
		}

		// Log audit within transaction
		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			oldAudit := s.buildAuditData(&oldData)
			newAudit := s.buildAuditData(vessel)
			if err := txAudit.Log(userID, "m_vessel", id, "UPDATE", oldAudit, newAudit, ip, ua); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log UPDATE for m_vessel ID %d: %v", id, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Get updated vessel with preloaded relations
	vessel, err = s.repo.GetByID(vessel.ID)
	if err != nil {
		return nil, err
	}

	return s.mapToResponse(vessel), nil
}

func (s *service) DeleteVessel(id uint, userID uint, ip, ua string) error {
	vessel, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("Vessel not found")
	}

	// Get old data for audit
	oldData := *vessel

	db := s.repo.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		// Set UpdatedBy before delete
		vessel.UpdatedBy = &userID
		if err := txRepo.Update(vessel); err != nil {
			return err
		}

		if err := txRepo.Delete(id); err != nil {
			return err
		}

		// Log audit within transaction
		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			oldAudit := s.buildAuditData(&oldData)
			if err := txAudit.Log(userID, "m_vessel", id, "DELETE", oldAudit, nil, ip, ua); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log DELETE for m_vessel ID %d: %v", id, err)
			}
		}

		return nil
	})

	return err
}

func (s *service) RestoreVessel(id uint, userID uint, ip, ua string) error {
	// Use Unscoped to find the soft-deleted vessel
	var vessel model.MVessel
	db := s.repo.GetDB()
	if err := db.Unscoped().First(&vessel, id).Error; err != nil {
		return errors.New("Vessel not found")
	}
	if vessel.DeletedAt.Time.IsZero() {
		return errors.New("Vessel is not deleted")
	}

	oldData := vessel

	err := db.Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if err := txRepo.Restore(id); err != nil {
			return err
		}

		if s.auditService != nil {
			txAudit := s.auditService.WithTx(tx)
			oldAudit := s.buildAuditData(&oldData)
			if err := txAudit.Log(userID, "m_vessel", id, "RESTORE", oldAudit, nil, ip, ua); err != nil {
				log.Printf("[AUDIT ERROR] Failed to log RESTORE for m_vessel ID %d: %v", id, err)
			}
		}

		return nil
	})

	return err
}

func (s *service) GetDeletedVessels() ([]VesselResponse, error) {
	vessels, err := s.repo.GetDeleted()
	if err != nil {
		return nil, err
	}

	if len(vessels) == 0 {
		return []VesselResponse{}, nil
	}

	var response []VesselResponse
	for _, vessel := range vessels {
		response = append(response, *s.mapToResponse(&vessel))
	}

	return response, nil
}

// buildAuditData builds a complete VesselAuditData from a vessel model
func (s *service) buildAuditData(vessel *model.MVessel) VesselAuditData {
	var mcuIDs []uint
	for _, mcu := range vessel.MCUs {
		mcuIDs = append(mcuIDs, mcu.ID)
	}

	return VesselAuditData{
		ID:             vessel.ID,
		Code:           vessel.Code,
		Name:           vessel.Name,
		IsActive:       vessel.IsActive,
		CompanyID:      vessel.CompanyID,
		VesselTypeID:   vessel.VesselTypeID,
		ClassID:        vessel.ClassID,
		MaterialID:     vessel.MaterialID,
		OperationID:    vessel.OperationID,
		FleetID:        vessel.FleetID,
		SquadID:        vessel.SquadID,
		CallSign:       vessel.CallSign,
		IMONumber:      vessel.IMONumber,
		MMSINumber:     vessel.MMSINumber,
		Flag:           vessel.Flag,
		PortOfRegistry: vessel.PortOfRegistry,
		ProductionYear: vessel.ProductionYear,
		Builder:        vessel.Builder,
		HullNumber:     vessel.HullNumber,
		ClassNotation:  vessel.ClassNotation,
		Nickname:       vessel.Nickname,
		TrackingSource: vessel.TrackingSource,
		MCUIDs:         mcuIDs,
	}
}

// Helper function to map model to response
func (s *service) mapToResponse(vessel *model.MVessel) *VesselResponse {
	mcus := []MCUInfo{}
	for _, mcu := range vessel.MCUs {
		mcus = append(mcus, MCUInfo{
			ID:       mcu.ID,
			Name:     mcu.Name,
			Location: mcu.Location,
		})
	}

	response := &VesselResponse{
		ID:             vessel.ID,
		Code:           vessel.Code,
		Name:           vessel.Name,
		IsActive:       vessel.IsActive,
		CallSign:       vessel.CallSign,
		IMONumber:      vessel.IMONumber,
		MMSINumber:     vessel.MMSINumber,
		Flag:           vessel.Flag,
		PortOfRegistry: vessel.PortOfRegistry,
		ProductionYear: vessel.ProductionYear,
		Builder:        vessel.Builder,
		HullNumber:     vessel.HullNumber,
		ClassNotation:  vessel.ClassNotation,
		Nickname:       vessel.Nickname,
		TrackingSource: vessel.TrackingSource,
		MCUs:           mcus,
		CreatedAt:      vessel.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      vessel.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// Add company info if available
	if vessel.CompanyID != nil && vessel.Company.ID > 0 {
		response.Company = &LookupInfo{
			ID:   vessel.Company.ID,
			Name: vessel.Company.Name,
		}
	}

	// Add lookup info if available
	if vessel.VesselTypeID != nil && vessel.VesselType.ID > 0 {
		response.VesselType = &LookupInfo{
			ID:   vessel.VesselType.ID,
			Name: vessel.VesselType.Name,
		}
	}

	if vessel.FleetID != nil && vessel.Fleet.ID > 0 {
		response.Fleet = &LookupInfo{
			ID:   vessel.Fleet.ID,
			Name: vessel.Fleet.Name,
		}
	}

	if vessel.SquadID != nil && vessel.Squad.ID > 0 {
		response.Squad = &LookupInfo{
			ID:   vessel.Squad.ID,
			Name: vessel.Squad.Name,
		}
	}

	if vessel.ClassID != nil && vessel.Classification.ID > 0 {
		response.Classification = &LookupInfo{
			ID:   vessel.Classification.ID,
			Name: vessel.Classification.Name,
		}
	}

	if vessel.MaterialID != nil && vessel.MaterialHull.ID > 0 {
		response.MaterialHull = &LookupInfo{
			ID:   vessel.MaterialHull.ID,
			Name: vessel.MaterialHull.Name,
		}
	}

	if vessel.OperationID != nil && vessel.OperationStatus.ID > 0 {
		response.OperationStatus = &LookupInfo{
			ID:   vessel.OperationStatus.ID,
			Name: vessel.OperationStatus.Name,
		}
	}

	// The fuel type carries the emission factors, so the client can turn a
	// volume of fuel into an emission figure without repeating the formula.
	if vessel.FuelTypeID != nil && vessel.FuelType.ID > 0 {
		fuelType := &FuelTypeInfo{
			ID:        vessel.FuelType.ID,
			Code:      vessel.FuelType.Code,
			Name:      vessel.FuelType.Name,
			Reference: vessel.FuelType.Reference,
			Method:    vessel.FuelType.Method,
		}
		if value, ok := vessel.FuelType.TtWCO2KgPerLitre(); ok {
			fuelType.TtWCO2KgPerLitre = &value
		}
		if value, ok := vessel.FuelType.TtWCO2eKgPerLitre(); ok {
			fuelType.TtWCO2eKgPerLitre = &value
		}
		if value, ok := vessel.FuelType.WtWCO2eKgPerLitre(); ok {
			fuelType.WtWCO2eKgPerLitre = &value
		}

		// The raw IMO parameters travel with the derived factors, so the panel
		// can show how a factor is built instead of asking the operator to
		// believe it. A parameter that was never filled in stays null.
		if vessel.FuelType.CfCO2KgPerKg != nil {
			fuelType.CfCO2KgPerKg = vessel.FuelType.CfCO2KgPerKg
		}
		if vessel.FuelType.CfCH4KgPerKg != nil {
			fuelType.CfCH4KgPerKg = vessel.FuelType.CfCH4KgPerKg
		}
		if vessel.FuelType.CfN2OKgPerKg != nil {
			fuelType.CfN2OKgPerKg = vessel.FuelType.CfN2OKgPerKg
		}
		if vessel.FuelType.MethaneSlipPc != nil {
			fuelType.MethaneSlipPercent = vessel.FuelType.MethaneSlipPc
		}
		if vessel.FuelType.WttIntensityGCO2eqPerMJ != nil {
			fuelType.WttIntensityGCO2eqPerMJ = vessel.FuelType.WttIntensityGCO2eqPerMJ
		}
		if vessel.FuelType.LhvMJPerKg != nil {
			fuelType.LhvMJPerKg = vessel.FuelType.LhvMJPerKg
		}
		if vessel.FuelType.DensityKgPerLitre != nil {
			fuelType.DensityKgPerLitre = vessel.FuelType.DensityKgPerLitre
		}
		response.FuelType = fuelType
	}

	// Add dimensions if available
	if vessel.Dimensions.ID > 0 {
		response.Dimensions = mapDimensionsToResponse(&vessel.Dimensions)
	}

	// Add capacity if available
	if vessel.Capacity.ID > 0 {
		response.Capacity = mapCapacityToResponse(&vessel.Capacity)
	}

	return response
}

// Helper function to convert decimal to string with 2 decimal places
func decimalToString(d *decimal.Decimal) *string {
	if d == nil {
		return nil
	}
	str := d.StringFixed(2)
	return &str
}

// Helper function to map dimensions to response
func mapDimensionsToResponse(dimensions *model.VesselDimensions) *vesseldimensions.VesselDimensionsResponse {
	return &vesseldimensions.VesselDimensionsResponse{
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
		CreatedAt:    dimensions.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    dimensions.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Helper function to map capacity to response
func mapCapacityToResponse(capacity *model.VesselCapacity) *vesselcapacity.VesselCapacityResponse {
	return &vesselcapacity.VesselCapacityResponse{
		ID:              capacity.ID,
		VesselID:        capacity.VesselID,
		LoadCapacity:    decimalToString(capacity.LoadCapacity),
		LoadUnit:        capacity.LoadUnit,
		DeckArea:        decimalToString(capacity.DeckArea),
		CargoHoldVolume: decimalToString(capacity.CargoHoldVolume),
		HasCrane:        capacity.HasCrane,
		SWLCrane:        decimalToString(capacity.SWLCrane),
		BollardPull:     decimalToString(capacity.BollardPull),
		CreatedAt:       capacity.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       capacity.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
