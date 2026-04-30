/**
 * Allowed hostnames for billing portal redirects (e.g. Paddle).
 * Reduces open-redirect risk when using API-returned portal URLs.
 */
const ALLOWED_PORTAL_HOSTS = [
  "portal.paddle.com",
  "checkout.paddle.com",
  "buy.paddle.com",
  "cdn.paddle.com",
  "vendors.paddle.com",
];

/**
 * Returns true if url is a valid https URL whose host is in the allowlist.
 * Use before assigning to window.location.href to mitigate open-redirect.
 */
export function isAllowedPortalUrl(url: string): boolean {
  if (!url || typeof url !== "string") return false;
  const trimmed = url.trim();
  if (!trimmed.startsWith("https://")) return false;
  try {
    const { hostname } = new URL(trimmed);
    return ALLOWED_PORTAL_HOSTS.some(
      (allowed) => hostname === allowed || hostname.endsWith(`.${allowed}`),
    );
  } catch {
    return false;
  }
}
