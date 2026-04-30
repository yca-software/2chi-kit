/**
 * Must stay aligned with `persist({ name })` in `states/user.ts`.
 * Used to read `selectedOrgId` before Zustand has finished rehydrating.
 */
const USER_STATE_STORAGE_NAME = "user-storage";

function readPersistedSelectedOrgId(): string | null {
  if (typeof localStorage === "undefined") return null;
  try {
    const raw = localStorage.getItem(USER_STATE_STORAGE_NAME);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as {
      state?: { selectedOrgId?: unknown };
    };
    const id = parsed.state?.selectedOrgId;
    return typeof id === "string" && id.length > 0 ? id : null;
  } catch {
    return null;
  }
}

/**
 * Picks an org for the user: last selected (from memory or persisted storage) if still in `roles`,
 * otherwise the first role. `roles` is assumed non-empty when a choice is required.
 */
export function resolveDefaultOrganizationId(
  roles: readonly { organizationId: string }[],
  inMemorySelectedOrgId: string | null,
): string | null {
  if (!roles.length) return null;
  if (
    inMemorySelectedOrgId &&
    roles.some((r) => r.organizationId === inMemorySelectedOrgId)
  ) {
    return inMemorySelectedOrgId;
  }
  const fromStorage = readPersistedSelectedOrgId();
  if (fromStorage && roles.some((r) => r.organizationId === fromStorage)) {
    return fromStorage;
  }
  return roles[0]?.organizationId ?? null;
}
