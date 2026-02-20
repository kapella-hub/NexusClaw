"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { KeyRound, Plus, Trash2, Loader2 } from "lucide-react";
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
import { DataTable, type Column } from "@/components/data-table";
import { useAuth } from "@/lib/auth-context";
import * as api from "@/lib/api";
import { ApiRequestError } from "@/lib/api";
import type { VaultEntry } from "@/lib/types";

export default function VaultPage() {
  const { isAuthenticated, loading: authLoading } = useAuth();
  const router = useRouter();

  const [entries, setEntries] = useState<VaultEntry[]>([]);
  const [dataLoading, setDataLoading] = useState(true);
  const [error, setError] = useState("");

  // Dialog state
  const [dialogOpen, setDialogOpen] = useState(false);
  const [provider, setProvider] = useState("");
  const [accessToken, setAccessToken] = useState("");
  const [secret, setSecret] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState("");

  // Delete loading state (keyed by entry id)
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const fetchEntries = useCallback(async () => {
    try {
      const data = await api.listVault();
      setEntries(data ?? []);
      setError("");
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.message);
      } else {
        setError("Failed to load vault entries.");
      }
    } finally {
      setDataLoading(false);
    }
  }, []);

  useEffect(() => {
    if (authLoading) return;
    if (!isAuthenticated) {
      router.push("/login");
      return;
    }
    fetchEntries();
  }, [isAuthenticated, authLoading, router, fetchEntries]);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setFormError("");
    setSubmitting(true);
    try {
      await api.createVaultEntry(
        provider,
        accessToken || undefined,
        secret || undefined,
      );
      setDialogOpen(false);
      setProvider("");
      setAccessToken("");
      setSecret("");
      await fetchEntries();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setFormError(err.message);
      } else {
        setFormError("Failed to create credential.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(entry: VaultEntry) {
    const confirmed = window.confirm(
      `Delete credential for "${entry.provider}"? This cannot be undone.`,
    );
    if (!confirmed) return;

    setDeletingId(entry.id);
    try {
      await api.deleteVaultEntry(entry.id);
      await fetchEntries();
    } catch (err) {
      if (err instanceof ApiRequestError) {
        setError(err.message);
      } else {
        setError("Failed to delete credential.");
      }
    } finally {
      setDeletingId(null);
    }
  }

  const columns: Column<VaultEntry>[] = [
    {
      key: "provider",
      label: "Provider",
      sortable: true,
      render: (entry) => (
        <span className="font-medium">{entry.provider}</span>
      ),
    },
    {
      key: "created_at",
      label: "Created At",
      sortable: true,
      render: (entry) => (
        <span className="text-muted-foreground">
          {new Date(entry.created_at).toLocaleDateString(undefined, {
            year: "numeric",
            month: "short",
            day: "numeric",
          })}
        </span>
      ),
    },
    {
      key: "actions",
      label: "Actions",
      render: (entry) => (
        <Button
          variant="ghost"
          size="sm"
          className="text-destructive hover:bg-destructive/10 hover:text-destructive"
          onClick={() => handleDelete(entry)}
          disabled={deletingId === entry.id}
          aria-label={`Delete credential for ${entry.provider}`}
        >
          {deletingId === entry.id ? (
            <Loader2 className="size-4 animate-spin" />
          ) : (
            <Trash2 className="size-4" />
          )}
        </Button>
      ),
    },
  ];

  if (authLoading || dataLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-4 border-muted border-t-primary" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <KeyRound className="size-6 text-primary" />
          <h1 className="text-2xl font-bold">Credential Vault</h1>
        </div>

        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 size-4" />
              Add Credential
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add Credential</DialogTitle>
              <DialogDescription>
                Store a new provider credential in your vault.
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={handleCreate} className="space-y-4">
              {formError && (
                <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
                  {formError}
                </p>
              )}
              <div className="space-y-2">
                <Label htmlFor="provider">Provider</Label>
                <Input
                  id="provider"
                  placeholder="e.g. openai, github, anthropic"
                  value={provider}
                  onChange={(e) => setProvider(e.target.value)}
                  required
                  autoComplete="off"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="access-token">Access Token</Label>
                <Input
                  id="access-token"
                  type="password"
                  placeholder="Optional"
                  value={accessToken}
                  onChange={(e) => setAccessToken(e.target.value)}
                  autoComplete="off"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="secret">Secret</Label>
                <Input
                  id="secret"
                  type="password"
                  placeholder="Optional"
                  value={secret}
                  onChange={(e) => setSecret(e.target.value)}
                  autoComplete="off"
                />
              </div>
              <DialogFooter>
                <Button type="submit" disabled={submitting}>
                  {submitting ? (
                    <>
                      <Loader2 className="mr-2 size-4 animate-spin" />
                      Saving...
                    </>
                  ) : (
                    "Save Credential"
                  )}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {error && (
        <p className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </p>
      )}

      <DataTable
        columns={columns}
        data={entries}
        emptyMessage="No credentials stored yet. Add your first credential above."
      />
    </div>
  );
}
