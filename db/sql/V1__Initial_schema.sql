-- ================================
-- Migration: V1__Initial_schema.sql
-- Description: Initial database schema with user_preferences table
-- Author: Generated from Alembic migration 20240601_initial
-- Date: 2025-06-10
-- ================================

-- Create user_preferences table
CREATE TABLE user_preferences (
    chat_id TEXT NOT NULL,
    notification_time TEXT,
    timezone TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (chat_id)
);

-- Create indexes for better performance
CREATE INDEX idx_user_preferences_notification_time ON user_preferences(notification_time) WHERE notification_time IS NOT NULL;
CREATE INDEX idx_user_preferences_timezone ON user_preferences(timezone) WHERE timezone IS NOT NULL;

-- Create trigger to automatically update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_preferences_updated_at 
    BEFORE UPDATE ON user_preferences 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE user_preferences IS 'Stores user preferences for Telegram bot users';
COMMENT ON COLUMN user_preferences.chat_id IS 'Telegram chat ID (primary key)';
COMMENT ON COLUMN user_preferences.notification_time IS 'Preferred notification time in HH:MM format';
COMMENT ON COLUMN user_preferences.timezone IS 'User timezone (e.g., Europe/Moscow)';
COMMENT ON COLUMN user_preferences.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN user_preferences.updated_at IS 'Timestamp when record was last updated';
