"use client";

import { use, useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Play, Search, Square, Trash2 } from "lucide-react";
import type { MCPServer } from "@/lib/types";
import { useAuth } from "@/lib/auth-context";
import {
  deleteServer,
  discoverServers,
  getServer,
  startServer,
  stopServer,
  ApiRequestError,
} from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { StatusBadge } from "@/components/status-badge";
import { MCPInspector } from "@/components/mcp-inspector";

export default function ServerDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const { isAuthenticated, loading: authLoading } = useAuth();
  const router = useRouter();

  const [server, setServer] = useState<MCPServer | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState("");

  // Discover state
  const [discoverQuery, setDiscoverQuery] = useState("");
  const [discoverResults, setDiscoverResults] = useState<MCPServer[]>([]);
  const [discoverLoading, setDiscoverLoading] = useState(false);
  const [discoverError, setDiscoverError] = useState("");

  const fetchServer = useCallback(async () => {
    try {
      const data = await getServer(id);
      setServer(data);
      setError("");
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.message);
      } else {
        setError("Failed to load server details.");
      }
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push("/login");
      return;
    }
    if (isAuthenticated) {
      fetchServer();
    }
  }, [authLoading, isAuthenticated, router, fetchServer]);

  async function handleStart() {
    setActionLoading(true);
    try {
      await startServer(id);
      await fetchServer();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.message);
      } else {
        setError("Failed to start server.");
      }
    } finally {
      setActionLoading(false);
    }
  }

  async function handleStop() {
    setActionLoading(true);
    try {
      await stopServer(id);
      await fetchServer();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.message);
      } else {
        setError("Failed to stop server.");
      }
    } finally {
      setActionLoading(false);
    }
  }

  async function handleDelete() {
    if (!server) return;
    if (!window.confirm(`Delete server "${server.name}"? This cannot be undone.`)) {
      return;
    }
    setActionLoading(true);
    try {
      await deleteServer(id);
      router.push("/servers");
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.message);
      } else {
        setError("Failed to delete server.");
      }
      setActionLoading(false);
    }
  }

  async function handleDiscover(e: React.FormEvent) {
    e.preventDefault();
    setDiscoverError("");
    setDiscoverLoading(true);
    try {
      const results = await discoverServers(discoverQuery || undefined);
      setDiscoverResults(results);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setDiscoverError(err.message);
      } else {
        setDiscoverError("Discovery request failed.");
      }
    } finally {
      setDiscoverLoading(false);
    }
  }

  if (authLoading || (!isAuthenticated && !authLoading)) {
    return null;
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="size-8 animate-spin rounded-full border-4 border-muted border-t-primary" />
      </div>
    );
  }

  if (error && !server) {
    return (
      <div className="space-y-4">
        <Link
          href="/servers"
          className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="size-4" />
          Back to Servers
        </Link>
        <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </p>
      </div>
    );
  }

  if (!server) return null;

  return (
    <div className="space-y-6">
      <Link
        href="/servers"
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="size-4" />
        Back to Servers
      </Link>

      {error && (
        <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </p>
      )}

      {/* Server info */}
      <Card>
        <CardHeader>
          <div className="flex items-start justify-between gap-4">
            <div className="space-y-1">
              <CardTitle className="text-xl">{server.name}</CardTitle>
              <p className="text-sm text-muted-foreground">{server.image}</p>
            </div>
            <StatusBadge status={server.status} />
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-3 text-sm sm:grid-cols-2">
            <div>
              <span className="text-muted-foreground">Container ID</span>
              <p className="font-mono">{server.container_id || "N/A"}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Created</span>
              <p>{new Date(server.created_at).toLocaleString()}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Updated</span>
              <p>{new Date(server.updated_at).toLocaleString()}</p>
            </div>
          </div>

          <div className="flex gap-2 pt-2">
            <Button
              variant="outline"
              size="sm"
              disabled={actionLoading || server.status !== "stopped"}
              onClick={handleStart}
            >
              <Play className="size-4" />
              Start
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={actionLoading || server.status !== "running"}
              onClick={handleStop}
            >
              <Square className="size-4" />
              Stop
            </Button>
            <Button
              variant="destructive"
              size="sm"
              disabled={actionLoading}
              onClick={handleDelete}
              className="ml-auto"
            >
              <Trash2 className="size-4" />
              Delete
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Config */}
      <Card>
        <CardHeader>
          <CardTitle>Configuration</CardTitle>
        </CardHeader>
        <CardContent>
          <pre className="overflow-x-auto rounded-md bg-muted p-4 text-sm font-mono">
            {JSON.stringify(server.config, null, 2)}
          </pre>
        </CardContent>
      </Card>

      {/* Discover */}
      <Card>
        <CardHeader>
          <CardTitle>Discover Servers</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <form onSubmit={handleDiscover} className="flex gap-2">
            <Input
              placeholder="Search query..."
              value={discoverQuery}
              onChange={(e) => setDiscoverQuery(e.target.value)}
              className="max-w-sm"
            />
            <Button type="submit" variant="outline" disabled={discoverLoading}>
              <Search className="size-4" />
              {discoverLoading ? "Searching..." : "Search"}
            </Button>
          </form>

          {discoverError && (
            <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {discoverError}
            </p>
          )}

          {discoverResults.length > 0 && (
            <ul className="divide-y rounded-md border">
              {discoverResults.map((s) => (
                <li key={s.id} className="flex items-center justify-between px-4 py-3">
                  <div className="min-w-0">
                    <p className="truncate font-medium">{s.name}</p>
                    <p className="truncate text-sm text-muted-foreground">
                      {s.image}
                    </p>
                  </div>
                  <StatusBadge status={s.status} />
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      {/* Interactive MCP Inspector */}
      {server.status === "running" && (
        <div className="mt-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
          <MCPInspector serverId={id} />
        </div>
      )}
    </div>
  );
}
