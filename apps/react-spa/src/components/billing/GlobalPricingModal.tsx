import { useGetOrganizationQuery } from "@/api";
import { PricingModal } from "./PricingModal";
import { usePricingModalStore } from "@/states";
import { useShallow } from "zustand/shallow";
import { Dialog, DialogContent } from "@yca-software/design-system";
import { Loader2 } from "lucide-react";

export function GlobalPricingModal() {
  const { open, orgId, checkoutIntent, priceId, close } = usePricingModalStore(
    useShallow((state) => ({
      open: state.open,
      orgId: state.orgId,
      checkoutIntent: state.checkoutIntent,
      priceId: state.priceId,
      close: state.close,
    })),
  );

  const enabled = open && !!orgId;
  const { data: organization, isLoading } = useGetOrganizationQuery(
    orgId ?? "",
  );

  if (!enabled) {
    return null;
  }

  return (
    <>
      <PricingModal
        open={open}
        onOpenChange={(nextOpen) => {
          if (!nextOpen) {
            close();
          }
        }}
        organization={organization ?? null}
        upgradeOnly={Boolean(organization?.paddleSubscriptionId)}
        initialCheckoutIntent={checkoutIntent}
        initialPriceId={priceId}
      />
      {open && isLoading && (
        <Dialog open>
          <DialogContent
            className="z-[200] flex max-w-sm items-center justify-center gap-2 border-0 bg-transparent shadow-none"
            overlayClassName="z-[199]"
            showCloseButton={false}
          >
            <Loader2
              className="h-8 w-8 animate-spin text-primary"
              aria-hidden
            />
            <span className="sr-only">Loading…</span>
          </DialogContent>
        </Dialog>
      )}
    </>
  );
}
