package audit

import (
	"encoding/json"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

type Service interface {
	Log(userID uint, tableName string, recordID uint, action string, oldData, newData interface{}, ipAddress, userAgent string) error
	GetHistory(tableName string, recordID *uint, action string, limit, offset int) ([]AuditLogResponse, int64, error)
	GetUserActivity(userID uint, limit int) ([]AuditLogResponse, error)
	GetCreator(tableName string, recordID uint) (*uint, error)
	WithTx(tx *gorm.DB) Service
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) WithTx(tx *gorm.DB) Service {
	return &service{repo: s.repo.WithTx(tx)}
}

func (s *service) Log(userID uint, tableName string, recordID uint, action string, oldData, newData interface{}, ipAddress, userAgent string) error {
	var oldJSON, newJSON *string

	if oldData != nil {
		data, err := json.Marshal(oldData)
		if err == nil {
			str := string(data)
			oldJSON = &str
		}
	}

	if newData != nil {
		data, err := json.Marshal(newData)
		if err == nil {
			str := string(data)
			newJSON = &str
		}
	}

	// Handle userID = 0 as NULL (e.g. failed login with unknown user)
	var userIDPtr *uint
	if userID > 0 {
		userIDPtr = &userID
	}

	auditLog := &model.AuditLog{
		UserID:     userIDPtr,
		AuditTable: tableName,
		RecordID:   recordID,
		Action:     action,
		OldValue:   oldJSON,
		NewValue:   newJSON,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}

	return s.repo.Create(auditLog)
}

func (s *service) GetHistory(tableName string, recordID *uint, action string, limit, offset int) ([]AuditLogResponse, int64, error) {
	logs, total, err := s.repo.GetByTableAndRecord(tableName, recordID, action, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return mapToResponses(logs), total, nil
}

func (s *service) GetUserActivity(userID uint, limit int) ([]AuditLogResponse, error) {
	logs, err := s.repo.GetByUser(userID, limit)
	if err != nil {
		return nil, err
	}
	return mapToResponses(logs), nil
}

func (s *service) GetCreator(tableName string, recordID uint) (*uint, error) {
	return s.repo.GetCreator(tableName, recordID)
}

// mapToResponses converts AuditLog models to AuditLogResponse with user name
func mapToResponses(logs []model.AuditLog) []AuditLogResponse {
	responses := make([]AuditLogResponse, len(logs))
	for i, l := range logs {
		var uid uint
		if l.UserID != nil {
			uid = *l.UserID
		}
		responses[i] = AuditLogResponse{
			ID:        l.ID,
			UserID:    uid,
			UserName:  l.User.Username,
			TableName: l.AuditTable,
			RecordID:  l.RecordID,
			Action:    l.Action,
			OldValue:  toRawMessage(l.OldValue),
			NewValue:  toRawMessage(l.NewValue),
			IPAddress: l.IPAddress,
			UserAgent: l.UserAgent,
			CreatedAt: l.CreatedAt,
		}
	}
	return responses
}

// toRawMessage converts *string (JSON text) to json.RawMessage so it won't be double-encoded
func toRawMessage(s *string) json.RawMessage {
	if s == nil {
		return nil
	}
	return json.RawMessage(*s)
}
