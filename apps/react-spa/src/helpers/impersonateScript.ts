import {
  ACCESS_TOKEN_COOKIE_NAME,
  REFRESH_TOKEN_COOKIE_NAME,
  COOKIE_DOMAIN,
} from "@/constants";

const IMPERSONATE_SESSION_MAX_AGE_SEC = 3600; // 1 hour

const USER_STORAGE_KEY = "user-storage";

/**
 * Builds a script string that when pasted and run in the browser console
 * sets the access and refresh token cookies, restores user-storage so the app
 * rehydrates with tokens and fetches the user, then reloads the page.
 */
export function buildImpersonateScript(
  accessToken: string,
  refreshToken: string
): string {
  const domainPart = COOKIE_DOMAIN ? `;domain=${COOKIE_DOMAIN}` : "";
  const maxAge = `;max-age=${IMPERSONATE_SESSION_MAX_AGE_SEC}`;
  const a = accessToken.replace(/\\/g, "\\\\").replace(/'/g, "\\'");
  const r = refreshToken.replace(/\\/g, "\\\\").replace(/'/g, "\\'");

  const state = {
    state: {
      accessInfoFromToken: null,
      userData: { user: null, admin: null, roles: null },
      tokens: { access: accessToken, refresh: refreshToken },
      isRefreshingAccessToken: false,
      selectedOrgId: null,
    },
    version: 1,
  };
  const stateStr = JSON.stringify(state).replace(/\\/g, "\\\\").replace(/'/g, "\\'");

  return `document.cookie='${ACCESS_TOKEN_COOKIE_NAME}=${a};path=/${maxAge}${domainPart}';document.cookie='${REFRESH_TOKEN_COOKIE_NAME}=${r};path=/${maxAge}${domainPart}';localStorage.setItem('${USER_STORAGE_KEY}','${stateStr}');location.reload()`;
}
