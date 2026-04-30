/**
 * Permission groups for role create/edit. Grouped by context; read is implied when write or delete is selected.
 */
export type PermissionAction = "read" | "write" | "delete";

export interface PermissionOption {
  key: string;
  action: PermissionAction;
  labelKey: string;
}

export interface RolePermissionGroup {
  contextKey: string;
  labelKey: string;
  /** Short description shown on info icon hover */
  descriptionKey: string;
  permissions: PermissionOption[];
}

export const ROLE_PERMISSION_GROUPS: RolePermissionGroup[] = [
  {
    contextKey: "org",
    labelKey: "settings:org.roles.permissionGroups.org.label",
    descriptionKey: "settings:org.roles.permissionGroups.org.description",
    permissions: [
      {
        key: "org:read",
        action: "read",
        labelKey: "settings:org.roles.permissionActions.read",
      },
      {
        key: "org:write",
        action: "write",
        labelKey: "settings:org.roles.permissionActions.write",
      },
      {
        key: "org:delete",
        action: "delete",
        labelKey: "settings:org.roles.permissionActions.delete",
      },
    ],
  },
  {
    contextKey: "members",
    labelKey: "settings:org.roles.permissionGroups.members.label",
    descriptionKey: "settings:org.roles.permissionGroups.members.description",
    permissions: [
      {
        key: "members:read",
        action: "read",
        labelKey: "settings:org.roles.permissionActions.read",
      },
      {
        key: "members:write",
        action: "write",
        labelKey: "settings:org.roles.permissionActions.write",
      },
      {
        key: "members:delete",
        action: "delete",
        labelKey: "settings:org.roles.permissionActions.delete",
      },
    ],
  },
  {
    contextKey: "audit",
    labelKey: "settings:org.roles.permissionGroups.audit.label",
    descriptionKey: "settings:org.roles.permissionGroups.audit.description",
    permissions: [
      {
        key: "audit:read",
        action: "read",
        labelKey: "settings:org.roles.permissionActions.read",
      },
    ],
  },
  {
    contextKey: "subscription",
    labelKey: "settings:org.roles.permissionGroups.subscription.label",
    descriptionKey:
      "settings:org.roles.permissionGroups.subscription.description",
    permissions: [
      {
        key: "subscription:read",
        action: "read",
        labelKey: "settings:org.roles.permissionActions.read",
      },
      {
        key: "subscription:write",
        action: "write",
        labelKey: "settings:org.roles.permissionActions.write",
      },
    ],
  },
  {
    contextKey: "role",
    labelKey: "settings:org.roles.permissionGroups.role.label",
    descriptionKey: "settings:org.roles.permissionGroups.role.description",
    permissions: [
      {
        key: "role:read",
        action: "read",
        labelKey: "settings:org.roles.permissionActions.read",
      },
      {
        key: "role:write",
        action: "write",
        labelKey: "settings:org.roles.permissionActions.write",
      },
      {
        key: "role:delete",
        action: "delete",
        labelKey: "settings:org.roles.permissionActions.delete",
      },
    ],
  },
  {
    contextKey: "team",
    labelKey: "settings:org.roles.permissionGroups.team.label",
    descriptionKey: "settings:org.roles.permissionGroups.team.description",
    permissions: [
      {
        key: "team:read",
        action: "read",
        labelKey: "settings:org.roles.permissionActions.read",
      },
      {
        key: "team:write",
        action: "write",
        labelKey: "settings:org.roles.permissionActions.write",
      },
      {
        key: "team:delete",
        action: "delete",
        labelKey: "settings:org.roles.permissionActions.delete",
      },
    ],
  },
  {
    contextKey: "team_member",
    labelKey: "settings:org.roles.permissionGroups.teamMember.label",
    descriptionKey:
      "settings:org.roles.permissionGroups.teamMember.description",
    permissions: [
      {
        key: "team_member:read",
        action: "read",
        labelKey: "settings:org.roles.permissionActions.read",
      },
      {
        key: "team_member:write",
        action: "write",
        labelKey: "settings:org.roles.permissionActions.write",
      },
      {
        key: "team_member:delete",
        action: "delete",
        labelKey: "settings:org.roles.permissionActions.delete",
      },
    ],
  },
  {
    contextKey: "api_key",
    labelKey: "settings:org.roles.permissionGroups.apiKeys.label",
    descriptionKey: "settings:org.roles.permissionGroups.apiKeys.description",
    permissions: [
      {
        key: "api_key:read",
        action: "read",
        labelKey: "settings:org.roles.permissionActions.read",
      },
      {
        key: "api_key:write",
        action: "write",
        labelKey: "settings:org.roles.permissionActions.write",
      },
      {
        key: "api_key:delete",
        action: "delete",
        labelKey: "settings:org.roles.permissionActions.delete",
      },
    ],
  },
];
