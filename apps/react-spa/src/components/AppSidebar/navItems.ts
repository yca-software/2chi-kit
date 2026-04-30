import { LayoutDashboard, Settings } from "lucide-react";
import type { LucideIcon } from "lucide-react";

export interface NavItem {
  titleKey: string;
  url: string;
  icon: LucideIcon;
}

export const navItems: NavItem[] = [
  { titleKey: "dashboard", url: "/dashboard", icon: LayoutDashboard },
  { titleKey: "settings", url: "/settings", icon: Settings },
];
