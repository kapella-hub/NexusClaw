"use client";

import Link from "next/link";
import { Play, Square, Trash2 } from "lucide-react";
import type { MCPServer } from "@/lib/types";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/status-badge";

interface ServerCardProps {
  server: MCPServer;
  onStart: (id: string) => void;
  onStop: (id: string) => void;
  onDelete: (id: string) => void;
  loading?: boolean;
}

export function ServerCard({
  server,
  onStart,
  onStop,
  onDelete,
  loading,
}: ServerCardProps) {
  function handleDelete() {
    if (window.confirm(`Delete server "${server.name}"? This cannot be undone.`)) {
      onDelete(server.id);
    }
  }

  return (
    <Card className="flex flex-col">
      <CardHeader>
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0 space-y-1">
            <CardTitle>
              <Link
                href={`/servers/${server.id}`}
                className="hover:underline"
              >
                {server.name}
              </Link>
            </CardTitle>
            <CardDescription className="truncate">
              {server.image}
            </CardDescription>
          </div>
          <StatusBadge status={server.status} />
        </div>
      </CardHeader>

      <CardContent className="flex-1">
        {server.container_id && (
          <p className="text-xs text-muted-foreground font-mono">
            Container: {server.container_id.slice(0, 12)}
          </p>
        )}
      </CardContent>

      <CardFooter className="gap-2">
        <Button
          variant="outline"
          size="icon-sm"
          disabled={loading || server.status !== "stopped"}
          onClick={() => onStart(server.id)}
          aria-label={`Start ${server.name}`}
        >
          <Play className="size-4" />
        </Button>
        <Button
          variant="outline"
          size="icon-sm"
          disabled={loading || server.status !== "running"}
          onClick={() => onStop(server.id)}
          aria-label={`Stop ${server.name}`}
        >
          <Square className="size-4" />
        </Button>
        <Button
          variant="outline"
          size="icon-sm"
          disabled={loading}
          onClick={handleDelete}
          aria-label={`Delete ${server.name}`}
          className="ml-auto text-destructive hover:bg-destructive/10"
        >
          <Trash2 className="size-4" />
        </Button>
      </CardFooter>
    </Card>
  );
}
