import { Navigate, Outlet } from "react-router";
import { useUserState } from "@/states/user";
import { useShallow } from "zustand/shallow";

/**
 * RequireOrganization ensures users have at least one organization.
 * If they don't, redirects them to the dashboard to complete onboarding.
 */
export const RequireOrganization = () => {
  const { userData } = useUserState(
    useShallow((state) => ({
      userData: state.userData,
    })),
  );

  const roles = userData.roles || [];

  // If user has no organizations, redirect to dashboard for onboarding
  if (roles.length === 0) {
    return <Navigate to="/dashboard" replace />;
  }

  return <Outlet />;
};
