import type {
  ApiError,
  AuditEntry,
  BudgetCap,
  Credential,
  MCPServer,
  Rule,
  Session,
  VaultEntry,
} from "./types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const TOKEN_KEY = "nexusclaw_token";

// ---------------------------------------------------------------------------
// Token helpers
// ---------------------------------------------------------------------------

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

// ---------------------------------------------------------------------------
// Base fetch wrapper
// ---------------------------------------------------------------------------

export class ApiRequestError extends Error {
  status: number;
  body: ApiError;

  constructor(status: number, body: ApiError) {
    super(body.error);
    this.name = "ApiRequestError";
    this.status = status;
    this.body = body;
  }
}

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  const token = getToken();
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      ...headers,
      ...options?.headers,
    },
  });

  if (res.status === 204) {
    return undefined as T;
  }

  const body = await res.json();

  if (!res.ok) {
    throw new ApiRequestError(res.status, body as ApiError);
  }

  return body as T;
}

// ---------------------------------------------------------------------------
// Auth (pass)
// ---------------------------------------------------------------------------

export async function register(
  email: string,
  password: string,
): Promise<{ id: string; email: string }> {
  return apiFetch("/api/v1/pass/register", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export async function login(
  email: string,
  password: string,
): Promise<Session> {
  return apiFetch("/api/v1/pass/sessions", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export async function logout(sessionId: string): Promise<void> {
  return apiFetch(`/api/v1/pass/sessions/${sessionId}`, {
    method: "DELETE",
  });
}

// ---------------------------------------------------------------------------
// Vault (pass)
// ---------------------------------------------------------------------------

export async function listVault(): Promise<VaultEntry[]> {
  return apiFetch("/api/v1/pass/vault");
}

export async function createVaultEntry(
  provider: string,
  access_token?: string,
  secret?: string,
): Promise<VaultEntry> {
  const cred: Credential = { provider, access_token, secret };
  return apiFetch("/api/v1/pass/vault", {
    method: "POST",
    body: JSON.stringify(cred),
  });
}

export async function deleteVaultEntry(id: string): Promise<void> {
  return apiFetch(`/api/v1/pass/vault/${id}`, {
    method: "DELETE",
  });
}

// ---------------------------------------------------------------------------
// Servers (nodes)
// ---------------------------------------------------------------------------

export async function listServers(): Promise<MCPServer[]> {
  return apiFetch("/api/v1/nodes/");
}

export async function createServer(
  name: string,
  image: string,
  config?: Record<string, unknown>,
): Promise<MCPServer> {
  return apiFetch("/api/v1/nodes/", {
    method: "POST",
    body: JSON.stringify({ name, image, config: config ?? {} }),
  });
}

export async function getServer(id: string): Promise<MCPServer> {
  return apiFetch(`/api/v1/nodes/${id}`);
}

export async function deleteServer(id: string): Promise<void> {
  return apiFetch(`/api/v1/nodes/${id}`, {
    method: "DELETE",
  });
}

export async function startServer(
  id: string,
): Promise<{ status: string }> {
  return apiFetch(`/api/v1/nodes/${id}/start`, {
    method: "POST",
  });
}

export async function stopServer(
  id: string,
): Promise<{ status: string }> {
  return apiFetch(`/api/v1/nodes/${id}/stop`, {
    method: "POST",
  });
}

export async function discoverServers(
  query?: string,
): Promise<MCPServer[]> {
  const q = query ? `?q=${encodeURIComponent(query)}` : "";
  return apiFetch(`/api/v1/nodes/discover${q}`);
}

// ---------------------------------------------------------------------------
// Sentry (rules, audit, budget)
// ---------------------------------------------------------------------------

export async function listRules(): Promise<Rule[]> {
  return apiFetch("/api/v1/sentry/rules");
}

export async function createRule(
  data: Omit<Rule, "id" | "created_at" | "updated_at">,
): Promise<Rule> {
  return apiFetch("/api/v1/sentry/rules", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateRule(
  id: string,
  data: Partial<Omit<Rule, "id" | "created_at" | "updated_at">>,
): Promise<Rule> {
  return apiFetch(`/api/v1/sentry/rules/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteRule(id: string): Promise<void> {
  return apiFetch(`/api/v1/sentry/rules/${id}`, {
    method: "DELETE",
  });
}

export async function listAudit(): Promise<AuditEntry[]> {
  return apiFetch("/api/v1/sentry/audit");
}

export async function getBudget(): Promise<BudgetCap> {
  return apiFetch("/api/v1/sentry/budget");
}

export async function updateBudget(
  data: Partial<Omit<BudgetCap, "user_id" | "used_tokens" | "created_at">>,
): Promise<BudgetCap> {
  return apiFetch("/api/v1/sentry/budget", {
    method: "PUT",
    body: JSON.stringify(data),
  });
}
