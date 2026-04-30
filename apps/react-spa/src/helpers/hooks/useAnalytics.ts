import { useCallback } from "react";
import { posthog } from "@/analytics";

export function useAnalytics() {
  const capture = useCallback(
    (eventName: string, properties?: Record<string, unknown>) => {
      posthog?.capture(eventName, properties);
    },
    [],
  );

  const identify = useCallback(
    (userId: string, traits?: Record<string, unknown>) => {
      posthog?.identify(userId, traits);
    },
    [],
  );

  const group = useCallback(
    (groupType: string, groupKey: string, traits?: Record<string, unknown>) => {
      posthog?.group(groupType, groupKey, traits);
    },
    [],
  );

  return { capture, identify, group };
}
