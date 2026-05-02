import { useState } from "react";
import { useMatch, useNavigate, useParams } from "react-router";
import { Helmet } from "react-helmet-async";
import { useTranslationNamespace } from "@/helpers";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Button,
  ConfirmDialog,
} from "@yca-software/design-system";
import { Building2, Pencil, Archive, RotateCcw } from "lucide-react";
import {
  ADMIN_SUBSCRIPTION_TYPE_BASIC,
  ADMIN_SUBSCRIPTION_TYPE_PRO,
  ADMIN_SUBSCRIPTION_TYPE_ENTERPRISE,
  useAdminArchivedOrganizationDetailQuery,
  useAdminOrganizationDetailQuery,
  useAdminRestoreOrganizationMutation,
} from "@/api";
import { AdminDetailPage } from "../AdminDetailPage";
import { DetailFieldList } from "../DetailFieldList";
import { AdminEditSubscriptionDialog } from "./AdminEditSubscriptionDialog";

function formatDate(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    dateStyle: "medium",
  });
}

function subscriptionTypeLabel(
  type: number,
  t: (key: string) => string,
): string {
  switch (type) {
    case 0:
      return t("admin:organizations.subscriptionTypeFree");
    case ADMIN_SUBSCRIPTION_TYPE_BASIC:
      return t("admin:organizations.subscriptionTypeBasic");
    case ADMIN_SUBSCRIPTION_TYPE_PRO:
      return t("admin:organizations.subscriptionTypePro");
    case ADMIN_SUBSCRIPTION_TYPE_ENTERPRISE:
      return t("admin:organizations.subscriptionTypeEnterprise");
    default:
      return String(type);
  }
}

export function AdminOrganizationDetail() {
  const { t } = useTranslationNamespace(["admin"]);
  const navigate = useNavigate();
  const { orgId } = useParams<{ orgId: string }>();
  const archivedMatch = useMatch({
    path: "/admin/organizations/archived/:orgId",
    end: true,
  });
  const isArchivedDetail = !!archivedMatch;

  const activeQuery = useAdminOrganizationDetailQuery(
    isArchivedDetail ? undefined : orgId,
  );
  const archivedQuery = useAdminArchivedOrganizationDetailQuery(
    isArchivedDetail ? orgId : undefined,
  );

  const { data, isLoading, isError, refetch } = isArchivedDetail
    ? archivedQuery
    : activeQuery;

  const [editSubscriptionOpen, setEditSubscriptionOpen] = useState(false);
  const [restoreDialogOpen, setRestoreDialogOpen] = useState(false);

  const restoreMutation = useAdminRestoreOrganizationMutation({
    onSuccess: () => {
      if (!orgId) return;
      navigate(`/admin/organizations/${orgId}`);
    },
  });

  if (!data && !isLoading && !isError) return null;

  const backHref = isArchivedDetail
    ? "/admin/organizations?scope=archived"
    : "/admin/organizations";
  const backLabel = isArchivedDetail
    ? t("admin:organizations.archived.backToArchivedList")
    : t("admin:organizations.backToList");

  return (
    <>
      {data && (
        <Helmet>
          <title>
            {data.name} –{" "}
            {isArchivedDetail
              ? t("admin:organizations.archived.detailsTitle")
              : t("admin:organizations.details")}
          </title>
        </Helmet>
      )}
      <AdminDetailPage
        backHref={backHref}
        backLabel={backLabel}
        isLoading={isLoading || !orgId}
        isError={!!isError || (!isLoading && !!orgId && !data)}
        notFoundMessage={t("admin:organizations.notFound")}
        headerActions={
          data ? (
            <div className="flex items-center gap-2">
              {!isArchivedDetail && data.customSubscription && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setEditSubscriptionOpen(true)}
                  aria-label={t("admin:organizations.editSubscription.title")}
                >
                  <Pencil className="mr-2 h-4 w-4" aria-hidden />
                  {t("admin:organizations.editSubscription.button")}
                </Button>
              )}
              {isArchivedDetail && (
                <Button
                  variant="default"
                  size="sm"
                  onClick={() => setRestoreDialogOpen(true)}
                  aria-label={t("admin:organizations.restore.button")}
                >
                  <RotateCcw className="mr-2 h-4 w-4" aria-hidden />
                  {t("admin:organizations.restore.button")}
                </Button>
              )}
            </div>
          ) : undefined
        }
      >
        {data && (
          <>
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Building2 className="h-5 w-5" aria-hidden />
                  {isArchivedDetail
                    ? t("admin:organizations.archived.detailsTitle")
                    : t("admin:organizations.details")}
                </CardTitle>
                <CardDescription>{data.name}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                <DetailFieldList
                  fields={[
                    {
                      label: t("admin:organizations.name"),
                      value: data.name,
                    },
                    {
                      label: t("admin:organizations.created"),
                      value: formatDate(data.createdAt),
                    },
                    {
                      label: t("admin:organizations.billingEmail"),
                      value: data.billingEmail,
                      span: 2,
                    },
                    {
                      label: t("admin:organizations.address"),
                      value:
                        [data.address, data.city, data.zip, data.country]
                          .filter(Boolean)
                          .join(", ") || "—",
                      span: 2,
                    },
                    {
                      label: t("admin:organizations.customSubscription"),
                      value: data.customSubscription
                        ? t("admin:organizations.yes")
                        : t("admin:organizations.no"),
                    },
                    {
                      label: t("admin:organizations.subscriptionType"),
                      value: subscriptionTypeLabel(data.subscriptionType, t),
                    },
                    {
                      label: t("admin:organizations.subscriptionSeats"),
                      value: String(data.subscriptionSeats),
                    },
                    ...(data.subscriptionExpiresAt
                      ? [
                          {
                            label: t("admin:organizations.subscriptionExpires"),
                            value: formatDate(data.subscriptionExpiresAt),
                          },
                        ]
                      : []),
                  ]}
                />
              </CardContent>
            </Card>
            {!isArchivedDetail && (
              <AdminEditSubscriptionDialog
                open={editSubscriptionOpen}
                onOpenChange={setEditSubscriptionOpen}
                organization={data}
                onSuccess={() => refetch()}
              />
            )}
            <ConfirmDialog
              open={archiveDialogOpen}
              onOpenChange={setArchiveDialogOpen}
              title={t("common:confirm")}
              description={t("admin:organizations.archive.confirmDescription")}
              cancelLabel={t("common:cancel")}
              confirmLabel={
                archiveMutation.isPending
                  ? t("common:archiving")
                  : t("common:archive")
              }
              variant="destructive"
              isPending={archiveMutation.isPending}
              onConfirm={() => {
                if (!orgId) return;
                archiveMutation.mutate(orgId);
              }}
              closeOnOutsideClick
            />
            <ConfirmDialog
              open={restoreDialogOpen}
              onOpenChange={setRestoreDialogOpen}
              title={t("common:confirm")}
              description={t("admin:organizations.restore.confirmDescription")}
              cancelLabel={t("common:cancel")}
              confirmLabel={
                restoreMutation.isPending
                  ? t("admin:organizations.restore.restoring")
                  : t("admin:organizations.restore.button")
              }
              variant="default"
              isPending={restoreMutation.isPending}
              onConfirm={() => {
                if (!orgId) return;
                restoreMutation.mutate(orgId);
              }}
              closeOnOutsideClick
            />
          </>
        )}
      </AdminDetailPage>
    </>
  );
}
