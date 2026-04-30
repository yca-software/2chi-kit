import { useEffect } from "react";
import { Navigate } from "react-router";
import { useUserState } from "@/states";
import { Loader2 } from "lucide-react";
import {
  getAccessTokenFromCookies,
  getRefreshTokenFromCookies,
} from "@/helpers";

interface ProtectedRouteProps {
  children: React.ReactNode;
}

/** Renders children only when authenticated and profile ready; otherwise redirects or shows loading. Wraps content so protected layout does not mount when redirecting. */
export const ProtectedRoute = ({ children }: ProtectedRouteProps) => {
  const tokens = useUserState((state) => state.tokens);
  const setTokens = useUserState((state) => state.setTokens);
  const reset = useUserState((state) => state.reset);
  const isUserProfileReady = useUserState((state) => state.isUserProfileReady);
  const userData = useUserState((state) => state.userData);

  // No refresh token cookie = not authenticated (cookie is source of truth; state may be stale)
  const refreshTokenCookie = getRefreshTokenFromCookies();
  const accessTokenCookie = getAccessTokenFromCookies();
  if (!refreshTokenCookie) {
    if (tokens.refresh) {
      reset();
    }
    return <Navigate to="/" replace />;
  }

  // Restore tokens from cookies after hard refresh (we intentionally don't persist tokens in localStorage)
  useEffect(() => {
    if (!refreshTokenCookie) return;
    if (tokens.refresh) return;
    setTokens({
      access: accessTokenCookie ?? "",
      refresh: refreshTokenCookie,
    });
  }, [accessTokenCookie, refreshTokenCookie, setTokens, tokens.refresh]);

  // Wait for hydration to avoid redirect loops on hard refresh
  if (!tokens.refresh) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  // Still loading user data - show loading spinner
  if (!isUserProfileReady) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  // Loaded but no user = invalid session
  if (!userData.user) {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
};
