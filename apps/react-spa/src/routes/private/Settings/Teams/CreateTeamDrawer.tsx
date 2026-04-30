import { useMemo } from "react";
import { useTranslationNamespace } from "@/helpers";
import { useCreateTeamMutation } from "@/api";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import type { MutationError } from "@/types";
import { FormDrawer } from "@yca-software/design-system";
import { TeamForm, type TeamFormData } from "./TeamForm";

interface CreateTeamDrawerProps {
  open: boolean;
  onClose: () => void;
  currentOrgId: string;
}

export function CreateTeamDrawer({
  open,
  onClose,
  currentOrgId,
}: CreateTeamDrawerProps) {
  const { t, isLoading: tLoading } = useTranslationNamespace([
    "settings",
    "common",
  ]);

  const teamSchema = useMemo(() => {
    if (tLoading) {
      return z.object({ name: z.string(), description: z.string() });
    }

    return z.object({
      name: z.string().min(1, t("settings:org.validation.nameRequired")),
      description: z.string(),
    });
  }, [t, tLoading]);

  const form = useForm<TeamFormData>({
    resolver: zodResolver(teamSchema),
    defaultValues: { name: "", description: "" },
  });

  const createMutation = useCreateTeamMutation(currentOrgId, {
    onSuccess: () => {
      onClose();
      form.reset({ name: "", description: "" });
    },
    onError: (err: MutationError) => {
      toast.error(err.error?.message ?? t("common:defaultError"));
    },
  });

  return (
    <FormDrawer
      open={open}
      onOpenChange={onClose}
      title={t("settings:org.teams.createTeam")}
      description={t("settings:org.teams.createTeamDescription")}
    >
      <TeamForm
        mode="create"
        form={form}
        onSubmit={createMutation.mutate}
        isPending={createMutation.isPending}
      />
    </FormDrawer>
  );
}
