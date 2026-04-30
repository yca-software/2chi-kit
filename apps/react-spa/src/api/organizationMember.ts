import { useAPI } from "@/helpers";
import type {
  OrganizationMemberWithUser,
  MutationCallbacks,
  MutationError,
} from "@/types";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { ORGANIZATION_QUERY_KEYS } from "@/constants";

export const useListOrganizationMembersQuery = (orgId: string) => {
  const fetchWrapper = useAPI();
  return useQuery<OrganizationMemberWithUser[]>({
    queryKey: [ORGANIZATION_QUERY_KEYS.MEMBERS, orgId],
    queryFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}/member`,
        method: "GET",
      }) as Promise<OrganizationMemberWithUser[]>,
    enabled: !!orgId,
  });
};

export interface UpdateMemberRoleRequest {
  roleId: string;
}

export const useUpdateMemberRoleMutation = (
  orgId: string,
  memberId: string,
  callbacks: MutationCallbacks<unknown> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<unknown, MutationError, UpdateMemberRoleRequest>({
    mutationFn: (body: UpdateMemberRoleRequest) =>
      fetchWrapper({
        endpoint: `organization/${orgId}/member/${memberId}/role`,
        method: "PATCH",
        body,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.MEMBERS, orgId],
      });
      callbacks.onSuccess?.({});
    },
    onError: callbacks.onError,
  });
};

export const useRemoveMemberMutation = (
  orgId: string,
  memberId: string,
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<void, MutationError, void>({
    mutationFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}/member/${memberId}`,
        method: "DELETE",
      }) as Promise<void>,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.MEMBERS, orgId],
      });
      callbacks.onSuccess?.();
    },
    onError: callbacks.onError,
  });
};
