import { useAPI } from "@/helpers";
import type { User, AdminAccess } from "@/types";
import type {
  Organization,
  OrganizationMemberWithOrganizationAndRole,
  MutationCallbacks,
  MutationError,
} from "@/types";
import {
  useQuery,
  useInfiniteQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import {
  ADMIN_USER_QUERY_KEYS,
  ADMIN_ORGANIZATION_QUERY_KEYS,
} from "@/constants";
import type { AuthenticateResponse } from "@/api";

/* ----- Users ----- */

export interface AdminUserListItem extends User {}

export interface AdminUserListParams {
  search?: string;
  limit?: number;
  offset?: number;
}

export interface AdminUserListResponse {
  items: AdminUserListItem[];
  hasNext: boolean;
}

const USER_PAGE_SIZE = 20;

export const useAdminUserListInfiniteQuery = (search: string = "") => {
  const fetchWrapper = useAPI();
  return useInfiniteQuery<AdminUserListResponse>({
    queryKey: [ADMIN_USER_QUERY_KEYS.ALL, "infinite", search],
    queryFn: ({ pageParam = 0 }) => {
      const searchParams = new URLSearchParams();
      searchParams.set("limit", String(USER_PAGE_SIZE));
      searchParams.set("offset", String(pageParam));
      if (search) searchParams.set("search", search);
      const qs = searchParams.toString();
      return fetchWrapper({
        endpoint: `admin/user?${qs}`,
        method: "GET",
      }) as Promise<AdminUserListResponse>;
    },
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) => {
      if (!lastPage.hasNext) return undefined;
      return allPages.length * USER_PAGE_SIZE;
    },
  });
};

export interface AdminUserDetailResponse {
  user: User;
  adminAccess: AdminAccess | null;
  roles: OrganizationMemberWithOrganizationAndRole[] | null;
}

export const useAdminUserDetailQuery = (userId: string | undefined) => {
  const fetchWrapper = useAPI();
  return useQuery<AdminUserDetailResponse>({
    queryKey: [ADMIN_USER_QUERY_KEYS.DETAIL, userId],
    queryFn: () =>
      fetchWrapper({
        endpoint: `admin/user/${userId}`,
        method: "GET",
      }) as Promise<AdminUserDetailResponse>,
    enabled: !!userId,
  });
};

export const useAdminImpersonateUserMutation = (
  callbacks: MutationCallbacks<AuthenticateResponse> = {},
) => {
  const fetchWrapper = useAPI();
  return useMutation<AuthenticateResponse, MutationError, string>({
    mutationFn: (userId: string) =>
      fetchWrapper({
        endpoint: `admin/user/${userId}/impersonate`,
        method: "POST",
      }) as Promise<AuthenticateResponse>,
    onSuccess: callbacks.onSuccess,
    onError: callbacks.onError,
  });
};

export const useAdminDeleteUserMutation = (
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<void, MutationError, string>({
    mutationFn: (userId: string) =>
      fetchWrapper({
        endpoint: `admin/user/${userId}`,
        method: "DELETE",
      }) as Promise<void>,
    onSuccess: (_, userId) => {
      queryClient.removeQueries({
        queryKey: [ADMIN_USER_QUERY_KEYS.DETAIL, userId],
      });
      queryClient.invalidateQueries({
        queryKey: [ADMIN_USER_QUERY_KEYS.ALL],
      });
      callbacks.onSuccess?.();
    },
    onError: callbacks.onError,
  });
};

/* ----- Organizations ----- */

export interface AdminOrganizationListItem extends Organization {}

export interface AdminOrganizationListParams {
  search?: string;
  limit?: number;
  offset?: number;
}

export interface AdminOrganizationListResponse {
  items: AdminOrganizationListItem[];
  hasNext: boolean;
}

const ORG_PAGE_SIZE = 20;

export const useAdminOrganizationListInfiniteQuery = (search: string = "") => {
  const fetchWrapper = useAPI();
  return useInfiniteQuery<AdminOrganizationListResponse>({
    queryKey: [ADMIN_ORGANIZATION_QUERY_KEYS.ALL, "infinite", search],
    queryFn: ({ pageParam = 0 }) => {
      const searchParams = new URLSearchParams();
      searchParams.set("limit", String(ORG_PAGE_SIZE));
      searchParams.set("offset", String(pageParam));
      if (search) searchParams.set("search", search);
      const qs = searchParams.toString();
      return fetchWrapper({
        endpoint: `admin/organization?${qs}`,
        method: "GET",
      }) as Promise<AdminOrganizationListResponse>;
    },
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) => {
      if (!lastPage.hasNext) return undefined;
      return allPages.length * ORG_PAGE_SIZE;
    },
  });
};

export const useAdminArchivedOrganizationListInfiniteQuery = (
  search: string = "",
) => {
  const fetchWrapper = useAPI();
  return useInfiniteQuery<AdminOrganizationListResponse>({
    queryKey: [ADMIN_ORGANIZATION_QUERY_KEYS.ALL_ARCHIVED, "infinite", search],
    queryFn: ({ pageParam = 0 }) => {
      const searchParams = new URLSearchParams();
      searchParams.set("limit", String(ORG_PAGE_SIZE));
      searchParams.set("offset", String(pageParam));
      if (search) searchParams.set("search", search);
      const qs = searchParams.toString();
      return fetchWrapper({
        endpoint: `admin/organization/archived?${qs}`,
        method: "GET",
      }) as Promise<AdminOrganizationListResponse>;
    },
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) => {
      if (!lastPage.hasNext) return undefined;
      return allPages.length * ORG_PAGE_SIZE;
    },
  });
};

export const useAdminOrganizationDetailQuery = (orgId: string | undefined) => {
  const fetchWrapper = useAPI();
  return useQuery<Organization>({
    queryKey: [ADMIN_ORGANIZATION_QUERY_KEYS.DETAIL, orgId],
    queryFn: () =>
      fetchWrapper({
        endpoint: `admin/organization/${orgId}`,
        method: "GET",
      }) as Promise<Organization>,
    enabled: !!orgId,
  });
};

export const useAdminArchivedOrganizationDetailQuery = (
  orgId: string | undefined,
) => {
  const fetchWrapper = useAPI();
  return useQuery<Organization>({
    queryKey: [ADMIN_ORGANIZATION_QUERY_KEYS.DETAIL_ARCHIVED, orgId],
    queryFn: () =>
      fetchWrapper({
        endpoint: `admin/organization/archived/${orgId}`,
        method: "GET",
      }) as Promise<Organization>,
    enabled: !!orgId,
  });
};

/** Subscription type: 1=Basic, 2=Pro, 3=Enterprise (backend constants). */
export const ADMIN_SUBSCRIPTION_TYPE_BASIC = 1;
export const ADMIN_SUBSCRIPTION_TYPE_PRO = 2;
export const ADMIN_SUBSCRIPTION_TYPE_ENTERPRISE = 3;

export interface AdminCreateOrganizationWithCustomSubscriptionRequest {
  name: string;
  placeId: string;
  billingEmail: string;
  ownerEmail: string;
  subscriptionType: number;
  subscriptionSeats: number;
  subscriptionExpiresAt?: string | null;
  language: string;
}

export const useAdminCreateOrganizationWithCustomSubscriptionMutation = (
  callbacks: MutationCallbacks<Organization> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<
    Organization,
    MutationError,
    AdminCreateOrganizationWithCustomSubscriptionRequest
  >({
    mutationFn: (body) =>
      fetchWrapper({
        endpoint: "admin/organization",
        method: "POST",
        body,
      }) as Promise<Organization>,
    onSuccess: (data) => {
      queryClient.invalidateQueries({
        queryKey: [ADMIN_ORGANIZATION_QUERY_KEYS.ALL],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export interface AdminUpdateOrganizationSubscriptionRequest {
  customSubscription?: boolean;
  subscriptionType?: number;
  subscriptionSeats?: number;
  subscriptionExpiresAt?: string | null; // RFC3339 or empty string to clear
}

export const useAdminUpdateOrganizationSubscriptionMutation = (
  orgId: string,
  callbacks: MutationCallbacks<Organization> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<
    Organization,
    MutationError,
    AdminUpdateOrganizationSubscriptionRequest
  >({
    mutationFn: (body) =>
      fetchWrapper({
        endpoint: `admin/organization/${orgId}/subscription`,
        method: "PATCH",
        body,
      }) as Promise<Organization>,
    onSuccess: (data) => {
      queryClient.setQueryData(
        [ADMIN_ORGANIZATION_QUERY_KEYS.DETAIL, orgId],
        data,
      );
      queryClient.invalidateQueries({
        queryKey: [ADMIN_ORGANIZATION_QUERY_KEYS.ALL],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export const useAdminRestoreOrganizationMutation = (
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<void, MutationError, string>({
    mutationFn: (orgId: string) =>
      fetchWrapper({
        endpoint: `admin/organization/archived/${orgId}/restore`,
        method: "POST",
      }) as Promise<void>,
    onSuccess: (_, orgId) => {
      queryClient.removeQueries({
        queryKey: [ADMIN_ORGANIZATION_QUERY_KEYS.DETAIL_ARCHIVED, orgId],
      });
      queryClient.invalidateQueries({
        queryKey: [ADMIN_ORGANIZATION_QUERY_KEYS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [ADMIN_ORGANIZATION_QUERY_KEYS.ALL_ARCHIVED],
      });
      callbacks.onSuccess?.();
    },
    onError: callbacks.onError,
  });
};
