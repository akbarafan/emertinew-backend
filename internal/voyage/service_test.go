package voyage

import (
	"errors"
	"testing"
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
)

type mockRepository struct {
	activeVoyage *model.Voyage
	voyages      []model.Voyage
	createErr    error
	updateErr    error
	getErr       error
}

func (m *mockRepository) GetAll(status string, vesselID uint, search string, companyID *uint) ([]model.Voyage, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	var res []model.Voyage
	if m.activeVoyage != nil {
		res = append(res, *m.activeVoyage)
	}
	res = append(res, m.voyages...)
	return res, nil
}

func (m *mockRepository) CompleteVoyage(id uint) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if m.activeVoyage != nil && m.activeVoyage.ID == id {
		m.activeVoyage.Status = "completed"
	}
	for i := range m.voyages {
		if m.voyages[i].ID == id {
			m.voyages[i].Status = "completed"
		}
	}
	return nil
}

func (m *mockRepository) GetAllActive(companyID *uint) ([]model.Voyage, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	var res []model.Voyage
	if m.activeVoyage != nil {
		res = append(res, *m.activeVoyage)
	}
	for _, v := range m.voyages {
		if v.Status != "completed" {
			res = append(res, v)
		}
	}
	return res, nil
}

func (m *mockRepository) GetActiveByVesselID(vesselID uint) (*model.Voyage, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.activeVoyage != nil && m.activeVoyage.VesselID == vesselID {
		return m.activeVoyage, nil
	}
	return nil, errors.New("record not found")
}

func (m *mockRepository) GetByID(id uint) (*model.Voyage, error) {
	if m.activeVoyage != nil && m.activeVoyage.ID == id {
		return m.activeVoyage, nil
	}
	for _, v := range m.voyages {
		if v.ID == id {
			return &v, nil
		}
	}
	return nil, errors.New("record not found")
}

func (m *mockRepository) GetByVesselID(vesselID uint) ([]model.Voyage, error) {
	var res []model.Voyage
	for _, v := range m.voyages {
		if v.VesselID == vesselID {
			res = append(res, v)
		}
	}
	return res, nil
}

func (m *mockRepository) Create(v *model.Voyage) error {
	if m.createErr != nil {
		return m.createErr
	}
	v.ID = 101
	m.activeVoyage = v
	m.voyages = append(m.voyages, *v)
	return nil
}

func (m *mockRepository) Update(v *model.Voyage) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.activeVoyage = v
	return nil
}

func (m *mockRepository) CompletePreviousActive(vesselID uint, excludeID uint) error {
	return nil
}

func (m *mockRepository) Delete(id uint) error {
	return nil
}

func (m *mockRepository) GetVesselByID(id uint) (*model.MVessel, error) {
	cid := uint(1)
	return &model.MVessel{
		ID:        id,
		Name:      "Test Vessel",
		CompanyID: &cid,
	}, nil
}

func TestGetActiveVoyage(t *testing.T) {
	now := time.Now()
	mockRepo := &mockRepository{
		activeVoyage: &model.Voyage{
			ID:              1,
			VoyageNumber:    "V-2026-0001",
			VesselID:        319,
			Status:          "en_route",
			OriginName:      "Tg. Perak",
			OriginCode:      "SRK",
			DestinationName: "Madura Offshore",
			DestinationCode: "MOP",
			DepartureTime:   &now,
		},
	}

	svc := NewService(mockRepo)
	res, err := svc.GetActiveVoyage(319)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.VoyageNumber != "V-2026-0001" {
		t.Errorf("expected voyage number V-2026-0001, got %s", res.VoyageNumber)
	}
	if res.OriginName != "Tg. Perak" {
		t.Errorf("expected origin Tg. Perak, got %s", res.OriginName)
	}
}

func TestCreateVoyage(t *testing.T) {
	mockRepo := &mockRepository{}
	svc := NewService(mockRepo)

	req := CreateVoyageRequest{
		VesselID:        319,
		OriginName:      "Gresik Port",
		OriginCode:      "GRK",
		DestinationName: "Kalianget",
		DestinationCode: "KLN",
		Status:          "en_route",
	}

	res, err := svc.CreateVoyage(req, nil)
	if err != nil {
		t.Fatalf("expected create success, got error %v", err)
	}
	if res.ID != 101 {
		t.Errorf("expected ID 101, got %d", res.ID)
	}
	if res.OriginName != "Gresik Port" {
		t.Errorf("expected origin Gresik Port, got %s", res.OriginName)
	}
	if res.Status != "en_route" {
		t.Errorf("expected status en_route, got %s", res.Status)
	}
}
