import { Component, type ReactNode } from "react";
import { Button } from "@yca-software/design-system";
import { useTranslationNamespace } from "@/helpers";
import { useUserState } from "@/states/user";
import { isAnalyticsEnabled, posthog } from "@/analytics";

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

function ErrorBoundaryFallback({ onRetry }: { onRetry: () => void }) {
  const { t } = useTranslationNamespace(["common"]);
  return (
    <div className="flex min-h-[400px] flex-col items-center justify-center gap-4 p-8 text-center">
      <h2 className="text-xl font-semibold">
        {t("errorBoundary.title", { ns: "common" })}
      </h2>
      <p className="text-muted-foreground max-w-md">
        {t("errorBoundary.description", { ns: "common" })}
      </p>
      <Button onClick={onRetry}>
        {t("errorBoundary.refreshButton", { ns: "common" })}
      </Button>
    </div>
  );
}

/** Error boundaries must be class components; React only supports componentDidCatch/getDerivedStateFromError in classes. */
export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error("ErrorBoundary caught:", error, errorInfo);

    const state = useUserState.getState();
    const selectedOrgId = state.selectedOrgId;
    const location: Location | undefined =
      typeof window !== "undefined" ? window.location : undefined;

    if (isAnalyticsEnabled()) {
      try {
        posthog.capture("client_error", {
          message: error.message,
          name: error.name,
          stack: error.stack,
          componentStack: errorInfo.componentStack ?? undefined,
          route:
            location && typeof location.pathname === "string"
              ? (location.pathname as string)
              : undefined,
          pageUrl:
            location && typeof location.href === "string"
              ? (location.href as string)
              : undefined,
          orgId: selectedOrgId || undefined,
        });
      } catch {}
    }
  }

  render() {
    if (this.state.hasError && this.state.error) {
      if (this.props.fallback) {
        return this.props.fallback;
      }
      return (
        <ErrorBoundaryFallback
          onRetry={() => {
            this.setState({ hasError: false, error: undefined });
            window.location.reload();
          }}
        />
      );
    }
    return this.props.children;
  }
}
