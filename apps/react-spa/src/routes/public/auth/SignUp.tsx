import { useSignUpMutation } from "@/api";
import {
  useAuthenticate,
  useTranslationNamespace,
  captureCheckoutIntentFromSearchParams,
} from "@/helpers";
import { useSearchParams } from "react-router";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useEffect, useMemo } from "react";
import { useNavigate } from "react-router";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Button,
  Form,
  Link,
  Alert,
  AlertDescription,
} from "@yca-software/design-system";
import {
  EmailField,
  FirstNameField,
  LastNameField,
  PasswordField,
  GoogleButton,
} from "@/components";
import { Trans } from "react-i18next";
import { LEGAL_VERSION } from "@/constants";
import { UserPlus } from "lucide-react";

export const SignUp = () => {
  const { t, isLoading } = useTranslationNamespace(["auth", "common"]);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const invitationToken = searchParams.get("invitationToken");
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
      { pathname: "/signup", search: next ? `?${next}` : "" },
      { replace: true },
    );
  }, [navigate]);

  const signUpSchema = useMemo(() => {
    if (isLoading) {
      return z.object({
        firstName: z.string(),
        lastName: z.string(),
        email: z.string(),
        password: z.string(),
      });
    }
    return z.object({
      firstName: z
        .string()
        .min(1, t("auth:signUp.validation.firstNameRequired")),
      lastName: z.string().min(1, t("auth:signUp.validation.lastNameRequired")),
      email: z
        .string()
        .min(1, t("auth:signUp.validation.emailRequired"))
        .refine((val) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(val), {
          message: t("auth:signUp.validation.emailInvalid"),
        }),
      password: z
        .string()
        .min(8, t("auth:signUp.validation.passwordMinLength")),
    });
  }, [t, isLoading]);

  type SignUpFormData = z.infer<typeof signUpSchema>;

  const form = useForm<SignUpFormData>({
    resolver: zodResolver(signUpSchema),
    defaultValues: {
      firstName: "",
      lastName: "",
      email: "",
      password: "",
    },
  });

  const {
    mutate: signUp,
    isPending,
    error,
  } = useSignUpMutation({
    onSuccess: (data) => {
      authenticate(data.accessToken, data.refreshToken);
    },
  });

  const onSubmit = (data: SignUpFormData) => {
    signUp({
      ...data,
      termsVersion: LEGAL_VERSION,
      invitationToken: invitationToken || undefined,
    });
  };

  if (isLoading) {
    return null;
  }

  return (
    <Card className="w-full">
      <CardHeader className="space-y-1">
        <CardTitle className="text-2xl">{t("auth:signUp.title")}</CardTitle>
        <CardDescription>
          <Trans
            i18nKey="auth:signUp.orSignIn"
            components={[<Link to="/" key="signin-link" />]}
          />
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            {error && (
              <Alert variant="destructive">
                <AlertDescription>
                  {error?.error?.message || t("defaultError", { ns: "common" })}
                </AlertDescription>
              </Alert>
            )}

            {invitationToken && (
              <Alert className="border-primary/40 bg-primary/5 [&>svg]:text-primary">
                <UserPlus aria-hidden />
                <AlertDescription>
                  {t("auth:signUp.invitationNotice")}
                </AlertDescription>
              </Alert>
            )}

            <div className="grid grid-cols-2 gap-4">
              <FirstNameField control={form.control} name="firstName" />
              <LastNameField control={form.control} name="lastName" />
            </div>

            <EmailField
              control={form.control}
              name="email"
              autoComplete="email"
            />

            <PasswordField
              control={form.control}
              name="password"
              autoComplete="new-password"
              placeholder={t("auth:signUp.passwordPlaceholder")}
            />

            <Button type="submit" className="w-full" disabled={isPending}>
              {isPending
                ? t("loading", { ns: "common" })
                : t("auth:signUp.submitButton")}
            </Button>
          </form>

          <div className="relative my-6">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t" />
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-card px-2 text-muted-foreground">
                {t("auth:signUp.orDivider")}
              </span>
            </div>
          </div>

          <GoogleButton
            invitationToken={invitationToken}
            disabled={isPending}
          />
        </Form>

        <p className="mt-4 text-center text-xs text-muted-foreground">
          <Trans
            i18nKey="auth:signUp.agreementText"
            components={[
              <a
                href="/terms-of-service"
                target="_blank"
                rel="noopener noreferrer"
                key="terms-link"
                className="underline underline-offset-4 hover:text-primary"
              />,
            ]}
          />
        </p>
      </CardContent>
    </Card>
  );
};
