import { useEffect } from "react";
import { useLocation } from "react-router";
import { posthog } from "../analytics";

/**
 * Captures pageview events on route changes for SPA navigation.
 * Must be rendered inside the router and PostHog provider.
 */
export function PageviewTracker() {
  const location = useLocation();

  useEffect(() => {
    if (posthog && location.pathname) {
      posthog.capture("$pageview", {
        $current_url: window.location.href,
        pathname: location.pathname,
      });
    }
  }, [location.pathname, location.search]);

  return null;
}
