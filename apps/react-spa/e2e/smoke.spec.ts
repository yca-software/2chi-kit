import { test, expect } from "@playwright/test";

test.describe("Smoke", () => {
  test("sign-in page loads", async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveURL(/\//);
    await expect(page.locator("body")).toBeVisible();
  });

  test("sign-up route is reachable", async ({ page }) => {
    await page.goto("/signup");
    await expect(page).toHaveURL(/\/signup/);
    await expect(page.locator("body")).toBeVisible();
  });
});
