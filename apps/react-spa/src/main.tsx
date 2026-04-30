import ReactDOM from "react-dom/client";
import { PostHogProvider } from "@posthog/react";
import { HelmetProvider } from "react-helmet-async";
import { Toaster } from "sonner";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider, TooltipProvider } from "@yca-software/design-system";

import { App } from "./App";
import { initAnalytics, isAnalyticsEnabled, posthog } from "./analytics";
import { evaluateRetry } from "./helpers";
import { useThemeStore } from "./states";
import "./index.css";
import "./i18n";

initAnalytics();

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      refetchOnReconnect: false,
      retry: (failureCount, error) => evaluateRetry(failureCount, error),
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      staleTime: 5 * 60 * 1000, // 5 minutes
    },
  },
});

const AppWithProviders = () => (
  <QueryClientProvider client={queryClient}>
    <HelmetProvider>
      <ThemeProvider useThemeStore={useThemeStore}>
        <TooltipProvider delayDuration={200}>
          <App />
          <Toaster richColors closeButton position="top-right" />
        </TooltipProvider>
      </ThemeProvider>
    </HelmetProvider>
  </QueryClientProvider>
);

const root = ReactDOM.createRoot(document.getElementById("root")!);
root.render(
  isAnalyticsEnabled() ? (
    <PostHogProvider client={posthog}>
      <AppWithProviders />
    </PostHogProvider>
  ) : (
    <AppWithProviders />
  ),
);
