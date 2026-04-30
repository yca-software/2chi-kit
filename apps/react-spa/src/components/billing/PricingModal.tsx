import { useEffect, useState } from "react";
import { toast } from "sonner";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  Button,
  Card,
  CardContent,
  CardHeader,
  Dialog,
  DialogContent,
  DialogTitle,
} from "@yca-software/design-system";
import { Check, Loader2 } from "lucide-react";
import type { Organization } from "@/types";
import {
  getPaddlePriceId,
  getPlanChangeEffectiveAt,
  getPlanIdBySubscriptionType,
  PAYMENT_INTERVAL,
  PLANS,
  KEY_FEATURES_TABLE,
  TRIAL_DAYS,
  ANNUAL_DISCOUNT_PERCENT,
  type PlanId,
} from "@/constants";
import {
  useCreateCheckoutSessionMutation,
  useProcessTransactionMutation,
  useChangePlanMutation,
} from "@/api";
import { useTranslationNamespace } from "@/helpers";
import { usePaddle } from "./usePaddle";
import { useUserState } from "@/states";
import { EnterpriseContactForm } from "./EnterpriseContactForm";

// --- PricingContent (embeddable plan picker; used in modal and onboarding) ---

export interface PricingContentProps {
  organizationId?: string;
  organization?: {
    name: string;
    subscriptionType: number;
    subscriptionPaymentInterval: number;
  } | null;
  currentPlanId?: PlanId | null;
  currentInterval?: "monthly" | "annual";
  onComplete?: () => void;
  onPlanSelect?: (
    planId: "basic" | "pro",
    interval: "monthly" | "annual",
  ) => Promise<void> | void;
  onContactUsClick?: () => void;
  isChangingPlan?: boolean;
  /** When true, only allow upgrades (e.g. from paywall). Downgrade options are disabled. */
  upgradeOnly?: boolean;
  /** Scroll this tier into view (e.g. marketing deep link). */
  initialPlanFocus?: "basic" | "pro" | null;
}

function formatPrice(cents: number): string {
  return new Intl.NumberFormat(undefined, {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(cents / 100);
}

const PLAN_TIER_ORDER: Record<string, number> = {
  basic: 1,
  pro: 2,
  enterprise: 3,
};

/** True if selection is a tier downgrade (e.g. Pro→Basic). Same-tier annual→monthly is not a downgrade. */
function isTierDowngrade(
  planId: string,
  currentPlanId: PlanId | null,
): boolean {
  if (!currentPlanId || currentPlanId === "enterprise") return false;
  const currentTier = PLAN_TIER_ORDER[currentPlanId] ?? 0;
  const targetTier = PLAN_TIER_ORDER[planId] ?? 0;
  return targetTier < currentTier;
}

export function PricingContent({
  organizationId,
  organization,
  currentPlanId,
  currentInterval = "monthly",
  onComplete,
  onPlanSelect,
  onContactUsClick,
  isChangingPlan = false,
  upgradeOnly = false,
  initialPlanFocus = null,
}: PricingContentProps) {
  const { t } = useTranslationNamespace(["pricing", "common"]);
  const userEmail = useUserState((s) => s.userData.user?.email ?? "");
  const { ready: paddleReady, openCheckoutWithItems } = usePaddle();
  const [enterpriseFormOpen, setEnterpriseFormOpen] = useState(false);
  const [loadingPriceId, setLoadingPriceId] = useState<string | null>(null);
  const [billingInterval, setBillingInterval] = useState<"monthly" | "annual">(
    "monthly",
  );

  useEffect(() => {
    if (!initialPlanFocus) return;
    const el = document.getElementById(`pricing-plan-${initialPlanFocus}`);
    el?.scrollIntoView({ block: "center", behavior: "smooth" });
  }, [initialPlanFocus]);

  const handlePaddleOpen = (planId: "basic" | "pro") => {
    const priceId = getPaddlePriceId(planId, billingInterval);
    if (!priceId) return;
    setLoadingPriceId(planId);
    openCheckoutWithItems(priceId, userEmail);
    setLoadingPriceId(null);
  };

  const handlePlanClick = async (planId: "basic" | "pro") => {
    if (onPlanSelect) {
      setLoadingPriceId(planId);
      try {
        await onPlanSelect(planId, billingInterval);
      } finally {
        setLoadingPriceId(null);
      }
      return;
    }
    handlePaddleOpen(planId);
  };

  const isCurrentPlan = (planId: string) =>
    currentPlanId != null &&
    planId === currentPlanId &&
    (planId === "enterprise" || billingInterval === currentInterval);

  const getPlanButtonLabel = (planId: string) => {
    if (isCurrentPlan(planId)) return t("pricing:currentPlan");
    if (upgradeOnly && isTierDowngrade(planId, currentPlanId ?? null))
      return t("pricing:downgradeNotAvailable");
    // Same plan, different interval (e.g. Basic monthly → Basic annual)
    if (
      currentPlanId != null &&
      planId === currentPlanId &&
      planId !== "enterprise" &&
      billingInterval !== currentInterval
    ) {
      return billingInterval === "annual"
        ? t("pricing:changeToAnnual")
        : t("pricing:changeToMonthly");
    }
    if (
      currentPlanId != null &&
      organization?.subscriptionType != null &&
      organization.subscriptionType !== 0
    ) {
      const currentTier = PLAN_TIER_ORDER[currentPlanId] ?? 0;
      const targetTier = PLAN_TIER_ORDER[planId] ?? 0;
      if (targetTier > currentTier) return t("pricing:upgradePlan");
      if (targetTier < currentTier) return t("pricing:downgradePlan");
    }
    return t("pricing:selectPlan");
  };

  return (
    <div className="space-y-8">
      <div className="space-y-4 text-center">
        <div className="space-y-1">
          <h2 className="text-3xl font-semibold tracking-tight">
            {t("pricing:title")}
          </h2>
          <p className="text-muted-foreground mx-auto max-w-xl text-sm">
            {t("pricing:subtitle")}
          </p>
        </div>

        <div className="flex flex-col items-center gap-2">
          <div className="inline-flex items-center gap-3 rounded-full border bg-background px-2 py-1 text-xs sm:text-sm">
            <button
              type="button"
              className={`cursor-pointer rounded-full px-3 py-1 font-medium transition-colors ${
                billingInterval === "monthly"
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground"
              }`}
              onClick={() => setBillingInterval("monthly")}
            >
              {t("pricing:monthlyBilling")}
            </button>
            <span className="h-4 w-px bg-border" aria-hidden />
            <button
              type="button"
              className={`cursor-pointer rounded-full px-3 py-1 font-medium transition-colors ${
                billingInterval === "annual"
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground"
              }`}
              onClick={() => setBillingInterval("annual")}
            >
              {t("pricing:annualBilling")}
            </button>
          </div>
          <div className="rounded-full border border-primary/40 bg-primary/10 px-3 py-1.5 text-sm font-semibold text-primary">
            {t("pricing:annualBenefit", { percent: ANNUAL_DISCOUNT_PERCENT })}
          </div>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        {PLANS.map((plan) => {
          const current = isCurrentPlan(plan.id);
          const paddlePlan = plan.id === "basic" || plan.id === "pro";
          return (
            <Card
              key={plan.id}
              id={`pricing-plan-${plan.id}`}
              className={`flex flex-col ${current ? "ring-2 ring-primary" : ""}`}
            >
              <CardHeader className="pb-2">
                <h3 className="text-2xl font-bold tracking-tight">
                  {t(plan.nameKey)}
                </h3>
                <div className="min-h-6 flex flex-wrap items-center gap-2">
                  {current && (
                    <span className="inline-block rounded-full bg-primary px-2 py-0.5 text-xs font-medium text-primary-foreground">
                      {t("pricing:currentPlan")}
                    </span>
                  )}
                  {TRIAL_DAYS > 0 && plan.usePaddle && !current && (
                    <span className="inline-block rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                      {t("pricing:trialBadge", { days: TRIAL_DAYS })}
                    </span>
                  )}
                </div>
              </CardHeader>
              <CardContent className="flex flex-1 flex-col gap-4 pt-0">
                {plan.usePaddle ? (
                  <>
                    <div className="flex items-baseline gap-1">
                      <span className="text-2xl font-bold">
                        {billingInterval === "annual"
                          ? formatPrice(plan.annualPriceCents)
                          : formatPrice(plan.monthlyPriceCents)}
                      </span>
                      <span className="text-muted-foreground text-sm">
                        {billingInterval === "annual"
                          ? t("pricing:perYear")
                          : t("pricing:perMonth")}
                      </span>
                    </div>
                    <ul className="space-y-2 text-sm">
                      {plan.features.map((f) => (
                        <li key={f.key} className="flex items-center gap-2">
                          <Check className="h-4 w-4 shrink-0 text-primary" />
                          {t(f.key)}
                        </li>
                      ))}
                    </ul>
                    <Button
                      className="mt-auto"
                      disabled={
                        current ||
                        !!loadingPriceId ||
                        isChangingPlan ||
                        (!onPlanSelect && !paddleReady) ||
                        (upgradeOnly &&
                          isTierDowngrade(plan.id, currentPlanId ?? null))
                      }
                      onClick={() =>
                        paddlePlan &&
                        handlePlanClick(plan.id as "basic" | "pro")
                      }
                    >
                      {loadingPriceId === plan.id ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        getPlanButtonLabel(plan.id)
                      )}
                    </Button>
                  </>
                ) : (
                  <>
                    <div className="text-2xl font-bold">
                      {t("pricing:contactUs")}
                    </div>
                    <ul className="space-y-2 text-sm">
                      {plan.features.map((f) => (
                        <li key={f.key} className="flex items-center gap-2">
                          <Check className="h-4 w-4 shrink-0 text-primary" />
                          {t(f.key)}
                        </li>
                      ))}
                    </ul>
                    <Button
                      className="mt-auto"
                      variant="outline"
                      onClick={() =>
                        onContactUsClick
                          ? onContactUsClick()
                          : setEnterpriseFormOpen(true)
                      }
                    >
                      {t("pricing:contactUs")}
                    </Button>
                  </>
                )}
              </CardContent>
            </Card>
          );
        })}
      </div>

      <div className="mb-10 overflow-x-auto sm:mb-12">
        <div className="min-w-[640px] rounded-lg border">
          <div className="grid grid-cols-[minmax(0,2fr)_repeat(3,minmax(0,1fr))] border-b bg-muted/40 text-sm font-medium">
            <div className="px-4 py-3 text-left">
              {t("pricing:keyFeatures")}
            </div>
            {PLANS.map((plan) => (
              <div key={plan.id} className="px-4 py-3 text-center">
                {t(plan.nameKey)}
              </div>
            ))}
          </div>
          {KEY_FEATURES_TABLE.map((row) => (
            <div
              key={row.rowLabelKey}
              className="grid grid-cols-[minmax(0,2fr)_repeat(3,minmax(0,1fr))] border-t text-sm"
            >
              <div className="px-4 py-3">{t(row.rowLabelKey)}</div>
              {([row.basic, row.pro, row.enterprise] as const).map(
                (value, i) => (
                  <div
                    key={i}
                    className="flex items-center justify-center px-4 py-3"
                  >
                    {typeof value === "string" ? (
                      <span className="text-muted-foreground">{t(value)}</span>
                    ) : value ? (
                      <Check className="h-4 w-4 text-primary" />
                    ) : (
                      <span className="inline-block h-1.5 w-1.5 rounded-full bg-muted-foreground/30" />
                    )}
                  </div>
                ),
              )}
            </div>
          ))}
        </div>
      </div>

      {!onContactUsClick && (
        <EnterpriseContactForm
          open={enterpriseFormOpen}
          onOpenChange={setEnterpriseFormOpen}
          organizationId={organizationId}
          organizationName={organization?.name}
          onSuccess={onComplete}
        />
      )}
    </div>
  );
}

// --- PricingModal (dialog wrapper for change plan / subscribe) ---

interface PricingModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  organization: Organization | null | undefined;
  onContactUsClick?: () => void;
  /** When true, only upgrades are allowed (e.g. from paywall). Downgrade options are disabled. */
  upgradeOnly?: boolean;
  initialCheckoutIntent?: string | null;
  initialPriceId?: string | null;
}

type PendingPlanChange = {
  planId: "basic" | "pro";
  interval: "monthly" | "annual";
};

function planFocusFromCheckoutIntent(
  intent: string | null | undefined,
): "basic" | "pro" | null {
  if (!intent) return null;
  const x = intent.trim().toLowerCase();
  if (x === "starter") return "basic";
  if (x === "basic" || x === "pro") return x;
  return null;
}

export function PricingModal({
  open,
  onOpenChange,
  organization,
  onContactUsClick,
  upgradeOnly = false,
  initialCheckoutIntent = null,
  initialPriceId: _initialPriceId = null,
}: PricingModalProps) {
  const { t } = useTranslationNamespace(["pricing", "common"]);
  const [pendingChange, setPendingChange] = useState<PendingPlanChange | null>(
    null,
  );
  const processTransactionMutation = useProcessTransactionMutation(
    organization?.id ?? "",
  );
  const { openCheckoutWithTransactionId } = usePaddle({
    onCheckoutCompleted: (data, context) => {
      if (context) {
        processTransactionMutation.mutate({
          transactionId: data.transaction_id,
          priceId: context.priceId,
        });
      }
    },
  });
  const checkoutMutation = useCreateCheckoutSessionMutation(
    organization?.id ?? "",
    {
      onError: (err) => {
        toast.error(err.error?.message ?? t("common:defaultError"));
      },
    },
  );
  const changePlanMutation = useChangePlanMutation(organization?.id ?? "", {
    onSuccess: () => {
      setPendingChange(null);
      onOpenChange(false);
    },
    onError: (err) => {
      toast.error(err.error?.message ?? t("common:defaultError"));
    },
  });

  if (!organization) return null;

  const hasExistingSubscription = Boolean(organization.paddleSubscriptionId);
  const currentPlanId = getPlanIdBySubscriptionType(
    organization.subscriptionType,
  );
  const currentInterval: "monthly" | "annual" =
    organization.subscriptionPaymentInterval === PAYMENT_INTERVAL.ANNUAL
      ? "annual"
      : "monthly";

  const handlePlanSelect = async (
    planId: "basic" | "pro",
    interval: "monthly" | "annual",
  ) => {
    const priceId = getPaddlePriceId(planId, interval);
    if (!priceId) {
      toast.error(t("common:defaultError"));
      return;
    }
    if (hasExistingSubscription) {
      setPendingChange({ planId, interval });
      return;
    }
    try {
      const data = await checkoutMutation.mutateAsync({ planId: priceId });
      if (!data.transactionId) {
        toast.error(t("common:defaultError"));
        return;
      }
      onOpenChange(false);
      setTimeout(() => {
        openCheckoutWithTransactionId(data.transactionId, {
          organizationId: organization.id,
          priceId,
        });
      }, 0);
    } catch {
      // onError already toasts
    }
  };

  const effectiveAt =
    pendingChange &&
    getPlanChangeEffectiveAt(
      currentPlanId,
      currentInterval,
      pendingChange.planId,
      pendingChange.interval,
    );
  const currentPlanLabel =
    currentPlanId != null
      ? `${t(`pricing:plans.${currentPlanId}.name`)} ${t(currentInterval === "annual" ? "pricing:annualBilling" : "pricing:monthlyBilling")}`
      : "";
  const newPlanLabel =
    pendingChange != null
      ? `${t(`pricing:plans.${pendingChange.planId}.name`)} ${t(pendingChange.interval === "annual" ? "pricing:annualBilling" : "pricing:monthlyBilling")}`
      : "";
  const billingEndDate =
    organization.subscriptionExpiresAt != null
      ? new Date(organization.subscriptionExpiresAt).toLocaleDateString(
          undefined,
          { dateStyle: "long" },
        )
      : "";

  const confirmDescription =
    pendingChange != null && effectiveAt === "next_billing_period"
      ? t("pricing:confirmPlanChangeDescNextBilling", {
          currentPlan: currentPlanLabel,
          newPlan: newPlanLabel,
          date: billingEndDate,
        })
      : pendingChange != null
        ? t("pricing:confirmPlanChangeDescImmediate", {
            currentPlan: currentPlanLabel,
            newPlan: newPlanLabel,
          })
        : "";

  const handleConfirmPlanChange = () => {
    if (pendingChange == null) return;
    const priceId = getPaddlePriceId(
      pendingChange.planId,
      pendingChange.interval,
    );
    if (!priceId) return;
    changePlanMutation.mutate({ planId: priceId });
  };

  return (
    <>
      <AlertDialog
        open={pendingChange != null}
        onOpenChange={(open: boolean) => !open && setPendingChange(null)}
      >
        <AlertDialogContent
          className="z-[210] max-w-md"
          overlayClassName="z-[210]"
          onOverlayClick={() => setPendingChange(null)}
        >
          <AlertDialogHeader className="gap-2">
            <AlertDialogTitle className="text-base sm:text-lg">
              {t("pricing:confirmPlanChangeTitle")}
            </AlertDialogTitle>
            <AlertDialogDescription className="text-sm">
              {confirmDescription}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter className="flex flex-col-reverse gap-2 pt-2 sm:flex-row sm:justify-end sm:gap-2 sm:pt-0">
            <AlertDialogCancel
              onClick={() => setPendingChange(null)}
              disabled={changePlanMutation.isPending}
              className="m-0 w-full sm:w-auto"
            >
              {t("common:cancel")}
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirmPlanChange}
              disabled={changePlanMutation.isPending}
              className="m-0 w-full sm:w-auto"
            >
              {changePlanMutation.isPending && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden />
              )}
              {changePlanMutation.isPending
                ? t("common:loading")
                : t("pricing:confirmPlanChangeConfirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent
          className="fixed inset-0 z-[200] max-h-screen w-screen max-w-none translate-x-0 translate-y-0 overflow-y-auto border-0 p-0 sm:max-w-none"
          overlayClassName="z-[200]"
          showCloseButton
        >
          <DialogTitle className="sr-only">Choose your plan</DialogTitle>
          <div className="mx-auto flex min-h-screen max-w-6xl flex-col px-4 py-8 pb-16 sm:py-12 sm:pb-20">
            <PricingContent
              organizationId={organization.id}
              organization={organization}
              currentPlanId={currentPlanId}
              currentInterval={currentInterval}
              onPlanSelect={handlePlanSelect}
              onContactUsClick={onContactUsClick}
              isChangingPlan={changePlanMutation.isPending}
              upgradeOnly={upgradeOnly}
              initialPlanFocus={planFocusFromCheckoutIntent(
                initialCheckoutIntent,
              )}
            />
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
