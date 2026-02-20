"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Pencil, Plus, Shield, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { RuleForm } from "@/components/rule-form";
import { useAuth } from "@/lib/auth-context";
import * as api from "@/lib/api";
import type { Rule, RuleAction } from "@/lib/types";

const ACTION_COLORS: Record<RuleAction, string> = {
  block: "bg-red-500/15 text-red-500",
  allow: "bg-green-500/15 text-green-500",
  alert: "bg-yellow-500/15 text-yellow-500",
};

export default function RulesPage() {
  const { isAuthenticated, loading: authLoading } = useAuth();
  const router = useRouter();

  const [rules, setRules] = useState<Rule[]>([]);
  const [loading, setLoading] = useState(true);
  const [formOpen, setFormOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<Rule | undefined>(undefined);

  const fetchRules = useCallback(async () => {
    try {
      const data = await api.listRules();
      setRules(data ?? []);
    } catch {
      // Silently handle; rules remain empty
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (authLoading) return;
    if (!isAuthenticated) {
      router.push("/login");
      return;
    }
    fetchRules();
  }, [isAuthenticated, authLoading, router, fetchRules]);

  function openCreate() {
    setEditingRule(undefined);
    setFormOpen(true);
  }

  function openEdit(rule: Rule) {
    setEditingRule(rule);
    setFormOpen(true);
  }

  async function handleFormSubmit(data: {
    name: string;
    description: string;
    pattern: string;
    action: RuleAction;
    enabled: boolean;
  }) {
    if (editingRule) {
      await api.updateRule(editingRule.id, data);
    } else {
      await api.createRule(data);
    }
    await fetchRules();
  }

  async function handleDelete(rule: Rule) {
    if (!window.confirm(`Delete rule "${rule.name}"?`)) return;
    await api.deleteRule(rule.id);
    await fetchRules();
  }

  async function handleToggleEnabled(rule: Rule) {
    await api.updateRule(rule.id, { enabled: !rule.enabled });
    await fetchRules();
  }

  if (authLoading || loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-4 border-muted border-t-primary" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Shield className="size-6 text-muted-foreground" />
          <h1 className="text-2xl font-bold">Firewall Rules</h1>
        </div>
        <Button onClick={openCreate}>
          <Plus className="size-4" />
          Create Rule
        </Button>
      </div>

      {rules.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16">
          <Shield className="mb-4 size-10 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">
            No firewall rules configured yet.
          </p>
          <Button variant="outline" className="mt-4" onClick={openCreate}>
            <Plus className="size-4" />
            Create your first rule
          </Button>
        </div>
      ) : (
        <div className="rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Pattern</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Enabled</TableHead>
                <TableHead className="w-24">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rules.map((rule) => (
                <TableRow key={rule.id}>
                  <TableCell className="font-medium">{rule.name}</TableCell>
                  <TableCell>
                    <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">
                      {rule.pattern}
                    </code>
                  </TableCell>
                  <TableCell>
                    <Badge className={ACTION_COLORS[rule.action]}>
                      {rule.action}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Checkbox
                      checked={rule.enabled}
                      onCheckedChange={() => handleToggleEnabled(rule)}
                      aria-label={`Toggle ${rule.name}`}
                    />
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      <Button
                        variant="ghost"
                        size="icon-xs"
                        onClick={() => openEdit(rule)}
                        aria-label={`Edit ${rule.name}`}
                      >
                        <Pencil />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon-xs"
                        onClick={() => handleDelete(rule)}
                        aria-label={`Delete ${rule.name}`}
                      >
                        <Trash2 className="text-destructive" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <RuleForm
        open={formOpen}
        onOpenChange={setFormOpen}
        rule={editingRule}
        onSubmit={handleFormSubmit}
      />
    </div>
  );
}
