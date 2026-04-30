import { useTranslationNamespace } from "@/helpers";
import { Button } from "@yca-software/design-system";
import { Archive } from "lucide-react";
import { ArchiveOrgDialog } from "./ArchiveOrgDialog";

interface SettingsHeaderProps {
  organizationName?: string;
  canArchiveOrg: boolean;
  archiveDialogOpen: boolean;
  setArchiveDialogOpen: (open: boolean) => void;
  onArchiveConfirm: () => void;
  isArchivePending: boolean;
}

export function SettingsHeader({
  organizationName,
  canArchiveOrg,
  archiveDialogOpen,
  setArchiveDialogOpen,
  onArchiveConfirm,
  isArchivePending,
}: SettingsHeaderProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  return (
    <div className="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div className="min-w-0">
        <h1 className="truncate text-xl font-bold min-[480px]:text-2xl">
          {organizationName}
        </h1>
        <p className="text-muted-foreground text-sm min-[480px]:text-base">
          {t("settings:org.description")}
        </p>
      </div>
      <div className="flex shrink-0 items-center gap-2">
        {canArchiveOrg && (
          <>
            <Button
              variant="outline"
              size="sm"
              className="text-destructive hover:text-destructive"
              onClick={() => setArchiveDialogOpen(true)}
            >
              <Archive className="mr-2 h-4 w-4" />
              {t("common:archive")}
            </Button>
            <ArchiveOrgDialog
              open={archiveDialogOpen}
              onOpenChange={setArchiveDialogOpen}
              onConfirm={onArchiveConfirm}
              isPending={isArchivePending}
            />
          </>
        )}
      </div>
    </div>
  );
}
