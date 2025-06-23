-- ================================
-- Migration: V2__Add_note_table.sql
-- Description: Add note table for storing user note from Telegram bot
-- Author: System Migration
-- Date: 2025-06-10
-- ================================

-- Create note table
CREATE TABLE note (
    note_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chat_id TEXT NOT NULL,
    user_id UUID NOT NULL,
    title TEXT,
    content TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'idea',
    status TEXT NOT NULL DEFAULT 'active', 
    source TEXT NOT NULL DEFAULT 'telegram',
    voice_file_id TEXT,
    transcription TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NULL
);

-- Create indexes for better performance
CREATE INDEX idx_note_uuid_id ON note(user_id);
CREATE INDEX idx_note_chat_id ON note(chat_id);

-- Create trigger to automatically update updated_at column
CREATE TRIGGER update_note_updated_at 
    BEFORE UPDATE ON note 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
