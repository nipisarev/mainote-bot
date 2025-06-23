-- ================================
-- Migration: V1__Initial_schema.sql
-- Description: Initial database schema with users and user_settings tables
-- Author: Generated from Alembic migration 20240601_initial
-- Date: 2025-06-10
-- ================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NULL
);

CREATE TABLE user_settings (
    user_settings_id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    chat_id TEXT NOT NULL,
    morning_notification_time TEXT,
    timezone TEXT,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NULL
);

-- Create trigger to automatically update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_settings_updated_at 
    BEFORE UPDATE ON user_settings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();