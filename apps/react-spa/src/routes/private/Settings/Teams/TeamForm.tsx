import { useTranslationNamespace } from "@/helpers";
import { Button, Form, SheetFooter } from "@yca-software/design-system";
import { NameField, DescriptionField } from "@/components";
import { Loader2 } from "lucide-react";
import { UseFormReturn } from "react-hook-form";

export type TeamFormData = {
  name: string;
  description: string;
};

interface TeamFormProps {
  mode: "create" | "edit";
  form: UseFormReturn<TeamFormData>;
  onSubmit: (data: TeamFormData) => void;
  onCancel?: () => void;
  isPending: boolean;
}

export function TeamForm({
  mode,
  form,
  onSubmit,
  onCancel,
  isPending,
}: TeamFormProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  const submitLabel = mode === "create" ? t("common:create") : t("common:save");

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="flex flex-col gap-6"
      >
        <NameField
          control={form.control}
          name="name"
          label={t("settings:org.teams.name")}
          placeholder={t("settings:org.teams.namePlaceholder")}
        />
        <DescriptionField
          control={form.control}
          name="description"
          label={t("settings:org.teams.descriptionLabel")}
          placeholder={t("settings:org.teams.descriptionPlaceholder")}
        />
        <SheetFooter>
          {onCancel && (
            <Button type="button" variant="outline" onClick={onCancel}>
              {t("common:cancel")}
            </Button>
          )}
          <Button type="submit" disabled={isPending}>
            {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {submitLabel}
          </Button>
        </SheetFooter>
      </form>
    </Form>
  );
}
