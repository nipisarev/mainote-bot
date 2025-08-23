-- ================================
-- Migration: V5__Add_unique_notification_constraint.sql
-- Description: Add unique constraint to prevent duplicate notifications for same user at same time
-- Author: System Migration
-- Date: 2025-01-27
-- ================================

-- Add unique constraint to prevent duplicate notifications for same user at same time
-- This prevents race conditions and duplicate notifications
ALTER TABLE scheduled_notifications 
ADD CONSTRAINT unique_user_notification_time 
UNIQUE (user_id, send_at);

-- Create index to support the constraint efficiently
CREATE INDEX idx_scheduled_notifications_user_send_at ON scheduled_notifications(user_id, send_at); 