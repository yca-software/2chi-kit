import { useAPI } from "@/helpers";
import { MutationCallbacks, MutationError, Role } from "@/types";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { ORGANIZATION_QUERY_KEYS } from "@/constants";

export type ListRolesResponse = Role[] | null;

const ROLES_LIST_KEY = "list" as const;

function rolesListQueryKey(orgId: string) {
  return [ORGANIZATION_QUERY_KEYS.ROLES, orgId, ROLES_LIST_KEY] as const;
}

export const useListRolesQuery = (
  orgId: string,
  options?: { enabled?: boolean },
) => {
  const fetchWrapper = useAPI();
  const enabled = options?.enabled ?? true;
  return useQuery<ListRolesResponse>({
    queryKey: rolesListQueryKey(orgId),
    queryFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}/role`,
        method: "GET",
      }) as Promise<ListRolesResponse>,
    enabled: !!orgId && enabled,
  });
};

export interface CreateRoleRequest {
  name: string;
  description?: string;
  permissions?: string[];
}

export const useCreateRoleMutation = (
  orgId: string,
  callbacks: MutationCallbacks<Role> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<Role, MutationError, CreateRoleRequest>({
    mutationFn: (body: CreateRoleRequest) =>
      fetchWrapper({
        endpoint: `organization/${orgId}/role`,
        method: "POST",
        body,
      }) as Promise<Role>,
    onSuccess: (data) => {
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.ROLES, orgId],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export interface UpdateRoleRequest {
  name?: string;
  description?: string;
  permissions?: string[];
}

export const useUpdateRoleMutation = (
  orgId: string,
  roleId: string,
  callbacks: MutationCallbacks<Role> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<Role, MutationError, UpdateRoleRequest>({
    mutationFn: (body: UpdateRoleRequest) =>
      fetchWrapper({
        endpoint: `organization/${orgId}/role/${roleId}`,
        method: "PATCH",
        body,
      }) as Promise<Role>,
    onSuccess: (data) => {
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.ROLES, orgId],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export const useDeleteRoleMutation = (
  orgId: string,
  roleId: string,
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<void, MutationError, void>({
    mutationFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}/role/${roleId}`,
        method: "DELETE",
      }) as Promise<void>,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.ROLES, orgId],
      });
      callbacks.onSuccess?.();
    },
    onError: callbacks.onError,
  });
};
