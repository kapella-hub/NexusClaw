export interface User {
  id: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface Session {
  id: string;
  user_id: string;
  token?: string;
  expires_at: string;
  created_at: string;
}

export interface VaultEntry {
  id: string;
  user_id: string;
  provider: string;
  created_at: string;
  updated_at: string;
}

export interface Credential {
  provider: string;
  access_token?: string;
  secret?: string;
}

export type ServerStatus = "stopped" | "starting" | "running" | "stopping" | "error";

export interface MCPServer {
  id: string;
  owner_id: string;
  name: string;
  image: string;
  status: ServerStatus;
  config: Record<string, unknown>;
  container_id?: string;
  tools?: any[];
  resources?: any[];
  created_at: string;
  updated_at: string;
}

export type RuleAction = "block" | "allow" | "alert";

export interface Rule {
  id: string;
  name: string;
  description?: string;
  pattern: string;
  action: RuleAction;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface AuditEntry {
  id: string;
  user_id?: string;
  action: string;
  resource?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export type BudgetPeriod = "daily" | "weekly" | "monthly";

export interface BudgetCap {
  id: string;
  user_id: string;
  period: BudgetPeriod;
  max_tokens: number;
  used_tokens: number;
  reset_at: string;
  created_at: string;
}

export interface ApiError {
  error: string;
}
