import { Link, useLocation } from "react-router";
import { useUserState } from "@/states/user";
import { useShallow } from "zustand/shallow";
import { Shield, HelpCircle } from "lucide-react";
import { Tooltip, cn } from "@yca-software/design-system";
import { OrgSelector } from "./OrgSelector";
import { useTranslationNamespace } from "@/helpers";
import { navItems } from "./navItems";

export interface AppSidebarNavigationProps {
  collapsed: boolean;
  onClose?: () => void;
  onOpenSupport: () => void;
}

export function AppSidebarNavigation({
  collapsed,
  onClose,
  onOpenSupport,
}: AppSidebarNavigationProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const location = useLocation();
  const { userData, selectedOrgId } = useUserState(
    useShallow((state) => ({
      userData: state.userData,
      selectedOrgId: state.selectedOrgId,
    })),
  );

  return (
    <>
      <div
        className={cn(
          "flex shrink-0 items-center justify-center border-b transition-[padding] duration-200",
          collapsed ? "py-3 px-2" : "px-4 py-4",
        )}
      >
        <OrgSelector variant={collapsed ? "iconOnly" : "default"} />
      </div>

      <nav
        className={cn(
          "flex-1 space-y-1 overflow-auto transition-[padding] duration-200",
          collapsed ? "p-2" : "p-4",
        )}
      >
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive =
            location.pathname === item.url ||
            location.pathname.startsWith(`${item.url}/`);

          let href = item.url;
          if (selectedOrgId) {
            if (item.url === "/dashboard") {
              href = `/dashboard/${selectedOrgId}`;
            } else if (item.url === "/settings") {
              href = `/settings/${selectedOrgId}`;
            }
          }

          const linkContent = (
            <Link
              to={href}
              onClick={onClose}
              className={cn(
                "flex items-center rounded-lg text-sm font-medium transition-colors cursor-pointer",
                collapsed ? "justify-center p-2" : "gap-3 px-3 py-2",
                isActive
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
              )}
            >
              <Icon
                className={cn("h-5 w-5 shrink-0", collapsed && "h-5 w-5")}
              />
              {!collapsed && t(`settings:nav.${item.titleKey}`)}
            </Link>
          );

          return collapsed ? (
            <Tooltip
              key={item.titleKey}
              content={t(`settings:nav.${item.titleKey}`)}
              side="right"
            >
              {linkContent}
            </Tooltip>
          ) : (
            <span key={item.titleKey}>{linkContent}</span>
          );
        })}

        {userData.admin &&
          (collapsed ? (
            <Tooltip content={t("settings:nav.admin")} side="right">
              <Link
                to="/admin"
                onClick={onClose}
                className={cn(
                  "flex items-center rounded-lg text-sm font-medium transition-colors cursor-pointer justify-center p-2",
                  location.pathname.startsWith("/admin")
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
                )}
              >
                <Shield className="h-5 w-5 shrink-0" />
              </Link>
            </Tooltip>
          ) : (
            <Link
              to="/admin"
              onClick={onClose}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors cursor-pointer",
                location.pathname.startsWith("/admin")
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
              )}
            >
              <Shield className="h-5 w-5" />
              {t("settings:nav.admin")}
            </Link>
          ))}

        {collapsed ? (
          <Tooltip content={t("settings:support.contactSupport")} side="right">
            <button
              type="button"
              onClick={onOpenSupport}
              className="flex w-full cursor-pointer items-center justify-center rounded-lg p-2 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            >
              <HelpCircle className="h-5 w-5 shrink-0" />
            </button>
          </Tooltip>
        ) : (
          <button
            type="button"
            onClick={onOpenSupport}
            className="w-full flex cursor-pointer items-center gap-3 rounded-lg px-3 py-2 text-left text-sm font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
          >
            <HelpCircle className="h-4 w-4 shrink-0" />
            {t("settings:support.contactSupport")}
          </button>
        )}
      </nav>
    </>
  );
}
