import {
  removeAccessTokenCookie,
  removeRefreshTokenCookie,
} from "@/helpers/cookie";
import { useUserState } from "@/states/user";
import { useLogoutMutation } from "@/api/auth";
import { useQueryClient } from "@tanstack/react-query";

const clearSession = (
  reset: () => void,
  queryClient: ReturnType<typeof useQueryClient>,
) => {
  reset();
  removeAccessTokenCookie();
  removeRefreshTokenCookie();
  queryClient.invalidateQueries();
  // Navigation will be handled by ProtectedRoute when it detects no tokens
};

export const useSignOut = () => {
  const reset = useUserState((state) => state.reset);
  const tokens = useUserState((state) => state.tokens);
  const queryClient = useQueryClient();

  const logoutMutation = useLogoutMutation({
    onSuccess: () => clearSession(reset, queryClient),
    onError: () => clearSession(reset, queryClient),
  });

  return () => {
    if (tokens?.refresh) {
      logoutMutation.mutate({ refreshToken: tokens.refresh });
    } else {
      clearSession(reset, queryClient);
    }
  };
};
