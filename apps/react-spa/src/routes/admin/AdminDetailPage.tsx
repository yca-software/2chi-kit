import { Link } from "react-router";
import { Button } from "@yca-software/design-system";
import { ArrowLeft, Loader2 } from "lucide-react";

export interface AdminDetailPageProps {
  backHref: string;
  backLabel: string;
  isLoading: boolean;
  isError: boolean;
  notFoundMessage: string;
  children: React.ReactNode;
  headerActions?: React.ReactNode;
}

/**
 * Reusable admin detail page: back button, loading/error states, content.
 * Mobile-friendly: stacked layout on small screens.
 */
export function AdminDetailPage({
  backHref,
  backLabel,
  isLoading,
  isError,
  notFoundMessage,
  children,
  headerActions,
}: AdminDetailPageProps) {
  if (isLoading) {
    return (
      <div
        className="flex min-h-[40vh] items-center justify-center"
        role="status"
        aria-live="polite"
      >
        <Loader2 className="h-8 w-8 animate-spin text-primary" aria-hidden />
        <span className="sr-only">Loading...</span>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="space-y-4 sm:space-y-6">
        <p className="text-destructive">{notFoundMessage}</p>
        <Link to={backHref}>
          <Button variant="outline">{backLabel}</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-4 sm:space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <Link to={backHref}>
          <Button variant="ghost" size="sm">
            <ArrowLeft className="mr-2 h-4 w-4" aria-hidden />
            {backLabel}
          </Button>
        </Link>
        {headerActions}
      </div>
      {children}
    </div>
  );
}
