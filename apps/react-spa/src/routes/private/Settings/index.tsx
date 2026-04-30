import { Navigate, useParams } from "react-router";
import { useUserState } from "@/states/user";
import { useShallow } from "zustand/shallow";
import {
  useSettingsPermissions,
  SETTINGS_SECTIONS,
} from "./useSettingsPermissions";

export const Settings = () => {
  const { orgId } = useParams<{ orgId: string }>();
  const { selectedOrgId } = useUserState(
    useShallow((state) => ({
      selectedOrgId: state.selectedOrgId,
    })),
  );

  const currentOrgId = orgId || selectedOrgId;

  if (!currentOrgId) {
    return <Navigate to="/dashboard" replace />;
  }

  const permissions = useSettingsPermissions();
  const firstAllowed = SETTINGS_SECTIONS.find(
    (s) => permissions[s.permissionKey]?.canRead,
  );

  if (!firstAllowed) {
    return <Navigate to="/dashboard" replace />;
  }

  return (
    <Navigate to={`/settings/${currentOrgId}/${firstAllowed.path}`} replace />
  );
};
