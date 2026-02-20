"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ScrollText } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useAuth } from "@/lib/auth-context";
import * as api from "@/lib/api";
import type { AuditEntry } from "@/lib/types";

export default function AuditPage() {
  const { isAuthenticated, loading: authLoading } = useAuth();
  const router = useRouter();

  const [entries, setEntries] = useState<AuditEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (authLoading) return;
    if (!isAuthenticated) {
      router.push("/login");
      return;
    }

    api
      .listAudit()
      .then((data) => setEntries(data ?? []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [isAuthenticated, authLoading, router]);

  if (authLoading || loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-4 border-muted border-t-primary" />
      </div>
    );
  }

  const sorted = [...entries].sort(
    (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <ScrollText className="size-6 text-muted-foreground" />
        <h1 className="text-2xl font-bold">Audit Log</h1>
      </div>

      {sorted.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16">
          <ScrollText className="mb-4 size-10 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">
            No audit entries recorded yet.
          </p>
        </div>
      ) : (
        <div className="rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Action</TableHead>
                <TableHead>Resource</TableHead>
                <TableHead>Metadata</TableHead>
                <TableHead>Timestamp</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {sorted.map((entry) => {
                const meta = entry.metadata
                  ? JSON.stringify(entry.metadata)
                  : "-";
                const truncatedMeta =
                  meta.length > 80 ? meta.slice(0, 80) + "..." : meta;

                return (
                  <TableRow key={entry.id}>
                    <TableCell className="font-medium">
                      {entry.action}
                    </TableCell>
                    <TableCell>{entry.resource || "-"}</TableCell>
                    <TableCell>
                      <span
                        className="font-mono text-xs text-muted-foreground"
                        title={meta}
                      >
                        {truncatedMeta}
                      </span>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {new Date(entry.created_at).toLocaleString()}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}
