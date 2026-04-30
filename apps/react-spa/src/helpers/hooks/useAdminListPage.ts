import { useState, useEffect, useRef, useCallback } from "react";
import type { UseInfiniteQueryResult, InfiniteData } from "@tanstack/react-query";

export interface AdminListPageResponse<T> {
  items: T[];
  hasNext: boolean;
}

export function useAdminListPage<T>(
  useInfiniteQuery: (
    search: string
  ) => UseInfiniteQueryResult<InfiniteData<AdminListPageResponse<T>>>,
  options?: { debounceMs?: number }
) {
  const debounceMs = options?.debounceMs ?? 300;
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const loadMoreRef = useRef<HTMLTableRowElement>(null);

  useEffect(() => {
    const id = window.setTimeout(() => setDebouncedSearch(search), debounceMs);
    return () => window.clearTimeout(id);
  }, [search, debounceMs]);

  const query = useInfiniteQuery(debouncedSearch);
  const items =
    query.data?.pages.flatMap((p: AdminListPageResponse<T>) => p.items) ?? [];

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
      { rootMargin: "100px", threshold: 0 }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [handleLoadMore]);

  return {
    search,
    setSearch,
    items,
    query,
    loadMoreRef,
  };
}
