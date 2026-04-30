import { SiteHeader } from "@/components/SiteHeader";
import { resolveDefaultOrganizationId, useTranslationNamespace } from "@/helpers";
import { useUserState } from "@/states/user";
import { Navigate, Outlet } from "react-router";
import { useMemo } from "react";

export const AuthLayout = () => {
  const { t, isLoading } = useTranslationNamespace("auth");
  const tokens = useUserState((state) => state.tokens);
  const isUserProfileReady = useUserState((state) => state.isUserProfileReady);
  const userData = useUserState((state) => state.userData);

  // Determine redirect path based on user's organization membership
  const redirectPath = useMemo(() => {
    if (!tokens.refresh || !isUserProfileReady || !userData.user) {
      return null;
    }

    const roles = userData.roles;

    if (roles && roles.length > 0) {
      const preferred = resolveDefaultOrganizationId(
        roles,
        useUserState.getState().selectedOrgId,
      );
      if (preferred) return `/dashboard/${preferred}`;
    }

    // If user has no organizations, they need onboarding
    // For now, redirect to dashboard which will handle onboarding
    return "/dashboard";
  }, [tokens.refresh, isUserProfileReady, userData]);

  if (redirectPath) {
    return <Navigate to={redirectPath} replace />;
  }

  return (
    <div className="grid min-h-svh lg:grid-cols-2">
      <div className="flex flex-col">
        <SiteHeader />
        <div className="flex min-w-0 flex-1 items-center justify-center px-3 pb-10 min-[400px]:px-4">
          <div className="w-full max-w-sm min-w-0">
            <Outlet />
          </div>
        </div>
      </div>
      <div className="bg-muted relative hidden lg:block">
        {!isLoading && (
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="max-w-lg space-y-6 p-10 text-center">
              <h2 className="text-3xl font-bold tracking-tight">
                {t("auth:heroTitle")}
              </h2>
              <p className="text-muted-foreground text-lg">
                {t("auth:heroDescription")}
              </p>
            </div>
          </div>
        )}
        <div className="absolute inset-0 bg-linear-to-br from-primary/5 to-primary/10" />
      </div>
    </div>
  );
};
