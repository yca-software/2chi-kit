import { useMemo, useState } from "react";
import {
  Button,
  Input,
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@yca-software/design-system";
import { Loader2, User } from "lucide-react";
import type { MutationError, OrganizationMemberWithUser, Team } from "@/types";
import { useTranslationNamespace } from "@/helpers";
import { useAddTeamMemberMutation } from "@/api";
import { toast } from "sonner";

interface AddTeamMemberDropdownProps {
  currentOrgId: string;
  team: Team;
  members: OrganizationMemberWithUser[];
}

function filterMembers(
  members: OrganizationMemberWithUser[],
  query: string,
): OrganizationMemberWithUser[] {
  const q = query.trim().toLowerCase();
  if (!q) return members;
  return members.filter(
    (m) =>
      m.userFirstName?.toLowerCase().includes(q) ||
      m.userLastName?.toLowerCase().includes(q) ||
      m.userEmail?.toLowerCase().includes(q),
  );
}

export function AddTeamMemberDropdown({
  currentOrgId,
  team,
  members,
}: AddTeamMemberDropdownProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const [searchQuery, setSearchQuery] = useState("");
  const [isOpen, setIsOpen] = useState(false);

  const addTeamMemberMutation = useAddTeamMemberMutation(
    currentOrgId,
    team.id,
    {
      onSuccess: () => {
        setSearchQuery("");
        setIsOpen(false);
      },
      onError: (err: MutationError) => {
        toast.error(err.error?.message ?? t("common:defaultError"));
      },
    },
  );
  const isPending = addTeamMemberMutation.isPending;

  const filteredMembers = useMemo(
    () => filterMembers(members, searchQuery),
    [members, searchQuery],
  );

  return (
    <div className="space-y-2">
      <label className="text-sm font-medium">
        {t("settings:org.teams.addMembersToTeamLabel")}
      </label>
      <Popover open={isOpen} onOpenChange={setIsOpen}>
        <PopoverTrigger asChild>
          <Button
            type="button"
            variant="outline"
            className="w-full justify-between"
            disabled={isPending}
          >
            <span className="truncate text-muted-foreground">
              {t("settings:org.teams.searchMemberToAdd")}
            </span>
            {isPending ? (
              <Loader2 className="ml-2 h-4 w-4 animate-spin" />
            ) : (
              <User className="ml-2 h-4 w-4" />
            )}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-(--radix-popover-trigger-width) p-2">
          <Input
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder={t("settings:org.teams.searchMemberToAdd")}
            disabled={isPending}
          />
          <div className="mt-2 max-h-60 overflow-auto rounded-md border">
            {filteredMembers.length === 0 ? (
              <div className="px-3 py-2 text-sm text-muted-foreground">
                {searchQuery.trim()
                  ? t("settings:org.teams.noMembersMatch")
                  : t("settings:org.teams.noMembersToAdd")}
              </div>
            ) : (
              filteredMembers.map((member) => (
                <button
                  key={member.id}
                  type="button"
                  className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground"
                  onClick={() => {
                    addTeamMemberMutation.mutate({ userId: member.userId });
                  }}
                >
                  <User className="h-4 w-4 shrink-0" />
                  <div className="min-w-0">
                    <div className="truncate font-medium">
                      {member.userFirstName} {member.userLastName}
                    </div>
                    <div className="truncate text-xs text-muted-foreground">
                      {member.userEmail}
                    </div>
                  </div>
                </button>
              ))
            )}
          </div>
        </PopoverContent>
      </Popover>
    </div>
  );
}
