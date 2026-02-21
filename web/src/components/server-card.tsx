"use client";

import { useState } from "react";
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
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
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [capabilitiesOpen, setCapabilitiesOpen] = useState(false);

  function handleDelete() {
    onDelete(server.id);
    setDeleteOpen(false);
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

      <CardContent className="flex-1 space-y-4">
        {server.container_id && (
          <p className="text-xs text-muted-foreground font-mono">
            Container: {server.container_id.slice(0, 12)}
          </p>
        )}
        {(server.tools && server.tools.length > 0) || (server.resources && server.resources.length > 0) ? (
          <div className="flex gap-2 text-xs">
            {server.tools && server.tools.length > 0 && (
              <button
                onClick={() => setCapabilitiesOpen(true)}
                className="inline-flex cursor-pointer hover:bg-primary/20 transition-colors items-center rounded-md bg-primary/10 px-2 py-1 text-xs font-medium text-primary ring-1 ring-inset ring-primary/20"
              >
                {server.tools.length} Tools
              </button>
            )}
            {server.resources && server.resources.length > 0 && (
              <button
                onClick={() => setCapabilitiesOpen(true)}
                className="inline-flex cursor-pointer hover:bg-secondary/20 transition-colors items-center rounded-md bg-secondary/10 px-2 py-1 text-xs font-medium text-secondary-foreground ring-1 ring-inset ring-secondary/20"
              >
                {server.resources.length} Resources
              </button>
            )}
          </div>
        ) : null}
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
          onClick={() => setDeleteOpen(true)}
          aria-label={`Delete ${server.name}`}
          className="ml-auto text-destructive hover:bg-destructive/10 hover:text-destructive"
        >
          <Trash2 className="size-4" />
        </Button>
      </CardFooter>

      {/* Capabilities Viewer Dialog */}
      <Dialog open={capabilitiesOpen} onOpenChange={setCapabilitiesOpen}>
        <DialogContent className="max-w-2xl max-h-[80vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>{server.name} Capabilities</DialogTitle>
            <DialogDescription>
              Available tools and resources exposed by this server.
            </DialogDescription>
          </DialogHeader>
          <ScrollArea className="flex-1 min-h-0 pr-4 mt-4">
            <div className="space-y-6">
              {server.tools && server.tools.length > 0 && (
                <div className="space-y-3">
                  <h3 className="font-semibold text-sm text-muted-foreground uppercase tracking-wider">Tools</h3>
                  <div className="space-y-3">
                    {server.tools.map((tool: any) => (
                      <div key={tool.name} className="bg-muted/50 rounded-lg p-3 text-sm">
                        <div className="font-mono font-medium text-primary mb-1">{tool.name}</div>
                        <div className="text-muted-foreground">{tool.description}</div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
              {server.resources && server.resources.length > 0 && (
                <div className="space-y-3">
                  <h3 className="font-semibold text-sm text-muted-foreground uppercase tracking-wider">Resources</h3>
                  <div className="space-y-3">
                    {server.resources.map((res: any) => (
                      <div key={res.uri} className="bg-muted/50 rounded-lg p-3 text-sm">
                        <div className="font-mono font-medium text-secondary-foreground mb-1">{res.name}</div>
                        <div className="text-xs text-muted-foreground break-all">{res.uri}</div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </ScrollArea>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Server</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete "{server.name}"? This action cannot be undone.
              Any active connections will be terminated.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="mt-4">
            <Button variant="outline" onClick={() => setDeleteOpen(false)}>Cancel</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={loading}>
              {loading ? "Deleting..." : "Delete Server"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
