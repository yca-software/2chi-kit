import {
  ACCESS_TOKEN_COOKIE_NAME,
  COOKIE_CONSENT_COOKIE_NAME,
  COOKIE_DOMAIN,
  REFRESH_TOKEN_COOKIE_NAME,
} from "@/constants";
import type { CookieConsent } from "@/states/cookieConsent";
import Cookies from "js-cookie";

export const getAccessTokenFromCookies = (): string | null => {
  return Cookies.get(ACCESS_TOKEN_COOKIE_NAME) || null;
};

export const getRefreshTokenFromCookies = (): string | null => {
  return Cookies.get(REFRESH_TOKEN_COOKIE_NAME) || null;
};

export const getCookieConsentFromCookies = (): CookieConsent | null => {
  const consentStr = Cookies.get(COOKIE_CONSENT_COOKIE_NAME);
  if (!consentStr) return null;
  try {
    return JSON.parse(consentStr) as CookieConsent;
  } catch {
    return null;
  }
};

const getTokenCookieOptions = (): Cookies.CookieAttributes => {
  const opts: Cookies.CookieAttributes = { path: "/" };
  if (COOKIE_DOMAIN) opts.domain = COOKIE_DOMAIN;
  return opts;
};

export const removeAccessTokenCookie = () => {
  Cookies.remove(ACCESS_TOKEN_COOKIE_NAME, getTokenCookieOptions());
};

export const removeRefreshTokenCookie = () => {
  Cookies.remove(REFRESH_TOKEN_COOKIE_NAME, getTokenCookieOptions());
};

export const removeCookieConsentCookie = () => {
  Cookies.remove(COOKIE_CONSENT_COOKIE_NAME);
};

export const setCookieConsentCookie = (
  timestamp: string,
  analytics: boolean,
) => {
  const consent: CookieConsent = { analytics, timestamp };
  const cookieOptions: Cookies.CookieAttributes = {};
  if (COOKIE_DOMAIN) {
    cookieOptions.domain = COOKIE_DOMAIN;
  }
  Cookies.set(
    COOKIE_CONSENT_COOKIE_NAME,
    JSON.stringify(consent),
    cookieOptions,
  );
};

export const setAccessTokenCookie = (token: string) => {
  const cookieOptions: Cookies.CookieAttributes = {};
  if (COOKIE_DOMAIN) {
    cookieOptions.domain = COOKIE_DOMAIN;
  }
  Cookies.set(ACCESS_TOKEN_COOKIE_NAME, token, cookieOptions);
};

export const setRefreshTokenCookie = (token: string) => {
  const cookieOptions: Cookies.CookieAttributes = {};
  if (COOKIE_DOMAIN) {
    cookieOptions.domain = COOKIE_DOMAIN;
  }
  Cookies.set(REFRESH_TOKEN_COOKIE_NAME, token, cookieOptions);
};

export const setTokens = (accessToken: string, refreshToken: string) => {
  setAccessTokenCookie(accessToken);
  setRefreshTokenCookie(refreshToken);
};
