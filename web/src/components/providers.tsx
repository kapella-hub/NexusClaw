"use client";

import { ThemeProvider } from "next-themes";
import { AuthProvider } from "@/lib/auth-context";
import { Toaster } from "@/components/ui/sonner";
import { AppShell } from "@/components/app-shell";
import { CommandPalette } from "@/components/command-palette";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
      <AuthProvider>
        <AppShell>{children}</AppShell>
        <CommandPalette />
        <Toaster />
      </AuthProvider>
    </ThemeProvider>
  );
}
