"use client";

import { useState } from "react";
import { usePathname } from "next/navigation";
import { SidebarContent } from "@/components/sidebar";
import { Header } from "@/components/header";
import {
  Sheet,
  SheetContent,
  SheetTitle,
} from "@/components/ui/sheet";
import { useAuth } from "@/lib/auth-context";

const PUBLIC_PATHS = ["/login", "/register"];

export function AppShell({ children }: { children: React.ReactNode }) {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const pathname = usePathname();
  const { isAuthenticated, loading } = useAuth();

  const isPublicPage = PUBLIC_PATHS.includes(pathname);

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-4 border-muted border-t-primary" />
      </div>
    );
  }

  if (isPublicPage || !isAuthenticated) {
    return <>{children}</>;
  }

  return (
    <div className="flex h-screen overflow-hidden">
      {/* Desktop sidebar */}
      <aside className="hidden w-60 shrink-0 lg:block">
        <SidebarContent />
      </aside>

      {/* Mobile sidebar */}
      <Sheet open={sidebarOpen} onOpenChange={setSidebarOpen}>
        <SheetContent side="left" className="w-60 p-0" showCloseButton={false}>
          <SheetTitle className="sr-only">Navigation</SheetTitle>
          <SidebarContent />
        </SheetContent>
      </Sheet>

      {/* Main area */}
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header onMenuClick={() => setSidebarOpen(true)} />
        <main className="flex-1 overflow-y-auto p-6">{children}</main>
      </div>
    </div>
  );
}
