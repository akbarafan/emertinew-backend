-- Alert history.
--
-- The command center derives alerts client-side from polling telemetry (high
-- temperature, door open, AC lost, stale data, geofence violation...). Those
-- derivations were previously thrown away every poll: `setMapEvents(events)`
-- replaced the whole array, so an alert that appeared and cleared between two
-- polls left no trace, and every surviving row carried the render timestamp
-- rather than the moment the condition was detected.
--
-- This table makes the feed an audit trail. The rules stay in one place — the
-- frontend — which POSTs an event only on the absent -> present transition.
-- `detected_at` is therefore the last tracking timestamp that showed the
-- condition, not the time the browser happened to poll.

CREATE TABLE IF NOT EXISTS m_alert_events (
    id          BIGSERIAL PRIMARY KEY,
    -- Stable id derived from the rule plus the vessel, e.g. `temp-alert-12`.
    -- Client supplies it so a retried POST is idempotent instead of duplicating.
    event_key   VARCHAR(120) NOT NULL,
    vessel_id   INTEGER      NOT NULL,
    vessel_name VARCHAR(150) NOT NULL,
    type        VARCHAR(40)  NOT NULL,
    severity    VARCHAR(10)  NOT NULL,
    message     TEXT         NOT NULL,

    -- When the condition was first observed, taken from the tracking record.
    detected_at TIMESTAMPTZ  NOT NULL,
    -- When a later poll showed the condition had cleared. NULL while ongoing.
    resolved_at TIMESTAMPTZ  NULL,
    -- Operator acknowledgement, independent of resolution.
    ack_at      TIMESTAMPTZ  NULL,

    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT m_alert_events_severity_chk
        CHECK (severity IN ('info', 'warning', 'error')),
    -- Same key twice only makes sense while the first one is still open; once
    -- resolved a fresh occurrence is a new row with a new detection time.
    CONSTRAINT m_alert_events_unique_open UNIQUE (event_key)
);

-- Primary read pattern: one vessel's feed, newest first.
CREATE INDEX IF NOT EXISTS idx_alert_events_vessel_detected
    ON m_alert_events (vessel_id, detected_at DESC);

-- Dashboard filter: severity over a time window.
CREATE INDEX IF NOT EXISTS idx_alert_events_severity_detected
    ON m_alert_events (severity, detected_at DESC);

-- Type breakdown for the widget tabs.
CREATE INDEX IF NOT EXISTS idx_alert_events_type_detected
    ON m_alert_events (type, detected_at DESC);

COMMENT ON TABLE  m_alert_events IS 'Client-derived alert audit trail; detection rules live in the frontend';
COMMENT ON COLUMN m_alert_events.event_key   IS 'Client-derived stable key, used for idempotent upsert';
COMMENT ON COLUMN m_alert_events.detected_at IS 'Tracking timestamp that showed the condition, not the poll time';
