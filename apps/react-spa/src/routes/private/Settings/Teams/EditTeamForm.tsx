import { TeamForm, TeamFormData } from "./TeamForm";
import { useUpdateTeamMutation } from "@/api";
import { MutationError, Team } from "@/types";
import { useTranslationNamespace } from "@/helpers";
import { useMemo } from "react";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { zodResolver } from "@hookform/resolvers/zod";

export interface EditTeamFormProps {
  currentOrgId: string;
  selectedTeam: Team;
  onTeamUpdated: (team: Team) => void;
  onClose: () => void;
}

export const EditTeamForm = ({
  currentOrgId,
  selectedTeam,
  onTeamUpdated,
  onClose,
}: EditTeamFormProps) => {
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
    defaultValues: {
      name: selectedTeam.name,
      description: selectedTeam.description ?? "",
    },
  });

  const updateMutation = useUpdateTeamMutation(currentOrgId, selectedTeam.id, {
    onSuccess: (updatedTeam: Team) => {
      form.reset({
        name: updatedTeam.name,
        description: updatedTeam.description ?? "",
      });
      onTeamUpdated(updatedTeam);
    },
    onError: (err: MutationError) => {
      toast.error(err.error?.message ?? t("common:defaultError"));
    },
  });

  return (
    <TeamForm
      mode="edit"
      form={form}
      onSubmit={(data) =>
        updateMutation.mutate({
          name: data.name,
          description: data.description,
        })
      }
      isPending={updateMutation.isPending}
      onCancel={onClose}
    />
  );
};
