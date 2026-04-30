import { useMemo } from "react";
import { useParams } from "react-router";
import { useUserState } from "@/states/user";
import { useShallow } from "zustand/shallow";
import {
  PERMISSION_ORG_READ,
  PERMISSION_ORG_WRITE,
  PERMISSION_ORG_DELETE,
  PERMISSION_MEMBERS_READ,
  PERMISSION_MEMBERS_WRITE,
  PERMISSION_MEMBERS_DELETE,
  PERMISSION_AUDIT_READ,
  PERMISSION_SUBSCRIPTION_READ,
  PERMISSION_SUBSCRIPTION_WRITE,
  PERMISSION_ROLE_READ,
  PERMISSION_ROLE_WRITE,
  PERMISSION_ROLE_DELETE,
  PERMISSION_TEAM_READ,
  PERMISSION_TEAM_WRITE,
  PERMISSION_TEAM_DELETE,
  PERMISSION_TEAM_MEMBER_READ,
  PERMISSION_TEAM_MEMBER_WRITE,
  PERMISSION_TEAM_MEMBER_DELETE,
  PERMISSION_API_KEY_READ,
  PERMISSION_API_KEY_WRITE,
  PERMISSION_API_KEY_DELETE,
} from "@/constants";

export type SectionPermission = {
  canRead: boolean;
  canWrite: boolean;
  canDelete: boolean;
};

export type SettingsPermissions = {
  generalMerged: SectionPermission;
  org: SectionPermission;
  subscription: SectionPermission;
  role: SectionPermission;
  team: SectionPermission;
  teamMember: SectionPermission;
  members: SectionPermission;
  apiKey: SectionPermission;
  audit: SectionPermission;
};

function hasPermission(
  permissions: string[] | undefined,
  permission: string,
): boolean {
  return !!permissions?.includes(permission);
}

export function useSettingsPermissions(): SettingsPermissions {
  const { orgId } = useParams<{ orgId: string }>();
  const { userData } = useUserState(
    useShallow((state) => ({ userData: state.userData })),
  );

  return useMemo(() => {
    const currentOrgId = orgId || "";
    const selectedOrg = userData.roles?.find(
      (r) => r.organizationId === currentOrgId,
    );
    const perms = selectedOrg?.rolePermissions ?? [];

    const orgRead = hasPermission(perms, PERMISSION_ORG_READ);
    const subRead = hasPermission(perms, PERMISSION_SUBSCRIPTION_READ);

    return {
      generalMerged: {
        canRead: orgRead || subRead,
        canWrite:
          hasPermission(perms, PERMISSION_ORG_WRITE) ||
          hasPermission(perms, PERMISSION_SUBSCRIPTION_WRITE),
        canDelete: false,
      },
      org: {
        canRead: orgRead,
        canWrite: hasPermission(perms, PERMISSION_ORG_WRITE),
        canDelete: hasPermission(perms, PERMISSION_ORG_DELETE),
      },
      subscription: {
        canRead: hasPermission(perms, PERMISSION_SUBSCRIPTION_READ),
        canWrite: hasPermission(perms, PERMISSION_SUBSCRIPTION_WRITE),
        canDelete: false,
      },
      role: {
        canRead: hasPermission(perms, PERMISSION_ROLE_READ),
        canWrite: hasPermission(perms, PERMISSION_ROLE_WRITE),
        canDelete: hasPermission(perms, PERMISSION_ROLE_DELETE),
      },
      team: {
        canRead: hasPermission(perms, PERMISSION_TEAM_READ),
        canWrite: hasPermission(perms, PERMISSION_TEAM_WRITE),
        canDelete: hasPermission(perms, PERMISSION_TEAM_DELETE),
      },
      teamMember: {
        canRead: hasPermission(perms, PERMISSION_TEAM_MEMBER_READ),
        canWrite: hasPermission(perms, PERMISSION_TEAM_MEMBER_WRITE),
        canDelete: hasPermission(perms, PERMISSION_TEAM_MEMBER_DELETE),
      },
      members: {
        canRead: hasPermission(perms, PERMISSION_MEMBERS_READ),
        canWrite: hasPermission(perms, PERMISSION_MEMBERS_WRITE),
        canDelete: hasPermission(perms, PERMISSION_MEMBERS_DELETE),
      },
      apiKey: {
        canRead: hasPermission(perms, PERMISSION_API_KEY_READ),
        canWrite: hasPermission(perms, PERMISSION_API_KEY_WRITE),
        canDelete: hasPermission(perms, PERMISSION_API_KEY_DELETE),
      },
      audit: {
        canRead: hasPermission(perms, PERMISSION_AUDIT_READ),
        canWrite: false,
        canDelete: false,
      },
    };
  }, [orgId, userData.roles]);
}

/** Ordered list of settings section paths and their read permission key */
export const SETTINGS_SECTIONS = [
  { path: "general", permissionKey: "generalMerged" as const },
  { path: "roles", permissionKey: "role" as const },
  { path: "teams", permissionKey: "team" as const },
  { path: "members", permissionKey: "members" as const },
  { path: "api-keys", permissionKey: "apiKey" as const },
  { path: "audit-log", permissionKey: "audit" as const },
] as const;
