import { useDeleteRoleMutation } from "@/api";
import { useTranslationNamespace } from "@/helpers";
import type { MutationError, Role } from "@/types";
import { ConfirmDialog } from "@yca-software/design-system";
import { toast } from "sonner";

interface DeleteRoleDialogProps {
  currentOrgId: string;
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
  selectedRole: Role | null;
}

export function DeleteRoleDialog({
  open,
  onClose,
  selectedRole,
  onSuccess,
  currentOrgId,
}: DeleteRoleDialogProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  const deleteMutation = useDeleteRoleMutation(
    currentOrgId,
    selectedRole?.id ?? "",
    {
      onSuccess,
      onError: (err: MutationError) => {
        const body = err.error as
          | {
              message?: string;
              extra?: { memberEmails?: string[] };
            }
          | undefined;
        const apiMessage = body?.message ?? t("common:defaultError");
        const emails = body?.extra?.memberEmails;
        const blockedByAssignedMembers =
          err.status === 409 && Array.isArray(emails) && emails.length > 0;

        if (blockedByAssignedMembers) {
          let detail = apiMessage;
          if (!emails.every((e) => detail.includes(e))) {
            detail = `${detail}\n\n${emails.join(", ")}`;
          }
          toast.error(t("settings:org.roles.deleteBlockedTitle"), {
            description: detail,
            duration: 16_000,
          });
        } else {
          toast.error(apiMessage);
        }
        onClose();
      },
    },
  );

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onClose}
      title={t("common:confirm")}
      description={t("settings:org.roles.confirmDelete")}
      cancelLabel={t("common:cancel")}
      confirmLabel={
        deleteMutation.isPending ? t("common:deleting") : t("common:delete")
      }
      variant="destructive"
      onConfirm={() => {
        if (selectedRole) {
          deleteMutation.mutate();
        }
      }}
      isPending={deleteMutation.isPending}
      closeOnOutsideClick
    />
  );
}
