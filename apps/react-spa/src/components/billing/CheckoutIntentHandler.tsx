import { useEffect, useRef } from "react";
import { useUserState } from "@/states/user";
import { useShallow } from "zustand/shallow";
import { consumeCheckoutIntent } from "@/helpers/billing/checkoutIntent";
import { resolveDefaultOrganizationId } from "@/helpers";
import { usePricingModalStore } from "@/states/pricingModal";

/**
 * After auth + org membership, opens the pricing modal if the user arrived from a marketing
 * pricing CTA (`/signup?checkoutIntent=…`).
 */
export function CheckoutIntentHandler() {
  const { roles, isUserProfileReady } = useUserState(
    useShallow((s) => ({
      roles: s.userData.roles,
      isUserProfileReady: s.isUserProfileReady,
    })),
  );
  const ran = useRef(false);

  useEffect(() => {
    if (!isUserProfileReady || ran.current) return;
    if (!roles || roles.length === 0) return;

    const orgId = resolveDefaultOrganizationId(
      roles,
      useUserState.getState().selectedOrgId,
    );
    if (!orgId) return;

    const intent = consumeCheckoutIntent();
    if (!intent) return;

    ran.current = true;
    usePricingModalStore.getState().openForOrg(orgId, {
      checkoutIntent: intent.checkoutIntent,
      priceId: intent.priceId,
    });
  }, [isUserProfileReady, roles]);

  return null;
}
