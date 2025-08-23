-- ================================
-- Migration: V6__Add_task_fields_to_note.sql
-- Description: Add due_at, effort_min, and priority fields to note table
-- Author: Automated Migration
-- Date: 2025-08-18
-- ================================

ALTER TABLE note
  ADD COLUMN IF NOT EXISTS due_at TIMESTAMP NULL,
  ADD COLUMN IF NOT EXISTS effort_min INTEGER NOT NULL DEFAULT 30,
  ADD COLUMN IF NOT EXISTS priority SMALLINT NOT NULL DEFAULT 0;

COMMENT ON COLUMN note.due_at IS
  'Optional deadline for actionable notes (category=''task''). Stored as UTC timestamp. Used for naive sorting in /brief (overdue/today/other).';

COMMENT ON COLUMN note.effort_min IS
  'Estimated effort in minutes for planning hints (default 30). Displayed in brief; also used as a tiebreaker in ordering.';

COMMENT ON COLUMN note.priority IS
  'Naive integer priority; higher is more important. Used as a tiebreaker in ordering.';
