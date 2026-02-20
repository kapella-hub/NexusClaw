"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Coins } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { BudgetGauge } from "@/components/budget-gauge";
import { useAuth } from "@/lib/auth-context";
import { ApiRequestError } from "@/lib/api";
import * as api from "@/lib/api";
import type { BudgetCap, BudgetPeriod } from "@/lib/types";

export default function BudgetPage() {
  const { isAuthenticated, loading: authLoading } = useAuth();
  const router = useRouter();

  const [budget, setBudget] = useState<BudgetCap | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  // Form state
  const [period, setPeriod] = useState<BudgetPeriod>("monthly");
  const [maxTokens, setMaxTokens] = useState("");

  const fetchBudget = useCallback(async () => {
    try {
      const data = await api.getBudget();
      setBudget(data);
      setPeriod(data.period);
      setMaxTokens(String(data.max_tokens));
    } catch (err) {
      if (err instanceof ApiRequestError && err.status === 404) {
        setBudget(null);
      }
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
    fetchBudget();
  }, [isAuthenticated, authLoading, router, fetchBudget]);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    try {
      await api.updateBudget({
        period,
        max_tokens: Number(maxTokens),
      });
      await fetchBudget();
    } finally {
      setSaving(false);
    }
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
      <div className="flex items-center gap-3">
        <Coins className="size-6 text-muted-foreground" />
        <h1 className="text-2xl font-bold">Token Budget</h1>
      </div>

      {budget ? (
        <Card>
          <CardHeader>
            <CardTitle>Current Usage</CardTitle>
            <CardDescription>
              {budget.period.charAt(0).toUpperCase() + budget.period.slice(1)}{" "}
              budget &middot; resets{" "}
              {budget.reset_at ? new Date(budget.reset_at).toLocaleDateString() : "N/A"}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <BudgetGauge
              used_tokens={budget.used_tokens}
              max_tokens={budget.max_tokens}
            />
            <div className="grid grid-cols-2 gap-4 text-sm sm:grid-cols-4">
              <div>
                <p className="text-muted-foreground">Used</p>
                <p className="font-medium">
                  {budget.used_tokens.toLocaleString()}
                </p>
              </div>
              <div>
                <p className="text-muted-foreground">Limit</p>
                <p className="font-medium">
                  {budget.max_tokens.toLocaleString()}
                </p>
              </div>
              <div>
                <p className="text-muted-foreground">Period</p>
                <p className="font-medium capitalize">{budget.period}</p>
              </div>
              <div>
                <p className="text-muted-foreground">Resets At</p>
                <p className="font-medium">
                  {budget.reset_at ? new Date(budget.reset_at).toLocaleString() : "N/A"}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardContent className="flex flex-col items-center py-16">
            <Coins className="mb-4 size-10 text-muted-foreground" />
            <p className="text-sm text-muted-foreground">
              No budget configured. Use the form below to set one up.
            </p>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>{budget ? "Update Budget" : "Configure Budget"}</CardTitle>
          <CardDescription>
            Set the token spending limit and billing period.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSave} className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="budget-period">Period</Label>
                <Select value={period} onValueChange={(v) => setPeriod(v as BudgetPeriod)}>
                  <SelectTrigger id="budget-period" className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="daily">Daily</SelectItem>
                    <SelectItem value="weekly">Weekly</SelectItem>
                    <SelectItem value="monthly">Monthly</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="budget-max-tokens">Max Tokens</Label>
                <Input
                  id="budget-max-tokens"
                  type="number"
                  min={1}
                  value={maxTokens}
                  onChange={(e) => setMaxTokens(e.target.value)}
                  placeholder="e.g. 100000"
                  required
                />
              </div>
            </div>
            <Button type="submit" disabled={saving}>
              {saving ? "Saving..." : "Save Budget"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
