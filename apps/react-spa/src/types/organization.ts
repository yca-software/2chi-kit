import type { Point } from "./geo";

export interface Organization {
  id: string;
  createdAt: string;
  deletedAt: string | null;

  name: string;

  address: string;
  city: string;
  zip: string;
  country: string;
  placeId: string;
  geo: Point;
  timezone: string;

  billingEmail: string;
  customSubscription: boolean;
  subscriptionExpiresAt: string | null;
  subscriptionPaymentInterval: number;
  subscriptionType: number;
  subscriptionSeats: number;
  subscriptionInTrial?: boolean;

  paddleCustomerId: string;
  paddleSubscriptionId: string | null;
  /** When set: switch to this plan at end of current period (e.g. annual→monthly). */
  scheduledPlanPriceId?: string | null;
}

export type RolePermissions = string[];

export interface Role {
  id: string;
  createdAt: string;

  organizationId: string;

  name: string;
  description: string;
  permissions: RolePermissions;
  locked: boolean;
}

export interface OrganizationMember {
  id: string;
  createdAt: string;

  organizationId: string;
  userId: string;
  roleId: string;
}

export interface OrganizationMemberWithOrganization extends OrganizationMember {
  organizationName: string;
}

export interface OrganizationMemberWithOrganizationAndRole extends OrganizationMemberWithOrganization {
  roleName: string;
  rolePermissions: RolePermissions;
}

export interface OrganizationMemberWithUser extends OrganizationMember {
  userEmail: string;
  userFirstName: string;
  userLastName: string;
}

export interface Team {
  id: string;
  createdAt: string;

  organizationId: string;

  name: string;
  description: string;
}

export interface TeamMember {
  id: string;
  createdAt: string;

  organizationId: string;
  teamId: string;
  userId: string;
}

export interface TeamMemberWithTeam extends TeamMember {
  teamName: string;
}

export interface TeamMemberWithUser extends TeamMember {
  userEmail: string;
  userFirstName: string;
  userLastName: string;
}

export interface Invitation {
  id: string;
  createdAt: string;
  expiresAt: string;
  acceptedAt: string | null;
  revokedAt: string | null;

  organizationId: string;
  roleId: string;
  email: string;

  invitedById: string | null;
  invitedByEmail: string;
}

export interface ApiKey {
  id: string;
  createdAt: string;
  expiresAt: string;

  name: string;
  keyPrefix: string;
  organizationId: string;
  permissions: RolePermissions;
}
