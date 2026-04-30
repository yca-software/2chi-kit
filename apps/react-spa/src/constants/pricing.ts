/**
 * Pricing and subscription constants.
 * Paddle price IDs are read from env (VITE_PADDLE_PRICE_*).
 * Trial days are defined here; actual trial is configured in Paddle.
 */

export const TRIAL_DAYS = 30;

/** Annual discount as a percentage (e.g. 15 = 15% off when paying annually). Matches Paddle: Basic $204/yr vs $240 (12×$20), Pro $408/yr vs $480 (12×$40). */
export const ANNUAL_DISCOUNT_PERCENT = 15;

export type PlanId = "basic" | "pro" | "enterprise";

export interface PlanFeature {
  key: string;
  included: boolean;
}

export interface PlanConfig {
  id: PlanId;
  nameKey: string;
  descriptionKey: string;
  features: PlanFeature[];
  /** Monthly price in cents (for display; actual price comes from Paddle) */
  monthlyPriceCents: number;
  /** Annual price in cents (for display) */
  annualPriceCents: number;
  /** Uses Paddle checkout; false for Free and Enterprise (contact us) */
  usePaddle: boolean;
}

export const PLANS: PlanConfig[] = [
  {
    id: "basic",
    nameKey: "pricing:plans.basic.name",
    descriptionKey: "pricing:plans.basic.description",
    monthlyPriceCents: 2000,
    annualPriceCents: 20400,
    usePaddle: true,
    features: [
      { key: "pricing:features.teamMembersBasic", included: true },
      { key: "pricing:features.auditLogBasic", included: true },
    ],
  },
  {
    id: "pro",
    nameKey: "pricing:plans.pro.name",
    descriptionKey: "pricing:plans.pro.description",
    monthlyPriceCents: 4000,
    annualPriceCents: 40800,
    usePaddle: true,
    features: [
      { key: "pricing:features.teamMembersPro", included: true },
      { key: "pricing:features.rolesTeams", included: true },
      { key: "pricing:features.apiAccess", included: true },
      { key: "pricing:features.auditLogPro", included: true },
      { key: "pricing:features.support", included: true },
      { key: "pricing:features.sla", included: false },
    ],
  },
  {
    id: "enterprise",
    nameKey: "pricing:plans.enterprise.name",
    descriptionKey: "pricing:plans.enterprise.description",
    monthlyPriceCents: 0,
    annualPriceCents: 0,
    usePaddle: false,
    features: [
      { key: "pricing:features.teamMembersEnterprise", included: true },
      { key: "pricing:features.rolesTeams", included: true },
      { key: "pricing:features.apiAccess", included: true },
      { key: "pricing:features.auditLogEnterprise", included: true },
      { key: "pricing:features.support", included: true },
      { key: "pricing:features.sla", included: true },
    ],
  },
];

/** Key features table: row label + per-plan value (boolean = check/dot, string = display value). */
export const KEY_FEATURES_TABLE: {
  rowLabelKey: string;
  basic: boolean | string;
  pro: boolean | string;
  enterprise: boolean | string;
}[] = [
  {
    rowLabelKey: "pricing:features.teamMemberCount",
    basic: "pricing:features.count3",
    pro: "pricing:features.count10",
    enterprise: "pricing:features.countUnlimited",
  },
  {
    rowLabelKey: "pricing:features.customRoles",
    basic: false,
    pro: true,
    enterprise: true,
  },
  {
    rowLabelKey: "pricing:features.teams",
    basic: false,
    pro: true,
    enterprise: true,
  },
  {
    rowLabelKey: "pricing:features.apiKeys",
    basic: false,
    pro: true,
    enterprise: true,
  },
  {
    rowLabelKey: "pricing:features.auditLogRetention",
    basic: "pricing:features.retention90",
    pro: "pricing:features.retention365",
    enterprise: "pricing:features.retentionUnlimited",
  },
];

export const PADDLE_PRICE_IDS = {
  basicMonthly: import.meta.env.VITE_PADDLE_PRICE_BASIC_MONTHLY as string,
  basicAnnual: import.meta.env.VITE_PADDLE_PRICE_BASIC_ANNUAL as string,
  proMonthly: import.meta.env.VITE_PADDLE_PRICE_PRO_MONTHLY as string,
  proAnnual: import.meta.env.VITE_PADDLE_PRICE_PRO_ANNUAL as string,
} as const;

export const PADDLE_CLIENT_TOKEN = import.meta.env
  .VITE_PADDLE_CLIENT_TOKEN as string;

/** "sandbox" or "production". Defaults to sandbox when not set. */
export const PADDLE_ENVIRONMENT = (import.meta.env.VITE_PADDLE_ENVIRONMENT as "sandbox" | "production") || "sandbox";

export const SUBSCRIPTION_TYPE = {
  INACTIVE: 0,
  FREE: 0,
  BASIC: 1,
  PRO: 2,
  ENTERPRISE: 3,
} as const;

export const PAYMENT_INTERVAL = {
  MONTHLY: 0,
  ANNUAL: 1,
} as const;

export type SubscriptionTypeValue =
  (typeof SUBSCRIPTION_TYPE)[keyof typeof SUBSCRIPTION_TYPE];

export function getPlanIdBySubscriptionType(type: number): PlanId | null {
  switch (type) {
    case SUBSCRIPTION_TYPE.BASIC:
      return "basic";
    case SUBSCRIPTION_TYPE.PRO:
      return "pro";
    case SUBSCRIPTION_TYPE.ENTERPRISE:
      return "enterprise";
    default:
      // No plan for FREE/unknown types.
      return null;
  }
}

export function getPlanBySubscriptionType(
  type: number,
): PlanConfig | undefined {
  const planId = getPlanIdBySubscriptionType(type);
  if (!planId) return undefined;
  return PLANS.find((p) => p.id === planId);
}

export function getPaddlePriceId(
  planId: "basic" | "pro",
  interval: "monthly" | "annual",
): string | undefined {
  if (planId === "basic") {
    return interval === "monthly"
      ? PADDLE_PRICE_IDS.basicMonthly
      : PADDLE_PRICE_IDS.basicAnnual;
  }
  if (planId === "pro") {
    return interval === "monthly"
      ? PADDLE_PRICE_IDS.proMonthly
      : PADDLE_PRICE_IDS.proAnnual;
  }
  return undefined;
}

/** Resolve Paddle price ID to plan id and interval (for scheduled change display). */
export function getPlanIdAndIntervalFromPriceId(
  priceId: string | null | undefined,
): { planId: PlanId; interval: "monthly" | "annual" } | null {
  if (!priceId) return null;
  if (priceId === PADDLE_PRICE_IDS.basicMonthly) return { planId: "basic", interval: "monthly" };
  if (priceId === PADDLE_PRICE_IDS.basicAnnual) return { planId: "basic", interval: "annual" };
  if (priceId === PADDLE_PRICE_IDS.proMonthly) return { planId: "pro", interval: "monthly" };
  if (priceId === PADDLE_PRICE_IDS.proAnnual) return { planId: "pro", interval: "annual" };
  return null;
}

/** Tier order for upgrade/downgrade (matches backend). */
const PLAN_TIER_ORDER: Record<string, number> = {
  basic: 1,
  pro: 2,
  enterprise: 3,
};

/**
 * Whether the change will take effect at next billing period (downgrade or annual→monthly).
 * Matches backend logic in change_plan.go.
 */
export function getPlanChangeEffectiveAt(
  currentPlanId: PlanId | null,
  currentInterval: "monthly" | "annual",
  targetPlanId: "basic" | "pro",
  targetInterval: "monthly" | "annual",
): "immediately" | "next_billing_period" {
  if (!currentPlanId || currentPlanId === "enterprise") return "immediately";
  const currentTier = PLAN_TIER_ORDER[currentPlanId] ?? 0;
  const targetTier = PLAN_TIER_ORDER[targetPlanId] ?? 0;
  const isDowngrade =
    targetTier < currentTier ||
    (targetTier === currentTier &&
      targetInterval === "monthly" &&
      currentInterval === "annual");
  return isDowngrade ? "next_billing_period" : "immediately";
}
