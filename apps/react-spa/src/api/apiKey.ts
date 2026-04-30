import { posthog } from "@/analytics";
import { useAPI } from "@/helpers";
import { MutationCallbacks, MutationError, ApiKey } from "@/types";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { ORGANIZATION_QUERY_KEYS } from "@/constants";

export const useListApiKeysQuery = (
  orgId: string,
  options?: { enabled?: boolean },
) => {
  const fetchWrapper = useAPI();
  const enabled = options?.enabled ?? true;
  return useQuery<ApiKey[]>({
    queryKey: [ORGANIZATION_QUERY_KEYS.API_KEYS, orgId],
    queryFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}/api-key`,
        method: "GET",
      }) as Promise<ApiKey[]>,
    enabled: !!orgId && enabled,
  });
};

export interface CreateApiKeyRequest {
  name: string;
  permissions: string[];
  expiresAt: string | null;
}

export interface CreateApiKeyResponse {
  apiKey: ApiKey;
  secret: string;
}

export interface UpdateApiKeyRequest {
  name: string;
  permissions: string[];
}

export const useCreateApiKeyMutation = (
  orgId: string,
  callbacks: MutationCallbacks<CreateApiKeyResponse>,
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<CreateApiKeyResponse, MutationError, CreateApiKeyRequest>({
    mutationFn: (body: CreateApiKeyRequest) =>
      fetchWrapper({
        endpoint: `organization/${orgId}/api-key`,
        method: "POST",
        body,
      }) as Promise<CreateApiKeyResponse>,
    onSuccess: (data) => {
      posthog?.capture?.("api_key_created", {
        organization_id: orgId,
        api_key_id: data.apiKey.id,
      });
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.API_KEYS, orgId],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export const useUpdateApiKeyMutation = (
  orgId: string,
  apiKeyId: string,
  callbacks: MutationCallbacks<ApiKey> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<ApiKey, MutationError, UpdateApiKeyRequest>({
    mutationFn: (body: UpdateApiKeyRequest) =>
      fetchWrapper({
        endpoint: `organization/${orgId}/api-key/${apiKeyId}`,
        method: "PATCH",
        body,
      }) as Promise<ApiKey>,
    onSuccess: (data) => {
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.API_KEYS, orgId],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export const useRevokeApiKeyMutation = (
  orgId: string,
  apiKeyId: string,
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<void, MutationError, void>({
    mutationFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}/api-key/${apiKeyId}`,
        method: "DELETE",
      }) as Promise<void>,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.API_KEYS, orgId],
      });
      callbacks.onSuccess?.();
    },
    onError: callbacks.onError,
  });
};
