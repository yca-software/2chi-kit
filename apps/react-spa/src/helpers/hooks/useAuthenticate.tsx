import { setTokens } from "@/helpers/cookie";
import { useUserState } from "@/states/user";

export const useAuthenticate = () => {
  const setAuthTokens = useUserState((state) => state.setTokens);

  return (accessToken: string, refreshToken: string) => {
    setTokens(accessToken, refreshToken);
    setAuthTokens({ access: accessToken, refresh: refreshToken });
  };
};
