import { SUBSCRIPTION_TYPE } from "@/constants";
import type { Organization } from "@/types";

/** Days after subscription expiry we still allow access (past-due grace). Must match backend SUBSCRIPTION_PAST_DUE_GRACE_DAYS. */
export const PAST_DUE_GRACE_DAYS = 7;

/**
 * True when the organization has a paid subscription that is past its expiry date
 * but still within the grace period (so the user can access the app but should see a warning).
 */
export function isSubscriptionPastDue(
  org: Organization | null | undefined,
): boolean {
  if (!org?.subscriptionExpiresAt) return false;
  if (
    org.subscriptionType === SUBSCRIPTION_TYPE.FREE &&
    !org.customSubscription
  )
    return false;
  if (
    org.subscriptionType === SUBSCRIPTION_TYPE.ENTERPRISE ||
    org.customSubscription
  )
    return false;
  const expiresAt = new Date(org.subscriptionExpiresAt).getTime();
  const now = Date.now();
  if (expiresAt > now) return false;
  const daysPast = (now - expiresAt) / (24 * 60 * 60 * 1000);
  return daysPast <= PAST_DUE_GRACE_DAYS;
}

/** -1 = unlimited seats */
const UNLIMITED_SEATS = -1;

/**
 * Capabilities derived from subscription tier.
 * Free: org-level only, 1 seat, no custom roles/teams, no API keys, no audit.
 * Basic: up to 3 seats, audit 90 days; no custom roles, no teams, no API keys.
 * Pro+: custom roles, teams, API keys; Pro audit 365 days. Enterprise: unlimited seats and audit.
 */
export interface SubscriptionCapabilities {
  canManageRoles: boolean;
  canManageTeams: boolean;
  /** True if org can invite more members (seats > current or unlimited) */
  canInviteMore: boolean;
  canUseApiKeys: boolean;
  canViewAudit: boolean;
  /** Max seats for display; -1 = unlimited */
  maxSeats: number;
}

export function getSubscriptionCapabilities(
  org: Organization | null | undefined,
  memberCount: number | null,
): SubscriptionCapabilities {
  if (!org) {
    return {
      canManageRoles: false,
      canManageTeams: false,
      canInviteMore: false,
      canUseApiKeys: false,
      canViewAudit: false,
      maxSeats: 1,
    };
  }

  const type = org.subscriptionType ?? SUBSCRIPTION_TYPE.FREE;
  const seats = org.subscriptionSeats ?? 1;

  const canManageRoles =
    type === SUBSCRIPTION_TYPE.PRO || type === SUBSCRIPTION_TYPE.ENTERPRISE;
  const canManageTeams =
    type === SUBSCRIPTION_TYPE.PRO || type === SUBSCRIPTION_TYPE.ENTERPRISE;
  const canUseApiKeys =
    type === SUBSCRIPTION_TYPE.PRO || type === SUBSCRIPTION_TYPE.ENTERPRISE;
  const canViewAudit =
    type === SUBSCRIPTION_TYPE.BASIC ||
    type === SUBSCRIPTION_TYPE.PRO ||
    type === SUBSCRIPTION_TYPE.ENTERPRISE;
  const isUnlimitedSeats = seats === UNLIMITED_SEATS;
  const canInviteMore =
    memberCount !== null && (isUnlimitedSeats || memberCount < seats);

  return {
    canManageRoles,
    canManageTeams,
    canInviteMore,
    canUseApiKeys,
    canViewAudit,
    maxSeats: seats,
  };
}

export function getAuditRetentionDays(
  subscriptionType: number | undefined,
): number {
  switch (subscriptionType) {
    case SUBSCRIPTION_TYPE.BASIC:
      return 180;
    case SUBSCRIPTION_TYPE.PRO:
      return 365;
    case SUBSCRIPTION_TYPE.ENTERPRISE:
      // Enterprise: 3 years of retention.
      return 365 * 3;
    case SUBSCRIPTION_TYPE.FREE:
    default:
      return 30;
  }
}
