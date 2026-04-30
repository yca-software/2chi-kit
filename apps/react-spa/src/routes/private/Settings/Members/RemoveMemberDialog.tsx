import { useRemoveMemberMutation } from "@/api";
import { useTranslationNamespace } from "@/helpers/hooks/useTranslationNamespace";
import { MutationError, OrganizationMemberWithUser } from "@/types";
import { ConfirmDialog } from "@yca-software/design-system";
import { toast } from "sonner";

interface RemoveMemberDialogProps {
  currentOrgId: string;
  open: boolean;
  onClose: () => void;
  selectedMember: OrganizationMemberWithUser | null;
}

export function RemoveMemberDialog({
  open,
  onClose,
  selectedMember,
  currentOrgId,
}: RemoveMemberDialogProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  const removeMutation = useRemoveMemberMutation(
    currentOrgId,
    selectedMember?.id ?? "",
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
      description={t("settings:org.members.confirmRemove")}
      cancelLabel={t("common:cancel")}
      confirmLabel={t("common:remove")}
      variant="destructive"
      closeOnOutsideClick
      onConfirm={() => {
        if (selectedMember) {
          removeMutation.mutate();
        }
      }}
      isPending={removeMutation.isPending}
    />
  );
}
