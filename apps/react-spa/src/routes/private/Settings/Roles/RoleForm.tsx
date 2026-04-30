import { UseFormReturn } from "react-hook-form";
import { useTranslationNamespace } from "@/helpers";
import { DescriptionField, NameField, PermissionsField } from "@/components";
import { Button, Form, SheetFooter } from "@yca-software/design-system";
import { ROLE_PERMISSION_GROUPS } from "./rolePermissionGroups";
import { Loader2 } from "lucide-react";

export type RoleFormData = {
  name: string;
  description: string;
  permissions: string[];
};

interface RoleFormProps {
  mode: "create" | "edit";
  isSystemRole?: boolean;
  form: UseFormReturn<RoleFormData>;
  onSubmit: (data: RoleFormData) => void;
  onCancel?: () => void;
  isPending: boolean;
}

export function RoleForm({
  mode,
  isSystemRole = false,
  form,
  onSubmit,
  onCancel,
  isPending,
}: RoleFormProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const selected = form.watch("permissions") ?? [];

  const submitLabel = mode === "create" ? t("common:create") : t("common:save");

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="flex flex-col gap-6"
      >
        {isSystemRole && (
          <div className="rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">
            {t("settings:org.roles.systemRoleReadonlyNotice")}
          </div>
        )}

        <NameField
          control={form.control}
          name="name"
          label={t("settings:org.roles.name")}
          placeholder={t("settings:org.roles.namePlaceholder")}
          disabled={isSystemRole}
        />

        <DescriptionField
          control={form.control}
          name="description"
          label={t("settings:org.roles.descriptionLabel")}
          placeholder={t("settings:org.roles.descriptionPlaceholder")}
          disabled={isSystemRole}
        />

        <PermissionsField
          control={form.control}
          name="permissions"
          label={t("settings:org.roles.permissions")}
          groups={ROLE_PERMISSION_GROUPS}
          t={t}
          disabled={isSystemRole}
        />

        <SheetFooter>
          {!isSystemRole && (
            <>
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
                {submitLabel}
              </Button>
            </>
          )}
        </SheetFooter>
      </form>
    </Form>
  );
}
