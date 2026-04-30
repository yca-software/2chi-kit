import { useState } from "react";
import {
  useTranslationNamespace,
  isAllowedPortalUrl,
} from "@/helpers";
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@yca-software/design-system";
import { Check, CreditCard } from "lucide-react";
import { toast } from "sonner";
import type { Organization } from "@/types";
import {
  getPlanBySubscriptionType,
  getPlanIdAndIntervalFromPriceId,
  PAYMENT_INTERVAL,
  SUBSCRIPTION_TYPE,
} from "@/constants";
import {
  useCreateCustomerPortalSessionMutation,
  useChangePlanMutation,
} from "@/api";
import { EnterpriseContactForm, PricingModal } from "@/components";

const UNLIMITED_SEATS = -1;

interface SubscriptionCardProps {
  organization: Organization | null | undefined;
  memberCount: number;
}

export function SubscriptionCard({
  organization,
  memberCount,
}: SubscriptionCardProps) {
  const { t } = useTranslationNamespace(["settings", "pricing"]);
  const [pricingModalOpen, setPricingModalOpen] = useState(false);
  const [enterpriseFormOpen, setEnterpriseFormOpen] = useState(false);

  const portalMutation = useCreateCustomerPortalSessionMutation(
    organization?.id ?? "",
    {
      onSuccess: (data) => {
        if (data.portalUrl && isAllowedPortalUrl(data.portalUrl)) {
          window.location.href = data.portalUrl;
        }
      },
    },
  );
  const changePlanMutation = useChangePlanMutation(organization?.id ?? "", {
    onError: (err) => {
      toast.error(err.error?.message ?? t("common:defaultError"));
    },
    onSuccess: (data) => {
      if (data.effectiveAt === "next_billing_period") {
        toast.success(t("settings:org.subscription.planChangeScheduled"));
      }
    },
  });

  const handleOpenPortal = () => {
    if (!organization?.id) return;
    portalMutation.mutate();
  };

  if (!organization) return null;

  const plan = getPlanBySubscriptionType(organization.subscriptionType);
  const isCustom = organization.customSubscription;
  const totalSeats = organization.subscriptionSeats;
  const isUnlimitedSeats = totalSeats === UNLIMITED_SEATS;
  const usedSeats = memberCount;
  const availableSeats = isUnlimitedSeats
    ? null
    : Math.max(0, totalSeats - usedSeats);

  const expiresAt = organization.subscriptionExpiresAt
    ? new Date(organization.subscriptionExpiresAt)
    : null;
  const now = new Date();
  const hasPaidPlan =
    organization.subscriptionType !== SUBSCRIPTION_TYPE.FREE || isCustom;
  const isExpired = expiresAt !== null && expiresAt.getTime() < now.getTime();
  const showNoActiveSubscription = hasPaidPlan && isExpired;
  const isFreePlan =
    organization.subscriptionType === SUBSCRIPTION_TYPE.FREE && !isCustom;

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>{t("settings:org.subscription.title")}</CardTitle>
          <CardDescription>
            {t("settings:org.subscription.description")}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center gap-4">
            <div className="bg-muted flex h-12 w-12 items-center justify-center rounded-lg">
              <CreditCard className="h-6 w-6" />
            </div>
            <div className="flex-1">
              <div className="flex items-center gap-2 flex-wrap">
                <p className="font-medium">
                  {isFreePlan
                    ? t("settings:org.subscription.noPlanLabel")
                    : plan
                      ? t(plan.nameKey)
                      : t("settings:org.subscription.paid")}
                </p>
                {organization.subscriptionInTrial && (
                  <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                    {t("settings:org.subscription.trialBadge")}
                  </span>
                )}
                {isCustom && (
                  <span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                    {t("settings:org.subscription.customBadge")}
                  </span>
                )}
              </div>
              <p className="text-sm text-muted-foreground">
                {organization.subscriptionExpiresAt
                  ? t(
                      organization.subscriptionInTrial
                        ? "settings:org.subscription.billingStartsOn"
                        : "settings:org.subscription.renewsOn",
                      {
                        date: new Date(
                          organization.subscriptionExpiresAt,
                        ).toLocaleDateString(),
                      },
                    )
                  : t("settings:org.subscription.noExpiration")}
              </p>
              {showNoActiveSubscription && (
                <p className="text-xs font-medium text-destructive">
                  {t("settings:org.subscription.noActiveSubscription")}
                </p>
              )}
              {(organization.subscriptionType !== 0 ||
                organization.customSubscription) && (
                <p className="text-sm text-muted-foreground">
                  {(organization.subscriptionPaymentInterval ?? 0) ===
                  PAYMENT_INTERVAL.ANNUAL
                    ? t("settings:org.subscription.paymentIntervalAnnual")
                    : t("settings:org.subscription.paymentIntervalMonthly")}
                </p>
              )}
              {organization.scheduledPlanPriceId &&
                organization.subscriptionExpiresAt &&
                (() => {
                  const scheduled = getPlanIdAndIntervalFromPriceId(
                    organization.scheduledPlanPriceId,
                  );
                  const planLabel = scheduled
                    ? `${t(`pricing:plans.${scheduled.planId}.name`)} (${t(scheduled.interval === "annual" ? "pricing:annualBilling" : "pricing:monthlyBilling")})`
                    : t("settings:org.subscription.yourNewPlan");
                  return (
                    <p className="text-sm font-medium text-amber-600 dark:text-amber-500">
                      {t("settings:org.subscription.changeScheduledOn", {
                        plan: planLabel,
                        date: new Date(
                          organization.subscriptionExpiresAt,
                        ).toLocaleDateString(undefined, { dateStyle: "long" }),
                      })}
                    </p>
                  );
                })()}
            </div>
          </div>

          {!isFreePlan && (
            <div>
              <p className="text-sm font-medium text-muted-foreground">
                {t("settings:org.subscription.seats")}
              </p>
              <p className="text-lg font-semibold">
                {isUnlimitedSeats
                  ? t("settings:org.subscription.seatsSummaryUnlimited", {
                      used: usedSeats,
                    })
                  : t("settings:org.subscription.seatsSummary", {
                      used: usedSeats,
                      total: totalSeats,
                      available: availableSeats,
                    })}
              </p>
            </div>
          )}

          <div>
            {isFreePlan && (
              <div className="rounded-lg border border-dashed border-primary/40 bg-primary/5 p-4">
                <p className="text-sm font-medium">
                  {t("settings:org.subscription.noSubscriptionMessage")}
                </p>
                <Button
                  size="sm"
                  className="mt-3"
                  onClick={() => setPricingModalOpen(true)}
                >
                  {t("settings:org.subscription.subscribeCta")}
                </Button>
              </div>
            )}
          </div>

          {plan && (
            <div>
              <p className="text-sm font-medium mb-2">
                {t("settings:org.subscription.yourPlanIncludes")}
              </p>
              <ul className="space-y-1.5 text-sm">
                {plan.features.map((f) => (
                  <li key={f.key} className="flex items-center gap-2">
                    <Check className="h-4 w-4 shrink-0 text-primary" />
                    {t(f.key)}
                  </li>
                ))}
              </ul>
            </div>
          )}

          <div className="flex flex-wrap items-center gap-2 border-t pt-4">
            {!isCustom && !isFreePlan && (
              <Button
                size="sm"
                disabled={changePlanMutation.isPending}
                onClick={() => setPricingModalOpen(true)}
              >
                {changePlanMutation.isPending
                  ? t("common:loading")
                  : t("settings:org.subscription.changePlan")}
              </Button>
            )}
            {organization.paddleCustomerId && (
              <Button
                variant="outline"
                size="sm"
                onClick={handleOpenPortal}
                disabled={portalMutation.isPending}
              >
                {portalMutation.isPending
                  ? t("common:loading")
                  : t("settings:org.subscription.manageBilling")}
              </Button>
            )}
            {isCustom && (
              <p className="text-sm text-muted-foreground">
                {t("settings:org.subscription.contactCsNote")}
              </p>
            )}
          </div>
        </CardContent>
      </Card>

      <PricingModal
        open={pricingModalOpen}
        onOpenChange={setPricingModalOpen}
        organization={organization}
        upgradeOnly={false}
        onContactUsClick={() => {
          setPricingModalOpen(false);
          setEnterpriseFormOpen(true);
        }}
      />
      <EnterpriseContactForm
        open={enterpriseFormOpen}
        onOpenChange={setEnterpriseFormOpen}
        organizationId={organization.id}
      />
    </div>
  );
}
