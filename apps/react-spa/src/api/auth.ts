import { posthog } from "@/analytics";
import { API_URL, REFRESH_TOKEN_QUERY_KEYS } from "@/constants";
import {
  getRefreshTokenFromCookies,
  setAccessTokenCookie,
  useAPI,
} from "@/helpers";
import { useUserState } from "@/states";
import { MutationCallbacks, MutationError } from "@/types";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { jwtDecode } from "jwt-decode";

export interface AuthenticateWithPasswordRequest {
  email: string;
  password: string;
}

export interface AuthenticateResponse {
  accessToken: string;
  refreshToken: string;
}

export const useAuthenticateWithPasswordMutation = (
  callbacks: MutationCallbacks<AuthenticateResponse>,
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<
    AuthenticateResponse,
    MutationError,
    AuthenticateWithPasswordRequest
  >({
    mutationFn: (body: AuthenticateWithPasswordRequest) =>
      fetchWrapper({
        endpoint: "auth/login",
        method: "POST",
        body,
      }) as Promise<AuthenticateResponse>,
    onSuccess: (data) => {
      const decoded = jwtDecode<{ sub: string }>(data.accessToken);
      posthog?.capture?.("login_success", { method: "password" });
      posthog?.identify?.(decoded.sub);
      queryClient.invalidateQueries({
        queryKey: [REFRESH_TOKEN_QUERY_KEYS.ACTIVE, decoded.sub],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export interface AuthenticateWithGoogleRequest {
  code: string;
  termsVersion: string;
  invitationToken?: string;
}

export const useAuthenticateWithGoogleMutation = (
  callbacks: MutationCallbacks<AuthenticateResponse>,
) => {
  const fetchWrapper = useAPI();
  const queryClient = useQueryClient();
  return useMutation<
    AuthenticateResponse,
    MutationError,
    AuthenticateWithGoogleRequest
  >({
    mutationFn: (body: AuthenticateWithGoogleRequest) =>
      fetchWrapper({
        endpoint: "auth/oauth/google",
        method: "POST",
        body,
      }) as Promise<AuthenticateResponse>,
    onSuccess: (data) => {
      const decoded = jwtDecode<{ sub: string }>(data.accessToken);
      posthog?.capture?.("login_success", { method: "google" });
      posthog?.identify?.(decoded.sub);
      queryClient.invalidateQueries({
        queryKey: [REFRESH_TOKEN_QUERY_KEYS.ACTIVE, decoded.sub],
      });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export interface RefreshAccessTokenResponse {
  accessToken: string;
}

export const useRefreshAccessTokenMutation = (
  callbacks: MutationCallbacks<RefreshAccessTokenResponse> = {},
) => {
  return useMutation<RefreshAccessTokenResponse, MutationError, void>({
    mutationFn: async () => {
      const refreshToken = getRefreshTokenFromCookies();
      if (!refreshToken) {
        throw new Error("No refresh token");
      }
      const res = await fetch(`${API_URL}/auth/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refreshToken }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        throw { error: data, status: res.status };
      }
      return { accessToken: data.accessToken };
    },
    onSuccess: (data) => {
      setAccessTokenCookie(data.accessToken);
      useUserState.getState().setAccessToken(data.accessToken);
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export interface LogoutRequest {
  refreshToken: string;
}

export const useLogoutMutation = (callbacks: MutationCallbacks<void> = {}) => {
  const fetchWrapper = useAPI();
  return useMutation<void, MutationError, LogoutRequest>({
    mutationFn: (body: LogoutRequest) =>
      fetchWrapper({
        endpoint: "auth/logout",
        method: "POST",
        body,
      }) as Promise<void>,
    onSuccess: callbacks.onSuccess,
    onError: callbacks.onError,
  });
};

export interface ForgotPasswordRequest {
  email: string;
}

export const useForgotPasswordMutation = (
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  return useMutation<void, MutationError, ForgotPasswordRequest>({
    mutationFn: (body: ForgotPasswordRequest) =>
      fetchWrapper({
        endpoint: "auth/forgot-password",
        method: "POST",
        body,
      }) as Promise<void>,
    onSuccess: callbacks.onSuccess,
    onError: callbacks.onError,
  });
};

export interface ResetPasswordRequest {
  token: string;
  password: string;
}

export const useResetPasswordMutation = (
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  return useMutation<void, MutationError, ResetPasswordRequest>({
    mutationFn: (body: ResetPasswordRequest) =>
      fetchWrapper({
        endpoint: "auth/reset-password",
        method: "POST",
        body,
      }) as Promise<void>,
    onSuccess: callbacks.onSuccess,
    onError: callbacks.onError,
  });
};

export interface SignUpRequest {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  language?: string;
  termsVersion: string;
  invitationToken?: string;
}

export interface SignUpResponse {
  accessToken: string;
  refreshToken: string;
}

export const useSignUpMutation = (
  callbacks: MutationCallbacks<SignUpResponse>,
) => {
  const fetchWrapper = useAPI();
  return useMutation<SignUpResponse, MutationError, SignUpRequest>({
    mutationFn: (body: SignUpRequest) =>
      fetchWrapper({
        endpoint: "auth/signup",
        method: "POST",
        body,
      }) as Promise<SignUpResponse>,
    onSuccess: (data) => {
      const decoded = jwtDecode<{ sub: string }>(data.accessToken);
      posthog?.capture?.("signup_success", { user_id: decoded.sub });
      callbacks.onSuccess?.(data);
    },
    onError: callbacks.onError,
  });
};

export interface VerifyEmailRequest {
  token: string;
}

export const useVerifyEmailMutation = (
  callbacks: MutationCallbacks<void> = {},
) => {
  const fetchWrapper = useAPI();
  return useMutation<void, MutationError, VerifyEmailRequest>({
    mutationFn: (body: VerifyEmailRequest) =>
      fetchWrapper({
        endpoint: "auth/verify-email",
        method: "POST",
        body,
      }) as Promise<void>,
    onSuccess: callbacks.onSuccess,
    onError: callbacks.onError,
  });
};
