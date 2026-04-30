import { useNavigate } from "react-router";
import { Helmet } from "react-helmet-async";
import { useTranslationNamespace, useAdminListPage } from "@/helpers";
import { AdminListPage } from "../AdminListPage";
import { useAdminUserListInfiniteQuery, AdminUserListItem } from "@/api";

function formatDate(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

function formatName(user: AdminUserListItem) {
  return [user.firstName, user.lastName].filter(Boolean).join(" ") || "—";
}

export function AdminUsers() {
  const { t } = useTranslationNamespace(["admin"]);
  const navigate = useNavigate();
  const { search, setSearch, items, query, loadMoreRef } = useAdminListPage(
    useAdminUserListInfiniteQuery,
  );

  return (
    <>
      <Helmet>
        <title>{t("admin:users.title")}</title>
      </Helmet>
      <AdminListPage<AdminUserListItem>
        title={t("admin:users.title")}
        description={t("admin:users.description")}
        cardTitle={t("admin:users.allUsers")}
        searchPlaceholder={t("admin:users.searchPlaceholder")}
        emptyMessage={t("admin:users.noUsers")}
        columns={[
          {
            key: "name",
            header: t("admin:users.name"),
            render: formatName,
            className: "font-medium",
          },
          {
            key: "email",
            header: t("admin:users.email"),
            render: (u) => u.email,
          },
          {
            key: "created",
            header: t("admin:users.created"),
            render: (u) => formatDate(u.createdAt),
          },
        ]}
        items={items}
        isLoading={query.isLoading}
        hasNextPage={query.hasNextPage ?? false}
        isFetchingNextPage={query.isFetchingNextPage}
        loadMoreRef={loadMoreRef}
        search={search}
        onSearchChange={setSearch}
        onRowClick={(u) => navigate(`/admin/users/${u.id}`)}
        getRowKey={(u) => u.id}
      />
    </>
  );
}
