import { removeCookieConsentCookie, setCookieConsentCookie } from "@/helpers";
import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface CookieConsent {
  analytics: boolean;
  timestamp: string;
}

export interface CookieConsentState {
  consent: CookieConsent | null;
  setCookieConsent: (timestamp: string, analytics: boolean) => void;
  clearConsentTimestamp: () => void;
}

export const useCookieConsent = create<CookieConsentState>()(
  persist(
    (set) => ({
      consent: null,
      setCookieConsent: (timestamp: string, analytics: boolean) => {
        set({ consent: { analytics, timestamp } });
        setCookieConsentCookie(timestamp, analytics);
      },
      clearConsentTimestamp: () => {
        set({ consent: null });
        removeCookieConsentCookie();
      },
    }),
    {
      name: "cookie-consent-storage",
    },
  ),
);
