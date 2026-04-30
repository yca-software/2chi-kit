export function getConstraintNameFromError(error: unknown): string {
  if (!error || typeof error !== "object") return "";

  if ("extra" in error) {
    const extra = (error as { extra?: unknown }).extra;
    if (extra && typeof extra === "object" && "constraint_name" in extra) {
      const constraintName = (extra as { constraint_name?: unknown })
        .constraint_name;
      return typeof constraintName === "string" ? constraintName : "";
    }
    return "";
  }

  if ("error" in error) {
    return getConstraintNameFromError((error as { error?: unknown }).error);
  }

  return "";
}
