import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@yca-software/design-system";
import { useTranslationNamespace, isAllowedPortalUrl } from "@/helpers";
import { useCreateCustomerPortalSessionMutation } from "@/api";
import { usePaymentGateStore } from "@/states";
import { useShallow } from "zustand/shallow";

export function PaymentGateModal() {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const { modalOpen, orgId, closeModal } = usePaymentGateStore(
    useShallow((state) => ({
      modalOpen: state.modalOpen,
      orgId: state.orgId,
      closeModal: state.closeModal,
    })),
  );
  const portalMutation = useCreateCustomerPortalSessionMutation(orgId ?? "", {
    onSuccess: (data) => {
      if (data.portalUrl && isAllowedPortalUrl(data.portalUrl)) {
        closeModal();
        window.location.href = data.portalUrl;
      }
    },
  });

  const handleManageBilling = () => {
    if (!orgId) return;
    portalMutation.mutate();
  };

  return (
    <AlertDialog open={modalOpen} onOpenChange={(o) => !o && closeModal()}>
      <AlertDialogContent
        data-slot="payment-gate-content"
        className="gap-4 sm:gap-6"
      >
        <AlertDialogHeader className="gap-2">
          <AlertDialogTitle className="text-base sm:text-lg">
            {t("settings:org.subscription.paymentGateTitle")}
          </AlertDialogTitle>
          <AlertDialogDescription className="text-sm">
            {t("settings:org.subscription.paymentGateDescription")}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter className="flex flex-col-reverse gap-2 pt-2 sm:flex-row sm:justify-end sm:gap-2 sm:pt-0">
          <AlertDialogCancel
            onClick={() => closeModal()}
            disabled={portalMutation.isPending}
            className="m-0 w-full sm:w-auto"
          >
            {t("common:close")}
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={handleManageBilling}
            disabled={portalMutation.isPending}
            className="w-full sm:w-auto"
          >
            {portalMutation.isPending
              ? t("common:loading")
              : t("settings:org.subscription.manageBilling")}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
