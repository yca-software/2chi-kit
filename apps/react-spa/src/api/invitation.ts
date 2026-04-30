import { posthog } from "@/analytics";
import { useAPI } from "@/helpers";
import type {
  Invitation,
  OrganizationMemberWithUser,
  MutationCallbacks,
  MutationError,
} from "@/types";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { INVITATION_QUERY_KEYS } from "@/constants";

export type ListInvitationsResponse = Invitation[] | null;

export const useListInvitationsQuery = (orgId: string) => {
  const fetchWrapper = useAPI();
  return useQuery<ListInvitationsResponse>({
    queryKey: [INVITATION_QUERY_KEYS.LIST, orgId],
    queryFn: () =>
      fetchWrapper({
        endpoint: `organization/${orgId}/invitation`,
        method: "GET",
      }) as Promise<ListInvitationsResponse>,
    enabled: !!orgId,
  });
};

export interface CreateInvitationRequest {
  email: string;
  roleId: string;
}

export interface CreateInvitationResponse {
  invitation: Invitation | null;
  member: OrganizationMemberWithUser | null;
}

export const useCreateInvitationMutation = (
  orgId: string,
  callbacks: MutationCallbacks<CreateInvitationResponse>,
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<
    CreateInvitationResponse,
    MutationError,
    CreateInvitationRequest
  >({
    mutationFn: (body: CreateInvitationRequest) =>
      fetchWrapper({
        endpoint: `organization/${orgId}/invitation`,
        method: "POST",
        body,
      }) as Promise<CreateInvitationResponse>,
    onSuccess: (data) => {
      posthog?.capture?.("invitation_created", {
        organization_id: orgId,
        invitation_id: data.invitation?.id,
        member_id: data.member?.id,
      });
      queryClient.invalidateQueries({
        queryKey: [INVITATION_QUERY_KEYS.LIST, orgId],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export const useRevokeInvitationMutation = (
  orgId: string,
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<void, MutationError, string>({
    mutationFn: (invitationId: string) =>
      fetchWrapper({
        endpoint: `organization/${orgId}/invitation/${invitationId}`,
        method: "DELETE",
      }) as Promise<void>,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [INVITATION_QUERY_KEYS.LIST, orgId],
      });
      callbacks.onSuccess?.();
    },
    onError: callbacks.onError,
  });
};
