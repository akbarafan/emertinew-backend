package sensor

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"github.com/Alugara-Inovasi-Utama/emerti/utils"
	"gorm.io/gorm"
)

type Repository interface {
	// Sensor Data (TimescaleDB hypertable)
	CreateData(data *model.SensorData) error
	GetAll(limit int, sensorID uint) ([]model.SensorData, error)
	GetBySensorID(sensorID uint) ([]model.SensorData, error)
	GetLatestBySensorID(sensorID uint) (*model.SensorData, error)

	// Sensor Master CRUD
	CreateSensor(sensor *model.MSensor) error
	GetSensorByID(id uint) (*model.MSensor, error)
	GetSensorByTopic(topic string) (*model.MSensor, error)
	GetAllSensors() ([]model.MSensor, error)
	UpdateSensor(sensor *model.MSensor) error
	DeleteSensor(id uint) error

	// Helper methods for vessel info
	GetMCUByID(mcuID uint) (*model.MMCU, error)
	GetVesselByID(vesselID uint) (*model.MVessel, error)
	GetSensorsByVesselID(vesselID uint) ([]model.MSensor, error)

	// DB access
	GetDB() *gorm.DB
	WithTx(tx *gorm.DB) Repository
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateData(data *model.SensorData) error {
	return r.db.Create(data).Error
}

func (r *repository) GetAll(limit int, sensorID uint) ([]model.SensorData, error) {
	var data []model.SensorData
	if limit <= 0 {
		limit = 100
	}

	query := r.db.Order("time DESC").Limit(limit)

	// Apply sensor_id filter if provided
	if sensorID > 0 {
		query = query.Where("sensor_id = ?", sensorID)
	}

	err := query.Find(&data).Error
	return data, err
}

func (r *repository) GetBySensorID(sensorID uint) ([]model.SensorData, error) {
	var data []model.SensorData
	err := r.db.Where("sensor_id = ?", sensorID).Order("time DESC").Find(&data).Error
	return data, err
}

func (r *repository) GetLatestBySensorID(sensorID uint) (*model.SensorData, error) {
	var data model.SensorData
	err := r.db.Where("sensor_id = ?", sensorID).Order("time DESC").First(&data).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &data, err
}

// Sensor Master CRUD implementations
func (r *repository) CreateSensor(sensor *model.MSensor) error {
	return r.db.Create(sensor).Error
}

func (r *repository) GetSensorByID(id uint) (*model.MSensor, error) {
	var sensor model.MSensor
	err := r.db.First(&sensor, id).Error
	if err != nil {
		return nil, err
	}
	return &sensor, nil
}

func (r *repository) GetSensorByTopic(topic string) (*model.MSensor, error) {
	var sensor model.MSensor
	err := r.db.Where("topic = ?", topic).First(&sensor).Error
	if err != nil {
		return nil, err
	}
	return &sensor, nil
}

func (r *repository) GetAllSensors() ([]model.MSensor, error) {
	var sensors []model.MSensor
	err := r.db.Order("id ASC").Find(&sensors).Error
	return sensors, err
}

func (r *repository) UpdateSensor(sensor *model.MSensor) error {
	return r.db.Save(sensor).Error
}

func (r *repository) DeleteSensor(id uint) error {
	// Get the record first
	var sensor model.MSensor
	if err := r.db.First(&sensor, id).Error; err != nil {
		return err
	}

	// Add suffix to topic and topic_key before soft delete to release constraints
	suffix := utils.GenerateDeletedSuffix()
	sensor.Topic = sensor.Topic + suffix
	sensor.TopicKey = sensor.TopicKey + suffix
	if err := r.db.Save(&sensor).Error; err != nil {
		return err
	}

	return r.db.Delete(&sensor).Error
}

func (r *repository) GetSensorsByVesselID(vesselID uint) ([]model.MSensor, error) {
	var sensors []model.MSensor
	err := r.db.Joins("JOIN m_mcu ON m_sensors.mcu_id = m_mcu.id").
		Where("m_mcu.vessel_id = ?", vesselID).
		Preload("MCU").
		Find(&sensors).Error
	return sensors, err
}

// Helper methods implementations
func (r *repository) GetMCUByID(mcuID uint) (*model.MMCU, error) {
	var mcu model.MMCU
	err := r.db.Preload("Vessel").First(&mcu, mcuID).Error
	if err != nil {
		return nil, err
	}
	return &mcu, nil
}

func (r *repository) GetVesselByID(vesselID uint) (*model.MVessel, error) {
	var vessel model.MVessel
	err := r.db.First(&vessel, vesselID).Error
	if err != nil {
		return nil, err
	}
	return &vessel, nil
}

func (r *repository) GetDB() *gorm.DB {
	return r.db
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}
