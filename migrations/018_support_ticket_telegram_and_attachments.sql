-- Migration: 018_support_ticket_telegram_and_attachments.sql
-- Description: Add Telegram integration fields, operational status, and attachment URLs to support tickets & replies

ALTER TABLE t_support_tickets
    ADD COLUMN IF NOT EXISTS telegram_message_id INTEGER,
    ADD COLUMN IF NOT EXISTS telegram_chat_id BIGINT,
    ADD COLUMN IF NOT EXISTS vessel_status VARCHAR(30) DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS port_location VARCHAR(100),
    ADD COLUMN IF NOT EXISTS attachment_url VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_t_support_tickets_telegram_msg_id ON t_support_tickets(telegram_message_id);

ALTER TABLE t_ticket_replies
    ADD COLUMN IF NOT EXISTS attachment_url VARCHAR(255),
    ADD COLUMN IF NOT EXISTS telegram_message_id INTEGER;
