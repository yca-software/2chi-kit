import { useUserState } from "@/states";
import { useShallow } from "zustand/shallow";
import {
  getRefreshTokenFromCookies,
  removeAccessTokenCookie,
  removeRefreshTokenCookie,
  setAccessTokenCookie,
} from "../cookie";
import { RequestConfig, RequestError } from "./types";
import { API_URL } from "@/constants";
import { getRequestOptions } from "./request";

export const useRefreshAccessToken = () => {
  const { reset, setIsRefreshingAccessToken, setAccessToken } = useUserState(
    useShallow((state) => ({
      reset: state.reset,
      setIsRefreshingAccessToken: state.setIsRefreshingAccessToken,
      setAccessToken: state.setAccessToken,
    })),
  );

  const forceSignOut = () => {
    reset();
    removeAccessTokenCookie();
    removeRefreshTokenCookie();
    window.location.href = "/";
  };

  return () => {
    setIsRefreshingAccessToken(true);
    const refreshToken = getRefreshTokenFromCookies();
    const requestConfig: RequestConfig = {
      endpoint: "auth/refresh",
      method: "POST",
      body: { refreshToken },
    };

    return fetch(
      `${API_URL}/${requestConfig.endpoint}`,
      getRequestOptions(requestConfig, null),
    )
      .then((response) => {
        return response
          .text()
          .then((responseText) => ({ response, responseText }));
      })
      .then(({ response, responseText }) => {
        let data: { accessToken?: string } | null = null;
        if (responseText) {
          try {
            data = JSON.parse(responseText);
          } catch {
            setIsRefreshingAccessToken(false);
            forceSignOut();
            return Promise.reject({
              error: new Error("invalid refresh response"),
              status: 401,
              retry: false,
            } as RequestError);
          }
        }
        if (!response.ok) {
          setIsRefreshingAccessToken(false);
          forceSignOut();
          return Promise.reject({
            error: new Error("failed to refresh access token"),
            status: 401,
            retry: false,
          } as RequestError);
        }
        if (data?.accessToken) {
          setAccessTokenCookie(data.accessToken);
          setAccessToken(data.accessToken);
        }
        setIsRefreshingAccessToken(false);
        return Promise.resolve((data?.accessToken ?? "") as string);
      })
      .finally(() => {
        setIsRefreshingAccessToken(false);
      });
  };
};
