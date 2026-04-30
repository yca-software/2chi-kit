import { useAPI } from "@/helpers";
import { MutationCallbacks, MutationError, Team } from "@/types";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { ORGANIZATION_QUERY_KEYS } from "@/constants";

export type ListTeamsResponse = Team[] | null;

const TEAMS_LIST_KEY = "list" as const;

function teamsListQueryKey(orgId: string) {
  return [ORGANIZATION_QUERY_KEYS.TEAMS, orgId, TEAMS_LIST_KEY] as const;
}

export const useListTeamsQuery = (
  orgId: string,
  options?: { enabled?: boolean },
) => {
  const fetchWrapper = useAPI();
  const enabled = options?.enabled ?? true;
  return useQuery<ListTeamsResponse>({
    queryKey: teamsListQueryKey(orgId),
    queryFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}/team`,
        method: "GET",
      }) as Promise<ListTeamsResponse>,
    enabled: !!orgId && enabled,
  });
};

export interface CreateTeamRequest {
  name: string;
  description: string;
}

export const useCreateTeamMutation = (
  orgId: string,
  callbacks: MutationCallbacks<Team> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<Team, MutationError, CreateTeamRequest>({
    mutationFn: (body: CreateTeamRequest) =>
      fetchWrapper({
        endpoint: `organization/${orgId}/team`,
        method: "POST",
        body,
      }) as Promise<Team>,
    onSuccess: (data) => {
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.TEAMS, orgId],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export interface UpdateTeamRequest {
  name?: string;
  description?: string;
}

export const useUpdateTeamMutation = (
  orgId: string,
  teamId: string,
  callbacks: MutationCallbacks<Team> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<Team, MutationError, UpdateTeamRequest>({
    mutationFn: (body) =>
      fetchWrapper({
        endpoint: `organization/${orgId}/team/${teamId}`,
        method: "PATCH",
        body,
      }) as Promise<Team>,
    onSuccess: (data) => {
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.TEAMS, orgId],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export const useDeleteTeamMutation = (
  orgId: string,
  teamId: string,
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<void, MutationError, void>({
    mutationFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}/team/${teamId}`,
        method: "DELETE",
      }) as Promise<void>,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [ORGANIZATION_QUERY_KEYS.TEAMS, orgId],
      });
      callbacks.onSuccess?.();
    },
    onError: callbacks.onError,
  });
};
