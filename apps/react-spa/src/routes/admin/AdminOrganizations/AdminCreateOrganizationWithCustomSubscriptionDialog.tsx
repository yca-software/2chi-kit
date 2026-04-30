import * as React from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  Button,
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
  Alert,
  AlertDescription,
  Input,
  Label,
  Select,
  DatePicker,
} from "@yca-software/design-system";
import { Loader2 } from "lucide-react";
import { useTranslationNamespace } from "@/helpers";
import { NameField, AddressField, EmailField } from "@/components";
import {
  useAdminCreateOrganizationWithCustomSubscriptionMutation,
  ADMIN_SUBSCRIPTION_TYPE_BASIC,
  ADMIN_SUBSCRIPTION_TYPE_PRO,
  ADMIN_SUBSCRIPTION_TYPE_ENTERPRISE,
  AdminCreateOrganizationWithCustomSubscriptionRequest,
} from "@/api";
import { LANGUAGES, DEFAULT_LANGUAGE } from "@/constants";
import { format } from "date-fns";

function getTomorrow(): Date {
  const d = new Date();
  d.setDate(d.getDate() + 1);
  d.setHours(0, 0, 0, 0);
  return d;
}

const baseSchema = z.object({
  name: z.string().min(1, "Name is required"),
  placeId: z.string().min(1, "Location is required"),
  billingEmail: z
    .string()
    .min(1, "Billing email is required")
    .email("Invalid email"),
  ownerEmail: z
    .string()
    .min(1, "Owner email is required")
    .email("Invalid email"),
  subscriptionType: z.number().min(1).max(3),
  subscriptionSeats: z.number().int().min(1, "At least 1 seat"),
  subscriptionExpiresAt: z.string().optional(),
  language: z.string().min(1, "Language is required"),
});

type FormData = z.infer<typeof baseSchema>;

export interface AdminCreateOrganizationWithCustomSubscriptionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: (orgId: string) => void;
}

export function AdminCreateOrganizationWithCustomSubscriptionDialog({
  open,
  onOpenChange,
  onSuccess,
}: AdminCreateOrganizationWithCustomSubscriptionDialogProps) {
  const { t } = useTranslationNamespace(["admin"]);
  const schema = React.useMemo(
    () =>
      baseSchema.refine(
        (data) => {
          if (!data.subscriptionExpiresAt?.trim()) return true;
          const chosen = new Date(data.subscriptionExpiresAt + "T00:00:00");
          const tomorrow = getTomorrow();
          return chosen.getTime() >= tomorrow.getTime();
        },
        {
          message: t(
            "admin:organizations.createCustom.subscriptionExpiresAtMinDateError",
          ),
          path: ["subscriptionExpiresAt"],
        },
      ),
    [t],
  );
  const createMutation =
    useAdminCreateOrganizationWithCustomSubscriptionMutation({
      onSuccess: (data) => {
        onOpenChange(false);
        onSuccess?.(data.id);
      },
    });

  const form = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      placeId: "",
      billingEmail: "",
      ownerEmail: "",
      subscriptionType: ADMIN_SUBSCRIPTION_TYPE_BASIC,
      subscriptionSeats: 5,
      subscriptionExpiresAt: "",
      language: DEFAULT_LANGUAGE,
    },
  });

  const onSubmit = (data: FormData) => {
    let subscriptionExpiresAt: string | undefined;
    if (data.subscriptionExpiresAt?.trim()) {
      const d = new Date(data.subscriptionExpiresAt + "T00:00:00");
      subscriptionExpiresAt = Number.isNaN(d.getTime())
        ? undefined
        : d.toISOString();
    }
    const body: AdminCreateOrganizationWithCustomSubscriptionRequest = {
      name: data.name,
      placeId: data.placeId,
      billingEmail: data.billingEmail,
      ownerEmail: data.ownerEmail,
      subscriptionType: data.subscriptionType,
      subscriptionSeats: data.subscriptionSeats,
      subscriptionExpiresAt,
      language: data.language,
    };
    createMutation.mutate(body);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {t("admin:organizations.createCustom.title")}
          </DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            {createMutation.error && (
              <Alert variant="destructive">
                <AlertDescription>
                  {(createMutation.error as { error?: { message?: string } })
                    ?.error?.message ?? "Something went wrong."}
                </AlertDescription>
              </Alert>
            )}

            <NameField
              control={form.control}
              name="name"
              label={t("admin:organizations.createCustom.nameLabel")}
              placeholder={t(
                "admin:organizations.createCustom.namePlaceholder",
              )}
            />

            <AddressField
              control={form.control}
              name="placeId"
              form={form}
              label={t("admin:organizations.createCustom.locationLabel")}
              placeholder={t(
                "admin:organizations.createCustom.locationPlaceholder",
              )}
            />

            <EmailField
              control={form.control}
              name="billingEmail"
              label={t("admin:organizations.createCustom.billingEmailLabel")}
              placeholder={t(
                "admin:organizations.createCustom.billingEmailPlaceholder",
              )}
            />

            <EmailField
              control={form.control}
              name="ownerEmail"
              label={t("admin:organizations.createCustom.ownerEmailLabel")}
              placeholder={t(
                "admin:organizations.createCustom.ownerEmailPlaceholder",
              )}
              description={t(
                "admin:organizations.createCustom.ownerEmailDescription",
              )}
            />

            <FormField
              control={form.control}
              name="language"
              render={({ field }) => (
                <FormItem>
                  <Label>
                    {t("admin:organizations.createCustom.inviteLanguageLabel")}
                  </Label>
                  <FormControl>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      options={Object.entries(LANGUAGES).map(
                        ([code, label]) => ({
                          value: code,
                          label,
                        }),
                      )}
                      aria-label={t(
                        "admin:organizations.createCustom.inviteLanguageLabel",
                      )}
                    />
                  </FormControl>
                  <p className="text-xs text-muted-foreground">
                    {t(
                      "admin:organizations.createCustom.inviteLanguageDescription",
                    )}
                  </p>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="subscriptionType"
              render={({ field }) => (
                <FormItem>
                  <Label>
                    {t(
                      "admin:organizations.createCustom.subscriptionTypeLabel",
                    )}
                  </Label>
                  <FormControl>
                    <Select
                      value={String(field.value)}
                      onValueChange={(v) => field.onChange(Number(v))}
                      options={[
                        {
                          value: String(ADMIN_SUBSCRIPTION_TYPE_BASIC),
                          label: t(
                            "admin:organizations.createCustom.subscriptionTypeBasic",
                          ),
                        },
                        {
                          value: String(ADMIN_SUBSCRIPTION_TYPE_PRO),
                          label: t(
                            "admin:organizations.createCustom.subscriptionTypePro",
                          ),
                        },
                        {
                          value: String(ADMIN_SUBSCRIPTION_TYPE_ENTERPRISE),
                          label: t(
                            "admin:organizations.createCustom.subscriptionTypeEnterprise",
                          ),
                        },
                      ]}
                      aria-label={t(
                        "admin:organizations.createCustom.subscriptionTypeLabel",
                      )}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="space-y-2">
              <Label>
                {t("admin:organizations.createCustom.subscriptionSeatsLabel")}
              </Label>
              <Input
                type="number"
                min={1}
                {...form.register("subscriptionSeats", { valueAsNumber: true })}
              />
              {form.formState.errors.subscriptionSeats && (
                <p className="text-sm text-destructive">
                  {form.formState.errors.subscriptionSeats.message}
                </p>
              )}
            </div>

            <FormField
              control={form.control}
              name="subscriptionExpiresAt"
              render={({ field }) => (
                <FormItem>
                  <Label className="text-muted-foreground">
                    {t(
                      "admin:organizations.createCustom.subscriptionExpiresAtLabel",
                    )}
                  </Label>
                  <FormControl>
                    <DatePicker
                      value={
                        field.value
                          ? new Date(field.value + "T00:00:00")
                          : undefined
                      }
                      onChange={(date) =>
                        field.onChange(date ? format(date, "yyyy-MM-dd") : "")
                      }
                      minDate={getTomorrow()}
                      placeholder={t(
                        "admin:organizations.createCustom.subscriptionExpiresAtPlaceholder",
                      )}
                      aria-label={t(
                        "admin:organizations.createCustom.subscriptionExpiresAtLabel",
                      )}
                    />
                  </FormControl>
                  <p className="text-xs text-muted-foreground">
                    {t(
                      "admin:organizations.createCustom.subscriptionExpiresAtDescription",
                    )}
                  </p>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter className="gap-2 pt-4">
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={createMutation.isPending}
              >
                {t("admin:organizations.createCustom.cancel")}
              </Button>
              <Button type="submit" disabled={createMutation.isPending}>
                {createMutation.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    {t("admin:organizations.createCustom.creating")}
                  </>
                ) : (
                  t("admin:organizations.createCustom.submitButton")
                )}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
