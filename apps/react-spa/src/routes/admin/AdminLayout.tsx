import { useState } from "react";
import { Outlet, Link, useLocation } from "react-router";
import { useTranslationNamespace } from "@/helpers";
import {
  LayoutDashboard,
  Users,
  Building2,
  ArrowLeft,
  PanelLeft,
  X,
  Loader2,
} from "lucide-react";
import { Button, Separator, cn } from "@yca-software/design-system";
import { ThemeToggle, LanguageSelector } from "@/components";

const navItems = [
  { path: "/admin", icon: LayoutDashboard, labelKey: "nav.dashboard" },
  { path: "/admin/users", icon: Users, labelKey: "nav.users" },
  {
    path: "/admin/organizations",
    icon: Building2,
    labelKey: "nav.organizations",
  },
];

export function AdminLayout() {
  const { t, isLoading } = useTranslationNamespace(["admin"]);
  const location = useLocation();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="flex min-h-screen min-w-0 flex-col bg-background">
      {/* Top Bar */}
      <header className="bg-card shrink-0 border-b">
        <div className="mx-auto flex min-w-0 max-w-7xl items-center justify-between gap-2 px-3 py-3 sm:px-6 lg:px-8">
          <div className="flex min-w-0 flex-1 items-center gap-2 sm:gap-4">
            <Button
              variant="ghost"
              size="icon"
              className="h-9 w-9 shrink-0 lg:hidden"
              onClick={() => setSidebarOpen((o) => !o)}
              aria-label={
                sidebarOpen ? t("admin:menuClose") : t("admin:menuOpen")
              }
            >
              {sidebarOpen ? (
                <X className="h-5 w-5" />
              ) : (
                <PanelLeft className="h-5 w-5" />
              )}
            </Button>
            <Link to="/dashboard" className="shrink-0">
              <Button
                variant="ghost"
                size="sm"
                className="text-muted-foreground hover:text-foreground cursor-pointer"
              >
                <ArrowLeft className="mr-2 h-4 w-4" />
                <span className="hidden sm:inline">{t("admin:backToApp")}</span>
              </Button>
            </Link>
            <Separator orientation="vertical" className="hidden h-6 sm:block" />
            <h1 className="truncate text-base font-semibold sm:text-lg">
              {t("admin:title")}
            </h1>
          </div>
          <div className="flex shrink-0 items-center gap-1">
            <LanguageSelector variant="ghost" size="icon" />
            <ThemeToggle />
          </div>
        </div>
      </header>

      <div className="relative flex min-w-0 flex-1">
        {/* Sidebar - overlay on mobile, in-flow on lg */}
        {sidebarOpen && (
          <div
            className="fixed inset-0 z-40 bg-black/50 lg:hidden"
            onClick={() => setSidebarOpen(false)}
            aria-hidden
          />
        )}
        <aside
          className={cn(
            "fixed inset-y-0 left-0 z-50 w-64 transform border-r bg-card pt-16 transition-transform duration-200 lg:static lg:block lg:translate-x-0 lg:shrink-0 lg:pt-0",
            sidebarOpen ? "translate-x-0" : "-translate-x-full",
          )}
        >
          <nav className="space-y-1 p-4">
            {navItems.map((item) => {
              const isActive =
                location.pathname === item.path ||
                (item.path !== "/admin" &&
                  location.pathname.startsWith(item.path));

              return (
                <Link
                  key={item.path}
                  to={item.path}
                  onClick={() => setSidebarOpen(false)}
                  className={cn(
                    "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors cursor-pointer",
                    isActive
                      ? "bg-accent text-accent-foreground"
                      : "text-muted-foreground hover:text-foreground hover:bg-accent/50",
                  )}
                >
                  <item.icon className="h-5 w-5 shrink-0" />
                  {t(`admin:${item.labelKey}`)}
                </Link>
              );
            })}
          </nav>
        </aside>

        {/* Main Content */}
        <main className="min-w-0 flex-1 p-4 sm:p-6 lg:p-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
