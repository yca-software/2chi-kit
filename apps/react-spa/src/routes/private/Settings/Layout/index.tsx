import { useState, useMemo } from "react";
import { Outlet, useNavigate, useParams } from "react-router";
import { useTranslationNamespace } from "@/helpers";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { useUserState } from "@/states";
import { useShallow } from "zustand/shallow";
import type { MutationError } from "@/types";
import { useArchiveOrganizationMutation } from "@/api";
import { SettingsHeader } from "./SettingsHeader";
import { SettingsNav } from "./SettingsNav";
import { useSettingsPermissions } from "../useSettingsPermissions";

export function SettingsLayout() {
  const { t, isLoading } = useTranslationNamespace(["settings", "common"]);
  const navigate = useNavigate();
  const { orgId } = useParams<{ orgId: string }>();
  const { selectedOrgId, userData } = useUserState(
    useShallow((state) => ({
      selectedOrgId: state.selectedOrgId,
      userData: state.userData,
    })),
  );

  const organizations = userData.roles || [];
  const [archiveDialogOpen, setArchiveDialogOpen] = useState(false);

  const currentOrgId = orgId || selectedOrgId;
  const selectedOrg = useMemo(() => {
    return organizations.find((r) => r.organizationId === currentOrgId);
  }, [currentOrgId, organizations]);

  const archiveMutation = useArchiveOrganizationMutation(currentOrgId || "", {
    onSuccess: (freshUser) => {
      setArchiveDialogOpen(false);
      toast.success(t("settings:org.archiveSuccess"));
      // Use refetched roles (archived org already excluded by backend)
      const roles = freshUser?.roles ?? [];
      const otherOrg = roles.find((r) => r.organizationId !== currentOrgId);
      if (otherOrg) {
        useUserState.getState().setSelectedOrgId(otherOrg.organizationId);
        navigate(`/dashboard/${otherOrg.organizationId}`, { replace: true });
      } else {
        useUserState.getState().setSelectedOrgId(null);
        navigate("/dashboard", { replace: true });
      }
    },
    onError: (err: MutationError) => {
      setArchiveDialogOpen(false);
      toast.error(err.error?.message ?? t("common:defaultError"));
    },
  });

  const permissions = useSettingsPermissions();
  const canArchiveOrg = !!selectedOrg && permissions.org.canDelete;
  const hasOrganizations = organizations && organizations.length > 0;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!hasOrganizations) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold">{t("settings:org.title")}</h1>
          <p className="text-muted-foreground">
            {t("settings:org.description")}
          </p>
        </div>
        <div className="flex items-center justify-center py-12">
          <p className="text-muted-foreground">
            {t("settings:org.noOrganization")}
          </p>
        </div>
      </div>
    );
  }

  if (!currentOrgId) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">
          {t("settings:org.noOrganization")}
        </p>
      </div>
    );
  }

  return (
    <div className="min-w-0 space-y-6">
      <SettingsHeader
        organizationName={selectedOrg?.organizationName}
        canArchiveOrg={!!canArchiveOrg}
        archiveDialogOpen={archiveDialogOpen}
        setArchiveDialogOpen={setArchiveDialogOpen}
        onArchiveConfirm={() => archiveMutation.mutate()}
        isArchivePending={archiveMutation.isPending}
      />

      <SettingsNav currentOrgId={currentOrgId} />

      <div className="pt-2">
        <Outlet />
      </div>
    </div>
  );
}
