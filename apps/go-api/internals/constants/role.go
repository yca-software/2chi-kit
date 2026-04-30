package constants

import "github.com/yca-software/2chi-kit/go-api/internals/models"

var DEFAULT_OWNER_PERMISSIONS = []string{
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
}

var DEFAULT_ADMIN_PERMISSIONS = []string{
	PERMISSION_ORG_READ,
	PERMISSION_ORG_WRITE,
	PERMISSION_MEMBERS_READ,
	PERMISSION_MEMBERS_WRITE,
	PERMISSION_MEMBERS_DELETE,
	PERMISSION_AUDIT_READ,
	PERMISSION_SUBSCRIPTION_READ,
	PERMISSION_ROLE_READ,
	PERMISSION_ROLE_WRITE,
	PERMISSION_ROLE_DELETE,
	PERMISSION_TEAM_READ,
	PERMISSION_TEAM_WRITE,
	PERMISSION_TEAM_DELETE,
	PERMISSION_TEAM_MEMBER_READ,
	PERMISSION_TEAM_MEMBER_WRITE,
	PERMISSION_TEAM_MEMBER_DELETE,
}

var DEFAULT_TEAM_MANAGER_PERMISSIONS = []string{
	PERMISSION_ORG_READ,
	PERMISSION_MEMBERS_READ,
	PERMISSION_ROLE_READ,
	PERMISSION_TEAM_READ,
	PERMISSION_TEAM_MEMBER_READ,
}

var DEFAULT_TEAM_MEMBER_PERMISSIONS = []string{
	PERMISSION_ORG_READ,
	PERMISSION_MEMBERS_READ,
	PERMISSION_ROLE_READ,
	PERMISSION_TEAM_READ,
	PERMISSION_TEAM_MEMBER_READ,
}

var DEFAULT_ROLES_TO_CREATE_FOR_ORGANIZATION = []models.Role{
	{
		Name:        "Owner",
		Description: "Full control over the organization: manage all settings, members, roles, billing, teams, and audit logs. Can transfer or delete the organization. Typically the account creator.",
		Permissions: DEFAULT_OWNER_PERMISSIONS,
		Locked:      true,
	},
	{
		Name:        "Admin",
		Description: "Can manage most organization-level settings, members (including inviting, updating, or removing), access billing info, manage all roles and teams, and audit logs. Cannot transfer or delete the organization.",
		Permissions: DEFAULT_ADMIN_PERMISSIONS,
		Locked:      true,
	},
	{
		Name:        "Team Manager",
		Description: "Can view organization information, list members and roles, view teams, and see team members. Cannot modify organization settings, invite or update members, or change teams or roles.",
		Permissions: DEFAULT_TEAM_MANAGER_PERMISSIONS,
		Locked:      true,
	},
	{
		Name:        "Team Member",
		Description: "Read-only access to basic organization, team, member, and role data. Cannot invite, update, or remove members, and cannot make any changes to teams, roles, or settings.",
		Permissions: DEFAULT_TEAM_MEMBER_PERMISSIONS,
		Locked:      true,
	},
}
