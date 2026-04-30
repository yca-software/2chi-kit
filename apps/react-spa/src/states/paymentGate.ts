import { create } from "zustand";

export interface PaymentGateState {
  bannerVisible: boolean;
  modalOpen: boolean;
  orgId: string | null;
  showBanner: (orgId: string) => void;
  hideBanner: () => void;
  openModal: () => void;
  closeModal: () => void;
}

export const usePaymentGateStore = create<PaymentGateState>((set) => ({
  bannerVisible: false,
  modalOpen: false,
  orgId: null,
  showBanner: (orgId) =>
    set((state) => ({
      ...state,
      bannerVisible: true,
      orgId,
    })),
  hideBanner: () =>
    set((state) => ({
      ...state,
      bannerVisible: false,
    })),
  openModal: () =>
    set((state) =>
      state.orgId
        ? {
            ...state,
            modalOpen: true,
          }
        : state,
    ),
  closeModal: () =>
    set((state) => ({
      ...state,
      modalOpen: false,
    })),
}));
