/**
 * Permission constants - must match backend (internals/constants/permission.go)
 */
export const PERMISSION_ORG_READ = "org:read";
export const PERMISSION_ORG_WRITE = "org:write";
export const PERMISSION_ORG_DELETE = "org:delete";

export const PERMISSION_MEMBERS_READ = "members:read";
export const PERMISSION_MEMBERS_WRITE = "members:write";
export const PERMISSION_MEMBERS_DELETE = "members:delete";

export const PERMISSION_AUDIT_READ = "audit:read";

export const PERMISSION_SUBSCRIPTION_READ = "subscription:read";
export const PERMISSION_SUBSCRIPTION_WRITE = "subscription:write";

export const PERMISSION_ROLE_READ = "role:read";
export const PERMISSION_ROLE_WRITE = "role:write";
export const PERMISSION_ROLE_DELETE = "role:delete";

export const PERMISSION_TEAM_READ = "team:read";
export const PERMISSION_TEAM_WRITE = "team:write";
export const PERMISSION_TEAM_DELETE = "team:delete";

export const PERMISSION_TEAM_MEMBER_READ = "team_member:read";
export const PERMISSION_TEAM_MEMBER_WRITE = "team_member:write";
export const PERMISSION_TEAM_MEMBER_DELETE = "team_member:delete";

export const PERMISSION_API_KEY_READ = "api_key:read";
export const PERMISSION_API_KEY_WRITE = "api_key:write";
export const PERMISSION_API_KEY_DELETE = "api_key:delete";

/**
 * Permissions that can be assigned to an API key.
 * Must match backend assignablePermissions (internals/services/api_key/main.go).
 */
export const ASSIGNABLE_API_KEY_PERMISSIONS: string[] = [
  PERMISSION_ORG_READ,
  PERMISSION_MEMBERS_READ,
  PERMISSION_AUDIT_READ,
  PERMISSION_SUBSCRIPTION_READ,
  PERMISSION_ROLE_READ,
  PERMISSION_TEAM_READ,
  PERMISSION_TEAM_MEMBER_READ,
];
