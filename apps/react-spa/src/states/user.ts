import { AccessInfo, getAccessInfoFromToken } from "@/helpers";
import {
  AdminAccess,
  OrganizationMemberWithOrganizationAndRole,
  User,
} from "@/types";
import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface UserState {
  /** From JWT: userId, email, permissions, isAdmin. Available as soon as we have an access token. */
  accessInfoFromToken: AccessInfo | null;
  userData: {
    user: User | null;
    admin: AdminAccess | null;
    roles: OrganizationMemberWithOrganizationAndRole[] | null;
  };
  tokens: {
    access: string;
    refresh: string;
  };
  isRefreshingAccessToken: boolean;
  selectedOrgId: string | null;
  /** True when get-current-user has finished (or when not authenticated). Used for route guards. */
  isUserProfileReady: boolean;

  reset: () => void;
  setUserProfileReady: (ready: boolean) => void;
  setUserData: (userData: {
    user: User | null;
    admin: AdminAccess | null;
    roles: OrganizationMemberWithOrganizationAndRole[] | null;
  }) => void;
  setAccessToken: (accessToken: string) => void;
  setTokens: (tokens: { access: string; refresh: string }) => void;
  setIsRefreshingAccessToken: (isRefreshingAccessToken: boolean) => void;
  setSelectedOrgId: (selectedOrgId: string | null) => void;
}

export const useUserState = create<UserState>()(
  persist(
    (set) => ({
      accessInfoFromToken: null,
      userData: {
        user: null,
        admin: null,
        roles: null,
      },
      tokens: {
        access: "",
        refresh: "",
      },
      isRefreshingAccessToken: false,
      selectedOrgId: null,
      isUserProfileReady: false,

      reset: () =>
        set({
          accessInfoFromToken: null,
          userData: {
            user: null,
            admin: null,
            roles: null,
          },
          tokens: {
            access: "",
            refresh: "",
          },
          isRefreshingAccessToken: false,
          selectedOrgId: null,
          isUserProfileReady: false,
        }),
      setUserProfileReady: (isUserProfileReady: boolean) =>
        set({ isUserProfileReady }),
      setUserData: (userData: {
        user: User | null;
        admin: AdminAccess | null;
        roles: OrganizationMemberWithOrganizationAndRole[] | null;
      }) => set({ userData }),
      setAccessToken: (accessToken: string) =>
        set((state) => ({
          tokens: { ...state.tokens, access: accessToken },
          accessInfoFromToken: getAccessInfoFromToken(accessToken),
        })),
      setTokens: (tokens: { access: string; refresh: string }) =>
        set({
          tokens,
          accessInfoFromToken: getAccessInfoFromToken(tokens.access),
        }),
      setIsRefreshingAccessToken: (isRefreshingAccessToken: boolean) =>
        set({ isRefreshingAccessToken }),
      setSelectedOrgId: (selectedOrgId: string | null) =>
        set({ selectedOrgId }),
    }),
    {
      name: "user-storage",
      version: 1,
      // Do not persist tokens to reduce XSS blast radius; tokens are in cookies and refreshed on load
      partialize: (state) => ({
        accessInfoFromToken: state.accessInfoFromToken,
        userData: state.userData,
        selectedOrgId: state.selectedOrgId,
      }),
      migrate: (persistedState, _version) => {
        // Persist passes the inner state; return as-is so old storage (version 0) loads without warning
        return (persistedState ?? {}) as typeof persistedState;
      },
    },
  ),
);
