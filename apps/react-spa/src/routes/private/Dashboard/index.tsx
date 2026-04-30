import { useParams, Navigate } from "react-router";
import { useEffect } from "react";
import { Onboarding } from "@/components";
import { useUserState } from "@/states";
import { useShallow } from "zustand/shallow";
import { LEGAL_VERSION } from "@/constants";
import { resolveDefaultOrganizationId } from "@/helpers";
import { DashboardContent } from "./DashboardContent";

export const Dashboard = () => {
  const { orgId } = useParams<{ orgId: string }>();
  const { selectedOrgId, userData } = useUserState(
    useShallow((state) => ({
      selectedOrgId: state.selectedOrgId,
      userData: state.userData,
    })),
  );

  // Get roles from user state (populated by Root.tsx)
  const roles = userData.roles || [];
  const hasAcceptedTerms = userData.user?.termsVersion === LEGAL_VERSION;

  // If orgId in URL doesn't match selected org, update selection
  useEffect(() => {
    if (orgId && orgId !== selectedOrgId) {
      const orgExists = roles.some((r) => r.organizationId === orgId);
      if (orgExists) {
        useUserState.getState().setSelectedOrgId(orgId);
      }
    }
  }, [orgId, selectedOrgId, roles]);

  // If user has no roles, show onboarding modal only after they've accepted terms
  // (avoids stacking legal gate and onboarding modals)
  if (roles.length === 0) {
    if (hasAcceptedTerms) {
      return (
        <div className="relative flex flex-1 flex-col">
          <Onboarding
            open={true}
            onOpenChange={() => {}}
            showCloseButton={false}
          />
        </div>
      );
    }
    return <DashboardContent />;
  }

  if (roles.length > 0) {
    const orgIdInRoles = Boolean(
      orgId && roles.some((r) => r.organizationId === orgId),
    );
    if (!orgId || !orgIdInRoles) {
      const store = useUserState.getState();
      const resolved = resolveDefaultOrganizationId(roles, store.selectedOrgId);
      if (resolved) {
        if (resolved !== store.selectedOrgId) {
          store.setSelectedOrgId(resolved);
        }
        return <Navigate to={`/dashboard/${resolved}`} replace />;
      }
    }
  }

  const currentOrg = roles.find(
    (r) => r.organizationId === orgId || r.organizationId === selectedOrgId,
  );

  return <DashboardContent organizationName={currentOrg?.organizationName} />;
};
