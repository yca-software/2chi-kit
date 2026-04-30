import {
  getSubscriptionCapabilities,
  useTranslationNamespace,
} from "@/helpers";
import { usePricingModalStore, useUserState } from "@/states";
import { OrganizationMemberWithUser } from "@/types";
import { useMemo, useState } from "react";
import { Navigate, useParams } from "react-router";
import { useShallow } from "zustand/shallow";
import {
  SETTINGS_SECTIONS,
  useSettingsPermissions,
} from "../useSettingsPermissions";
import { Loader2, UserPlus, Users2 } from "lucide-react";
import {
  useGetOrganizationQuery,
  useListInvitationsQuery,
  useListOrganizationMembersQuery,
  useListRolesQuery,
} from "@/api";
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@yca-software/design-system";
import { MemberRow } from "./MemberRow";
import { InviteDrawer } from "./InviteDrawer";
import { PendingInvitations } from "./PendingInvitations";
import { RemoveMemberDialog } from "./RemoveMemberDialog";
import { MemberDetailsDrawer } from "./MemberDetailsDrawer";

type MembersUIState = {
  selectedMember: OrganizationMemberWithUser | null;
  drawerMode: string; // "view", "create"
  isDeleteMemberDialogOpen: boolean;
};

export const MembersSettings = () => {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const { orgId } = useParams<{ orgId: string }>();
  const { openForOrg } = usePricingModalStore(
    useShallow((state) => ({
      openForOrg: state.openForOrg,
    })),
  );
  const { userData } = useUserState(
    useShallow((state) => ({ userData: state.userData })),
  );
  const currentUserId = userData.user?.id;
  const currentOrgId = orgId || "";

  const selectedCurrentUserRole = useMemo(
    () => userData.roles?.find((r) => r.organizationId === currentOrgId),
    [currentOrgId, userData.roles],
  );
  const currentUserRoleName = selectedCurrentUserRole?.roleName ?? "";

  const [state, setState] = useState<MembersUIState>({
    selectedMember: null,
    drawerMode: "",
    isDeleteMemberDialogOpen: false,
  });

  const permissions = useSettingsPermissions();
  const canManageMembers = permissions.members.canWrite;
  const canDeleteMembers = permissions.members.canDelete;

  const { data: organization } = useGetOrganizationQuery(currentOrgId);
  const { data: members, isLoading: membersLoading } =
    useListOrganizationMembersQuery(currentOrgId);
  const { data: invitationsData, isLoading: invitationsLoading } =
    useListInvitationsQuery(currentOrgId);
  const invitations = invitationsData ?? [];
  const { data: rolesData, isLoading: rolesLoading } =
    useListRolesQuery(currentOrgId);
  const roles = rolesData ?? [];

  const capabilities = getSubscriptionCapabilities(
    organization,
    members ? members.length : null,
  );

  const isLoading = membersLoading || rolesLoading || invitationsLoading;

  if (!permissions.members.canRead) {
    const firstAllowed = SETTINGS_SECTIONS.find(
      (s) => permissions[s.permissionKey]?.canRead,
    );
    return firstAllowed ? (
      <Navigate to={`/settings/${orgId}/${firstAllowed.path}`} replace />
    ) : (
      <Navigate to={`/settings/${orgId}`} replace />
    );
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <InviteDrawer
        currentOrgId={currentOrgId}
        open={state.drawerMode === "invite"}
        onClose={() => {
          setState((s) => ({ ...s, drawerMode: "" }));
        }}
        roles={roles}
      />

      <RemoveMemberDialog
        open={state.isDeleteMemberDialogOpen}
        onClose={() => {
          setState((s) => ({ ...s, isDeleteMemberDialogOpen: false }));
        }}
        selectedMember={state.selectedMember}
        currentOrgId={currentOrgId}
      />

      <MemberDetailsDrawer
        currentOrgId={currentOrgId}
        open={state.drawerMode === "view" && !!state.selectedMember}
        onClose={() => {
          setState((s) => ({ ...s, drawerMode: "", selectedMember: null }));
        }}
        selectedMember={state.selectedMember}
        roles={roles}
        canUpdateRole={canManageMembers}
        canRemove={canDeleteMembers}
        isSelf={
          !!currentUserId &&
          !!state.selectedMember &&
          state.selectedMember.userId === currentUserId
        }
        onDeleteClick={(member) => {
          setState((s) => ({
            ...s,
            selectedMember: member,
            isDeleteMemberDialogOpen: true,
          }));
        }}
        onMemberUpdated={(member) => {
          setState((s) => ({
            ...s,
            selectedMember: member,
            drawerMode: "view",
          }));
        }}
      />

      <Card>
        <CardHeader>
          <div className="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="min-w-0">
              <CardTitle>{t("settings:org.members.title")}</CardTitle>
              <CardDescription>
                {t("settings:org.members.description")}
              </CardDescription>
            </div>
            <div className="shrink-0 flex flex-col items-end gap-2">
              {canManageMembers && roles.length > 0 && (
                <Button
                  onClick={() => {
                    setState((s) => ({ ...s, drawerMode: "invite" }));
                  }}
                  disabled={!capabilities.canInviteMore}
                >
                  <UserPlus className="mr-2 h-4 w-4" />
                  {t("common:invite")}
                </Button>
              )}
              {!capabilities.canInviteMore && (
                <button
                  type="button"
                  onClick={() => openForOrg(currentOrgId)}
                  className="text-xs text-primary underline-offset-2 hover:underline cursor-pointer"
                >
                  {t("settings:org.upsell.membersDescription")}
                </button>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {members && members.length > 0 ? (
            <div className="space-y-2">
              {members.map((member) => (
                <MemberRow
                  key={member.id}
                  member={member}
                  roleName={
                    roles.find((r) => String(r.id) === String(member.roleId))
                      ?.name ??
                    (member.userId === currentUserId ? currentUserRoleName : "")
                  }
                  isSelf={!!currentUserId && member.userId === currentUserId}
                  onSelect={() =>
                    setState((s) => ({
                      ...s,
                      selectedMember: member,
                      drawerMode: "view",
                    }))
                  }
                  canRemove={canDeleteMembers}
                  onRemoveClick={() =>
                    setState((s) => ({
                      ...s,
                      selectedMember: member,
                      isDeleteMemberDialogOpen: true,
                    }))
                  }
                />
              ))}
            </div>
          ) : (
            <div className="rounded-lg border border-dashed bg-muted/20 px-6 py-10 text-center">
              <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary">
                <Users2 className="h-6 w-6" />
              </div>
              <h3 className="mt-4 text-base font-semibold">
                {t("settings:org.members.noMembersTitle")}
              </h3>
              <p className="mx-auto mt-2 max-w-md text-sm text-muted-foreground">
                {t("settings:org.members.noMembersDescription")}
              </p>
              {canManageMembers && roles.length > 0 && (
                <Button
                  size="sm"
                  className="mt-5"
                  disabled={!capabilities.canInviteMore}
                  onClick={() => {
                    setState((s) => ({ ...s, drawerMode: "invite" }));
                  }}
                >
                  <UserPlus className="mr-2 h-4 w-4" />
                  {t("common:invite")}
                </Button>
              )}
            </div>
          )}

          <PendingInvitations
            currentOrgId={currentOrgId}
            invitations={invitations}
            roles={roles}
            canRemove={canDeleteMembers}
          />
        </CardContent>
      </Card>
    </div>
  );
};
