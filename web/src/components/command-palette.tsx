"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Calculator, Calendar, CreditCard, Settings, Smile, User, Server, KeyRound, ShieldAlert } from "lucide-react";

import {
    CommandDialog,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList,
    CommandSeparator,
    CommandShortcut,
} from "@/components/ui/command";

export function CommandPalette() {
    const [open, setOpen] = useState(false);
    const router = useRouter();

    useEffect(() => {
        const down = (e: KeyboardEvent) => {
            if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
                e.preventDefault();
                setOpen((open) => !open);
            }
        };

        document.addEventListener("keydown", down);
        return () => document.removeEventListener("keydown", down);
    }, []);

    const runCommand = (command: () => void) => {
        setOpen(false);
        command();
    };

    return (
        <CommandDialog open={open} onOpenChange={setOpen}>
            <CommandInput placeholder="Type a command or search..." />
            <CommandList>
                <CommandEmpty>No results found.</CommandEmpty>
                <CommandGroup heading="Navigation">
                    <CommandItem onSelect={() => runCommand(() => router.push("/"))}>
                        <Calendar className="mr-2 h-4 w-4" />
                        <span>Dashboard</span>
                    </CommandItem>
                    <CommandItem onSelect={() => runCommand(() => router.push("/servers"))}>
                        <Server className="mr-2 h-4 w-4" />
                        <span>Servers</span>
                    </CommandItem>
                    <CommandItem onSelect={() => runCommand(() => router.push("/vault"))}>
                        <KeyRound className="mr-2 h-4 w-4" />
                        <span>Vault</span>
                    </CommandItem>
                    <CommandItem onSelect={() => runCommand(() => router.push("/sentry"))}>
                        <ShieldAlert className="mr-2 h-4 w-4" />
                        <span>Sentry (Firewall)</span>
                    </CommandItem>
                </CommandGroup>
                <CommandSeparator />
                <CommandGroup heading="Settings">
                    <CommandItem onSelect={() => runCommand(() => router.push("/sentry?tab=budget"))}>
                        <CreditCard className="mr-2 h-4 w-4" />
                        <span>API Token Budget</span>
                    </CommandItem>
                    <CommandItem onSelect={() => runCommand(() => router.push("/login"))}>
                        <User className="mr-2 h-4 w-4" />
                        <span>Logout / Switch User</span>
                    </CommandItem>
                </CommandGroup>
            </CommandList>
        </CommandDialog>
    );
}
