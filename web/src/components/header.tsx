"use client";

import { LogOut, Menu } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/lib/auth-context";
import { ThemeToggle } from "@/components/theme-toggle";

interface HeaderProps {
  onMenuClick: () => void;
}

export function Header({ onMenuClick }: HeaderProps) {
  const { user, logout } = useAuth();

  return (
    <header className="flex h-14 items-center gap-4 border-b bg-background px-4">
      <Button
        variant="ghost"
        size="icon"
        className="lg:hidden"
        onClick={onMenuClick}
      >
        <Menu className="size-5" />
        <span className="sr-only">Toggle sidebar</span>
      </Button>

      <div className="flex-1" />

      <ThemeToggle />

      {user && (
        <div className="flex items-center gap-3">
          <span className="text-sm text-muted-foreground">{user.email}</span>
          <Button variant="ghost" size="icon" onClick={logout}>
            <LogOut className="size-4" />
            <span className="sr-only">Log out</span>
          </Button>
        </div>
      )}
    </header>
  );
}
