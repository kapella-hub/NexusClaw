ALTER TABLE mcp_servers ADD COLUMN tools JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE mcp_servers ADD COLUMN resources JSONB NOT NULL DEFAULT '[]'::jsonb;
