package mcu

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(mcu *model.MMCU) error
	GetByID(id uint) (*model.MMCU, error)
	GetAll() ([]model.MMCU, error)
	Update(mcu *model.MMCU) error
	Delete(id uint) error
	AssignSensorsToMCU(mcuID uint, sensorIDs []uint) ([]uint, error)
	ValidateSensorIDs(sensorIDs []uint) ([]uint, []uint, error)
	GetDB() *gorm.DB
	WithTx(tx *gorm.DB) Repository

	// Bulk sensor operations
	GetSensorsByMCUID(mcuID uint) ([]model.MSensor, error)
	GetSensorByTopic(topic string) (*model.MSensor, error)
	BatchUpdateSensorTopic(sensorUpdates []SensorTopicUpdate) error
	GetVesselByID(vesselID uint) (*model.MVessel, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(mcu *model.MMCU) error {
	return r.db.Create(mcu).Error
}

func (r *repository) GetByID(id uint) (*model.MMCU, error) {
	var mcu model.MMCU
	err := r.db.Preload("Sensors").First(&mcu, id).Error
	if err != nil {
		return nil, err
	}
	return &mcu, nil
}

func (r *repository) GetAll() ([]model.MMCU, error) {
	var mcus []model.MMCU
	err := r.db.Preload("Sensors").Order("id ASC").Find(&mcus).Error
	return mcus, err
}

func (r *repository) Update(mcu *model.MMCU) error {
	return r.db.Omit("Sensors").Save(mcu).Error
}

func (r *repository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.MMCU{}).Error
}

func (r *repository) ValidateSensorIDs(sensorIDs []uint) ([]uint, []uint, error) {
	var existingIDs []uint
	err := r.db.Model(&model.MSensor{}).Where("id IN ?", sensorIDs).Pluck("id", &existingIDs).Error
	if err != nil {
		return nil, nil, err
	}

	// Find missing IDs
	missingIDs := []uint{}
	for _, id := range sensorIDs {
		found := false
		for _, existingID := range existingIDs {
			if id == existingID {
				found = true
				break
			}
		}
		if !found {
			missingIDs = append(missingIDs, id)
		}
	}

	return existingIDs, missingIDs, nil
}

func (r *repository) AssignSensorsToMCU(mcuID uint, sensorIDs []uint) ([]uint, error) {
	// Validate that all sensor IDs exist
	if len(sensorIDs) > 0 {
		_, missingIDs, err := r.ValidateSensorIDs(sensorIDs)
		if err != nil {
			return nil, err
		}
		if len(missingIDs) > 0 {
			return missingIDs, gorm.ErrRecordNotFound
		}
	}

	// Use transaction for atomicity
	return nil, r.db.Transaction(func(tx *gorm.DB) error {
		// Unassign all sensors from this MCU first (set mcu_id to NULL)
		// Use raw SQL to ensure NULL is properly set
		if err := tx.Exec("UPDATE m_sensors SET mcu_id = NULL WHERE mcu_id = ?", mcuID).Error; err != nil {
			return err
		}

		// Assign new sensors to this MCU
		if len(sensorIDs) > 0 {
			// Use raw SQL to ensure update happens correctly
			if err := tx.Exec("UPDATE m_sensors SET mcu_id = ? WHERE id IN ?", mcuID, sensorIDs).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}

// SensorTopicUpdate represents a batch sensor topic update
type SensorTopicUpdate struct {
	ID       uint
	NewTopic string
	NewKey   string
}

func (r *repository) GetSensorsByMCUID(mcuID uint) ([]model.MSensor, error) {
	var sensors []model.MSensor
	err := r.db.Where("mcu_id = ?", mcuID).Find(&sensors).Error
	return sensors, err
}

func (r *repository) GetSensorByTopic(topic string) (*model.MSensor, error) {
	var sensor model.MSensor
	err := r.db.Where("topic = ?", topic).First(&sensor).Error
	if err != nil {
		return nil, err
	}
	return &sensor, nil
}

func (r *repository) BatchUpdateSensorTopic(updates []SensorTopicUpdate) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, u := range updates {
			if err := tx.Model(&model.MSensor{}).Where("id = ?", u.ID).Updates(map[string]interface{}{
				"topic":     u.NewTopic,
				"topic_key": u.NewKey,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repository) GetVesselByID(vesselID uint) (*model.MVessel, error) {
	var vessel model.MVessel
	err := r.db.Preload("Fleet").Preload("Squad").First(&vessel, vesselID).Error
	if err != nil {
		return nil, err
	}
	return &vessel, nil
}
