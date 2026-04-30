import { Alert, AlertDescription, Button } from "@yca-software/design-system";
import { AlertCircle } from "lucide-react";
import { useTranslationNamespace } from "@/helpers";
import {
  useUserState,
  usePaymentGateStore,
  usePricingModalStore,
} from "@/states";
import { useShallow } from "zustand/shallow";
import { useGetOrganizationQuery } from "@/api";
import { isSubscriptionPastDue } from "@/helpers";

export function SubscriptionPaymentBanner() {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const { bannerVisible, orgId, openModal, showBanner } = usePaymentGateStore(
    useShallow((state) => ({
      bannerVisible: state.bannerVisible,
      orgId: state.orgId,
      openModal: state.openModal,
      showBanner: state.showBanner,
    })),
  );
  const { openForOrg } = usePricingModalStore(
    useShallow((state) => ({
      openForOrg: state.openForOrg,
    })),
  );
  const selectedOrgId = useUserState((s) => s.selectedOrgId);
  const { data: selectedOrg } = useGetOrganizationQuery(selectedOrgId ?? "");
  const pastDue = selectedOrg && isSubscriptionPastDue(selectedOrg);

  const effectiveOrgId =
    bannerVisible && orgId ? orgId : pastDue ? (selectedOrgId ?? null) : null;

  if (!effectiveOrgId) {
    return null;
  }

  const isPastDueBanner = pastDue && effectiveOrgId === selectedOrgId;

  const handleViewPlans = () => {
    openForOrg(effectiveOrgId);
  };

  const handleManageBilling = () => {
    if (bannerVisible && orgId) {
      openModal();
    } else {
      showBanner(effectiveOrgId);
      openModal();
    }
  };

  return (
    <div className="border-b bg-muted/40 px-3 py-2 text-xs sm:px-4 sm:text-sm">
      <Alert className="border-none bg-transparent p-0">
        <AlertCircle className="text-amber-500" />
        <AlertDescription className="flex flex-wrap items-center gap-2">
          <span className="font-medium">
            {isPastDueBanner
              ? t("settings:org.subscription.pastDueTitle")
              : t("settings:org.subscription.paymentGateTitle")}
          </span>
          <span className="text-muted-foreground">
            {isPastDueBanner
              ? t("settings:org.subscription.pastDueDescription")
              : t("settings:org.subscription.paymentGateDescription")}
          </span>
          <div className="ml-auto flex flex-wrap gap-2">
            <Button
              size="sm"
              variant="outline"
              onClick={handleViewPlans}
              className="h-8"
            >
              {t("settings:org.upsell.viewPlans")}
            </Button>
            <Button size="sm" onClick={handleManageBilling} className="h-8">
              {t("settings:org.subscription.manageBilling")}
            </Button>
          </div>
        </AlertDescription>
      </Alert>
    </div>
  );
}
