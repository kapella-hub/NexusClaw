"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Coins, KeyRound, Server } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { BudgetGauge } from "@/components/budget-gauge";
import { useAuth } from "@/lib/auth-context";
import * as api from "@/lib/api";
import type { AuditEntry, BudgetCap, MCPServer, VaultEntry } from "@/lib/types";

export default function DashboardPage() {
  const { isAuthenticated, loading: authLoading } = useAuth();
  const router = useRouter();

  const [servers, setServers] = useState<MCPServer[]>([]);
  const [vault, setVault] = useState<VaultEntry[]>([]);
  const [budget, setBudget] = useState<BudgetCap | null>(null);
  const [audit, setAudit] = useState<AuditEntry[]>([]);
  const [dataLoading, setDataLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (authLoading) return;
    if (!isAuthenticated) {
      router.push("/login");
      return;
    }

    Promise.allSettled([
      api.listServers(),
      api.listVault(),
      api.getBudget(),
      api.listAudit(),
    ]).then(([sResult, vResult, bResult, aResult]) => {
      if (sResult.status === "fulfilled") setServers(sResult.value ?? []);
      if (vResult.status === "fulfilled") setVault(vResult.value ?? []);
      if (bResult.status === "fulfilled") setBudget(bResult.value);
      if (aResult.status === "fulfilled") setAudit(aResult.value ?? []);

      // Check if all critical fetches failed (servers is the baseline)
      if (sResult.status === "rejected" && vResult.status === "rejected") {
        setError("Failed to load dashboard data. Please try again.");
      }

      setDataLoading(false);
    });
  }, [isAuthenticated, authLoading, router]);

  // Show spinner while auth is loading or while redirecting
  if (authLoading || (!isAuthenticated && !error)) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-4 border-muted border-t-primary" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex h-64 items-center justify-center">
        <p className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </p>
      </div>
    );
  }

  const runningCount = servers.filter((s) => s.status === "running").length;

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>

      {/* Stat cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardDescription>Total Servers</CardDescription>
            <Server className="size-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {dataLoading ? "..." : servers.length}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardDescription>Running Servers</CardDescription>
            <span className="relative flex size-3">
              <span className="absolute inline-flex size-full animate-ping rounded-full bg-green-400 opacity-75" />
              <span className="relative inline-flex size-3 rounded-full bg-green-500" />
            </span>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {dataLoading ? "..." : runningCount}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardDescription>Vault Entries</CardDescription>
            <KeyRound className="size-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {dataLoading ? "..." : vault.length}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardDescription>Budget Usage</CardDescription>
            <Coins className="size-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">
              {dataLoading
                ? "..."
                : budget
                  ? `${budget.used_tokens.toLocaleString()} / ${budget.max_tokens.toLocaleString()}`
                  : "Not set"}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Budget gauge */}
      <Card>
        <CardHeader>
          <CardTitle>Token Budget</CardTitle>
          {budget && (
            <CardDescription>{budget.period} usage</CardDescription>
          )}
        </CardHeader>
        <CardContent>
          {dataLoading ? (
            <div className="h-3 w-full animate-pulse rounded-full bg-muted" />
          ) : budget ? (
            <BudgetGauge
              used_tokens={budget.used_tokens}
              max_tokens={budget.max_tokens}
            />
          ) : (
            <p className="text-sm text-muted-foreground">
              No budget configured. Set one in the Budget settings.
            </p>
          )}
        </CardContent>
      </Card>

      {/* Recent audit log */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Activity</CardTitle>
          <CardDescription>Last 5 audit entries</CardDescription>
        </CardHeader>
        <CardContent>
          {dataLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 3 }).map((_, i) => (
                <div
                  key={i}
                  className="h-8 w-full animate-pulse rounded bg-muted"
                />
              ))}
            </div>
          ) : audit.length === 0 ? (
            <p className="py-4 text-center text-sm text-muted-foreground">
              No activity recorded yet.
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Action</TableHead>
                  <TableHead>Resource</TableHead>
                  <TableHead className="text-right">Time</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {audit.slice(0, 5).map((entry) => (
                  <TableRow key={entry.id}>
                    <TableCell className="font-medium">
                      {entry.action}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {entry.resource || "-"}
                    </TableCell>
                    <TableCell className="text-right text-muted-foreground">
                      {formatTimestamp(entry.created_at)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function formatTimestamp(iso: string): string {
  try {
    const date = new Date(iso);
    return date.toLocaleString(undefined, {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return iso;
  }
}
