import { useState, useMemo, useCallback, useRef, useEffect } from "react";
import { useTranslationNamespace } from "@/helpers/hooks/useTranslationNamespace";
import {
  Button,
  Input,
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  Separator,
} from "@yca-software/design-system";
import { Loader2, Trash2, User } from "lucide-react";
import type { OrganizationMemberWithUser, TeamMemberWithUser } from "@/types/organization";

interface TeamMembersDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  teamName: string;
  teamMembers: TeamMemberWithUser[];
  availableMembersToAdd: OrganizationMemberWithUser[];
  teamMembersLoading: boolean;
  addMemberPending: boolean;
  onAddMember: (userId: string) => void;
  onRemoveMemberClick: (memberId: string) => void;
  canAddMember?: boolean;
  canRemoveMember?: boolean;
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

export function TeamMembersDrawer({
  open,
  onOpenChange,
  teamName,
  teamMembers,
  availableMembersToAdd,
  teamMembersLoading,
  addMemberPending,
  onAddMember,
  onRemoveMemberClick,
  canAddMember = true,
  canRemoveMember = true,
}: TeamMembersDrawerProps) {
  const { t } = useTranslationNamespace(["settings"]);
  const containerRef = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  const [searchQuery, setSearchQuery] = useState("");
  const [showPredictions, setShowPredictions] = useState(false);
  const [highlightedIndex, setHighlightedIndex] = useState(-1);

  useEffect(() => {
    if (!open) {
      setSearchQuery("");
      setShowPredictions(false);
      setHighlightedIndex(-1);
    } else {
      (document.activeElement as HTMLElement)?.blur();
    }
  }, [open]);

  const filteredMembers = useMemo(
    () => filterMembers(availableMembersToAdd, searchQuery),
    [availableMembersToAdd, searchQuery],
  );
  const selectableCount = filteredMembers.length;

  useEffect(() => {
    setHighlightedIndex(-1);
  }, [filteredMembers]);

  useEffect(() => {
    if (highlightedIndex >= 0 && listRef.current) {
      const item = listRef.current.querySelector(
        `[data-index="${highlightedIndex}"]`,
      );
      item?.scrollIntoView({ block: "nearest" });
    }
  }, [highlightedIndex]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Node;
      if (containerRef.current && !containerRef.current.contains(target)) {
        setShowPredictions(false);
        setHighlightedIndex(-1);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(e.target.value);
    setShowPredictions(true);
  };

  const handleFocus = () => {
    setShowPredictions(true);
  };

  const selectMember = useCallback(
    (member: OrganizationMemberWithUser) => {
      onAddMember(member.userId);
      setSearchQuery("");
      setShowPredictions(false);
      setHighlightedIndex(-1);
    },
    [onAddMember],
  );

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        setShowPredictions(false);
        setHighlightedIndex(-1);
        return;
      }
      if (!showPredictions || selectableCount === 0) return;
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setHighlightedIndex((prev) =>
          prev < selectableCount - 1 ? prev + 1 : 0,
        );
        return;
      }
      if (e.key === "ArrowUp") {
        e.preventDefault();
        setHighlightedIndex((prev) =>
          prev <= 0 ? selectableCount - 1 : prev - 1,
        );
        return;
      }
      if (
        e.key === "Enter" &&
        highlightedIndex >= 0 &&
        filteredMembers[highlightedIndex]
      ) {
        e.preventDefault();
        selectMember(filteredMembers[highlightedIndex]);
      }
    },
    [
      showPredictions,
      selectableCount,
      highlightedIndex,
      filteredMembers,
      selectMember,
    ],
  );

  const showAddMemberSection =
    canAddMember && availableMembersToAdd.length > 0;

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        className="sm:max-w-lg"
        onOpenAutoFocus={(e) => e.preventDefault()}
      >
        <SheetHeader>
          <SheetTitle>
            {teamName} – {t("settings:org.teams.members")}
          </SheetTitle>
          <SheetDescription>
            {t("settings:org.teams.membersDescription")}
          </SheetDescription>
        </SheetHeader>
        <div>
          {teamMembersLoading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
            </div>
          ) : (
            <div className="space-y-4">
              {showAddMemberSection && (
                <div ref={containerRef} className="relative space-y-2">
                  <label htmlFor="add-member-input" className="text-sm font-medium">
                    {t("settings:org.teams.addMembersToTeamLabel")}
                  </label>
                  <div className="relative">
                    {addMemberPending && (
                      <div className="absolute right-3 top-1/2 z-10 -translate-y-1/2">
                        <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                      </div>
                    )}
                    <Input
                      id="add-member-input"
                      type="text"
                      value={searchQuery}
                      onChange={handleInputChange}
                      onKeyDown={handleKeyDown}
                      placeholder={t("settings:org.teams.searchMemberToAdd")}
                      disabled={addMemberPending}
                      className="pl-10"
                      onFocus={handleFocus}
                      aria-autocomplete="list"
                      aria-expanded={showPredictions && selectableCount > 0}
                      aria-controls={
                        showPredictions
                          ? "add-member-predictions-list"
                          : undefined
                      }
                      aria-activedescendant={
                        highlightedIndex >= 0
                          ? `add-member-prediction-${highlightedIndex}`
                          : undefined
                      }
                      role="combobox"
                      aria-label={t("settings:org.teams.addMembersToTeamLabel")}
                    />
                    <User className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  </div>
                  {showPredictions && (
                    <div
                      ref={listRef}
                      id="add-member-predictions-list"
                      role="listbox"
                      className="absolute z-110 mt-1 max-h-60 w-full overflow-auto rounded-md border bg-background shadow-lg"
                    >
                      {filteredMembers.length === 0 ? (
                        <div className="px-4 py-3 text-sm text-muted-foreground">
                          {searchQuery.trim()
                            ? t("settings:org.teams.noMembersMatch")
                            : t("settings:org.teams.noMembersToAdd")}
                        </div>
                      ) : (
                        filteredMembers.map((member, index) => (
                          <button
                            key={member.id}
                            type="button"
                            role="option"
                            id={`add-member-prediction-${index}`}
                            data-index={index}
                            aria-selected={highlightedIndex === index}
                            className={`flex w-full cursor-pointer items-center gap-2 px-4 py-2 text-left transition-colors ${
                              highlightedIndex === index
                                ? "bg-accent text-accent-foreground"
                                : "hover:bg-accent hover:text-accent-foreground"
                            }`}
                            onMouseEnter={() => setHighlightedIndex(index)}
                            onClick={() => selectMember(member)}
                          >
                            <User className="h-4 w-4 shrink-0" />
                            <div>
                              <div className="font-medium">
                                {member.userFirstName} {member.userLastName}
                              </div>
                              <div className="text-xs text-muted-foreground">
                                {member.userEmail}
                              </div>
                            </div>
                          </button>
                        ))
                      )}
                    </div>
                  )}
                </div>
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
                          onClick={() => onRemoveMemberClick(member.id)}
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
      </SheetContent>
    </Sheet>
  );
}
