import { useGetCurrentUserQuery, useUpdateLanguageMutation } from "@/api";
import {
  getBrowserLanguage,
  getStoredLanguage,
  resolveDefaultOrganizationId,
  setStoredLanguage,
} from "@/helpers";
import { useUserState } from "@/states";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Outlet } from "react-router";
import { useShallow } from "zustand/shallow";

export const Root = () => {
  const { i18n } = useTranslation();

  const {
    tokens,
    setUserData,
    setSelectedOrgId,
    setUserProfileReady,
    userData,
  } = useUserState(
    useShallow((state) => ({
      tokens: state.tokens,
      setUserData: state.setUserData,
      setSelectedOrgId: state.setSelectedOrgId,
      setUserProfileReady: state.setUserProfileReady,
      userData: state.userData,
    })),
  );

  const {
    data: userProfile,
    isLoading,
    isFetched,
    isSuccess,
  } = useGetCurrentUserQuery(!!tokens.refresh);

  const updateLanguageMutation = useUpdateLanguageMutation({
    onSuccess: (updatedUser) => {
      useUserState.getState().setUserData({
        ...userData,
        user: updatedUser,
      });
    },
  });

  useEffect(() => {
    const userLanguage = userData.user?.language || userProfile?.user?.language;
    const storedLanguage = getStoredLanguage();

    // Priority: user language > stored language > current i18n language
    const targetLanguage = userLanguage || storedLanguage;

    if (targetLanguage && i18n.language !== targetLanguage) {
      i18n.changeLanguage(targetLanguage);
      const namespaces = i18n.options.ns || [];
      i18n.reloadResources(targetLanguage, namespaces);

      if (userLanguage) {
        setStoredLanguage(userLanguage);
      }
    }

    if (isSuccess && userProfile?.user && !userProfile.user.language) {
      const browserLanguage = getBrowserLanguage();
      if (browserLanguage) {
        updateLanguageMutation.mutate({ language: browserLanguage });
      }
    }
  }, [
    userData.user?.language,
    userProfile?.user?.language,
    i18n,
    isSuccess,
    userProfile,
  ]);

  useEffect(() => {
    if (userProfile && isSuccess) {
      setUserData({
        user: userProfile.user,
        admin: userProfile.adminAccess,
        roles: userProfile.roles,
      });

      if (userProfile.roles && userProfile.roles.length > 0) {
        const currentSelectedOrgId = useUserState.getState().selectedOrgId;
        const resolved = resolveDefaultOrganizationId(
          userProfile.roles,
          currentSelectedOrgId,
        );
        if (resolved && resolved !== currentSelectedOrgId) {
          setSelectedOrgId(resolved);
        }
      }
    }
  }, [userProfile, isSuccess, setUserData, setSelectedOrgId]);

  const isReady = (isFetched && !isLoading) || !tokens.refresh;
  useEffect(() => {
    setUserProfileReady(isReady);
  }, [isReady, setUserProfileReady]);

  return <Outlet />;
};
