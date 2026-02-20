"use client";

import { Progress } from "@/components/ui/progress";
import { cn } from "@/lib/utils";

interface BudgetGaugeProps {
  used_tokens: number;
  max_tokens: number;
}

export function BudgetGauge({ used_tokens, max_tokens }: BudgetGaugeProps) {
  const pct = max_tokens > 0 ? Math.min((used_tokens / max_tokens) * 100, 100) : 0;

  const colorClass =
    pct > 80
      ? "[&_[data-slot=progress-indicator]]:bg-red-500"
      : pct > 60
        ? "[&_[data-slot=progress-indicator]]:bg-yellow-500"
        : "[&_[data-slot=progress-indicator]]:bg-green-500";

  return (
    <div className="space-y-2">
      <Progress value={pct} className={cn("h-3", colorClass)} />
      <p className="text-sm text-muted-foreground">
        {used_tokens.toLocaleString()} / {max_tokens.toLocaleString()} tokens
        {max_tokens > 0 && (
          <span className="ml-1">({Math.round(pct)}%)</span>
        )}
      </p>
    </div>
  );
}
