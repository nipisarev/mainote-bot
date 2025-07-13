-- ================================
-- Migration: V4__Add_scheduled_notifications_table.sql
-- Description: Add scheduled_notifications table for delayed message delivery
-- Author: System Migration
-- Date: 2025-01-27
-- ================================

CREATE TABLE scheduled_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    message TEXT NOT NULL,
    send_at TIMESTAMP WITH TIME ZONE NOT NULL,
    delivered BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_scheduled_notifications_user_id ON scheduled_notifications(user_id);
CREATE INDEX idx_scheduled_notifications_send_at ON scheduled_notifications(send_at);
CREATE INDEX idx_scheduled_notifications_delivered ON scheduled_notifications(delivered);
CREATE INDEX idx_scheduled_notifications_pending ON scheduled_notifications(send_at, delivered) WHERE delivered = false;

-- Create trigger to automatically update updated_at column
CREATE TRIGGER update_scheduled_notifications_updated_at
    BEFORE UPDATE ON scheduled_notifications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 