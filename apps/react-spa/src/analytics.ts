import posthog from "posthog-js";

const posthogKey = import.meta.env.VITE_PUBLIC_POSTHOG_KEY;
const posthogHost =
  import.meta.env.VITE_PUBLIC_POSTHOG_HOST ?? "https://eu.i.posthog.com";

const sessionRecordingEnabled =
  import.meta.env.VITE_PUBLIC_POSTHOG_SESSION_RECORDING === "true";

let initialized = false;

export const isAnalyticsEnabled = (): boolean =>
  typeof posthogKey === "string" && posthogKey.length > 0;

export function initAnalytics(): void {
  if (!isAnalyticsEnabled() || initialized) {
    return;
  }

  if (typeof window !== "undefined" && !navigator.onLine) {
    const onOnline = () => {
      window.removeEventListener("online", onOnline);
      initAnalytics();
    };
    window.addEventListener("online", onOnline);
    return;
  }

  try {
    posthog.init(posthogKey, {
      api_host: posthogHost,
      capture_pageview: false,
      person_profiles: "identified_only",
      respect_dnt: true,
      disable_session_recording: !sessionRecordingEnabled,
    });
    initialized = true;
  } catch {
    // Keep app usable if init throws in constrained environments.
  }
}

export { posthog };
