import { Outlet, useLocation } from "react-router";
import {
  AppSidebar,
  PaymentGateModal,
  SubscriptionPaymentBanner,
  GlobalPricingModal,
} from "@/components";
import { LanguageSelector, ThemeToggle } from "@/components";
import { useCallback, useState } from "react";
import { Button } from "@yca-software/design-system";
import { PanelLeft, PanelLeftClose } from "lucide-react";
import { useTranslationNamespace, useMinWidth } from "@/helpers";
import { BREAKPOINT_PX } from "@/constants";

const SIDEBAR_STORAGE_KEY = "sidebar-collapsed";

function getStoredSidebarCollapsed(): boolean {
  try {
    return localStorage.getItem(SIDEBAR_STORAGE_KEY) === "true";
  } catch {
    return false;
  }
}

export const AppLayout = () => {
  const location = useLocation();
  const isLg = useMinWidth(BREAKPOINT_PX.lg);
  const { t, isLoading } = useTranslationNamespace(["settings", "common"]);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(
    getStoredSidebarCollapsed,
  );

  const collapsedOnDesktop = sidebarCollapsed && isLg;

  const handleToggleCollapse = useCallback(() => {
    setSidebarCollapsed((prev) => {
      const next = !prev;
      try {
        localStorage.setItem(SIDEBAR_STORAGE_KEY, String(next));
      } catch {}
      return next;
    });
  }, [setSidebarCollapsed]);

  if (isLoading) {
    return null;
  }

  return (
    <div className="flex h-screen min-h-0 w-full overflow-hidden">
      <PaymentGateModal />
      <GlobalPricingModal />
      <SubscriptionPaymentBanner />

      {/* Sidebar */}
      <aside
        className={`fixed inset-y-0 left-0 z-50 flex w-64 transform border-r bg-background transition-[transform,width] duration-200 ease-in-out lg:translate-x-0 ${
          sidebarOpen ? "translate-x-0" : "-translate-x-full"
        } ${collapsedOnDesktop ? "lg:w-16" : "lg:w-64"}`}
      >
        <AppSidebar
          onClose={() => setSidebarOpen(false)}
          collapsed={collapsedOnDesktop}
        />
      </aside>

      {/* Overlay for mobile */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-40 cursor-pointer bg-black/50 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Main content */}
      <div
        className={`flex min-w-0 flex-1 flex-col overflow-hidden transition-[padding] duration-200 ${
          collapsedOnDesktop ? "lg:pl-16" : "lg:pl-64"
        }`}
      >
        {/* Header */}
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-2 min-[480px]:px-4">
          <Button
            variant="ghost"
            size="icon"
            className="h-9 w-9 shrink-0 lg:hidden"
            onClick={() => setSidebarOpen(true)}
          >
            <PanelLeft className="h-5 w-5" />
          </Button>
          {isLg && (
            <Button
              variant="ghost"
              size="icon"
              className="hidden h-9 w-9 shrink-0 lg:flex"
              onClick={handleToggleCollapse}
              aria-label={
                collapsedOnDesktop ? "Expand sidebar" : "Collapse sidebar"
              }
            >
              {collapsedOnDesktop ? (
                <PanelLeft className="h-5 w-5" />
              ) : (
                <PanelLeftClose className="h-5 w-5" />
              )}
            </Button>
          )}
          <div className="min-w-0 flex-1">
            <h1 className="truncate text-base font-semibold min-[480px]:text-lg">
              {location.pathname.includes("/settings")
                ? t("settings:nav.settings")
                : t("settings:nav.dashboard")}
            </h1>
          </div>
          <div className="flex shrink-0 items-center gap-1 min-[480px]:gap-3">
            <ThemeToggle />
            <LanguageSelector variant="ghost" />
          </div>
        </header>

        <main className="min-h-0 flex-1 overflow-auto p-3 min-[480px]:p-4 md:p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
};
