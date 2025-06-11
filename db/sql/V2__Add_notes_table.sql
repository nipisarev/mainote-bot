-- ================================
-- Migration: V2__Add_notes_table.sql
-- Description: Add notes table for storing user notes from Telegram bot
-- Author: System Migration
-- Date: 2025-06-10
-- ================================

-- Create notes table
CREATE TABLE notes (
    id SERIAL PRIMARY KEY,
    chat_id TEXT NOT NULL,
    content TEXT NOT NULL,
    notion_page_id TEXT,
    voice_file_id TEXT,
    transcription TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint
    CONSTRAINT fk_notes_chat_id 
        FOREIGN KEY (chat_id) 
        REFERENCES user_preferences(chat_id) 
        ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX idx_notes_chat_id ON notes(chat_id);
CREATE INDEX idx_notes_created_at ON notes(created_at DESC);
CREATE INDEX idx_notes_notion_page_id ON notes(notion_page_id) WHERE notion_page_id IS NOT NULL;
CREATE INDEX idx_notes_voice_file_id ON notes(voice_file_id) WHERE voice_file_id IS NOT NULL;

-- Create full-text search index for content
CREATE INDEX idx_notes_content_fts ON notes USING gin(to_tsvector('russian', content));

-- Create trigger to automatically update updated_at column
CREATE TRIGGER update_notes_updated_at 
    BEFORE UPDATE ON notes 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE notes IS 'User notes stored from Telegram bot with Notion integration';
COMMENT ON COLUMN notes.id IS 'Unique identifier for the note';
COMMENT ON COLUMN notes.chat_id IS 'Telegram chat ID (foreign key to user_preferences)';
COMMENT ON COLUMN notes.content IS 'Note content in text format';
COMMENT ON COLUMN notes.notion_page_id IS 'Notion page ID if note was saved to Notion';
COMMENT ON COLUMN notes.voice_file_id IS 'Telegram voice file ID if note was created from voice message';
COMMENT ON COLUMN notes.transcription IS 'Voice message transcription from OpenAI Whisper';
COMMENT ON COLUMN notes.created_at IS 'Timestamp when note was created';
COMMENT ON COLUMN notes.updated_at IS 'Timestamp when note was last updated';
