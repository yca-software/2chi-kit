// Auth
export const REFRESH_TOKEN_QUERY_KEYS = {
  ACTIVE: "user-active-refresh-tokens",
};

// User
export const USER_QUERY_KEYS = {
  CURRENT: "current-user",
};

// Organization
export const ORGANIZATION_QUERY_KEYS = {
  DETAIL: "organization",
  ALL: "organizations",
  MEMBERS: "organization-members",
  ROLES: "organization-roles",
  TEAMS: "organization-teams",
  TEAM_MEMBERS: "team-members",
  API_KEYS: "organization-api-keys",
  AUDIT_LOGS: "organization-audit-logs",
};

// Invitation
export const INVITATION_QUERY_KEYS = {
  LIST: "invitations",
};

// Admin
export const ADMIN_USER_QUERY_KEYS = {
  ALL: "admin-users",
  DETAIL: "admin-user",
  AUDIT_LOGS: "admin-user-audit-logs",
};

export const ADMIN_ORGANIZATION_QUERY_KEYS = {
  ALL: "admin-organizations",
  ALL_ARCHIVED: "admin-organizations-archived",
  DETAIL: "admin-organization",
  DETAIL_ARCHIVED: "admin-organization-archived",
  AUDIT_LOGS: "admin-organization-audit-logs",
};
