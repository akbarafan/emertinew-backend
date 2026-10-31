package voyage

import (
	"errors"
	"fmt"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

type Service interface {
	GetAllVoyages(status string, vesselID uint, search string, companyID *uint) ([]VoyageResponse, error)
	GetAllActiveVoyages(companyID *uint) ([]VoyageResponse, error)
	GetActiveVoyage(vesselID uint) (*VoyageResponse, error)
	GetVoyagesByVessel(vesselID uint) ([]VoyageResponse, error)
	CreateVoyage(req CreateVoyageRequest, companyID *uint) (*VoyageResponse, error)
	UpdateVoyage(id uint, req UpdateVoyageRequest, companyID *uint) (*VoyageResponse, error)
	CompleteVoyage(id uint, companyID *uint) (*VoyageResponse, error)
	DeleteVoyage(id uint, companyID *uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAllVoyages(status string, vesselID uint, search string, companyID *uint) ([]VoyageResponse, error) {
	list, err := s.repo.GetAll(status, vesselID, search, companyID)
	if err != nil {
		return nil, err
	}
	res := make([]VoyageResponse, len(list))
	for i, v := range list {
		res[i] = *s.toResponse(&v)
	}
	return res, nil
}

func (s *service) CompleteVoyage(id uint, companyID *uint) (*VoyageResponse, error) {
	v, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if companyID != nil && *companyID > 0 && v.Vessel.CompanyID != nil && *v.Vessel.CompanyID != *companyID {
		return nil, errors.New("access denied to this voyage")
	}
	if err := s.repo.CompleteVoyage(id); err != nil {
		return nil, err
	}
	updated, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.toResponse(updated), nil
}

func (s *service) GetAllActiveVoyages(companyID *uint) ([]VoyageResponse, error) {
	list, err := s.repo.GetAllActive(companyID)
	if err != nil {
		return nil, err
	}
	seen := make(map[uint]bool)
	var res []VoyageResponse
	for _, v := range list {
		if !seen[v.VesselID] {
			seen[v.VesselID] = true
			res = append(res, *s.toResponse(&v))
		}
	}
	return res, nil
}

func (s *service) GetActiveVoyage(vesselID uint) (*VoyageResponse, error) {
	v, err := s.repo.GetActiveByVesselID(vesselID)
	if err != nil {
		return nil, err
	}
	return s.toResponse(v), nil
}

func (s *service) GetVoyagesByVessel(vesselID uint) ([]VoyageResponse, error) {
	list, err := s.repo.GetByVesselID(vesselID)
	if err != nil {
		return nil, err
	}
	res := make([]VoyageResponse, len(list))
	for i, v := range list {
		res[i] = *s.toResponse(&v)
	}
	return res, nil
}

func (s *service) CreateVoyage(req CreateVoyageRequest, companyID *uint) (*VoyageResponse, error) {
	if companyID != nil && *companyID > 0 {
		vessel, err := s.repo.GetVesselByID(req.VesselID)
		if err != nil {
			return nil, errors.New("vessel not found")
		}
		if vessel.CompanyID != nil && *vessel.CompanyID != *companyID {
			return nil, errors.New("vessel does not belong to your company")
		}
	}

	status := req.Status
	if status == "" {
		status = "en_route"
	}

	voyageNumber := req.VoyageNumber
	if voyageNumber == "" {
		now := time.Now()
		voyageNumber = fmt.Sprintf("V-%d-%04d", now.Year(), now.Unix()%10000)
	}

	originType := req.OriginType
	if originType == "" {
		originType = "port"
	}

	destType := req.DestinationType
	if destType == "" {
		destType = "offshore"
	}

	v := &model.Voyage{
		VoyageNumber:    voyageNumber,
		VesselID:        req.VesselID,
		Status:          status,
		OriginType:      originType,
		OriginID:        req.OriginID,
		OriginName:      req.OriginName,
		OriginCode:      req.OriginCode,
		OriginLat:       req.OriginLat,
		OriginLong:      req.OriginLong,
		DepartureTime:   req.DepartureTime,
		DestinationType: destType,
		DestinationID:   req.DestinationID,
		DestinationName: req.DestinationName,
		DestinationCode: req.DestinationCode,
		DestinationLat:  req.DestinationLat,
		DestinationLong: req.DestinationLong,
		ArrivalTime:     req.ArrivalTime,
		Notes:           req.Notes,
	}

	if err := s.repo.Create(v); err != nil {
		return nil, err
	}

	// Reload with vessel preload
	loaded, err := s.repo.GetByID(v.ID)
	if err == nil {
		return s.toResponse(loaded), nil
	}
	return s.toResponse(v), nil
}

func (s *service) UpdateVoyage(id uint, req UpdateVoyageRequest, companyID *uint) (*VoyageResponse, error) {
	v, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if companyID != nil && *companyID > 0 && v.Vessel.CompanyID != nil && *v.Vessel.CompanyID != *companyID {
		return nil, errors.New("access denied to this voyage")
	}

	if req.Status != nil {
		v.Status = *req.Status
	}
	if req.OriginType != nil {
		v.OriginType = *req.OriginType
	}
	if req.OriginID != nil {
		v.OriginID = req.OriginID
	}
	if req.OriginName != nil {
		v.OriginName = *req.OriginName
	}
	if req.OriginCode != nil {
		v.OriginCode = *req.OriginCode
	}
	if req.OriginLat != nil {
		v.OriginLat = req.OriginLat
	}
	if req.OriginLong != nil {
		v.OriginLong = req.OriginLong
	}
	if req.DepartureTime != nil {
		v.DepartureTime = req.DepartureTime
	}
	if req.DestinationType != nil {
		v.DestinationType = *req.DestinationType
	}
	if req.DestinationID != nil {
		v.DestinationID = req.DestinationID
	}
	if req.DestinationName != nil {
		v.DestinationName = *req.DestinationName
	}
	if req.DestinationCode != nil {
		v.DestinationCode = *req.DestinationCode
	}
	if req.DestinationLat != nil {
		v.DestinationLat = req.DestinationLat
	}
	if req.DestinationLong != nil {
		v.DestinationLong = req.DestinationLong
	}
	if req.ArrivalTime != nil {
		v.ArrivalTime = req.ArrivalTime
	}
	if req.ActualDeparture != nil {
		v.ActualDeparture = req.ActualDeparture
	}
	if req.ActualArrival != nil {
		v.ActualArrival = req.ActualArrival
	}
	if req.Notes != nil {
		v.Notes = *req.Notes
	}

	if err := s.repo.Update(v); err != nil {
		return nil, err
	}

	loaded, err := s.repo.GetByID(v.ID)
	if err == nil {
		return s.toResponse(loaded), nil
	}
	return s.toResponse(v), nil
}

func (s *service) DeleteVoyage(id uint, companyID *uint) error {
	if companyID != nil && *companyID > 0 {
		v, err := s.repo.GetByID(id)
		if err != nil {
			return err
		}
		if v.Vessel.CompanyID != nil && *v.Vessel.CompanyID != *companyID {
			return errors.New("access denied to this voyage")
		}
	}
	return s.repo.Delete(id)
}

func (s *service) toResponse(v *model.Voyage) *VoyageResponse {
	if v == nil {
		return nil
	}
	vesselName := ""
	if v.Vessel.Name != "" {
		vesselName = v.Vessel.Name
	}

	return &VoyageResponse{
		ID:              v.ID,
		VoyageNumber:    v.VoyageNumber,
		VesselID:        v.VesselID,
		VesselName:      vesselName,
		Status:          v.Status,
		OriginType:      v.OriginType,
		OriginID:        v.OriginID,
		OriginName:      v.OriginName,
		OriginCode:      v.OriginCode,
		OriginLat:       v.OriginLat,
		OriginLong:      v.OriginLong,
		DepartureTime:   v.DepartureTime,
		DestinationType: v.DestinationType,
		DestinationID:   v.DestinationID,
		DestinationName: v.DestinationName,
		DestinationCode: v.DestinationCode,
		DestinationLat:  v.DestinationLat,
		DestinationLong: v.DestinationLong,
		ArrivalTime:     v.ArrivalTime,
		ActualDeparture: v.ActualDeparture,
		ActualArrival:   v.ActualArrival,
		Notes:           v.Notes,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}
