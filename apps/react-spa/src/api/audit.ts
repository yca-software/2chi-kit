import { useQuery } from "@tanstack/react-query";
import { ORGANIZATION_QUERY_KEYS } from "@/constants";
import { useAPI } from "@/helpers";
import type { AuditLog } from "@/types";

export interface ListAuditLogsParams {
  orgId: string;
  limit?: number;
  offset?: number;
  startDate?: string;
  endDate?: string;
}

export interface ListAuditLogsResponse {
  items: AuditLog[];
  hasNext: boolean;
}

export const useListAuditLogsQuery = (
  params: ListAuditLogsParams,
  options?: { enabled?: boolean },
) => {
  const fetchWrapper = useAPI();
  const enabled = options?.enabled ?? true;
  const { orgId, limit = 50, offset = 0, startDate, endDate } = params;

  const search = new URLSearchParams();
  search.set("limit", String(limit));
  search.set("offset", String(offset));
  if (startDate) search.set("startDate", startDate);
  if (endDate) search.set("endDate", endDate);
  const query = search.toString();

  return useQuery<ListAuditLogsResponse>({
    queryKey: [
      ORGANIZATION_QUERY_KEYS.AUDIT_LOGS,
      orgId,
      limit,
      offset,
      startDate,
      endDate,
    ],
    queryFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}/audit-log${query ? `?${query}` : ""}`,
        method: "GET",
      }) as Promise<ListAuditLogsResponse>,
    enabled: !!orgId && enabled,
  });
};
