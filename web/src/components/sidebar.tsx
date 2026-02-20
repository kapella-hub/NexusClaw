"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Activity,
  Coins,
  KeyRound,
  LayoutDashboard,
  ScrollText,
  Server,
  Shield,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Separator } from "@/components/ui/separator";

interface NavItem {
  label: string;
  href: string;
  icon: React.ElementType;
}

const mainNav: NavItem[] = [
  { label: "Dashboard", href: "/", icon: LayoutDashboard },
  { label: "Servers", href: "/servers", icon: Server },
  { label: "Vault", href: "/vault", icon: KeyRound },
];

const sentryNav: NavItem[] = [
  { label: "Rules", href: "/sentry/rules", icon: ScrollText },
  { label: "Audit Log", href: "/sentry/audit", icon: Activity },
  { label: "Budget", href: "/sentry/budget", icon: Coins },
];

function NavLink({ item }: { item: NavItem }) {
  const pathname = usePathname();
  const active =
    item.href === "/" ? pathname === "/" : pathname.startsWith(item.href);

  return (
    <Link
      href={item.href}
      className={cn(
        "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
        active
          ? "bg-zinc-800 text-white"
          : "text-zinc-400 hover:bg-zinc-800/50 hover:text-white",
      )}
    >
      <item.icon className="size-4" />
      {item.label}
    </Link>
  );
}

export function SidebarContent() {
  return (
    <div className="flex h-full flex-col bg-zinc-950 text-zinc-100">
      <div className="flex h-14 items-center gap-2 px-4">
        <Shield className="size-6 text-blue-500" />
        <span className="text-lg font-bold tracking-tight">NexusClaw</span>
      </div>

      <Separator className="bg-zinc-800" />

      <nav className="flex-1 space-y-1 px-2 py-4">
        {mainNav.map((item) => (
          <NavLink key={item.href} item={item} />
        ))}

        <div className="px-3 pb-1 pt-4 text-xs font-semibold uppercase tracking-wider text-zinc-500">
          Sentry
        </div>
        {sentryNav.map((item) => (
          <NavLink key={item.href} item={item} />
        ))}
      </nav>
    </div>
  );
}
