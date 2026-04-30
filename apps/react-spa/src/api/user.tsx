import { USER_QUERY_KEYS } from "@/constants";
import { useAPI } from "@/helpers";
import {
  AdminAccess,
  MutationCallbacks,
  MutationError,
  OrganizationMemberWithOrganizationAndRole,
  User,
} from "@/types";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export interface GetCurrentUserResponse {
  user: User;
  adminAccess: AdminAccess | null;
  roles: OrganizationMemberWithOrganizationAndRole[] | null;
}

export const useGetCurrentUserQuery = (enabled = true) => {
  const fetchWrapper = useAPI();
  return useQuery<GetCurrentUserResponse>({
    queryKey: [USER_QUERY_KEYS.CURRENT],
    queryFn: () =>
      fetchWrapper({
        endpoint: "user",
        method: "GET",
      }) as Promise<GetCurrentUserResponse>,
    enabled,
  });
};

export interface UpdateProfileRequest {
  firstName?: string;
  lastName?: string;
  avatarURL?: string;
}

export const useUpdateProfileMutation = (
  callbacks: MutationCallbacks<User> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<User, MutationError, UpdateProfileRequest>({
    mutationFn: (body: UpdateProfileRequest) =>
      fetchWrapper({
        endpoint: "user/profile",
        method: "PATCH",
        body,
      }) as Promise<User>,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: [USER_QUERY_KEYS.CURRENT] });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export interface UpdateLanguageRequest {
  language: string;
}

export const useUpdateLanguageMutation = (
  callbacks: MutationCallbacks<User> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<User, MutationError, UpdateLanguageRequest>({
    mutationFn: (body: UpdateLanguageRequest) =>
      fetchWrapper({
        endpoint: "user/language",
        method: "PATCH",
        body,
      }) as Promise<User>,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: [USER_QUERY_KEYS.CURRENT] });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export interface ChangePasswordRequest {
  currentPassword: string;
  newPassword: string;
}

export const useChangePasswordMutation = (
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  return useMutation<void, MutationError, ChangePasswordRequest>({
    mutationFn: (body: ChangePasswordRequest) =>
      fetchWrapper({
        endpoint: "user/password",
        method: "PATCH",
        body,
      }) as Promise<void>,
    onSuccess: callbacks.onSuccess,
    onError: callbacks.onError,
  });
};

export const useResendVerificationEmailMutation = (
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  return useMutation<void, MutationError, void>({
    mutationFn: () =>
      fetchWrapper({
        endpoint: "user/resend-verification-email",
        method: "POST",
      }) as Promise<void>,
    onSuccess: callbacks.onSuccess,
    onError: callbacks.onError,
  });
};

export interface AcceptTermsRequest {
  termsVersion: string;
}

export const useAcceptTermsMutation = (
  callbacks: MutationCallbacks<User> = {},
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<User, MutationError, AcceptTermsRequest>({
    mutationFn: (body: AcceptTermsRequest) =>
      fetchWrapper({
        endpoint: "user/terms",
        method: "PATCH",
        body,
      }) as Promise<User>,
    onSuccess: (data) => {
      queryClient.setQueryData([USER_QUERY_KEYS.CURRENT], (old: unknown) =>
        old && typeof old === "object" && "user" in old
          ? { ...old, user: data }
          : old,
      );
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};
