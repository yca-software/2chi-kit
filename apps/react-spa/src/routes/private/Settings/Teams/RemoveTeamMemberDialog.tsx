import { useRemoveTeamMemberMutation } from "@/api";
import { useTranslationNamespace } from "@/helpers/hooks/useTranslationNamespace";
import { MutationError, Team, TeamMemberWithUser } from "@/types";
import { ConfirmDialog } from "@yca-software/design-system";
import { toast } from "sonner";

interface RemoveTeamMemberDialogProps {
  currentOrgId: string;
  open: boolean;
  onClose: () => void;
  selectedTeamMember: TeamMemberWithUser | null;
  selectedTeam: Team | null;
}

export function RemoveTeamMemberDialog({
  currentOrgId,
  open,
  onClose,
  selectedTeamMember,
  selectedTeam,
}: RemoveTeamMemberDialogProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const memberName = selectedTeamMember
    ? `${selectedTeamMember.userFirstName} ${selectedTeamMember.userLastName}`.trim()
    : "";
  const memberEmail = selectedTeamMember?.userEmail ?? "";

  const removeTeamMemberMutation = useRemoveTeamMemberMutation(
    currentOrgId,
    selectedTeam?.id || "",
    {
      onSuccess: () => onClose(),
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
      description={t("settings:org.teams.confirmRemoveMember", {
        memberName,
        memberEmail,
      })}
      cancelLabel={t("common:cancel")}
      confirmLabel={t("common:remove")}
      variant="destructive"
      onConfirm={() => {
        if (selectedTeamMember && selectedTeam) {
          removeTeamMemberMutation.mutate(selectedTeamMember.id);
        }
      }}
      isPending={removeTeamMemberMutation.isPending}
    />
  );
}
