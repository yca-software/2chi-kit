import { NavLink, useLocation } from "react-router";
import { useTranslationNamespace } from "@/helpers";
import { cn } from "@yca-software/design-system";
import { Settings, Users, Users2, Shield, Key, FileText } from "lucide-react";
import {
  useSettingsPermissions,
  SETTINGS_SECTIONS,
} from "../useSettingsPermissions";

const settingsNavItems = [
  { path: "general", icon: Settings, labelKey: "general", exact: true },
  { path: "roles", icon: Shield, labelKey: "roles" },
  { path: "teams", icon: Users2, labelKey: "teams" },
  { path: "members", icon: Users, labelKey: "members" },
  { path: "api-keys", icon: Key, labelKey: "apiKeys" },
  { path: "audit-log", icon: FileText, labelKey: "auditLog" },
];

interface SettingsNavProps {
  currentOrgId: string;
}

export function SettingsNav({ currentOrgId }: SettingsNavProps) {
  const { t } = useTranslationNamespace(["settings"]);
  const location = useLocation();
  const permissions = useSettingsPermissions();

  const visibleItems = settingsNavItems.filter((item) => {
    const section = SETTINGS_SECTIONS.find((s) => s.path === item.path);
    if (!section) return true;
    const sectionPerms = permissions[section.permissionKey];
    return sectionPerms?.canRead === true;
  });

  return (
    <nav className="flex min-w-0 overflow-x-auto border-b">
      {visibleItems.map((item) => {
        const Icon = item.icon;
        const basePath = `/settings/${currentOrgId}`;
        const fullPath = `${basePath}/${item.path}`;
        const isActive = item.exact
          ? location.pathname === fullPath
          : location.pathname.startsWith(fullPath);

        return (
          <NavLink
            key={item.path}
            to={fullPath}
            end={item.exact}
            className={cn(
              "flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 -mb-px transition-colors whitespace-nowrap",
              isActive
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground hover:border-muted-foreground/30",
            )}
          >
            <Icon className="h-4 w-4" />
            {t(`settings:nav.${item.labelKey}`)}
          </NavLink>
        );
      })}
    </nav>
  );
}
