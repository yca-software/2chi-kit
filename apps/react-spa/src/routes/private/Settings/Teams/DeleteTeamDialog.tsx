import { useDeleteTeamMutation } from "@/api";
import { useTranslationNamespace } from "@/helpers";
import type { MutationError, Team } from "@/types";
import { ConfirmDialog } from "@yca-software/design-system";
import { toast } from "sonner";

interface DeleteTeamDialogProps {
  currentOrgId: string;
  open: boolean;
  onClose: () => void;
  selectedTeam: Team | null;
  onSuccess: () => void;
}

export function DeleteTeamDialog({
  currentOrgId,
  open,
  onClose,
  selectedTeam,
  onSuccess,
}: DeleteTeamDialogProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  const deleteMutation = useDeleteTeamMutation(
    currentOrgId,
    selectedTeam?.id ?? "",
    {
      onSuccess,
      onError: (err: MutationError) => {
        toast.error(err.error?.message ?? t("common:defaultError"));
        onClose();
      },
    },
  );

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onClose}
      title={t("common:confirm")}
      description={t("settings:org.teams.confirmDelete")}
      cancelLabel={t("common:cancel")}
      confirmLabel={
        deleteMutation.isPending ? t("common:deleting") : t("common:delete")
      }
      variant="destructive"
      onConfirm={() => {
        if (selectedTeam) {
          deleteMutation.mutate();
        }
      }}
      isPending={deleteMutation.isPending}
      closeOnOutsideClick
    />
  );
}
