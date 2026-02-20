"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { Rule, RuleAction } from "@/lib/types";

interface RuleFormData {
  name: string;
  description: string;
  pattern: string;
  action: RuleAction;
  enabled: boolean;
}

interface RuleFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  rule?: Rule;
  onSubmit: (data: RuleFormData) => Promise<void>;
}

export function RuleForm({ open, onOpenChange, rule, onSubmit }: RuleFormProps) {
  const isEdit = !!rule;
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [pattern, setPattern] = useState("");
  const [action, setAction] = useState<RuleAction>("block");
  const [enabled, setEnabled] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (open) {
      setName(rule?.name ?? "");
      setDescription(rule?.description ?? "");
      setPattern(rule?.pattern ?? "");
      setAction(rule?.action ?? "block");
      setEnabled(rule?.enabled ?? true);
    }
  }, [open, rule]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    try {
      await onSubmit({ name, description, pattern, action, enabled });
      onOpenChange(false);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit Rule" : "Create Rule"}</DialogTitle>
          <DialogDescription>
            {isEdit
              ? "Update the firewall rule configuration."
              : "Define a new firewall rule for the gateway."}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="rule-name">Name</Label>
            <Input
              id="rule-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Block prompt injection"
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="rule-description">Description</Label>
            <textarea
              id="rule-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description of what this rule does"
              rows={3}
              className="border-input placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-ring/50 dark:bg-input/30 w-full rounded-md border bg-transparent px-3 py-2 text-sm shadow-xs outline-none focus-visible:ring-[3px]"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="rule-pattern">Pattern</Label>
            <Input
              id="rule-pattern"
              value={pattern}
              onChange={(e) => setPattern(e.target.value)}
              placeholder="e.g. .*malicious.*"
              className="font-mono"
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="rule-action">Action</Label>
            <Select value={action} onValueChange={(v) => setAction(v as RuleAction)}>
              <SelectTrigger id="rule-action" className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="block">Block</SelectItem>
                <SelectItem value="allow">Allow</SelectItem>
                <SelectItem value="alert">Alert</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-2">
            <Checkbox
              id="rule-enabled"
              checked={enabled}
              onCheckedChange={(checked) => setEnabled(checked === true)}
            />
            <Label htmlFor="rule-enabled">Enabled</Label>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? (isEdit ? "Saving..." : "Creating...") : isEdit ? "Save Changes" : "Create Rule"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
