import { useRevokeApiKeyMutation } from "@/api";
import { useTranslationNamespace } from "@/helpers/hooks/useTranslationNamespace";
import { ApiKey, MutationError } from "@/types";
import { ConfirmDialog } from "@yca-software/design-system";
import { toast } from "sonner";

interface RevokeApiKeyDialogProps {
  currentOrgId: string;
  open: boolean;
  onClose: () => void;
  selectedApiKey: ApiKey | null;
  onSuccess: () => void;
}

export function RevokeApiKeyDialog({
  currentOrgId,
  open,
  onClose,
  selectedApiKey,
  onSuccess,
}: RevokeApiKeyDialogProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  const revokeMutation = useRevokeApiKeyMutation(
    currentOrgId,
    selectedApiKey?.id ?? "",
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
      title={t("settings:org.apiKeys.confirmDeleteTitle")}
      description={t("settings:org.apiKeys.confirmDelete")}
      cancelLabel={t("common:cancel")}
      confirmLabel={
        revokeMutation.isPending ? t("common:deleting") : t("common:delete")
      }
      variant="destructive"
      onConfirm={() => {
        if (selectedApiKey) {
          revokeMutation.mutate();
        }
      }}
      isPending={revokeMutation.isPending}
      closeOnOutsideClick
    />
  );
}
