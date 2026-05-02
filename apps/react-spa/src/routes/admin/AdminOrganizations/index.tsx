import { useState, useEffect, useRef, useCallback } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { Helmet } from "react-helmet-async";
import { useTranslationNamespace } from "@/helpers";
import { Button } from "@yca-software/design-system";
import { Plus } from "lucide-react";
import { AdminListPage } from "../AdminListPage";
import {
  useAdminOrganizationListInfiniteQuery,
  useAdminArchivedOrganizationListInfiniteQuery,
  AdminOrganizationListItem,
} from "@/api";
import { AdminCreateOrganizationWithCustomSubscriptionDialog } from "./AdminCreateOrganizationWithCustomSubscriptionDialog";

function formatDate(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    dateStyle: "medium",
  });
}

const LIST_DEBOUNCE_MS = 300;

export function AdminOrganizations() {
  const { t } = useTranslationNamespace(["admin"]);
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const scope =
    searchParams.get("scope") === "archived" ? "archived" : "active";
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const loadMoreRef = useRef<HTMLTableRowElement>(null);

  useEffect(() => {
    const id = window.setTimeout(
      () => setDebouncedSearch(search),
      LIST_DEBOUNCE_MS,
    );
    return () => window.clearTimeout(id);
  }, [search]);

  const activeQuery = useAdminOrganizationListInfiniteQuery(debouncedSearch);
  const archivedQuery =
    useAdminArchivedOrganizationListInfiniteQuery(debouncedSearch);
  const query = scope === "archived" ? archivedQuery : activeQuery;
  const items =
    query.data?.pages.flatMap((p) => p.items) ?? [];

  const handleLoadMore = useCallback(() => {
    if (query.hasNextPage && !query.isFetchingNextPage) {
      query.fetchNextPage();
    }
  }, [query.hasNextPage, query.isFetchingNextPage, query.fetchNextPage]);

  useEffect(() => {
    const el = loadMoreRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) handleLoadMore();
      },
      { rootMargin: "100px", threshold: 0 },
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [handleLoadMore]);

  const setScope = (next: "active" | "archived") => {
    if (next === "archived") {
      setSearchParams({ scope: "archived" });
    } else {
      setSearchParams({});
    }
  };

  return (
    <>
      <Helmet>
        <title>{t("admin:organizations.title")}</title>
      </Helmet>
      <AdminListPage<AdminOrganizationListItem>
        title={t("admin:organizations.title")}
        description={
          scope === "archived"
            ? t("admin:organizations.archived.description")
            : t("admin:organizations.description")
        }
        headerActions={
          <div className="flex flex-wrap items-center gap-2">
            <div
              className="inline-flex rounded-lg border border-border bg-muted/40 p-0.5"
              role="group"
              aria-label={t("admin:organizations.title")}
            >
              <Button
                variant={scope === "active" ? "secondary" : "ghost"}
                size="sm"
                className="rounded-md"
                onClick={() => setScope("active")}
              >
                {t("admin:organizations.filterActive")}
              </Button>
              <Button
                variant={scope === "archived" ? "secondary" : "ghost"}
                size="sm"
                className="rounded-md"
                onClick={() => setScope("archived")}
              >
                {t("admin:organizations.filterArchived")}
              </Button>
            </div>
            {scope === "active" ? (
              <Button onClick={() => setCreateDialogOpen(true)} size="sm">
                <Plus className="mr-2 h-4 w-4" aria-hidden />
                {t("admin:organizations.createCustom.button")}
              </Button>
            ) : null}
          </div>
        }
        cardTitle={
          scope === "archived"
            ? t("admin:organizations.archived.allArchived")
            : t("admin:organizations.allOrganizations")
        }
        searchPlaceholder={
          scope === "archived"
            ? t("admin:organizations.archived.searchPlaceholder")
            : t("admin:organizations.searchPlaceholder")
        }
        emptyMessage={
          scope === "archived"
            ? t("admin:organizations.archived.noOrganizations")
            : t("admin:organizations.noOrganizations")
        }
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
        onRowClick={(org) =>
          navigate(
            scope === "archived"
              ? `/admin/organizations/archived/${org.id}`
              : `/admin/organizations/${org.id}`,
          )
        }
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
