import { useNavigate } from "react-router";
import { useQueryClient } from "@tanstack/react-query";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@yca-software/design-system";
import { Building2 } from "lucide-react";
import { CreateOrganizationResponse } from "@/api";
import { USER_QUERY_KEYS } from "@/constants";
import { useTranslationNamespace } from "@/helpers";
import { useUserState } from "@/states";
import { usePricingModalStore } from "@/states/pricingModal";
import { CreateOrgForm, buildRoleFromCreateResponse } from "./CreateOrgForm";

export interface OnboardingProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  showCloseButton?: boolean;
  title?: string;
  description?: string;
  /** Called after onboarding applies user/org state changes. */
  onComplete?: (data: CreateOrganizationResponse) => void;
  /** Defaults to true and redirects to the created org dashboard. */
  navigateToDashboard?: boolean;
}

/**
 * Canonical onboarding entry for organization creation.
 * Future steps (invite users, etc.) should be added here.
 */
export function Onboarding({
  open,
  onOpenChange,
  showCloseButton = true,
  title,
  description,
  onComplete,
  navigateToDashboard = true,
}: OnboardingProps) {
  const { t } = useTranslationNamespace(["dashboard", "common"]);
  const navigate = useNavigate();
  const applyCreatedOrganization = useApplyCreatedOrganization();
  const openPricingForOrg = usePricingModalStore((state) => state.openForOrg);

  const handleOrgCreated = (data: CreateOrganizationResponse) => {
    applyCreatedOrganization(data);
    if (navigateToDashboard) {
      navigate(`/dashboard/${data.organization.id}`, { replace: true });
    }
    onOpenChange(false);
    openPricingForOrg(data.organization.id);
    onComplete?.(data);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="w-full max-w-2xl max-h-[90vh] overflow-y-auto"
        showCloseButton={showCloseButton}
      >
        <DialogHeader>
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
            <Building2 className="h-8 w-8 text-primary" />
          </div>
          <DialogTitle className="text-2xl text-center">
            {title ?? t("dashboard:onboarding.title")}
          </DialogTitle>
          <DialogDescription className="text-center">
            {description ?? t("dashboard:onboarding.description")}
          </DialogDescription>
        </DialogHeader>
        <div className="mx-auto max-w-2xl space-y-8 px-4 py-6">
          <div className="space-y-4">
            <CreateOrgForm onCreated={handleOrgCreated} />
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

/** Applies created org to user store and invalidates current-user query. */
export function useApplyCreatedOrganization() {
  const setUserData = useUserState((state) => state.setUserData);
  const setSelectedOrgId = useUserState((state) => state.setSelectedOrgId);
  const queryClient = useQueryClient();

  return (data: CreateOrganizationResponse) => {
    const newRole = buildRoleFromCreateResponse(data);
    if (newRole) {
      const { userData } = useUserState.getState();
      setUserData({
        ...userData,
        roles: [...(userData.roles ?? []), newRole],
      });
    }
    setSelectedOrgId(data.organization.id);
    queryClient.invalidateQueries({ queryKey: [USER_QUERY_KEYS.CURRENT] });
  };
}
