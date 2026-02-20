"use client";

import type { ServerStatus } from "@/lib/types";
import { Badge } from "@/components/ui/badge";

const statusConfig: Record<
  ServerStatus,
  { variant: "default" | "secondary" | "destructive" | "outline"; className?: string; label: string }
> = {
  running: {
    variant: "default",
    className: "bg-green-600 text-white hover:bg-green-600",
    label: "Running",
  },
  stopped: {
    variant: "secondary",
    label: "Stopped",
  },
  starting: {
    variant: "outline",
    className: "border-yellow-500 text-yellow-600",
    label: "Starting",
  },
  stopping: {
    variant: "outline",
    className: "border-yellow-500 text-yellow-600",
    label: "Stopping",
  },
  error: {
    variant: "destructive",
    label: "Error",
  },
};

interface StatusBadgeProps {
  status: ServerStatus;
}

export function StatusBadge({ status }: StatusBadgeProps) {
  const config = statusConfig[status];
  return (
    <Badge variant={config.variant} className={config.className}>
      {config.label}
    </Badge>
  );
}
