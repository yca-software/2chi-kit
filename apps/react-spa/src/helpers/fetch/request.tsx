import { RequestConfig } from "./types";
import i18n from "@/i18n";

export const getRequestHeaders = (
  accessToken: string | null,
  multipart?: boolean,
): HeadersInit | undefined => {
  const headers: HeadersInit = {};
  if (!multipart) {
    headers["Content-Type"] = "application/json";
  }
  if (accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`;
  }
  // Include Accept-Language header with current i18n language
  const currentLanguage = i18n.language || "en";
  headers["Accept-Language"] = currentLanguage;
  return headers;
};

export const getRequestOptions = (
  config: RequestConfig,
  token: string | null,
): RequestInit | undefined => {
  const requestHeaders = getRequestHeaders(token, config.multipart);
  const requestOptions = {
    method: config.method,
    headers: requestHeaders,
    body: config.multipart ? config.body : JSON.stringify(config.body),
  };
  return requestOptions;
};
