import { useUpdateApiKeyMutation } from "@/api";
import { ApiKey, MutationError } from "@/types";
import { useTranslationNamespace } from "@/helpers";
import { useMemo } from "react";
import { z } from "zod";
import { Resolver, useForm } from "react-hook-form";
import { toast } from "sonner";
import { zodResolver } from "@hookform/resolvers/zod";
import { ApiKeyForm, ApiKeyFormData } from "./ApiKeyForm";

export interface EditApiKeyFormProps {
  currentOrgId: string;
  selectedApiKey: ApiKey;
  onApiKeyUpdated: (apiKey: ApiKey) => void;
  onClose: () => void;
}

export const EditApiKeyForm = ({
  currentOrgId,
  selectedApiKey,
  onApiKeyUpdated,
  onClose,
}: EditApiKeyFormProps) => {
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
            expiresAtDate: z.string().optional(),
          })
        : z.object({
            name: z.string().min(1, t("settings:org.validation.nameRequired")),
            permissions: z
              .array(z.string())
              .min(1, t("settings:org.validation.permissionsRequired")),
            expiresAtDate: z.string().optional(),
          }),
    [t, tLoading],
  );

  const form = useForm<ApiKeyFormData>({
    resolver: zodResolver(apiKeySchema) as Resolver<ApiKeyFormData>,
    defaultValues: {
      name: selectedApiKey.name,
      permissions: selectedApiKey.permissions,
    },
  });

  const updateMutation = useUpdateApiKeyMutation(
    currentOrgId,
    selectedApiKey.id,
    {
      onSuccess: (data) => {
        form.reset({
          name: data.name,
          permissions: data.permissions,
        });
        onApiKeyUpdated(data);
      },
      onError: (err: MutationError) => {
        toast.error(err.error?.message ?? t("common:defaultError"));
      },
    },
  );

  return (
    <ApiKeyForm
      mode="edit"
      form={form}
      onSubmit={(data) =>
        updateMutation.mutate({
          name: data.name,
          permissions: data.permissions,
        })
      }
      isPending={updateMutation.isPending}
      onCancel={onClose}
    />
  );
};
