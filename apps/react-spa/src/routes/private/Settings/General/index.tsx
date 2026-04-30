import { Navigate, useParams } from "react-router";
import { Loader2 } from "lucide-react";
import { GeneralInfoDisplay } from "./GeneralInfoDisplay";
import { GeneralEditDrawer } from "./GeneralEditDrawer";
import { SubscriptionCard } from "./SubscriptionCard";
import {
  useSettingsPermissions,
  SETTINGS_SECTIONS,
} from "../useSettingsPermissions";
import {
  useGetOrganizationQuery,
  useListOrganizationMembersQuery,
} from "@/api";
import { useState } from "react";

export const GeneralSettings = () => {
  const { orgId } = useParams<{ orgId: string }>();
  const [editDrawerOpen, setEditDrawerOpen] = useState(false);

  const permissions = useSettingsPermissions();

  const orgIdStr = orgId || "";
  const { data: organization, isLoading: orgLoading } =
    useGetOrganizationQuery(orgIdStr);
  const { data: members = [], isLoading: membersLoading } =
    useListOrganizationMembersQuery(orgIdStr);

  const showOrg = permissions.org.canRead;
  const showSubscription = permissions.subscription.canRead;

  const memberCount = members.length;
  const isPageLoading = orgLoading || membersLoading;

  if (!permissions.generalMerged.canRead) {
    const firstAllowed = SETTINGS_SECTIONS.find(
      (s) => permissions[s.permissionKey]?.canRead,
    );
    return firstAllowed ? (
      <Navigate to={`/settings/${orgId}/${firstAllowed.path}`} replace />
    ) : (
      <Navigate to={`/settings/${orgId}`} replace />
    );
  }

  if (isPageLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="flex min-h-0 flex-col gap-6">
      {showOrg ? (
        <div className="grid gap-4">
          <GeneralInfoDisplay
            organization={organization}
            onEdit={() => setEditDrawerOpen(true)}
            canEdit={permissions.org.canWrite}
          />
          <GeneralEditDrawer
            open={editDrawerOpen}
            onOpenChange={setEditDrawerOpen}
            organization={organization}
          />
        </div>
      ) : null}

      {showSubscription ? (
        <SubscriptionCard
          organization={organization}
          memberCount={memberCount}
        />
      ) : null}
    </div>
  );
};
