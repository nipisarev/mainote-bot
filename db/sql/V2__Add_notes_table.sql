-- ================================
-- Migration: V2__Add_notes_table.sql
-- Description: Add notes table for storing user notes from Telegram bot
-- Author: System Migration
-- Date: 2025-06-10
-- ================================

-- Create notes table
CREATE TABLE notes (
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
    created_at TIMESTAMPT WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPT WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPT WITHOUT TIME ZONE DEFAULT NULL
);

-- Create indexes for better performance
CREATE INDEX idx_notes_uuid_id ON notes(uuid_id);
CREATE INDEX idx_notes_chat_id ON notes(chat_id);
CREATE INDEX idx_notes_notion_page_id ON notes(notion_page_id) WHERE notion_page_id IS NOT NULL;

-- Create trigger to automatically update updated_at column
CREATE TRIGGER update_notes_updated_at 
    BEFORE UPDATE ON notes 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
