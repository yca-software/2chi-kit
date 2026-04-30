import { useState } from "react";
import { useNavigate } from "react-router";
import { Helmet } from "react-helmet-async";
import { useTranslationNamespace, useAdminListPage } from "@/helpers";
import { Button } from "@yca-software/design-system";
import { Plus } from "lucide-react";
import { AdminListPage } from "../AdminListPage";
import {
  useAdminOrganizationListInfiniteQuery,
  AdminOrganizationListItem,
} from "@/api";
import { AdminCreateOrganizationWithCustomSubscriptionDialog } from "./AdminCreateOrganizationWithCustomSubscriptionDialog";

function formatDate(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    dateStyle: "medium",
  });
}

export function AdminOrganizations() {
  const { t } = useTranslationNamespace(["admin"]);
  const navigate = useNavigate();
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const { search, setSearch, items, query, loadMoreRef } = useAdminListPage(
    useAdminOrganizationListInfiniteQuery,
  );

  return (
    <>
      <Helmet>
        <title>{t("admin:organizations.title")}</title>
      </Helmet>
      <AdminListPage<AdminOrganizationListItem>
        title={t("admin:organizations.title")}
        description={t("admin:organizations.description")}
        headerActions={
          <Button onClick={() => setCreateDialogOpen(true)} size="sm">
            <Plus className="mr-2 h-4 w-4" aria-hidden />
            {t("admin:organizations.createCustom.button")}
          </Button>
        }
        cardTitle={t("admin:organizations.allOrganizations")}
        searchPlaceholder={t("admin:organizations.searchPlaceholder")}
        emptyMessage={t("admin:organizations.noOrganizations")}
        columns={[
          {
            key: "name",
            header: t("admin:organizations.name"),
            render: (org) => org.name,
            className: "font-medium",
          },
          {
            key: "created",
            header: t("admin:organizations.created"),
            render: (org) => formatDate(org.createdAt),
          },
        ]}
        items={items}
        isLoading={query.isLoading}
        hasNextPage={query.hasNextPage ?? false}
        isFetchingNextPage={query.isFetchingNextPage}
        loadMoreRef={loadMoreRef}
        search={search}
        onSearchChange={setSearch}
        onRowClick={(org) => navigate(`/admin/organizations/${org.id}`)}
        getRowKey={(org) => org.id}
      />
      <AdminCreateOrganizationWithCustomSubscriptionDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
        onSuccess={(orgId) => navigate(`/admin/organizations/${orgId}`)}
      />
    </>
  );
}
