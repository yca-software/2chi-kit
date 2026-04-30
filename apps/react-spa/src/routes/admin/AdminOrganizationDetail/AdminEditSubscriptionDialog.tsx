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
  Checkbox,
  Select,
  DatePicker,
} from "@yca-software/design-system";
import { Loader2 } from "lucide-react";
import { useTranslationNamespace } from "@/helpers";
import {
  useAdminUpdateOrganizationSubscriptionMutation,
  ADMIN_SUBSCRIPTION_TYPE_BASIC,
  ADMIN_SUBSCRIPTION_TYPE_PRO,
  ADMIN_SUBSCRIPTION_TYPE_ENTERPRISE,
} from "@/api";
import type { Organization } from "@/types";
import { format } from "date-fns";

const schema = z.object({
  customSubscription: z.boolean(),
  subscriptionType: z.number().min(0).max(3),
  subscriptionSeats: z.number().int().min(1, "At least 1 seat"),
  subscriptionExpiresAt: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

export interface AdminEditSubscriptionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  organization: Organization;
  onSuccess?: () => void;
}

export function AdminEditSubscriptionDialog({
  open,
  onOpenChange,
  organization,
  onSuccess,
}: AdminEditSubscriptionDialogProps) {
  const { t } = useTranslationNamespace(["admin"]);
  const updateMutation = useAdminUpdateOrganizationSubscriptionMutation(
    organization.id,
    {
      onSuccess: () => {
        onOpenChange(false);
        onSuccess?.();
      },
    },
  );

  const form = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      customSubscription: organization.customSubscription,
      subscriptionType: organization.subscriptionType,
      subscriptionSeats: organization.subscriptionSeats,
      subscriptionExpiresAt: organization.subscriptionExpiresAt
        ? organization.subscriptionExpiresAt.slice(0, 10)
        : "",
    },
  });

  const onSubmit = (data: FormData) => {
    const body = {
      customSubscription: data.customSubscription,
      subscriptionType: data.subscriptionType,
      subscriptionSeats: data.subscriptionSeats,
      subscriptionExpiresAt: data.subscriptionExpiresAt?.trim()
        ? new Date(data.subscriptionExpiresAt + "T00:00:00").toISOString()
        : "",
    };
    updateMutation.mutate(body);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {t("admin:organizations.editSubscription.title")}
          </DialogTitle>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            {updateMutation.error && (
              <Alert variant="destructive">
                <AlertDescription>
                  {(updateMutation.error as { error?: { message?: string } })
                    ?.error?.message ?? "Something went wrong."}
                </AlertDescription>
              </Alert>
            )}

            <FormField
              control={form.control}
              name="customSubscription"
              render={({ field }) => (
                <FormItem className="flex flex-row items-center gap-2 space-y-0">
                  <FormControl>
                    <Checkbox
                      checked={field.value}
                      onCheckedChange={(checked) =>
                        field.onChange(checked === true)
                      }
                      aria-label={t(
                        "admin:organizations.editSubscription.customSubscriptionLabel",
                      )}
                    />
                  </FormControl>
                  <Label className="cursor-pointer font-normal">
                    {t(
                      "admin:organizations.editSubscription.customSubscriptionLabel",
                    )}
                  </Label>
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
                      "admin:organizations.editSubscription.subscriptionTypeLabel",
                    )}
                  </Label>
                  <FormControl>
                    <Select
                      value={String(field.value)}
                      onValueChange={(v) => field.onChange(Number(v))}
                      options={[
                        {
                          value: "0",
                          label: t(
                            "admin:organizations.editSubscription.subscriptionTypeFree",
                          ),
                        },
                        {
                          value: String(ADMIN_SUBSCRIPTION_TYPE_BASIC),
                          label: t(
                            "admin:organizations.editSubscription.subscriptionTypeBasic",
                          ),
                        },
                        {
                          value: String(ADMIN_SUBSCRIPTION_TYPE_PRO),
                          label: t(
                            "admin:organizations.editSubscription.subscriptionTypePro",
                          ),
                        },
                        {
                          value: String(ADMIN_SUBSCRIPTION_TYPE_ENTERPRISE),
                          label: t(
                            "admin:organizations.editSubscription.subscriptionTypeEnterprise",
                          ),
                        },
                      ]}
                      aria-label={t(
                        "admin:organizations.editSubscription.subscriptionTypeLabel",
                      )}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="space-y-2">
              <Label>
                {t(
                  "admin:organizations.editSubscription.subscriptionSeatsLabel",
                )}
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
                      "admin:organizations.editSubscription.subscriptionExpiresAtLabel",
                    )}
                  </Label>
                  <FormControl>
                    <DatePicker
                      value={
                        field.value
                          ? new Date(field.value.slice(0, 10) + "T00:00:00")
                          : undefined
                      }
                      onChange={(date) =>
                        field.onChange(date ? format(date, "yyyy-MM-dd") : "")
                      }
                      placeholder={t(
                        "admin:organizations.createCustom.subscriptionExpiresAtPlaceholder",
                      )}
                      aria-label={t(
                        "admin:organizations.editSubscription.subscriptionExpiresAtLabel",
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
                disabled={updateMutation.isPending}
              >
                {t("admin:organizations.createCustom.cancel")}
              </Button>
              <Button type="submit" disabled={updateMutation.isPending}>
                {updateMutation.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    {t("admin:organizations.editSubscription.saving")}
                  </>
                ) : (
                  t("admin:organizations.editSubscription.save")
                )}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
