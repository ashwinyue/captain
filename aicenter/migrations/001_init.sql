-- AI Center Database Schema
-- Compatible with tgo-ai Python version

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- LLM Providers table
CREATE TABLE IF NOT EXISTS ai_llm_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    provider_kind VARCHAR(50) NOT NULL,  -- openai, ark, anthropic, google, openai_compatible
    api_key VARCHAR(500),
    api_base_url VARCHAR(500),
    model VARCHAR(255),
    organization VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    timeout FLOAT DEFAULT 60,
    config JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_llm_providers_project_id ON ai_llm_providers(project_id);
CREATE INDEX idx_llm_providers_deleted_at ON ai_llm_providers(deleted_at);

-- Teams table
CREATE TABLE IF NOT EXISTS ai_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    supervisor_llm_id UUID REFERENCES ai_llm_providers(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    supervisor_instruction TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    is_enabled BOOLEAN DEFAULT TRUE,
    config JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_teams_project_id ON ai_teams(project_id);
CREATE INDEX idx_teams_deleted_at ON ai_teams(deleted_at);

-- Agents table
CREATE TABLE IF NOT EXISTS ai_agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    team_id UUID REFERENCES ai_teams(id) ON DELETE SET NULL,
    llm_provider_id UUID REFERENCES ai_llm_providers(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    instruction TEXT,
    model VARCHAR(255),
    is_default BOOLEAN DEFAULT FALSE,
    is_enabled BOOLEAN DEFAULT TRUE,
    config JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_agents_project_id ON ai_agents(project_id);
CREATE INDEX idx_agents_team_id ON ai_agents(team_id);
CREATE INDEX idx_agents_deleted_at ON ai_agents(deleted_at);

-- Agent Tools table
CREATE TABLE IF NOT EXISTS ai_agent_tools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES ai_agents(id) ON DELETE CASCADE,
    tool_provider VARCHAR(100) NOT NULL,  -- mcp, rag, builtin
    tool_name VARCHAR(255) NOT NULL,
    is_enabled BOOLEAN DEFAULT TRUE,
    config JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_agent_tools_agent_id ON ai_agent_tools(agent_id);
CREATE INDEX idx_agent_tools_deleted_at ON ai_agent_tools(deleted_at);

-- Agent Collections table (RAG collections)
CREATE TABLE IF NOT EXISTS ai_agent_collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES ai_agents(id) ON DELETE CASCADE,
    collection_id VARCHAR(255) NOT NULL,
    is_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_agent_collections_agent_id ON ai_agent_collections(agent_id);
CREATE INDEX idx_agent_collections_deleted_at ON ai_agent_collections(deleted_at);

-- Updated at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers
CREATE TRIGGER update_ai_llm_providers_updated_at BEFORE UPDATE ON ai_llm_providers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ai_teams_updated_at BEFORE UPDATE ON ai_teams FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ai_agents_updated_at BEFORE UPDATE ON ai_agents FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ai_agent_tools_updated_at BEFORE UPDATE ON ai_agent_tools FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ai_agent_collections_updated_at BEFORE UPDATE ON ai_agent_collections FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
