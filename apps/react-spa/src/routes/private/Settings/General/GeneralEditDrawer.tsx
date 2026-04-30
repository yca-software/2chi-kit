import { useTranslationNamespace } from "@/helpers/hooks/useTranslationNamespace";
import { useEffect, useMemo } from "react";
import {
  Button,
  Form,
  Alert,
  AlertDescription,
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetFooter,
} from "@yca-software/design-system";
import { Loader2 } from "lucide-react";
import { NameField, AddressField } from "@/components/fields";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import type { Organization } from "@/types/organization";
import { useParams } from "react-router";
import { useUpdateOrganizationMutation } from "@/api";

interface GeneralEditDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  organization: Organization | null | undefined;
}

export function GeneralEditDrawer({
  open,
  onOpenChange,
  organization,
}: GeneralEditDrawerProps) {
  const { t, isLoading: tLoading } = useTranslationNamespace(["settings", "common"]);
  const { orgId } = useParams<{ orgId: string }>();

  const generalSchema = useMemo(
    () =>
      z.object({
        name: z
          .string()
          .min(
            1,
            tLoading ? "Required" : t("settings:org.validation.nameRequired"),
          ),
        placeId: z
          .string()
          .min(
            1,
            tLoading
              ? "Required"
              : t("settings:org.validation.locationRequired"),
          ),
      }),
    [t, tLoading],
  );

  type GeneralFormData = z.infer<typeof generalSchema>;

  const form = useForm<GeneralFormData>({
    resolver: zodResolver(generalSchema),
    defaultValues: {
      name: "",
      placeId: "",
    },
  });

  useEffect(() => {
    if (organization && open) {
      form.reset({
        name: organization.name,
        placeId: organization.placeId,
      });
    }
  }, [organization, open, form]);

  const updateMutation = useUpdateOrganizationMutation(orgId || "", {
    onSuccess: () => onOpenChange(false),
  });

  const handleSubmit = (data: GeneralFormData) => {
    if (!organization || !orgId) return;
    if (!data.placeId) {
      form.setError("placeId", {
        type: "manual",
        message: t("settings:org.validation.locationRequired"),
      });
      return;
    }
    updateMutation.mutate({
      name: data.name,
      placeId: data.placeId,
    });
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="sm:max-w-lg">
        <SheetHeader>
          <SheetTitle>{t("settings:org.organization.editTitle")}</SheetTitle>
          <SheetDescription>
            {t("settings:org.organization.editDescription")}
          </SheetDescription>
        </SheetHeader>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(handleSubmit)}
            className="flex flex-col gap-4"
          >
            {updateMutation.error && (
              <Alert variant="destructive">
                <AlertDescription>
                  {updateMutation.error.error?.message || t("common:defaultError")}
                </AlertDescription>
              </Alert>
            )}

            <NameField
              control={form.control}
              name="name"
              label={t("settings:org.general.name")}
            />

            <AddressField
              control={form.control}
              name="placeId"
              form={form}
              label={t("settings:org.general.address")}
              displayValue={organization?.address}
            />

            <SheetFooter>
              <Button type="submit" disabled={updateMutation.isPending}>
                {updateMutation.isPending && (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                )}
                {t("common:save")}
              </Button>
            </SheetFooter>
          </form>
        </Form>
      </SheetContent>
    </Sheet>
  );
}
