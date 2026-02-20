"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useRouter } from "next/navigation";
import type { User } from "./types";
import * as api from "./api";

interface AuthState {
  user: User | null;
  sessionId: string | null;
  loading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
}

const AuthContext = createContext<AuthState | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const token = api.getToken();
    if (!token) {
      setLoading(false);
      return;
    }

    // Validate the token by making a lightweight API call.
    // listServers acts as a token validation check since it requires auth.
    api
      .listServers()
      .then(() => {
        // Token is valid. We don't get user info from this call, so we
        // reconstruct minimal user state from localStorage.
        const stored = localStorage.getItem("nexusclaw_user");
        if (stored) {
          try {
            setUser(JSON.parse(stored));
            setSessionId(localStorage.getItem("nexusclaw_session_id"));
          } catch {
            api.clearToken();
            localStorage.removeItem("nexusclaw_user");
            localStorage.removeItem("nexusclaw_session_id");
          }
        } else {
          // Token works but no stored user info; clear and force re-login.
          api.clearToken();
        }
      })
      .catch(() => {
        api.clearToken();
        localStorage.removeItem("nexusclaw_user");
        localStorage.removeItem("nexusclaw_session_id");
      })
      .finally(() => setLoading(false));
  }, []);

  const login = useCallback(
    async (email: string, password: string) => {
      const session = await api.login(email, password);
      api.setToken(session.token!);
      const u: User = {
        id: session.user_id,
        email,
        created_at: "",
        updated_at: "",
      };
      setUser(u);
      setSessionId(session.id);
      localStorage.setItem("nexusclaw_user", JSON.stringify(u));
      localStorage.setItem("nexusclaw_session_id", session.id);
      router.push("/");
    },
    [router],
  );

  const logout = useCallback(async () => {
    try {
      if (sessionId) {
        await api.logout(sessionId);
      }
    } catch {
      // Best-effort: proceed with local cleanup even if the API call fails.
    }
    api.clearToken();
    localStorage.removeItem("nexusclaw_user");
    localStorage.removeItem("nexusclaw_session_id");
    setUser(null);
    setSessionId(null);
    router.push("/login");
  }, [sessionId, router]);

  const register = useCallback(
    async (email: string, password: string) => {
      await api.register(email, password);
      await login(email, password);
    },
    [login],
  );

  const value = useMemo<AuthState>(
    () => ({
      user,
      sessionId,
      loading,
      isAuthenticated: !!user,
      login,
      logout,
      register,
    }),
    [user, sessionId, loading, login, logout, register],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}
