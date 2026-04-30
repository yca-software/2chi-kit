import { useAPI } from "@/helpers";
import {
  MutationCallbacks,
  MutationError,
  TeamMember,
  TeamMemberWithUser,
} from "@/types";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { ORGANIZATION_QUERY_KEYS } from "@/constants";

export type ListTeamMembersResponse = TeamMemberWithUser[] | null;

export const useListTeamMembersQuery = (orgId: string, teamId: string) => {
  const fetchWrapper = useAPI();
  return useQuery<ListTeamMembersResponse>({
    queryKey: [ORGANIZATION_QUERY_KEYS.TEAM_MEMBERS, orgId, teamId],
    queryFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}/team/${teamId}/member`,
        method: "GET",
      }) as Promise<ListTeamMembersResponse>,
    enabled: !!orgId && !!teamId,
  });
};

export interface AddTeamMemberRequest {
  userId: string;
}

export const useAddTeamMemberMutation = (
  orgId: string,
  teamId: string,
  callbacks: MutationCallbacks<TeamMember> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<TeamMember, MutationError, AddTeamMemberRequest>({
    mutationFn: (body: AddTeamMemberRequest) =>
      fetchWrapper({
        endpoint: `organization/${orgId}/team/${teamId}/member`,
        method: "POST",
        body,
      }) as Promise<TeamMember>,
    onSuccess: (data) => {
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.TEAM_MEMBERS, orgId, teamId],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export const useRemoveTeamMemberMutation = (
  orgId: string,
  teamId: string,
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<void, MutationError, string>({
    mutationFn: (memberId: string) =>
      fetchWrapper({
        endpoint: `organization/${orgId}/team/${teamId}/member/${memberId}`,
        method: "DELETE",
      }) as Promise<void>,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.TEAM_MEMBERS, orgId, teamId],
      });
      callbacks.onSuccess?.();
    },
    onError: callbacks.onError,
  });
};
