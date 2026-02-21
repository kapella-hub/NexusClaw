"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Plus, Server } from "lucide-react";
import type { MCPServer } from "@/lib/types";
import { useAuth } from "@/lib/auth-context";
import { toast } from "sonner";
import {
  createServer,
  deleteServer,
  listServers,
  startServer,
  stopServer,
  ApiRequestError,
} from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { ServerCard } from "@/components/server-card";

const SERVER_TEMPLATES = [
  {
    id: "custom",
    name: "Custom Server",
    image: "",
    config: "{}",
    description: "Provide your own Docker image and config",
  },
  {
    id: "sqlite",
    name: "SQLite Integration",
    image: "mcp/sqlite:latest",
    config: '{\n  "db_path": "/data/mcp.db"\n}',
    description: "Read and write to a local SQLite database",
  },
  {
    id: "github",
    name: "GitHub Plugin",
    image: "mcp/github:latest",
    config: '{\n  "github_token": "YOUR_TOKEN_HERE"\n}',
    description: "Interact with GitHub APIs",
  },
  {
    id: "filesystem",
    name: "Local Filesystem",
    image: "mcp/filesystem:latest",
    config: '{\n  "allowed_dirs": ["/data"]\n}',
    description: "Read and write files in allowed directories",
  },
];

export default function ServersPage() {
  const { isAuthenticated, loading: authLoading } = useAuth();
  const router = useRouter();

  const [servers, setServers] = useState<MCPServer[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);

  // Dialog state
  const [dialogOpen, setDialogOpen] = useState(false);
  const [formName, setFormName] = useState("");
  const [formImage, setFormImage] = useState("");
  const [formConfig, setFormConfig] = useState("");
  const [formError, setFormError] = useState("");
  const [formLoading, setFormLoading] = useState(false);
  const [template, setTemplate] = useState("custom");

  function handleTemplateChange(e: React.ChangeEvent<HTMLSelectElement>) {
    const value = e.target.value;
    setTemplate(value);
    const tpl = SERVER_TEMPLATES.find(t => t.id === value);
    if (tpl && value !== "custom") {
      if (!formName) setFormName(tpl.name);
      setFormImage(tpl.image);
      setFormConfig(tpl.config);
    }
  }

  const fetchServers = useCallback(async () => {
    try {
      const data = await listServers();
      setServers(data);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        toast.error(err.message);
      } else {
        toast.error("Failed to load servers.");
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push("/login");
      return;
    }
    if (isAuthenticated) {
      fetchServers();
    }
  }, [authLoading, isAuthenticated, router, fetchServers]);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setFormLoading(true);

    let config: Record<string, unknown> = {};
    if (formConfig.trim()) {
      try {
        config = JSON.parse(formConfig);
      } catch {
        toast.error("Config must be valid JSON.");
        setFormLoading(false);
        return;
      }
    }

    try {
      await createServer(formName, formImage, config);
      setDialogOpen(false);
      setFormName("");
      setFormImage("");
      setFormConfig("");
      setTemplate("custom");
      toast.success("Server registered successfully!");
      await fetchServers();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        toast.error(err.message);
      } else {
        toast.error("Failed to register server.");
      }
    } finally {
      setFormLoading(false);
    }
  }

  async function handleStart(id: string) {
    setActionLoading(true);
    try {
      await startServer(id);
      toast.success("Server started!");
      await fetchServers();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        toast.error(err.message);
      } else {
        toast.error("Failed to start server.");
      }
    } finally {
      setActionLoading(false);
    }
  }

  async function handleStop(id: string) {
    setActionLoading(true);
    try {
      await stopServer(id);
      toast.success("Server stopped!");
      await fetchServers();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        toast.error(err.message);
      } else {
        toast.error("Failed to stop server.");
      }
    } finally {
      setActionLoading(false);
    }
  }

  async function handleDelete(id: string) {
    setActionLoading(true);
    try {
      await deleteServer(id);
      toast.success("Server removed!");
      await fetchServers();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        toast.error(err.message);
      } else {
        toast.error("Failed to remove server.");
      }
    } finally {
      setActionLoading(false);
    }
  }

  if (authLoading || (!isAuthenticated && !authLoading)) {
    return null;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Servers</h1>
          <p className="text-sm text-muted-foreground">
            Manage your MCP server instances
          </p>
        </div>

        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="size-4" />
              Register Server
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Register Server</DialogTitle>
              <DialogDescription>
                Add a new MCP server to your gateway.
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={handleCreate} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="server-template">Template</Label>
                <select
                  id="server-template"
                  className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  value={template}
                  onChange={handleTemplateChange}
                >
                  {SERVER_TEMPLATES.map((tpl) => (
                    <option key={tpl.id} value={tpl.id}>
                      {tpl.name} - {tpl.description}
                    </option>
                  ))}
                </select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="server-name">Name</Label>
                <Input
                  id="server-name"
                  placeholder="my-mcp-server"
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="server-image">Image</Label>
                <Input
                  id="server-image"
                  placeholder="mcp/server:latest"
                  value={formImage}
                  onChange={(e) => setFormImage(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="server-config">Config (JSON)</Label>
                <textarea
                  id="server-config"
                  className="border-input bg-transparent placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] dark:bg-input/30 w-full min-h-[80px] rounded-md border px-3 py-2 text-sm shadow-xs outline-none disabled:opacity-50"
                  placeholder='{"key": "value"}'
                  value={formConfig}
                  onChange={(e) => setFormConfig(e.target.value)}
                  rows={3}
                />
              </div>
              <DialogFooter>
                <Button type="submit" disabled={formLoading}>
                  {formLoading ? "Creating..." : "Create"}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="size-8 animate-spin rounded-full border-4 border-muted border-t-primary" />
        </div>
      ) : servers.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12 text-center">
          <Server className="mb-3 size-10 text-muted-foreground" />
          <h2 className="text-lg font-semibold">No servers registered yet</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Register your first MCP server to get started.
          </p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {servers.map((server) => (
            <ServerCard
              key={server.id}
              server={server}
              onStart={handleStart}
              onStop={handleStop}
              onDelete={handleDelete}
              loading={actionLoading}
            />
          ))}
        </div>
      )}
    </div>
  );
}
