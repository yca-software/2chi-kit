import { NavLink, useLocation, useNavigate } from "react-router";
import type { ComponentType } from "react";
import { useTranslationNamespace } from "@/helpers";
import * as DesignSystem from "@yca-software/design-system";
import {
  useSettingsPermissions,
  SETTINGS_SECTIONS,
} from "../useSettingsPermissions";

type IconProps = {
  className?: string;
};

type SettingsSelectProps = {
  value: string;
  onValueChange: (value: string) => void;
  options: Array<{ value: string; label: string }>;
  placeholder?: string;
  "aria-label"?: string;
};

const cn = (...classes: Array<string | false | null | undefined>) =>
  classes.filter(Boolean).join(" ");

const Select = (
  DesignSystem as unknown as { Select: ComponentType<SettingsSelectProps> }
).Select;

const SettingsIcon = ({ className }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    className={className}
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <circle cx="12" cy="12" r="3" />
    <path d="M19.4 15a1.7 1.7 0 0 0 .34 1.87l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.7 1.7 0 0 0-1.87-.34 1.7 1.7 0 0 0-1 1.55V22a2 2 0 1 1-4 0v-.09a1.7 1.7 0 0 0-1-1.55 1.7 1.7 0 0 0-1.87.34l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.7 1.7 0 0 0 .34-1.87 1.7 1.7 0 0 0-1.55-1H2a2 2 0 1 1 0-4h.09a1.7 1.7 0 0 0 1.55-1 1.7 1.7 0 0 0-.34-1.87l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.7 1.7 0 0 0 1.87.34h.01a1.7 1.7 0 0 0 1-1.55V2a2 2 0 1 1 4 0v.09a1.7 1.7 0 0 0 1 1.55h.01a1.7 1.7 0 0 0 1.87-.34l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.7 1.7 0 0 0-.34 1.87v.01a1.7 1.7 0 0 0 1.55 1H22a2 2 0 1 1 0 4h-.09a1.7 1.7 0 0 0-1.55 1V15z" />
  </svg>
);

const UsersIcon = ({ className }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    className={className}
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
    <circle cx="9" cy="7" r="4" />
    <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
    <path d="M16 3.13a4 4 0 0 1 0 7.75" />
  </svg>
);

const TeamsIcon = ({ className }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    className={className}
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
    <circle cx="9" cy="7" r="4" />
    <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
    <circle cx="17" cy="8" r="3" />
  </svg>
);

const ShieldIcon = ({ className }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    className={className}
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <path d="M12 22s8-4 8-10V6l-8-4-8 4v6c0 6 8 10 8 10z" />
  </svg>
);

const KeyIcon = ({ className }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    className={className}
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <circle cx="7.5" cy="15.5" r="5.5" />
    <path d="m21 2-9.6 9.6" />
    <path d="m15.5 7.5 3 3L22 7l-3-3" />
  </svg>
);

const FileTextIcon = ({ className }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    className={className}
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
    <path d="M14 2v6h6" />
    <path d="M16 13H8" />
    <path d="M16 17H8" />
    <path d="M10 9H8" />
  </svg>
);

const settingsNavItems = [
  { path: "general", icon: SettingsIcon, labelKey: "general", exact: true },
  { path: "roles", icon: ShieldIcon, labelKey: "roles" },
  { path: "teams", icon: TeamsIcon, labelKey: "teams" },
  { path: "members", icon: UsersIcon, labelKey: "members" },
  { path: "api-keys", icon: KeyIcon, labelKey: "apiKeys" },
  { path: "audit-log", icon: FileTextIcon, labelKey: "auditLog" },
];

interface SettingsNavProps {
  currentOrgId: string;
}

export function SettingsNav({ currentOrgId }: SettingsNavProps) {
  const { t } = useTranslationNamespace(["settings"]);
  const location = useLocation();
  const navigate = useNavigate();
  const permissions = useSettingsPermissions();
  const basePath = `/settings/${currentOrgId}`;

  const visibleItems = settingsNavItems.filter((item) => {
    const section = SETTINGS_SECTIONS.find((s) => s.path === item.path);
    if (!section) return true;
    const sectionPerms = permissions[section.permissionKey];
    return sectionPerms?.canRead === true;
  });

  const navItemsWithPath = visibleItems.map((item) => {
    const fullPath = `${basePath}/${item.path}`;
    const isActive = item.exact
      ? location.pathname === fullPath
      : location.pathname.startsWith(fullPath);
    return { ...item, fullPath, isActive };
  });

  const activePath =
    navItemsWithPath.find((item) => item.isActive)?.fullPath ??
    navItemsWithPath[0]?.fullPath ??
    "";

  return (
    <>
      <div className="sm:hidden">
        <Select
          value={activePath}
          onValueChange={(value: string) => navigate(value)}
          options={navItemsWithPath.map((item) => ({
            value: item.fullPath,
            label: t(`settings:nav.${item.labelKey}`),
          }))}
          placeholder={t("settings:org.title")}
          aria-label={t("settings:org.title")}
        />
      </div>
      <nav className="hidden min-w-0 overflow-x-auto border-b sm:flex">
        {navItemsWithPath.map((item) => {
          const Icon = item.icon;

          return (
            <NavLink
              key={item.path}
              to={item.fullPath}
              end={item.exact}
              className={cn(
                "flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 -mb-px transition-colors whitespace-nowrap",
                item.isActive
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
    </>
  );
}
