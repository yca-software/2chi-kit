export * from "./queryKeys";
export * from "./legal";
export * from "./permissions";
export * from "./pricing";
export * from "./breakpoints";
export * from "./constraints";

export const API_URL =
  import.meta.env.VITE_API_URL || "http://localhost:1337/api/v1";
export const APP_NAME = import.meta.env.VITE_APP_NAME || "";
export const ENV = import.meta.env.VITE_APP_ENV || "local";
export const COOKIE_DOMAIN = import.meta.env.VITE_APP_COOKIE_DOMAIN || "";

/* Languages */
export const LANGUAGES = {
  en: "English",
};
export const DEFAULT_LANGUAGE = "en";

/* Cookies */
export const ACCESS_TOKEN_COOKIE_NAME = `@${APP_NAME}-${ENV}/access-token`;
export const REFRESH_TOKEN_COOKIE_NAME = `@${APP_NAME}-${ENV}/refresh-token`;
export const COOKIE_CONSENT_COOKIE_NAME = `@${APP_NAME}-${ENV}/cookie-consent`;

/* Google OAuth */
export const GOOGLE_CLIENT_ID =
  import.meta.env.VITE_OAUTH_GOOGLE_CLIENT_ID || "";
export const GOOGLE_REDIRECT_URI = `${window.location.origin}/oauth/google`;

/* Paddle */
export const PADDLE_CLIENT_TOKEN =
  import.meta.env.VITE_PADDLE_CLIENT_TOKEN || "";
export const PADDLE_ENVIRONMENT =
  import.meta.env.VITE_PADDLE_ENVIRONMENT || "sandbox";
