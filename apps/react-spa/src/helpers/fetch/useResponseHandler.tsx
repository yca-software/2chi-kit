import { API_URL } from "@/constants";
import { getRefreshTokenFromCookies } from "../cookie";
import { RequestConfig, RequestError } from "./types";
import { useRefreshAccessToken } from "./useRefreshAccessToken";
import { getRequestOptions } from "./request";

export const useResponseHandler = () => {
  const refreshAccessToken = useRefreshAccessToken();

  function handleResponse(
    response: Response,
    options?: { config?: RequestConfig; isRetry?: boolean },
  ): Promise<any> {
    return response.text().then((responseText) => {
      let data;
      if (responseText) {
        try {
          data = JSON.parse(responseText);
        } catch (e) {
          data = responseText;
        }
      }

      if (!response.ok) {
        // Read cookies at 401 time — a value captured on an earlier render can be
        // stale right after login, causing us to skip refresh and surface a false
        // "session expired" for the first private API call.
        const refreshToken = getRefreshTokenFromCookies();
        if (
          response.status === 401 &&
          refreshToken &&
          options?.config &&
          !options?.isRetry
        ) {
          return refreshAccessToken().then((newToken) =>
            fetch(
              `${API_URL}/${options.config!.endpoint}`,
              getRequestOptions(options.config!, newToken),
            ).then((retryResponse) =>
              handleResponse(retryResponse, {
                ...options,
                isRetry: true,
              }),
            ),
          );
        }

        return Promise.reject({
          error: data,
          status: response.status,
          retry: response.status >= 500,
        } as RequestError);
      }

      return Promise.resolve(data);
    });
  }

  return handleResponse;
};
