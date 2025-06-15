-- ================================
-- Migration: V3__Update_notes_schema_for_api.sql
-- Description: Update notes table to support new API fields (category, status, source, metadata)
-- Author: System Migration
-- Date: 2025-06-15
-- ================================

-- Add UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Add new columns to notes table
ALTER TABLE notes 
ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT 'idea',
ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active',
ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'telegram',
ADD COLUMN IF NOT EXISTS metadata JSONB,
ADD COLUMN IF NOT EXISTS title TEXT;

-- Change id column from SERIAL to UUID
-- First, add a new UUID column
ALTER TABLE notes ADD COLUMN IF NOT EXISTS uuid_id UUID DEFAULT uuid_generate_v4();

-- Update existing rows to have UUID values
UPDATE notes SET uuid_id = uuid_generate_v4() WHERE uuid_id IS NULL;

-- Make uuid_id NOT NULL and add unique constraint
ALTER TABLE notes 
ALTER COLUMN uuid_id SET NOT NULL;

-- Create indexes on new columns (only if they don't exist)
CREATE INDEX IF NOT EXISTS idx_notes_category ON notes(category);
CREATE INDEX IF NOT EXISTS idx_notes_status ON notes(status);
CREATE INDEX IF NOT EXISTS idx_notes_source ON notes(source);
CREATE INDEX IF NOT EXISTS idx_notes_uuid_id ON notes(uuid_id);
CREATE INDEX IF NOT EXISTS idx_notes_metadata ON notes USING gin(metadata) WHERE metadata IS NOT NULL;

-- Add comments for new columns
COMMENT ON COLUMN notes.category IS 'Note category: idea, task, personal';
COMMENT ON COLUMN notes.status IS 'Note status: active, archived, deleted';
COMMENT ON COLUMN notes.source IS 'Note source: telegram_bot, text, voice';
COMMENT ON COLUMN notes.metadata IS 'Additional metadata in JSON format';
COMMENT ON COLUMN notes.uuid_id IS 'UUID identifier for API compatibility';
COMMENT ON COLUMN notes.title IS 'Optional title for the note';