import { useForgotPasswordMutation } from "@/api";
import { useState, useMemo } from "react";
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
import { EmailField } from "@/components";
import { Mail, ArrowLeft } from "lucide-react";

export const ForgotPassword = () => {
  const { t, isLoading } = useTranslationNamespace(["auth", "common"]);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [submittedEmail, setSubmittedEmail] = useState("");

  const forgotPasswordSchema = useMemo(() => {
    if (isLoading) {
      return z.object({
        email: z.string(),
      });
    }
    return z.object({
      email: z
        .string()
        .min(1, t("auth:forgotPassword.validation.emailRequired"))
        .refine((val) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(val), {
          message: t("auth:forgotPassword.validation.emailInvalid"),
        }),
    });
  }, [t, isLoading]);

  type ForgotPasswordFormData = z.infer<typeof forgotPasswordSchema>;

  const form = useForm<ForgotPasswordFormData>({
    resolver: zodResolver(forgotPasswordSchema),
    defaultValues: {
      email: "",
    },
  });

  const {
    mutate: forgotPassword,
    isPending,
    error,
  } = useForgotPasswordMutation({
    onSuccess: () => {
      setSubmittedEmail(form.getValues("email"));
      setIsSubmitted(true);
    },
  });

  const onSubmit = (data: ForgotPasswordFormData) => {
    forgotPassword(data);
  };

  if (isLoading) {
    return null;
  }

  if (isSubmitted) {
    return (
      <Card className="w-full">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
            <Mail className="h-6 w-6 text-primary" />
          </div>
          <CardTitle className="text-2xl">
            {t("auth:forgotPassword.checkEmailTitle")}
          </CardTitle>
          <CardDescription>
            {t("auth:forgotPassword.checkEmailDescription")}{" "}
            <strong>{submittedEmail}</strong>
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-center text-sm text-muted-foreground">
            {t("auth:forgotPassword.checkEmailInstructions")}
          </p>
          <Button variant="outline" asChild className="w-full">
            <Link to="/">
              <ArrowLeft className="mr-2 h-4 w-4" />
              {t("auth:forgotPassword.backToSignIn")}
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
          {t("auth:forgotPassword.title")}
        </CardTitle>
        <CardDescription>
          {t("auth:forgotPassword.description")}
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
            <Button type="submit" className="w-full" disabled={isPending}>
              {isPending
                ? t("loading", { ns: "common" })
                : t("auth:forgotPassword.submitButton")}
            </Button>
          </form>
        </Form>
        <div className="mt-4 text-center">
          <Link to="/" className="text-sm inline-flex items-center gap-1">
            <ArrowLeft className="h-3 w-3" />
            {t("auth:forgotPassword.backToSignIn")}
          </Link>
        </div>
      </CardContent>
    </Card>
  );
};
