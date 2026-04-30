import { useEffect, useMemo } from "react";
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
  Input,
  Label,
} from "@yca-software/design-system";
import { FirstNameField, LastNameField } from "@/components";
import { Loader2, Save, Check, X } from "lucide-react";
import type { User } from "@/types";
import { useTranslationNamespace } from "@/helpers";

type ProfileFormData = {
  firstName: string;
  lastName: string;
};

interface ProfileFormProps {
  user: User | null;
  onSubmit: (data: ProfileFormData) => void;
  onCancel: () => void;
  isPending: boolean;
  isSuccess: boolean;
}

export function ProfileForm({
  user,
  onSubmit,
  onCancel,
  isPending,
  isSuccess,
}: ProfileFormProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const profileSchema = useMemo(
    () =>
      z.object({
        firstName: z
          .string()
          .min(1, t("settings:user.validation.firstNameRequired")),
        lastName: z
          .string()
          .min(1, t("settings:user.validation.lastNameRequired")),
      }),
    [t],
  );

  const form = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      firstName: user?.firstName || "",
      lastName: user?.lastName || "",
    },
  });

  // Reset form when user data changes
  useEffect(() => {
    if (user) {
      form.reset({
        firstName: user.firstName || "",
        lastName: user.lastName || "",
      });
    }
  }, [user, form]);

  return (
    <Card className="border-0 shadow-none bg-muted/30">
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-medium">
          {t("settings:profile.editProfile")}
        </CardTitle>
      </CardHeader>
      <CardContent className="pt-0">
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FirstNameField
              control={form.control}
              name="firstName"
              className="[&_input]:bg-background"
            />
            <LastNameField
              control={form.control}
              name="lastName"
              className="[&_input]:bg-background"
            />

            <div className="space-y-2 pt-1">
              <div className="space-y-2">
                <Label className="text-muted-foreground">
                  {t("settings:user.email")}
                </Label>
                <Input value={user?.email ?? ""} disabled readOnly />
                <p className="text-xs text-muted-foreground">
                  {t("settings:profile.emailDescription")}
                </p>
              </div>
            </div>

            <div className="flex items-center gap-2 pt-3">
              <Button type="submit" size="sm" disabled={isPending}>
                {isPending ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : isSuccess ? (
                  <Check className="mr-2 h-4 w-4" />
                ) : (
                  <Save className="mr-2 h-4 w-4" />
                )}
                {isSuccess ? t("settings:user.saved") : t("settings:user.save")}
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={onCancel}
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
