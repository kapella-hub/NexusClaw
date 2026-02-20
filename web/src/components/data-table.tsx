"use client";

import { useState, useMemo, useCallback } from "react";
import { ChevronDown, ChevronUp } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export interface Column<T> {
  key: string;
  label: string;
  sortable?: boolean;
  render?: (item: T) => React.ReactNode;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  emptyMessage?: string;
}

type SortDirection = "asc" | "desc";

export function DataTable<T>({
  columns,
  data,
  emptyMessage = "No data available.",
}: DataTableProps<T>) {
  const [sortKey, setSortKey] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<SortDirection>("asc");

  const handleSort = useCallback(
    (key: string) => {
      if (sortKey === key) {
        setSortDir((prev) => (prev === "asc" ? "desc" : "asc"));
      } else {
        setSortKey(key);
        setSortDir("asc");
      }
    },
    [sortKey],
  );

  const sortedData = useMemo(() => {
    if (!sortKey) return data;
    return [...data].sort((a, b) => {
      const aVal = (a as Record<string, unknown>)[sortKey];
      const bVal = (b as Record<string, unknown>)[sortKey];
      if (aVal == null && bVal == null) return 0;
      if (aVal == null) return sortDir === "asc" ? -1 : 1;
      if (bVal == null) return sortDir === "asc" ? 1 : -1;
      if (typeof aVal === "string" && typeof bVal === "string") {
        const cmp = aVal.localeCompare(bVal);
        return sortDir === "asc" ? cmp : -cmp;
      }
      if (aVal < bVal) return sortDir === "asc" ? -1 : 1;
      if (aVal > bVal) return sortDir === "asc" ? 1 : -1;
      return 0;
    });
  }, [data, sortKey, sortDir]);

  if (data.length === 0) {
    return (
      <p className="py-8 text-center text-sm text-muted-foreground">
        {emptyMessage}
      </p>
    );
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          {columns.map((col) => (
            <TableHead key={col.key}>
              {col.sortable ? (
                <button
                  type="button"
                  className="inline-flex items-center gap-1 hover:text-foreground"
                  onClick={() => handleSort(col.key)}
                >
                  {col.label}
                  {sortKey === col.key ? (
                    sortDir === "asc" ? (
                      <ChevronUp className="size-3.5" />
                    ) : (
                      <ChevronDown className="size-3.5" />
                    )
                  ) : (
                    <ChevronDown className="size-3.5 opacity-0 group-hover:opacity-50" />
                  )}
                </button>
              ) : (
                col.label
              )}
            </TableHead>
          ))}
        </TableRow>
      </TableHeader>
      <TableBody>
        {sortedData.map((item, idx) => (
          <TableRow key={(item as Record<string, unknown>)["id"] as string ?? idx}>
            {columns.map((col) => (
              <TableCell key={col.key}>
                {col.render
                  ? col.render(item)
                  : String(
                      (item as Record<string, unknown>)[col.key] ?? "",
                    )}
              </TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
