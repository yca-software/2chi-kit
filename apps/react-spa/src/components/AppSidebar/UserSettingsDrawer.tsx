import { useState } from "react";
import { useUserState } from "@/states";
import { useShallow } from "zustand/shallow";
import { useUpdateProfileMutation, useChangePasswordMutation } from "@/api";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from "@yca-software/design-system";
import { Button } from "@yca-software/design-system";
import { Separator } from "@yca-software/design-system";
import { Key, Shield, Pencil } from "lucide-react";
import { useTranslationNamespace } from "@/helpers";
import { USER_QUERY_KEYS } from "@/constants";
import { ProfileForm } from "./ProfileForm";
import { ProfileDisplay } from "./ProfileDisplay";
import { PasswordForm } from "./PasswordForm";
import { SessionsSection } from "./SessionsSection";

interface UserSettingsDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function UserSettingsDrawer({
  open,
  onOpenChange,
}: UserSettingsDrawerProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const { userData } = useUserState(
    useShallow((state) => ({
      userData: state.userData,
    })),
  );
  const user = userData.user;
  const queryClient = useQueryClient();

  const [uiState, setUIState] = useState({
    isEditingProfile: false,
    profileSuccess: false,
    showPasswordForm: false,
    passwordSuccess: false,
  });

  const updateProfileMutation = useUpdateProfileMutation({
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [USER_QUERY_KEYS.CURRENT],
      });
      setUIState((prev) => ({ ...prev, profileSuccess: true }));
      setTimeout(() => {
        setUIState((prev) => ({
          ...prev,
          profileSuccess: false,
          isEditingProfile: false,
        }));
      }, 1500);
    },
    onError: () => {
      toast.error(t("common:defaultError"));
    },
  });

  const changePasswordMutation = useChangePasswordMutation({
    onSuccess: () => {
      setUIState((prev) => ({ ...prev, passwordSuccess: true }));
      setTimeout(() => {
        setUIState((prev) => ({
          ...prev,
          passwordSuccess: false,
          showPasswordForm: false,
        }));
      }, 1500);
    },
    onError: () => {
      toast.error(t("common:defaultError"));
    },
  });

  const handleProfileSubmit = (data: {
    firstName: string;
    lastName: string;
  }) => {
    updateProfileMutation.mutate({
      firstName: data.firstName,
      lastName: data.lastName,
    });
  };

  const handlePasswordSubmit = (data: {
    currentPassword: string;
    newPassword: string;
  }) => {
    changePasswordMutation.mutate({
      currentPassword: data.currentPassword,
      newPassword: data.newPassword,
    });
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-full sm:max-w-xl p-0 flex flex-col overflow-hidden">
        <SheetHeader className="shrink-0 px-6 pt-6 pb-2">
          <SheetTitle className="text-xl">
            {t("settings:user.title")}
          </SheetTitle>
          <SheetDescription>{t("settings:user.description")}</SheetDescription>
        </SheetHeader>

        <div className="px-6 pb-6 min-w-0 flex-1 min-h-0 overflow-y-auto">
          <div className="space-y-6">
            {/* Profile Section */}
            <section>
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                  {t("settings:user.profile")}
                </h3>
                {!uiState.isEditingProfile && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() =>
                      setUIState((prev) => ({
                        ...prev,
                        isEditingProfile: true,
                      }))
                    }
                    className="h-8 px-2"
                  >
                    <Pencil className="h-4 w-4 mr-1.5" />
                    {t("common:edit")}
                  </Button>
                )}
              </div>
              {uiState.isEditingProfile ? (
                <ProfileForm
                  user={user}
                  onSubmit={handleProfileSubmit}
                  onCancel={() =>
                    setUIState((prev) => ({
                      ...prev,
                      isEditingProfile: false,
                    }))
                  }
                  isPending={updateProfileMutation.isPending}
                  isSuccess={uiState.profileSuccess}
                />
              ) : (
                <ProfileDisplay user={user} />
              )}
            </section>

            <Separator />

            {/* Security Section */}
            <section>
              <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider mb-3 flex items-center gap-2">
                <Shield className="h-4 w-4" />
                {t("settings:user.security")}
              </h3>

              {uiState.showPasswordForm ? (
                <PasswordForm
                  onSubmit={handlePasswordSubmit}
                  onCancel={() =>
                    setUIState((prev) => ({
                      ...prev,
                      showPasswordForm: false,
                    }))
                  }
                  isPending={changePasswordMutation.isPending}
                  isSuccess={uiState.passwordSuccess}
                />
              ) : (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    setUIState((prev) => ({
                      ...prev,
                      showPasswordForm: true,
                    }))
                  }
                  className="w-full justify-start"
                >
                  <Key className="mr-2 h-4 w-4" />
                  {t("settings:user.changePassword")}
                </Button>
              )}

              {user?.id && <SessionsSection userId={user.id} />}
            </section>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}
