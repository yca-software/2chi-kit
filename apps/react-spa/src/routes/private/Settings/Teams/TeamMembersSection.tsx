import { useMemo, useState } from "react";
import { useTranslationNamespace } from "@/helpers";
import { Button, Separator } from "@yca-software/design-system";
import { Loader2, Trash2, User } from "lucide-react";
import type {
  OrganizationMemberWithUser,
  Team,
  TeamMemberWithUser,
} from "@/types";
import {
  useListOrganizationMembersQuery,
  useListTeamMembersQuery,
} from "@/api";
import { AddTeamMemberDropdown } from "./AddTeamMemberDropdown";
import { RemoveTeamMemberDialog } from "./RemoveTeamMemberDialog";

interface TeamMembersSectionProps {
  currentOrgId: string;
  selectedTeam: Team;
  canAddMember: boolean;
  canRemoveMember: boolean;
}

export function TeamMembersSection({
  currentOrgId,
  selectedTeam,
  canAddMember,
  canRemoveMember,
}: TeamMembersSectionProps) {
  const { t } = useTranslationNamespace(["settings"]);

  const { data: orgMembers } = useListOrganizationMembersQuery(currentOrgId);

  const { data: teamMembersData, isLoading: teamMembersLoading } =
    useListTeamMembersQuery(currentOrgId, selectedTeam.id);
  const teamMembers = teamMembersData ?? [];

  const availableMembersToAdd = useMemo(() => {
    if (!orgMembers || !teamMembers) return [];
    const teamUserIds = new Set(
      (teamMembers as TeamMemberWithUser[]).map((member) => member.userId),
    );
    return (orgMembers as OrganizationMemberWithUser[]).filter(
      (member) => !teamUserIds.has(member.userId),
    );
  }, [orgMembers, teamMembers]);

  const [isRemoveDialogOpen, setIsRemoveDialogOpen] = useState(false);
  const [selectedTeamMember, setSelectedTeamMember] =
    useState<TeamMemberWithUser | null>(null);

  const showAddMemberSection = canAddMember && availableMembersToAdd.length > 0;

  return (
    <div>
      <RemoveTeamMemberDialog
        open={isRemoveDialogOpen && !!selectedTeamMember}
        onClose={() => {
          setIsRemoveDialogOpen(false);
          setSelectedTeamMember(null);
        }}
        selectedTeamMember={selectedTeamMember}
        selectedTeam={selectedTeam}
        currentOrgId={currentOrgId}
      />

      {teamMembersLoading ? (
        <div className="flex justify-center py-8">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : (
        <div className="space-y-4">
          {showAddMemberSection && (
            <AddTeamMemberDropdown
              currentOrgId={currentOrgId}
              team={selectedTeam}
              members={availableMembersToAdd}
            />
          )}
          {showAddMemberSection && <Separator />}
          <div className="flex items-center justify-between">
            <h4 className="text-sm font-medium">
              {t("settings:org.teams.teamMembers")}
            </h4>
          </div>
          {teamMembers.length > 0 ? (
            <div className="space-y-2">
              {teamMembers.map((member) => (
                <div
                  key={member.id}
                  className="flex items-center justify-between rounded-lg border p-3"
                >
                  <div className="flex items-center gap-3">
                    <div className="bg-primary/10 flex h-8 w-8 items-center justify-center rounded-full">
                      <User className="h-4 w-4 text-primary" />
                    </div>
                    <div>
                      <p className="text-sm font-medium">
                        {member.userFirstName} {member.userLastName}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {member.userEmail}
                      </p>
                    </div>
                  </div>
                  {canRemoveMember && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => {
                        setSelectedTeamMember(member);
                        setIsRemoveDialogOpen(true);
                      }}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <p className="py-4 text-center text-sm text-muted-foreground">
              {t("settings:org.teams.noTeamMembers")}
            </p>
          )}
        </div>
      )}
    </div>
  );
}
