-- ================================
-- Migration: V3__Add_integrations_table.sql
-- Description: Add integrations table for external applications (Notion, TickTick, Jira, etc.) with flexible auth and config
-- Author: System Migration
-- Date: 2025-06-15
-- ================================

CREATE TABLE app (
    app_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider TEXT NOT NULL, -- 'notion', 'ticktick', 'jira', 'todoist', 'asana', etc.
    name TEXT NOT NULL, -- User-friendly name for this integration
    description TEXT, -- Optional description
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NULL
);

CREATE TABLE integration (
    integration_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_id UUID NOT NULL,
    chat_id TEXT NOT NULL, -- Changed from user_id to chat_id to match project conventions
    status TEXT NOT NULL DEFAULT 'active', -- 'active', 'inactive', 'error', 'pending_auth'
    config JSONB NOT NULL DEFAULT '{}',
    auth_type TEXT NOT NULL, -- 'oauth2', 'api_key', 'basic_auth', 'token', 'custom'
    auth_data JSONB NOT NULL, -- Provider-specific settings
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NULL
);

CREATE TABLE note_integration (
    note_integration_id SERIAL PRIMARY KEY,
    note_id UUID NOT NULL, -- Changed from INTEGER to UUID to match notes.note_id
    integration_id UUID NOT NULL,
    external_id TEXT NOT NULL, -- ID/LINK in the external system
    sync_status TEXT NOT NULL DEFAULT 'synced', -- 'synced', 'pending', 'error', 'conflict'
    synced_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    retry_count INTEGER DEFAULT 0,
    last_error TEXT DEFAULT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NULL
);


-- Create indexes for better performance
CREATE INDEX idx_app_provider ON app(provider);
CREATE INDEX idx_integration_chat_id ON integration(chat_id);
CREATE INDEX idx_integration_app_id ON integration(app_id);
CREATE INDEX idx_integration_status ON integration(status);
CREATE INDEX idx_note_integration_note_id ON note_integration(note_id);
CREATE INDEX idx_note_integration_integration_id ON note_integration(integration_id);
CREATE INDEX idx_note_integration_sync_status ON note_integration(sync_status);

-- Create triggers to automatically update updated_at columns
CREATE TRIGGER update_app_updated_at
    BEFORE UPDATE ON app
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_integration_updated_at
    BEFORE UPDATE ON integration
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_note_integration_updated_at
    BEFORE UPDATE ON note_integration
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default Notion integration app
INSERT INTO app (provider, name, description) VALUES
(
    'notion',
    'Notion',
    '## 🧠 Notion Integration: Notes & Tasks Sync

    This app connects to your Notion workspace to store and manage your notes and tasks. It enables automatic syncing between Telegram input and a structured Notion database, helping you organize ideas, prioritize work, and plan your day.

    ### 🔐 Setup Instructions

    1. **Create Notion Integration Token**
       - Go to [https://www.notion.so/my-integrations](https://www.notion.so/my-integrations)
       - Click **"New integration"**
       - Give it a name (e.g., `Mainote Assistant`)
       - Copy the **Internal Integration Token**

    2. **Create a Database in Notion**
       - Create a new database (Table view recommended)
       - Add the following required fields (exact field names):
         - `Title` – (Title field, default)
         - `Content` – (Text)
         - `Status` – (Select, values: `Inbox`, `Ready`, `In Progress`, `Done`)
         - `Category` – (Select, values: `Task`, `Idea`, `Personal`, `Note`)
         - `Created Time` – (Created time, auto-generated)

    3. **Share the Database with the Integration**
       - Open the database in Notion
       - Click **Share** → **Invite**
       - Select your integration (e.g., `Mainote Assistant`) and click **Invite**

    4. **Provide These Details to the App**
       - Your **Notion Integration Token**
       - Your **Database ID** (found in the URL: `https://www.notion.so/yourworkspace/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx?v=...`)

    ### ✅ After Setup
    Once configured, all your voice/text notes sent through the assistant will be automatically structured and stored in Notion. You can view, filter, and manage them directly in your Notion workspace.'
);
