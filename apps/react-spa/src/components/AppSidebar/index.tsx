import { useState } from "react";
import { useUserState } from "@/states";
import { useShallow } from "zustand/shallow";
import { useSignOut, useTranslationNamespace } from "@/helpers";
import { UserSettingsDrawer } from "@/components/AppSidebar/UserSettingsDrawer";
import { SupportForm } from "./SupportForm";
import { AppSidebarHeader } from "./AppSidebarHeader";
import { AppSidebarNavigation } from "./AppSidebarNavigation";
import { AppSidebarFooter } from "./AppSidebarFooter";

export interface AppSidebarProps {
  onClose?: () => void;
  collapsed?: boolean;
}

export function AppSidebar({ onClose, collapsed = false }: AppSidebarProps) {
  const { isLoading } = useTranslationNamespace(["settings", "common"]);
  const { userData } = useUserState(
    useShallow((state) => ({ userData: state.userData })),
  );
  const signOut = useSignOut();
  const [isUserSettingsOpen, setIsUserSettingsOpen] = useState(false);
  const [supportFormOpen, setSupportFormOpen] = useState(false);

  const user = userData.user;

  if (isLoading) {
    return null;
  }

  return (
    <div className="flex h-full w-full min-w-0 flex-col">
      <AppSidebarHeader collapsed={collapsed} onClose={onClose} />
      <AppSidebarNavigation
        collapsed={collapsed}
        onClose={onClose}
        onOpenSupport={() => setSupportFormOpen(true)}
      />
      <AppSidebarFooter
        collapsed={collapsed}
        user={user}
        onOpenUserSettings={() => setIsUserSettingsOpen(true)}
        onSignOut={signOut}
      />

      <SupportForm open={supportFormOpen} onOpenChange={setSupportFormOpen} />
      <UserSettingsDrawer
        open={isUserSettingsOpen}
        onOpenChange={setIsUserSettingsOpen}
      />
    </div>
  );
}
