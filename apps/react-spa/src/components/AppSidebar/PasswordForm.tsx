import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Form,
  Button,
} from "@yca-software/design-system";
import { PasswordField } from "@/components";
import { Loader2, Check, X, Key } from "lucide-react";
import { useTranslationNamespace } from "@/helpers";
import { useMemo } from "react";

type PasswordFormData = {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
};

interface PasswordFormProps {
  onSubmit: (data: { currentPassword: string; newPassword: string }) => void;
  onCancel: () => void;
  isPending: boolean;
  isSuccess: boolean;
}

export function PasswordForm({
  onSubmit,
  onCancel,
  isPending,
  isSuccess,
}: PasswordFormProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const passwordSchema = useMemo(
    () =>
      z
        .object({
          currentPassword: z
            .string()
            .min(1, t("settings:user.validation.currentPasswordRequired")),
          newPassword: z
            .string()
            .min(8, t("settings:user.validation.passwordMinLength")),
          confirmPassword: z
            .string()
            .min(1, t("settings:user.validation.confirmPasswordRequired")),
        })
        .refine((data) => data.newPassword === data.confirmPassword, {
          message: t("settings:user.validation.passwordsDoNotMatch"),
          path: ["confirmPassword"],
        }),
    [t],
  );

  const form = useForm<PasswordFormData>({
    resolver: zodResolver(passwordSchema),
    defaultValues: {
      currentPassword: "",
      newPassword: "",
      confirmPassword: "",
    },
  });

  const handleSubmit = (data: PasswordFormData) => {
    onSubmit({
      currentPassword: data.currentPassword,
      newPassword: data.newPassword,
    });
  };

  return (
    <Card className="border-0 shadow-none bg-muted/30">
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-medium flex items-center gap-2">
          <Key className="h-4 w-4" />
          {t("settings:user.changePassword")}
        </CardTitle>
      </CardHeader>
      <CardContent className="pt-0">
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(handleSubmit)}
            className="space-y-4"
          >
            <PasswordField
              control={form.control}
              name="currentPassword"
              label={t("settings:user.currentPassword")}
              className="[&_input]:bg-background"
            />
            <PasswordField
              control={form.control}
              name="newPassword"
              label={t("settings:user.newPassword")}
              autoComplete="new-password"
              className="[&_input]:bg-background"
            />
            <PasswordField
              control={form.control}
              name="confirmPassword"
              label={t("settings:user.confirmPassword")}
              autoComplete="new-password"
              className="[&_input]:bg-background"
            />

            <div className="flex items-center gap-2 pt-3">
              <Button type="submit" size="sm" disabled={isPending}>
                {isPending ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : isSuccess ? (
                  <Check className="mr-2 h-4 w-4" />
                ) : null}
                {isSuccess
                  ? t("settings:user.passwordChanged")
                  : t("settings:user.save")}
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => {
                  form.reset();
                  onCancel();
                }}
              >
                <X className="mr-2 h-4 w-4" />
                {t("common:cancel")}
              </Button>
            </div>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
