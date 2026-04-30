import { useEffect, useRef } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { toast } from "sonner";
import { useAuthenticateWithGoogleMutation } from "@/api";
import { useAuthenticate, useTranslationNamespace } from "@/helpers";
import { LEGAL_VERSION } from "@/constants";
import type { MutationError } from "@/types";
import { Loader2 } from "lucide-react";

export const GoogleOAuthCallback = () => {
  const { t } = useTranslationNamespace(["common"]);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const code = searchParams.get("code");
  const error = searchParams.get("error");
  const authenticate = useAuthenticate();
  const exchangeKeyRef = useRef<string | null>(null);

  // Get the origin page to redirect back on error (sign-in is at "/")
  const getOriginPage = () => {
    const origin = sessionStorage.getItem("oauth_origin");
    if (!origin) {
      const hasInvitationToken = sessionStorage.getItem(
        "oauth_invitation_token",
      );
      return hasInvitationToken ? "/signup" : "/";
    }
    return origin;
  };

  const mutation = useAuthenticateWithGoogleMutation({
    onSuccess: (data) => {
      if (exchangeKeyRef.current) {
        sessionStorage.removeItem(exchangeKeyRef.current);
      }
      sessionStorage.removeItem("oauth_invitation_token");
      sessionStorage.removeItem("oauth_origin");
      authenticate(data.accessToken, data.refreshToken);
      navigate("/dashboard", { replace: true });
    },
    onError: (err: MutationError) => {
      if (err.error?.message) toast.error(err.error.message);
      if (exchangeKeyRef.current) {
        sessionStorage.removeItem(exchangeKeyRef.current);
      }
      const originPage = getOriginPage();
      sessionStorage.removeItem("oauth_invitation_token");
      sessionStorage.removeItem("oauth_origin");
      navigate(originPage);
    },
  });

  useEffect(() => {
    const storedInvitationToken = sessionStorage.getItem(
      "oauth_invitation_token",
    );

    if (!code || error) return;

    // React Strict Mode (dev) can remount/effect-twice. Prevent exchanging the
    // same Google `code` multiple times (the token exchange is one-shot).
    const exchangeKey = `oauth_google_exchange_inflight:${code}`;
    if (sessionStorage.getItem(exchangeKey) === "1") return;

    exchangeKeyRef.current = exchangeKey;
    sessionStorage.setItem(exchangeKey, "1");

    mutation.mutate({
      code,
      termsVersion: LEGAL_VERSION,
      invitationToken: storedInvitationToken || undefined,
    });
  }, [code, error]);

  useEffect(() => {
    if (error) {
      const originPage = getOriginPage();
      sessionStorage.removeItem("oauth_invitation_token");
      sessionStorage.removeItem("oauth_origin");
      navigate(originPage);
      return;
    }
  }, [error]);

  return (
    <div className="flex h-screen w-full items-center justify-center">
      <div className="flex items-center gap-3">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <span>{t("authenticating", { ns: "common" })}</span>
      </div>
    </div>
  );
};
