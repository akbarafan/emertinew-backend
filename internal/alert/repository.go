package alert

import (
	"time"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"gorm.io/gorm"
)

// Repository is the persistence boundary for the alert audit trail.
type Repository interface {
	// UpsertRecords inserts a transition that was not seen before and, when
	// `complete` is set, resolves every open occurrence the client no longer
	// reports.
	//
	// Doing both in one call keeps the client's request to a single POST: it
	// sends what is currently true, and the server works out what opened and
	// what closed.
	UpsertRecords(events []UpsertEventRequest, keysStillPresent map[string]bool, complete bool) (created, resolved int, err error)

	ListEvents(query ListEventsQuery) ([]model.MAlertEvent, int, int, error)
	AckEvent(id uint, at time.Time) (*model.MAlertEvent, error)
	GetEventByID(id uint) (*model.MAlertEvent, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) UpsertRecords(
	events []UpsertEventRequest,
	keysStillPresent map[string]bool,
	complete bool,
) (int, int, error) {
	if len(events) == 0 && !complete {
		// Nothing reported and the client has not declared the batch
		// authoritative, so there is nothing to write and nothing to close.
		return 0, 0, nil
	}

	created := 0
	resolved := 0
	now := time.Now()

	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, e := range events {
			var existing model.MAlertEvent
			err := tx.Where("event_key = ?", e.EventKey).First(&existing).Error

			switch {
			case err == nil:
				// Already known. Refresh the message and severity so a condition
				// that worsens (warning -> error) is reflected, but keep the
				// original detection time: the point of the table is when it
				// started, not when we last looked.
				updates := map[string]interface{}{
					"message":     e.Message,
					"severity":    e.Severity,
					"vessel_name": e.VesselName,
					"updated_at":  now,
				}
				if existing.IsOpen() {
					updates["resolved_at"] = nil
				}
				if err := tx.Model(&existing).Updates(updates).Error; err != nil {
					return err
				}

			case err == gorm.ErrRecordNotFound:
				row := model.MAlertEvent{
					EventKey:   e.EventKey,
					VesselID:   e.VesselID,
					VesselName: e.VesselName,
					Type:       e.Type,
					Severity:   e.Severity,
					Message:    e.Message,
					DetectedAt: e.DetectedAt,
					ResolvedAt: e.ResolvedAt,
				}
				if err := tx.Create(&row).Error; err != nil {
					return err
				}
				created++

			default:
				return err
			}
		}

		// Close anything open that the client stopped reporting. A vessel that
		// drops off the network produces no "all clear" event, so this is the
		// only place stale_data rows ever get closed.
		if complete {
			openQuery := tx.Model(&model.MAlertEvent{}).
				Where("resolved_at IS NULL")
			if len(keysStillPresent) > 0 {
				openQuery = openQuery.Where("event_key NOT IN ?", keysToSlice(keysStillPresent))
			}

			var open []model.MAlertEvent
			if err := openQuery.Find(&open).Error; err != nil {
				return err
			}
			if len(open) > 0 {
				ids := make([]uint, 0, len(open))
				for _, o := range open {
					ids = append(ids, o.ID)
				}
				if err := tx.Model(&model.MAlertEvent{}).
					Where("id IN ?", ids).
					Updates(map[string]interface{}{"resolved_at": now, "updated_at": now}).Error; err != nil {
					return err
				}
				resolved = len(ids)
			}
		}
		return nil
	})

	return created, resolved, err
}

// ListEvents returns one page of history, the unpaginated total for that filter,
// and how many alerts are open fleet-wide (the header counter).
func (r *repository) ListEvents(query ListEventsQuery) ([]model.MAlertEvent, int, int, error) {
	tx := r.db.Model(&model.MAlertEvent{})

	if query.VesselID > 0 {
		tx = tx.Where("vessel_id = ?", query.VesselID)
	}
	if query.Type != "" {
		tx = tx.Where("type = ?", query.Type)
	}
	if query.Severity != "" {
		tx = tx.Where("severity = ?", query.Severity)
	}
	if query.OpenOnly {
		tx = tx.Where("resolved_at IS NULL")
	}
	if !query.FromTime.IsZero() {
		tx = tx.Where("detected_at >= ?", query.FromTime)
	}
	if !query.ToTime.IsZero() {
		tx = tx.Where("detected_at < ?", query.ToTime)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	limit := query.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	page := query.Page
	if page <= 0 {
		page = 1
	}

	var rows []model.MAlertEvent
	if err := tx.Order("detected_at DESC, id DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, 0, 0, err
	}

	var openCount int64
	if err := r.db.Model(&model.MAlertEvent{}).Where("resolved_at IS NULL").Count(&openCount).Error; err != nil {
		return nil, 0, 0, err
	}

	return rows, int(total), int(openCount), nil
}

func (r *repository) AckEvent(id uint, at time.Time) (*model.MAlertEvent, error) {
	var row model.MAlertEvent
	if err := r.db.Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	if row.AckAt == nil {
		if err := r.db.Model(&row).Update("ack_at", at).Error; err != nil {
			return nil, err
		}
		row.AckAt = &at
	}
	return &row, nil
}

func (r *repository) GetEventByID(id uint) (*model.MAlertEvent, error) {
	var row model.MAlertEvent
	if err := r.db.Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func keysToSlice(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
