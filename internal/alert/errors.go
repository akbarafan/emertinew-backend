package alert

import "errors"

// Validation errors, mapped to 400 by the handler.
var (
	ErrMissingEventKey    = errors.New("event_key is required")
	ErrDuplicateEventKey  = errors.New("duplicate event_key in the same batch")
	ErrMissingDetectedAt  = errors.New("detected_at is required")
	ErrDetectedAtInFuture = errors.New("detected_at is in the future; check the client clock")
	ErrInvalidSeverity    = errors.New("severity must be one of info, warning, error")
	ErrInvalidDate        = errors.New("from and to must be YYYY-MM-DD")
)
