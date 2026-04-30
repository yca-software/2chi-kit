import { useTranslationNamespace } from "@/helpers";
import { FormDrawer } from "@yca-software/design-system";
import { ApiKeyForm, type ApiKeyFormData } from "./ApiKeyForm";
import { useMemo } from "react";
import z from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useCreateApiKeyMutation } from "@/api";
import type { MutationError } from "@/types";
import { toast } from "sonner";

interface CreateApiKeyDrawerProps {
  open: boolean;
  onClose: () => void;
  currentOrgId: string;
  onKeyCreated: (keyValue: string) => void;
}

export function CreateApiKeyDrawer({
  open,
  onClose,
  currentOrgId,
  onKeyCreated,
}: CreateApiKeyDrawerProps) {
  const { t, isLoading: tLoading } = useTranslationNamespace([
    "settings",
    "common",
  ]);

  const apiKeySchema = useMemo(
    () =>
      tLoading
        ? z.object({
            name: z.string(),
            permissions: z.array(z.string()),
            expiresAtDate: z.string(),
          })
        : z.object({
            name: z.string().min(1, t("settings:org.validation.nameRequired")),
            permissions: z
              .array(z.string())
              .min(1, t("settings:org.validation.permissionsRequired")),
            expiresAtDate: z.string(),
          }),
    [t, tLoading],
  );

  const createForm = useForm<ApiKeyFormData>({
    resolver: zodResolver(apiKeySchema),
    defaultValues: {
      name: "",
      permissions: [],
      expiresAtDate: "30d",
    },
  });

  const createMutation = useCreateApiKeyMutation(currentOrgId, {
    onSuccess: (data) => {
      createForm.reset();
      onKeyCreated(data.secret);
    },
    onError: (err: MutationError) => {
      toast.error(err.error?.message ?? t("common:defaultError"));
    },
  });

  const onSubmit = (data: ApiKeyFormData) => {
    const daysMap: Record<string, number> = {
      "30d": 30,
      "60d": 60,
      "90d": 90,
      "180d": 180,
      "365d": 365,
    };
    const days = daysMap[data.expiresAtDate];
    if (!days) return;

    const expiresAt = new Date();
    expiresAt.setDate(expiresAt.getDate() + days);

    createMutation.mutate({
      name: data.name,
      permissions: data.permissions,
      expiresAt: expiresAt.toISOString(),
    });
  };

  return (
    <FormDrawer
      open={open}
      onOpenChange={onClose}
      title={t("settings:org.apiKeys.createApiKey")}
      description={t("settings:org.apiKeys.createDescription")}
    >
      <ApiKeyForm
        mode="create"
        form={createForm}
        onSubmit={onSubmit}
        isPending={createMutation.isPending}
      />
    </FormDrawer>
  );
}
