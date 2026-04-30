import { create } from "zustand";

interface OpenForOrgOptions {
  checkoutIntent?: string | null;
  priceId?: string | null;
}

export interface PricingModalState {
  open: boolean;
  orgId: string | null;
  checkoutIntent: string | null;
  priceId: string | null;
  openForOrg: (orgId: string, options?: OpenForOrgOptions) => void;
  close: () => void;
}

export const usePricingModalStore = create<PricingModalState>((set) => ({
  open: false,
  orgId: null,
  checkoutIntent: null,
  priceId: null,
  openForOrg: (orgId, options) =>
    set({
      open: true,
      orgId,
      checkoutIntent: options?.checkoutIntent ?? null,
      priceId: options?.priceId ?? null,
    }),
  close: () =>
    set({
      open: false,
      orgId: null,
      checkoutIntent: null,
      priceId: null,
    }),
}));
