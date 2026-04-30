import { useEffect, useMemo } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Button,
  Form,
  Alert,
  AlertDescription,
} from "@yca-software/design-system";
import { Loader2 } from "lucide-react";
import { useTranslationNamespace, getConstraintNameFromError } from "@/helpers";
import {
  useCreateOrganizationMutation,
  CreateOrganizationResponse,
} from "@/api";
import type { OrganizationMemberWithOrganizationAndRole } from "@/types";
import { useUserState } from "@/states";
import { ORGANIZATION_PADDLE_CUSTOMER_CONSTRAINT } from "@/constants";
import { NameField, AddressField, EmailField } from "@/components/fields";

export function buildRoleFromCreateResponse(
  data: CreateOrganizationResponse,
): OrganizationMemberWithOrganizationAndRole | null {
  const { organization, roles, members } = data;
  if (!members || !organization) return null;

  const role = roles?.find((r) => r.id === members.roleId);
  return {
    ...members,
    organizationName: organization.name,
    roleName: role?.name ?? "",
    rolePermissions: role?.permissions ?? [],
  };
}

interface CreateOrgFormProps {
  /** When provided, called with created org data without navigation. */
  onCreated: (data: CreateOrganizationResponse) => void;
}

export const CreateOrgForm = ({ onCreated }: CreateOrgFormProps) => {
  const { t, isLoading } = useTranslationNamespace(["dashboard", "common"]);
  const userEmail = useUserState((state) => state.userData.user?.email ?? "");

  const onboardingSchema = useMemo(() => {
    if (isLoading) {
      return z.object({
        name: z.string(),
        placeId: z.string().optional(),
        billingEmail: z.string(),
      });
    }

    return z.object({
      name: z
        .string()
        .min(1, t("dashboard:onboardingForm.validation.nameRequired")),
      placeId: z
        .string()
        .min(1, t("dashboard:onboardingForm.validation.locationRequired"))
        .optional(),
      billingEmail: z
        .string()
        .min(1, t("dashboard:onboardingForm.validation.billingEmailRequired"))
        .email(t("dashboard:onboardingForm.validation.billingEmailInvalid")),
    });
  }, [t, isLoading]);

  type OnboardingFormData = z.infer<typeof onboardingSchema>;

  const form = useForm<OnboardingFormData>({
    // `@hookform/resolvers` sometimes type-checks against a different branded
    // Zod minor/instance than the one produced by `zod` in this app.
    // The runtime validation is still correct, so we relax the compile-time
    // constraint here to unblock form usage.
    resolver: zodResolver(onboardingSchema as any),
    defaultValues: {
      name: "",
      placeId: "",
      billingEmail: userEmail || "",
    },
  });

  // Keep billing email synced from the authenticated user.
  useEffect(() => {
    if (userEmail && !form.getValues("billingEmail")) {
      form.setValue("billingEmail", userEmail);
    }
  }, [userEmail, form]);

  const createMutation = useCreateOrganizationMutation({
    onSuccess: (data) => {
      onCreated(data);
    },
    onError: () => {
      // Error toast can be handled globally if desired.
    },
  });

  const constraintName = getConstraintNameFromError(createMutation.error);
  const billingEmailErrorMessage =
    constraintName === ORGANIZATION_PADDLE_CUSTOMER_CONSTRAINT
      ? t("dashboard:onboardingForm.validation.billingEmailAlreadyUsed")
      : "";

  useEffect(() => {
    if (billingEmailErrorMessage) {
      form.setError("billingEmail", {
        type: "server",
        message: billingEmailErrorMessage,
      });
    }
  }, [billingEmailErrorMessage, form]);

  const onSubmit = (data: OnboardingFormData) => {
    if (!data.placeId) {
      form.setError("placeId", {
        type: "manual",
        message: t("dashboard:onboardingForm.validation.locationRequired"),
      });
      return;
    }

    createMutation.mutate({
      name: data.name,
      placeId: data.placeId,
      billingEmail: data.billingEmail,
    });
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        {createMutation.error && (
          <Alert variant="destructive">
            <AlertDescription className="break-words whitespace-normal">
              {createMutation.error.error?.message ||
                t("defaultError", { ns: "common" })}
            </AlertDescription>
          </Alert>
        )}

        <NameField
          control={form.control}
          name="name"
          label={t("dashboard:onboardingForm.orgNameLabel")}
          placeholder={t("dashboard:onboardingForm.orgNamePlaceholder")}
        />

        <AddressField
          control={form.control}
          name="placeId"
          form={form}
          label={t("dashboard:onboardingForm.locationLabel")}
          placeholder={t("dashboard:onboardingForm.locationPlaceholder")}
        />

        <EmailField
          control={form.control}
          name="billingEmail"
          label={t("dashboard:onboardingForm.billingEmailLabel")}
          placeholder={t("dashboard:onboardingForm.billingEmailPlaceholder")}
          description={t("dashboard:onboardingForm.billingEmailDescription")}
        />

        <Button
          type="submit"
          className="w-full"
          disabled={createMutation.isPending}
        >
          {createMutation.isPending ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              {t("loading", { ns: "common" })}
            </>
          ) : (
            t("dashboard:onboardingForm.createButton")
          )}
        </Button>
      </form>
    </Form>
  );
};
