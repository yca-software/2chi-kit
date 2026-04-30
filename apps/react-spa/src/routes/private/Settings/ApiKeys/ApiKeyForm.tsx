import { useTranslationNamespace } from "@/helpers";
import { Button, Form, SheetFooter } from "@yca-software/design-system";
import { ExpiresAtField, NameField, PermissionsField } from "@/components";
import { Loader2 } from "lucide-react";
import { ROLE_PERMISSION_GROUPS } from "../Roles/rolePermissionGroups";
import { ASSIGNABLE_API_KEY_PERMISSIONS } from "@/constants";
import { UseFormReturn } from "react-hook-form";

const assignableSet = new Set(ASSIGNABLE_API_KEY_PERMISSIONS);
/** Permission groups filtered to only assignable API key permissions (matches backend). */
function getApiKeyPermissionGroups() {
  return ROLE_PERMISSION_GROUPS.map((group) => ({
    ...group,
    permissions: group.permissions.filter((p) => assignableSet.has(p.key)),
  })).filter((g) => g.permissions.length > 0);
}

const API_KEY_PERMISSION_GROUPS = getApiKeyPermissionGroups();

export interface ApiKeyFormData {
  name: string;
  permissions: string[];
  expiresAtDate: string;
}

interface ApiKeyFormProps {
  mode: "create" | "edit";
  form: UseFormReturn<ApiKeyFormData>;
  onSubmit: (data: ApiKeyFormData) => void;
  onCancel?: () => void;
  isPending: boolean;
}

export function ApiKeyForm({
  mode,
  form,
  onSubmit,
  onCancel,
  isPending,
}: ApiKeyFormProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const selected = form.watch("permissions") ?? [];

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="flex flex-col gap-6"
      >
        <NameField
          control={form.control}
          name="name"
          label={t("settings:org.apiKeys.name")}
          placeholder={t("settings:org.apiKeys.namePlaceholder")}
        />

        {mode === "create" && (
          <ExpiresAtField
            control={form.control}
            name="expiresAtDate"
            label={t("settings:org.apiKeys.expiration")}
            placeholder={t("settings:org.apiKeys.expiration")}
            ariaLabel={t("settings:org.apiKeys.expiration")}
            options={[
              {
                value: "30d",
                label: t("settings:org.apiKeys.expiration30Days"),
              },
              {
                value: "60d",
                label: t("settings:org.apiKeys.expiration60Days"),
              },
              {
                value: "90d",
                label: t("settings:org.apiKeys.expiration90Days"),
              },
              {
                value: "180d",
                label: t("settings:org.apiKeys.expiration180Days"),
              },
              {
                value: "365d",
                label: t("settings:org.apiKeys.expiration1Year"),
              },
            ]}
          />
        )}

        <PermissionsField
          control={form.control}
          name="permissions"
          label={t("settings:org.apiKeys.permissions")}
          groups={API_KEY_PERMISSION_GROUPS}
          t={t}
        />

        <SheetFooter>
          {onCancel && (
            <Button type="button" variant="outline" onClick={onCancel}>
              {t("common:cancel")}
            </Button>
          )}
          <Button
            type="submit"
            disabled={isPending || (selected?.length ?? 0) === 0}
          >
            {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {mode === "create" ? t("common:create") : t("common:save")}
          </Button>
        </SheetFooter>
      </form>
    </Form>
  );
}
