import { describe, it, expect, vi } from "vitest";

vi.mock("@/constants", () => ({
  ACCESS_TOKEN_COOKIE_NAME: "@test/access-token",
  REFRESH_TOKEN_COOKIE_NAME: "@test/refresh-token",
  COOKIE_DOMAIN: "",
}));

import { buildImpersonateScript } from "./impersonateScript";

describe("buildImpersonateScript", () => {
  it("returns a script string containing cookie and localStorage calls", () => {
    const script = buildImpersonateScript("access-123", "refresh-456");
    expect(script).toContain("document.cookie=");
    expect(script).toContain("localStorage.setItem(");
    expect(script).toContain("location.reload()");
  });

  it("escapes single quotes in tokens", () => {
    const script = buildImpersonateScript("token'with'quotes", "refresh");
    expect(script).toContain("token\\'with\\'quotes");
    expect(script).not.toMatch(/token'with'quotes/);
  });

  it("escapes backslashes in tokens", () => {
    const script = buildImpersonateScript("token\\with\\slashes", "refresh");
    expect(script).toContain("token\\\\with\\\\slashes");
  });

  it("includes max-age of 3600 seconds", () => {
    const script = buildImpersonateScript("a", "r");
    expect(script).toContain("max-age=3600");
  });

  it("includes user-storage key for localStorage", () => {
    const script = buildImpersonateScript("a", "r");
    expect(script).toContain("user-storage");
  });
});
