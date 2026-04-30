export function getOrgIdFromEndpoint(endpoint: string): string | null {
  const parts = endpoint.split("/");
  if (parts[0] === "organization" && parts[1]) return parts[1];
  return null;
}
