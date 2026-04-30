import { Fragment, useMemo, useState } from "react";
import { Navigate, useParams } from "react-router";
import {
  getSubscriptionCapabilities,
  useTranslationNamespace,
} from "@/helpers";
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  Tooltip,
  type DateRange,
} from "@yca-software/design-system";
import { ChevronDown, Loader2 } from "lucide-react";
import { useGetOrganizationQuery, useListAuditLogsQuery } from "@/api";
import type { AuditLog } from "@/types";
import {
  getDefaultLast7DaysRange,
  toEndOfDay,
  toStartOfDay,
} from "@/helpers/date";
import { getAuditRetentionDays } from "@/helpers/billing/subscriptionCapabilities";
import { usePricingModalStore } from "@/states";
import { useShallow } from "zustand/shallow";
import { DateRangeFilter } from "@/components";
import {
  SETTINGS_SECTIONS,
  useSettingsPermissions,
} from "../useSettingsPermissions";

const PAGE_SIZE = 20;

function formatDateParts(iso: string): { date: string; time: string } {
  if (!iso) return { date: "—", time: "—" };
  const parsed = new Date(iso);
  if (isNaN(parsed.getTime())) return { date: "—", time: "—" };
  return {
    date: parsed.toLocaleDateString(),
    time: parsed.toLocaleTimeString(),
  };
}

function truncateId(id: string | undefined): string {
  if (!id) return "—";
  if (id.length <= 12) return id;
  return `${id.slice(0, 8)}…`;
}

function getResourceLabel(log: AuditLog): string {
  if (log.resourceName && log.resourceName.trim()) return log.resourceName;
  return truncateId(log.resourceId);
}

function getAuditLogJSON(log: AuditLog): string {
  return JSON.stringify(log, null, 2);
}

export const AuditLogSettings = () => {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const { orgId } = useParams<{ orgId: string }>();
  const { openForOrg } = usePricingModalStore(
    useShallow((state) => ({
      openForOrg: state.openForOrg,
    })),
  );
  const permissions = useSettingsPermissions();
  const currentOrgId = orgId || "";

  const [draftRange, setDraftRange] = useState<DateRange | undefined>(
    getDefaultLast7DaysRange,
  );
  const [appliedRange, setAppliedRange] = useState<DateRange | undefined>(
    getDefaultLast7DaysRange,
  );
  const [offset, setOffset] = useState(0);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const { data: organization, isLoading: isOrganizationLoading } =
    useGetOrganizationQuery(currentOrgId);
  const capabilities = getSubscriptionCapabilities(organization, null);
  const retentionDays = getAuditRetentionDays(organization?.subscriptionType);
  const maxDate = useMemo(() => {
    const date = new Date();
    date.setHours(0, 0, 0, 0);
    return date;
  }, []);
  const minDate = useMemo(() => {
    const date = new Date(maxDate);
    date.setDate(date.getDate() - (retentionDays - 1));
    return date;
  }, [maxDate, retentionDays]);

  const queryParams = useMemo(() => {
    const startDate = appliedRange?.from
      ? toStartOfDay(appliedRange.from)?.toISOString()
      : undefined;
    const endDate = appliedRange?.to
      ? toEndOfDay(appliedRange.to)?.toISOString()
      : undefined;

    return {
      orgId: currentOrgId,
      limit: PAGE_SIZE,
      offset,
      startDate,
      endDate,
    };
  }, [appliedRange, currentOrgId, offset]);

  const { data, isLoading, isError, error } = useListAuditLogsQuery(
    queryParams,
    {
      enabled: !!orgId && !!organization && capabilities.canViewAudit,
    },
  );

  if (!permissions.audit?.canRead) {
    const firstAllowed = SETTINGS_SECTIONS.find(
      (section) => permissions[section.permissionKey]?.canRead,
    );
    return firstAllowed ? (
      <Navigate to={`/settings/${orgId}/${firstAllowed.path}`} replace />
    ) : (
      <Navigate to={`/settings/${orgId}`} replace />
    );
  }

  if (isOrganizationLoading && !organization) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!capabilities.canViewAudit) {
    return (
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>{t("settings:org.auditLog.title")}</CardTitle>
            <CardDescription>
              {t("settings:org.auditLog.description")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="mb-4 rounded-lg border border-dashed border-primary/40 bg-primary/5 p-4 text-left">
              <h3 className="text-sm font-semibold">
                {t("settings:org.upsell.auditTitle")}
              </h3>
              <p className="mt-1 text-xs text-muted-foreground">
                {t("settings:org.upsell.auditDescription")}
              </p>
              <Button
                variant="outline"
                size="sm"
                className="mt-3"
                onClick={() => openForOrg(currentOrgId)}
              >
                {t("settings:org.upsell.viewPlans")}
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  const items = data?.items ?? [];
  const hasNext = data?.hasNext ?? false;
  const hasPrevious = offset > 0;
  const isFeatureUnavailable =
    isError &&
    (error as { error?: { code?: string } })?.error?.code ===
      "FEATURE_NOT_AVAILABLE";

  const toggleExpanded = (id: string) => {
    setExpandedId((currentId) => (currentId === id ? null : id));
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>{t("settings:org.auditLog.title")}</CardTitle>
          <CardDescription>
            {t("settings:org.auditLog.description")}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {isFeatureUnavailable ? (
            <p className="text-sm text-muted-foreground">
              {t("common:notAvailable")}
            </p>
          ) : (
            <>
              <DateRangeFilter
                label={t("settings:org.auditLog.timeRange")}
                value={draftRange}
                onChange={setDraftRange}
                minDate={minDate}
                maxDate={maxDate}
                onApply={(nextRange) => {
                  setAppliedRange(nextRange);
                  setDraftRange(nextRange);
                  setOffset(0);
                  setExpandedId(null);
                }}
              />

              <div className="space-y-3">
                <Table className="rounded-md border">
                  <TableHeader>
                    <TableRow className="bg-muted/50">
                      <TableHead
                        className="w-8 px-4 py-3"
                        aria-label={t("settings:org.auditLog.details")}
                      />
                      <TableHead className="px-4 py-3">
                        {t("settings:org.auditLog.date")}
                      </TableHead>
                      <TableHead className="px-4 py-3">
                        {t("settings:org.auditLog.actor")}
                      </TableHead>
                      <TableHead className="px-4 py-3">
                        {t("settings:org.auditLog.action")}
                      </TableHead>
                      <TableHead className="px-4 py-3">
                        {t("settings:org.auditLog.resource")}
                      </TableHead>
                      <TableHead className="px-4 py-3">
                        {t("settings:org.auditLog.resourceName")}
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {isLoading ? (
                      <TableRow>
                        <TableCell
                          colSpan={6}
                          className="px-4 py-10 text-center"
                        >
                          <div className="inline-flex items-center gap-2 text-muted-foreground">
                            <Loader2 className="h-4 w-4 animate-spin" />
                            {t("common:loading")}
                          </div>
                        </TableCell>
                      </TableRow>
                    ) : items.length === 0 ? (
                      <TableRow>
                        <TableCell
                          colSpan={6}
                          className="px-4 py-10 text-center text-muted-foreground"
                        >
                          {t("settings:org.auditLog.noLogs")}
                        </TableCell>
                      </TableRow>
                    ) : (
                      items.map((log) => {
                        const createdAt = formatDateParts(log.createdAt);
                        return (
                          <Fragment key={log.id}>
                            <TableRow
                              role="button"
                              tabIndex={0}
                              className="cursor-pointer border-b last:border-0 hover:bg-muted/50 focus:bg-muted/50 focus:outline-none"
                              onClick={() => toggleExpanded(log.id)}
                              onKeyDown={(event) => {
                                if (
                                  event.key === "Enter" ||
                                  event.key === " "
                                ) {
                                  event.preventDefault();
                                  toggleExpanded(log.id);
                                }
                              }}
                            >
                              <TableCell className="w-8 px-4 py-3">
                                <ChevronDown
                                  className={`h-4 w-4 text-muted-foreground ${
                                    expandedId === log.id ? "" : "-rotate-90"
                                  }`}
                                />
                              </TableCell>
                              <TableCell className="whitespace-nowrap px-4 py-3">
                                <div className="leading-tight">
                                  <p>{createdAt.date}</p>
                                  <p className="text-xs text-muted-foreground">
                                    {createdAt.time}
                                  </p>
                                </div>
                              </TableCell>
                              <TableCell className="w-[260px] min-w-[260px] px-4 py-3">
                                <Tooltip content={log.actorInfo || "—"}>
                                  <p className="truncate">
                                    {log.actorInfo || "—"}
                                  </p>
                                </Tooltip>
                              </TableCell>
                              <TableCell className="px-4 py-3">
                                {t(
                                  `settings:org.auditLog.actions.${log.action}`,
                                  {
                                    defaultValue: log.action,
                                  },
                                )}
                              </TableCell>
                              <TableCell className="px-4 py-3">
                                {t(
                                  `settings:org.auditLog.resources.${log.resourceType}`,
                                  {
                                    defaultValue: log.resourceType,
                                  },
                                )}
                              </TableCell>
                              <TableCell className="px-4 py-3">
                                <Tooltip content={getResourceLabel(log)}>
                                  <span className="block truncate text-foreground">
                                    {getResourceLabel(log)}
                                  </span>
                                </Tooltip>
                                {log.resourceName && log.resourceId ? (
                                  <Tooltip content={log.resourceId}>
                                    <span className="mt-0.5 block truncate font-mono text-xs text-muted-foreground">
                                      {truncateId(log.resourceId)}
                                    </span>
                                  </Tooltip>
                                ) : null}
                              </TableCell>
                            </TableRow>
                            {expandedId === log.id && (
                              <TableRow className="border-b bg-muted/30 last:border-0">
                                <TableCell colSpan={6} className="px-4 py-3">
                                  <div className="rounded-md border bg-background p-3">
                                    <p className="mb-2 text-xs font-medium text-muted-foreground">
                                      {t("settings:org.auditLog.details")}
                                    </p>
                                    <pre className="wrap-break-word overflow-x-auto whitespace-pre-wrap text-xs text-muted-foreground">
                                      {getAuditLogJSON(log)}
                                    </pre>
                                  </div>
                                </TableCell>
                              </TableRow>
                            )}
                          </Fragment>
                        );
                      })
                    )}
                  </TableBody>
                </Table>

                <div className="flex items-center justify-between">
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={!hasPrevious || isLoading}
                    onClick={() => {
                      setOffset((currentOffset) =>
                        Math.max(0, currentOffset - PAGE_SIZE),
                      );
                      setExpandedId(null);
                    }}
                  >
                    {t("settings:org.auditLog.previous")}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={!hasNext || isLoading}
                    onClick={() => {
                      setOffset((currentOffset) => currentOffset + PAGE_SIZE);
                      setExpandedId(null);
                    }}
                  >
                    {t("settings:org.auditLog.next")}
                  </Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
};
