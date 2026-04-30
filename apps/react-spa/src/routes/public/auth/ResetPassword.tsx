import { useResetPasswordMutation } from "@/api";
import { useState, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { useTranslationNamespace } from "@/helpers";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
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
import { PasswordField } from "@/components";
import { CheckCircle, XCircle, ArrowLeft } from "lucide-react";

export const ResetPassword = () => {
  const { t, isLoading } = useTranslationNamespace(["auth", "common"]);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const [isSubmitted, setIsSubmitted] = useState(false);

  const resetPasswordSchema = useMemo(() => {
    if (isLoading) {
      return z.object({
        password: z.string(),
        confirmPassword: z.string(),
      });
    }
    return z
      .object({
        password: z.string().min(8, {
          message: t("auth:resetPassword.validation.passwordMinLength"),
        }),
        confirmPassword: z.string().min(1, {
          message: t("auth:resetPassword.validation.confirmPasswordRequired"),
        }),
      })
      .refine((data) => data.password === data.confirmPassword, {
        message: t("auth:resetPassword.validation.passwordsDoNotMatch"),
        path: ["confirmPassword"],
      });
  }, [t, isLoading]);

  type ResetPasswordFormData = z.infer<typeof resetPasswordSchema>;

  const form = useForm<ResetPasswordFormData>({
    resolver: zodResolver(resetPasswordSchema),
    defaultValues: {
      password: "",
      confirmPassword: "",
    },
  });

  const {
    mutate: resetPassword,
    isPending,
    error,
  } = useResetPasswordMutation({
    onSuccess: () => {
      setIsSubmitted(true);
      setTimeout(() => {
        navigate("/");
      }, 3000);
    },
  });

  const onSubmit = (data: ResetPasswordFormData) => {
    if (!token) {
      return;
    }
    resetPassword({ token, password: data.password });
  };

  if (isLoading) {
    return null;
  }

  if (!token) {
    return (
      <Card className="w-full">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10">
            <XCircle className="h-6 w-6 text-destructive" />
          </div>
          <CardTitle className="text-2xl">
            {t("auth:resetPassword.invalidTokenTitle")}
          </CardTitle>
          <CardDescription>
            {t("auth:resetPassword.invalidTokenDescription")}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="outline" asChild className="w-full">
            <Link to="/forgot-password">
              {t("auth:resetPassword.requestNewLink")}
            </Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (isSubmitted) {
    return (
      <Card className="w-full">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
            <CheckCircle className="h-6 w-6 text-primary" />
          </div>
          <CardTitle className="text-2xl">
            {t("auth:resetPassword.successTitle")}
          </CardTitle>
          <CardDescription>
            {t("auth:resetPassword.successDescription")}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <Button variant="outline" asChild className="w-full">
            <Link to="/">
              <ArrowLeft className="mr-2 h-4 w-4" />
              {t("auth:resetPassword.backToSignIn")}
            </Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full">
      <CardHeader className="space-y-1">
        <CardTitle className="text-2xl">
          {t("auth:resetPassword.title")}
        </CardTitle>
        <CardDescription>{t("auth:resetPassword.description")}</CardDescription>
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
            <PasswordField
              control={form.control}
              name="password"
              label={t("auth:resetPassword.newPasswordLabel")}
              placeholder={t("auth:resetPassword.newPasswordPlaceholder")}
              autoComplete="new-password"
            />
            <PasswordField
              control={form.control}
              name="confirmPassword"
              label={t("auth:resetPassword.confirmPasswordLabel")}
              placeholder={t("auth:resetPassword.confirmPasswordPlaceholder")}
              autoComplete="new-password"
            />
            <Button type="submit" className="w-full" disabled={isPending}>
              {isPending
                ? t("loading", { ns: "common" })
                : t("auth:resetPassword.submitButton")}
            </Button>
          </form>
        </Form>
        <div className="mt-4 text-center">
          <Link to="/" className="text-sm inline-flex items-center gap-1">
            <ArrowLeft className="h-3 w-3" />
            {t("auth:resetPassword.backToSignIn")}
          </Link>
        </div>
      </CardContent>
    </Card>
  );
};
