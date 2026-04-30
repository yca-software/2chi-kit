import { getAccessTokenFromCookies } from "../cookie";
import { useResponseHandler } from "./useResponseHandler";
import { RequestConfig } from "./types";
import { API_URL } from "@/constants";
import { getRequestOptions } from "./request";

export const useAPI = () => {
  const responseHandler = useResponseHandler();

  return (config: RequestConfig) => {
    const accessToken = getAccessTokenFromCookies();
    return fetch(
      `${API_URL}/${config.endpoint}`,
      getRequestOptions(config, accessToken),
    ).then((response) => responseHandler(response, { config }));
  };
};
