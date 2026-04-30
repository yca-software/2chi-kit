import { JWTAccessTokenPermissionData } from "@/types";
import { jwtDecode } from "jwt-decode";

export const getInitials = (firstName?: string, lastName?: string) => {
  if (!firstName && !lastName) return "U";
  return `${firstName?.[0] || ""}${lastName?.[0] || ""}`.toUpperCase();
};

export interface DecodedAccessToken {
  sub: string;
  email?: string;
  permissions?: JWTAccessTokenPermissionData[];
  isAdmin?: boolean;
  impersonatedBy?: string | null;
  impersonatedByEmail?: string | null;
}

export interface AccessInfo {
  userId: string;
  email: string;
  permissions: JWTAccessTokenPermissionData[];
  isAdmin: boolean;
  impersonatedBy: string | null;
  impersonatedByEmail: string | null;
}

export function getAccessInfoFromToken(accessToken: string): AccessInfo | null {
  if (!accessToken?.trim()) return null;
  try {
    const decoded = jwtDecode<DecodedAccessToken>(accessToken);
    return {
      userId: decoded.sub,
      email: decoded.email ?? "",
      permissions: decoded.permissions ?? [],
      isAdmin: decoded.isAdmin ?? false,
      impersonatedBy: decoded.impersonatedBy ?? null,
      impersonatedByEmail: decoded.impersonatedByEmail ?? null,
    };
  } catch {
    return null;
  }
}
