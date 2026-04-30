import { Trans } from "react-i18next";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useEffect, useMemo } from "react";
import { useNavigate } from "react-router";
import {
  captureCheckoutIntentFromSearchParams,
  useAuthenticate,
  useTranslationNamespace,
} from "@/helpers";
import {
  useAuthenticateWithPasswordMutation,
  AuthenticateResponse,
} from "@/api";
import { Loader2 } from "lucide-react";
import {
  Alert,
  AlertDescription,
  Button,
  Form,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Link,
} from "@yca-software/design-system";
import { EmailField, PasswordField, GoogleButton } from "@/components";

type SignInFormData = {
  email: string;
  password: string;
};

export const SignIn = () => {
  const { t, isLoading } = useTranslationNamespace(["auth", "common"]);
  const navigate = useNavigate();
  const authenticate = useAuthenticate();

  useEffect(() => {
    const search = window.location.search;
    if (!search) return;
    captureCheckoutIntentFromSearchParams(search);
    const params = new URLSearchParams(search);
    const had =
      params.has("checkoutIntent") ||
      params.has("plan") ||
      params.has("priceId");
    if (!had) return;
    params.delete("checkoutIntent");
    params.delete("plan");
    params.delete("priceId");
    const next = params.toString();
    navigate(
      { pathname: "/", search: next ? `?${next}` : "" },
      { replace: true },
    );
  }, [navigate]);

  const signInSchema = useMemo(() => {
    if (isLoading) {
      return z.object({
        email: z.string(),
        password: z.string(),
      });
    }
    return z.object({
      email: z
        .string()
        .min(1, t("auth:signIn.validation.emailRequired"))
        .refine((val) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(val), {
          message: t("auth:signIn.validation.emailInvalid"),
        }),
      password: z.string().min(1, t("auth:signIn.validation.passwordRequired")),
    });
  }, [t, isLoading]);

  const form = useForm<SignInFormData>({
    resolver: zodResolver(signInSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  const {
    mutate: signIn,
    isPending,
    error,
  } = useAuthenticateWithPasswordMutation({
    onSuccess: (data: AuthenticateResponse) => {
      authenticate(data.accessToken, data.refreshToken);
      // Language will be updated in Root.tsx after user data loads
    },
  });

  const onSubmit = (data: SignInFormData) => {
    signIn(data);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <Card className="w-full">
      <CardHeader className="space-y-1">
        <CardTitle className="text-2xl">{t("auth:signIn.title")}</CardTitle>
        <CardDescription>
          <Trans
            i18nKey="auth:signIn.orCreateAccount"
            components={[<Link to="/signup" key="signup-link" />]}
          />
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            {error && (
              <Alert variant="destructive">
                <AlertDescription>
                  {error.error?.message || t("defaultError", { ns: "common" })}
                </AlertDescription>
              </Alert>
            )}

            <EmailField
              control={form.control}
              name="email"
              autoComplete="email"
            />

            <PasswordField
              control={form.control}
              name="password"
              autoComplete="current-password"
              rightLabel={
                <Link to="/forgot-password" className="text-sm">
                  {t("auth:signIn.forgotPassword")}
                </Link>
              }
            />

            <Button type="submit" className="w-full" disabled={isPending}>
              {isPending
                ? t("loading", { ns: "common" })
                : t("auth:signIn.submitButton")}
            </Button>
          </form>

          <div className="relative my-6">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-card px-2 text-muted-foreground">
                {t("auth:signIn.orDivider")}
              </span>
            </div>
          </div>

          <GoogleButton />
        </Form>
      </CardContent>
    </Card>
  );
};
