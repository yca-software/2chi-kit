import { useAPI, getRefreshTokenFromCookies } from "@/helpers";
import type {
  UserRefreshToken,
  MutationCallbacks,
  MutationError,
} from "@/types";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { REFRESH_TOKEN_QUERY_KEYS } from "@/constants";

export type ListActiveRefreshTokensResponse = UserRefreshToken[] | null;

export const useListActiveRefreshTokensQuery = (enabled = true) => {
  const fetchWrapper = useAPI();
  return useQuery<ListActiveRefreshTokensResponse>({
    queryKey: [REFRESH_TOKEN_QUERY_KEYS.ACTIVE],
    queryFn: () =>
      fetchWrapper({
        endpoint: "user/token",
        method: "GET",
      }) as Promise<ListActiveRefreshTokensResponse>,
    enabled,
  });
};

export const useRevokeRefreshTokenMutation = (
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<void, MutationError, string>({
    mutationFn: (tokenId: string) =>
      fetchWrapper({
        endpoint: `user/token/${tokenId}`,
        method: "DELETE",
      }) as Promise<void>,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [REFRESH_TOKEN_QUERY_KEYS.ACTIVE],
      });
      callbacks.onSuccess?.();
    },
    onError: callbacks.onError,
  });
};

export const useRevokeAllRefreshTokensMutation = (
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<void, MutationError, void>({
    mutationFn: () => {
      const keep = getRefreshTokenFromCookies();
      return fetchWrapper({
        endpoint: "user/token",
        method: "DELETE",
        body:
          keep != null && keep !== "" ? { keepRefreshToken: keep } : undefined,
      }) as Promise<void>;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [REFRESH_TOKEN_QUERY_KEYS.ACTIVE],
      });
      callbacks.onSuccess?.();
    },
    onError: callbacks.onError,
  });
};
