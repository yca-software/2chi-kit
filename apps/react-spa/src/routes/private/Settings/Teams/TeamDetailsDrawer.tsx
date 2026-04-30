import { useTranslationNamespace } from "@/helpers/hooks/useTranslationNamespace";
import { Team } from "@/types";
import {
  Button,
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@yca-software/design-system";
import { EditTeamForm } from "./EditTeamForm";
import { Pencil, Trash2 } from "lucide-react";
import { TeamMembersSection } from "./TeamMembersSection";

export interface TeamDetailsDrawerProps {
  currentOrgId: string;
  open: boolean;
  onClose: () => void;
  onEditClose: () => void;
  selectedTeam: Team | null;
  onEditClick: (team: Team) => void;
  onDeleteClick: (team: Team) => void;
  onTeamUpdated: (team: Team) => void;
  mode: string;
  canEdit: boolean;
  canDelete: boolean;
}

export const TeamDetailsDrawer = ({
  currentOrgId,
  open,
  onClose,
  onEditClose,
  selectedTeam,
  onEditClick,
  onDeleteClick,
  onTeamUpdated,
  mode,
  canEdit,
  canDelete,
}: TeamDetailsDrawerProps) => {
  const { t } = useTranslationNamespace(["settings", "common"]);

  if (!selectedTeam) {
    return null;
  }

  return (
    <Sheet open={open} onOpenChange={onClose}>
      <SheetContent className="sm:max-w-lg">
        <SheetHeader>
          <SheetTitle>{selectedTeam.name}</SheetTitle>
          <SheetDescription>
            {t("settings:org.teams.detailsDescription")}
          </SheetDescription>
        </SheetHeader>

        {mode === "view" && (
          <div className="mt-1 flex flex-col gap-6">
            {(canEdit || canDelete) && (
              <div className="flex items-center justify-end gap-2">
                {canEdit && (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      onEditClick(selectedTeam);
                    }}
                  >
                    <Pencil className="mr-2 h-4 w-4" />
                    {t("common:edit")}
                  </Button>
                )}
                {canDelete && (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="text-destructive hover:text-destructive"
                    onClick={() => onDeleteClick(selectedTeam)}
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    {t("common:delete")}
                  </Button>
                )}
              </div>
            )}

            {selectedTeam.description && (
              <div className="mt-1">
                <p className="text-xs font-semibold uppercase text-muted-foreground">
                  {t("settings:org.teams.descriptionLabel")}
                </p>
                <p className="mt-1 text-sm">{selectedTeam.description}</p>
              </div>
            )}

            <div className="mt-2">
              <TeamMembersSection
                currentOrgId={currentOrgId}
                selectedTeam={selectedTeam}
                canAddMember={canEdit}
                canRemoveMember={canDelete}
              />
            </div>
          </div>
        )}
        {mode === "edit" && (
          <div className="mt-6">
            <EditTeamForm
              currentOrgId={currentOrgId}
              selectedTeam={selectedTeam}
              onTeamUpdated={onTeamUpdated}
              onClose={onEditClose}
            />
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
};
