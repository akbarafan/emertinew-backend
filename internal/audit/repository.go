package audit

import (
	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(auditLog *model.AuditLog) error
	GetByTableAndRecord(tableName string, recordID *uint, action string, limit, offset int) ([]model.AuditLog, int64, error)
	GetByUser(userID uint, limit int) ([]model.AuditLog, error)
	GetCreator(tableName string, recordID uint) (*uint, error)
	WithTx(tx *gorm.DB) Repository
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) WithTx(tx *gorm.DB) Repository {
	return &repository{db: tx}
}

func (r *repository) Create(auditLog *model.AuditLog) error {
	return r.db.Create(auditLog).Error
}

func (r *repository) GetByTableAndRecord(tableName string, recordID *uint, action string, limit, offset int) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	// Build dynamic where clause
	baseQuery := r.db.Model(&model.AuditLog{}).Where("table_name = ?", tableName)
	if recordID != nil {
		baseQuery = baseQuery.Where("record_id = ?", *recordID)
	}
	if action != "" {
		baseQuery = baseQuery.Where("action = ?", action)
	}

	// Count total records
	baseQuery.Count(&total)

	query := r.db.Preload("User").Where("table_name = ?", tableName)
	if recordID != nil {
		query = query.Where("record_id = ?", *recordID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	query = query.Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&logs).Error
	return logs, total, err
}

func (r *repository) GetByUser(userID uint, limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	query := r.db.Preload("User").Where("user_id = ?", userID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

func (r *repository) GetCreator(tableName string, recordID uint) (*uint, error) {
	var log model.AuditLog
	err := r.db.Where("table_name = ? AND record_id = ? AND action = ?",
		tableName, recordID, "CREATE").
		First(&log).Error

	if err != nil {
		return nil, err
	}
	return log.UserID, nil
}
