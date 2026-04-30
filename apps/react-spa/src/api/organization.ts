import { posthog } from "@/analytics";
import { useAPI } from "@/helpers";
import type {
  Organization,
  Role,
  OrganizationMember,
  MutationCallbacks,
  MutationError,
} from "@/types";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { ORGANIZATION_QUERY_KEYS, USER_QUERY_KEYS } from "@/constants";
import { GetCurrentUserResponse, useRefreshAccessTokenMutation } from "@/api";
import { useUserState } from "@/states";
import { useShallow } from "zustand/shallow";

export const useGetOrganizationQuery = (orgId: string) => {
  const fetchWrapper = useAPI();
  const userOrgIds = useUserState(
    useShallow(
      (state) => state.userData?.roles?.map((r) => r.organizationId) ?? [],
    ),
  );
  const isOrgInUserList = userOrgIds.length === 0 || userOrgIds.includes(orgId);

  return useQuery<Organization>({
    queryKey: [ORGANIZATION_QUERY_KEYS.DETAIL, orgId],
    queryFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}`,
        method: "GET",
      }) as Promise<Organization>,
    enabled: !!orgId && isOrgInUserList,
  });
};

export interface CreateOrganizationRequest {
  name: string;
  placeId: string;
  billingEmail: string;
}

export interface CreateOrganizationResponse {
  organization: Organization;
  roles?: Role[];
  members?: OrganizationMember;
}

export const useCreateOrganizationMutation = (
  callbacks: MutationCallbacks<CreateOrganizationResponse>,
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  const refreshTokenMutation = useRefreshAccessTokenMutation();

  return useMutation<
    CreateOrganizationResponse,
    MutationError,
    CreateOrganizationRequest
  >({
    mutationFn: (body: CreateOrganizationRequest) =>
      fetchWrapper({
        endpoint: "organization",
        method: "POST",
        body,
      }) as Promise<CreateOrganizationResponse>,
    onSuccess: (data) => {
      const applySuccess = () => {
        posthog?.capture?.("organization_created", {
          organization_id: data.organization.id,
        });
        queryClient.invalidateQueries({
          queryKey: [ORGANIZATION_QUERY_KEYS.ALL],
        });
        queryClient.invalidateQueries({
          queryKey: [USER_QUERY_KEYS.CURRENT],
        });
        callbacks.onSuccess?.(data);
      };
      refreshTokenMutation.mutateAsync().then(applySuccess).catch(applySuccess);
    },
    onError: callbacks.onError,
  });
};

export interface UpdateOrganizationRequest {
  name?: string;
  placeId?: string;
}

export const useUpdateOrganizationMutation = (
  orgId: string,
  callbacks: MutationCallbacks<Organization> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<Organization, MutationError, UpdateOrganizationRequest>({
    mutationFn: (body: UpdateOrganizationRequest) =>
      fetchWrapper({
        endpoint: `organization/${orgId}`,
        method: "PATCH",
        body,
      }) as Promise<Organization>,
    onSuccess: (data) => {
      queryClient.setQueryData([ORGANIZATION_QUERY_KEYS.DETAIL, orgId], data);
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [USER_QUERY_KEYS.CURRENT],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export const useDeleteOrganizationMutation = (
  orgId: string,
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<void, MutationError, void>({
    mutationFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}`,
        method: "DELETE",
      }) as Promise<void>,
    onSuccess: () => {
      queryClient.removeQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.DETAIL, orgId],
      });
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.ALL],
      });
      callbacks.onSuccess?.();
    },
    onError: callbacks.onError,
  });
};

export const useArchiveOrganizationMutation = (
  orgId: string,
  callbacks: MutationCallbacks<GetCurrentUserResponse | undefined> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  const refreshTokenMutation = useRefreshAccessTokenMutation();

  return useMutation<void, MutationError, void>({
    mutationFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}/archive`,
        method: "POST",
      }) as Promise<void>,
    onSuccess: () => {
      const applySuccess = () => {
        queryClient.removeQueries({
          queryKey: [ORGANIZATION_QUERY_KEYS.DETAIL, orgId],
        });
        queryClient.invalidateQueries({
          queryKey: [ORGANIZATION_QUERY_KEYS.ALL],
        });
        queryClient.invalidateQueries({
          queryKey: [USER_QUERY_KEYS.CURRENT],
        });
        // Refetch current user so org list and store are updated (archived org excluded by backend)
        queryClient
          .refetchQueries({ queryKey: [USER_QUERY_KEYS.CURRENT] })
          .then(() => {
            const fresh = queryClient.getQueryData<GetCurrentUserResponse>([
              USER_QUERY_KEYS.CURRENT,
            ]);
            if (fresh) {
              useUserState.getState().setUserData({
                user: fresh.user,
                admin: fresh.adminAccess ?? null,
                roles: fresh.roles ?? null,
              });
            }
            callbacks.onSuccess?.(fresh ?? undefined);
          });
      };
      refreshTokenMutation.mutateAsync().then(applySuccess).catch(applySuccess);
    },
    onError: callbacks.onError,
  });
};

export interface CreateCheckoutSessionRequest {
  planId: string;
}

export interface CreateCheckoutSessionResponse {
  transactionId: string;
}

export const useCreateCheckoutSessionMutation = (
  orgId: string,
  callbacks: MutationCallbacks<CreateCheckoutSessionResponse> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<
    CreateCheckoutSessionResponse,
    MutationError,
    CreateCheckoutSessionRequest
  >({
    mutationFn: (body: CreateCheckoutSessionRequest) =>
      fetchWrapper({
        endpoint: `organization/${orgId}/subscription/checkout`,
        method: "POST",
        body,
      }) as Promise<CreateCheckoutSessionResponse>,
    onSuccess: (data) => {
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.DETAIL, orgId],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export interface CreateCustomerPortalSessionResponse {
  portalUrl: string;
}

export const useCreateCustomerPortalSessionMutation = (
  orgId: string,
  callbacks: MutationCallbacks<CreateCustomerPortalSessionResponse> = {},
) => {
  const fetchWrapper = useAPI();
  return useMutation<CreateCustomerPortalSessionResponse, MutationError, void>({
    mutationFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}/subscription/portal`,
        method: "POST",
      }) as Promise<CreateCustomerPortalSessionResponse>,
    onSuccess: callbacks.onSuccess,
    onError: callbacks.onError,
  });
};

export interface ChangePlanRequest {
  planId: string;
}

/** "immediately" = plan updated now; "next_billing_period" = change scheduled for end of current period (e.g. annual→monthly). */
export type ChangePlanEffectiveAt = "immediately" | "next_billing_period";

export interface ChangePlanResponse {
  organization: Organization;
  effectiveAt: ChangePlanEffectiveAt;
}

export const useChangePlanMutation = (
  orgId: string,
  callbacks: MutationCallbacks<ChangePlanResponse> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<ChangePlanResponse, MutationError, ChangePlanRequest>({
    mutationFn: (body: ChangePlanRequest) =>
      fetchWrapper({
        endpoint: `organization/${orgId}/subscription/change-plan`,
        method: "POST",
        body,
      }) as Promise<ChangePlanResponse>,
    onSuccess: (data) => {
      queryClient.setQueryData(
        [ORGANIZATION_QUERY_KEYS.DETAIL, orgId],
        data.organization,
      );
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.DETAIL, orgId],
      });
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [USER_QUERY_KEYS.CURRENT],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export interface ProcessTransactionRequest {
  transactionId: string;
  priceId: string;
}

export const useProcessTransactionMutation = (
  orgId: string,
  callbacks: MutationCallbacks<Organization> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<Organization, MutationError, ProcessTransactionRequest>({
    mutationFn: (body: ProcessTransactionRequest) =>
      fetchWrapper({
        endpoint: `organization/${orgId}/subscription/process-transaction`,
        method: "POST",
        body,
      }) as Promise<Organization>,
    onSuccess: (data) => {
      queryClient.setQueryData([ORGANIZATION_QUERY_KEYS.DETAIL, orgId], data);
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.ALL],
      });
      queryClient.invalidateQueries({
        queryKey: [USER_QUERY_KEYS.CURRENT],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};
