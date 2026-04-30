import { useCallback, useSyncExternalStore } from "react";

function useMediaQuery(
  query: string,
  getServerSnapshot: () => boolean = () => false,
): boolean {
  const subscribe = useCallback(
    (onStoreChange: () => void) => {
      const mql = window.matchMedia(query);
      mql.addEventListener("change", onStoreChange);
      return () => mql.removeEventListener("change", onStoreChange);
    },
    [query],
  );

  const getSnapshot = useCallback(
    () => window.matchMedia(query).matches,
    [query],
  );

  return useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
}

/** Viewport width is at least `minWidthPx` (Tailwind-style min-width breakpoint). */
export function useMinWidth(minWidthPx: number): boolean {
  return useMediaQuery(`(min-width: ${minWidthPx}px)`);
}

/** Viewport width is at most `maxWidthPx` (e.g. "mobile" vs `md - 1`). */
export function useMaxWidth(maxWidthPx: number): boolean {
  return useMediaQuery(`(max-width: ${maxWidthPx}px)`);
}
