import { useTranslationNamespace } from "@/helpers";
import { ConfirmDialog } from "@yca-software/design-system";

interface ArchiveOrgDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  isPending: boolean;
}

export function ArchiveOrgDialog({
  open,
  onOpenChange,
  onConfirm,
  isPending,
}: ArchiveOrgDialogProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("common:confirm")}
      description={t("settings:org.archiveConfirm")}
      cancelLabel={t("common:cancel")}
      confirmLabel={isPending ? t("common:archiving") : t("common:archive")}
      variant="destructive"
      onConfirm={onConfirm}
      isPending={isPending}
      closeOnOutsideClick
    />
  );
}
