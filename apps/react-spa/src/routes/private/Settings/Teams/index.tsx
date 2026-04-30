import {
  getSubscriptionCapabilities,
  useTranslationNamespace,
} from "@/helpers";
import { Team } from "@/types";
import { useState } from "react";
import { Navigate, useParams } from "react-router";
import {
  SETTINGS_SECTIONS,
  useSettingsPermissions,
} from "../useSettingsPermissions";
import { useGetOrganizationQuery, useListTeamsQuery } from "@/api";
import { Loader2, Plus, Users2 } from "lucide-react";
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@yca-software/design-system";
import { TeamRow } from "./TeamRow";
import { usePricingModalStore } from "@/states";
import { useShallow } from "zustand/react/shallow";
import { CreateTeamDrawer } from "./CreateTeamDrawer";
import { DeleteTeamDialog } from "./DeleteTeamDialog";
import { TeamDetailsDrawer } from "./TeamDetailsDrawer";

type TeamsUIState = {
  selectedTeam: Team | null;
  drawerMode: string; // "view", "edit", "create"
  isDeleteTeamDialogOpen: boolean;
};

export const TeamsSettings = () => {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const { orgId } = useParams<{ orgId: string }>();
  const { openForOrg } = usePricingModalStore(
    useShallow((state) => ({
      openForOrg: state.openForOrg,
    })),
  );

  const [state, setState] = useState<TeamsUIState>({
    selectedTeam: null,
    drawerMode: "",
    isDeleteTeamDialogOpen: false,
  });

  const permissions = useSettingsPermissions();

  const currentOrgId = orgId || "";

  const { data: organization } = useGetOrganizationQuery(currentOrgId);
  const capabilities = getSubscriptionCapabilities(organization, null);
  const canManageTeams =
    permissions.team.canWrite && capabilities.canManageTeams;
  const canDeleteTeams =
    permissions.team.canDelete && capabilities.canManageTeams;

  const { data: teamsData, isLoading } = useListTeamsQuery(currentOrgId);
  const teams = teamsData ?? [];

  if (!permissions.team.canRead) {
    const firstAllowed = SETTINGS_SECTIONS.find(
      (s) => permissions[s.permissionKey]?.canRead,
    );
    return firstAllowed ? (
      <Navigate to={`/settings/${orgId}/${firstAllowed.path}`} replace />
    ) : (
      <Navigate to={`/settings/${orgId}`} replace />
    );
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <CreateTeamDrawer
        open={state.drawerMode === "create"}
        onClose={() => {
          setState((s) => ({
            ...s,
            drawerMode: "",
          }));
        }}
        currentOrgId={currentOrgId}
      />

      <DeleteTeamDialog
        currentOrgId={currentOrgId}
        open={state.isDeleteTeamDialogOpen && !!state.selectedTeam}
        selectedTeam={state.selectedTeam}
        onClose={() =>
          setState((s) => ({
            ...s,
            isDeleteTeamDialogOpen: false,
          }))
        }
        onSuccess={() => {
          setState((s) => ({
            ...s,
            isDeleteTeamDialogOpen: false,
            drawerMode: "",
            selectedTeam: null,
          }));
        }}
      />

      <TeamDetailsDrawer
        currentOrgId={currentOrgId}
        open={
          ["view", "edit"].includes(state.drawerMode) && !!state.selectedTeam
        }
        onClose={() => {
          setState((s) => ({
            ...s,
            selectedTeam: null,
            drawerMode: "",
          }));
        }}
        onEditClose={() => {
          setState((s) => ({
            ...s,
            drawerMode: "view",
          }));
        }}
        selectedTeam={state.selectedTeam}
        onTeamUpdated={(team) => {
          setState((s) => ({
            ...s,
            selectedTeam: team,
            drawerMode: "view",
          }));
        }}
        mode={state.drawerMode}
        canEdit={canManageTeams}
        canDelete={canDeleteTeams}
        onEditClick={(team) => {
          setState((s) => ({
            ...s,
            selectedTeam: team,
            drawerMode: "edit",
          }));
        }}
        onDeleteClick={(team) => {
          setState((s) => ({
            ...s,
            selectedTeam: team,
            isDeleteTeamDialogOpen: true,
          }));
        }}
      />

      <Card>
        <CardHeader>
          <div className="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="min-w-0">
              <CardTitle>{t("settings:org.teams.title")}</CardTitle>
              <CardDescription>
                {t("settings:org.teams.description")}
              </CardDescription>
            </div>
            <div className="flex shrink-0 flex-wrap items-center gap-2">
              {canManageTeams && (
                <Button
                  size="sm"
                  onClick={() => {
                    setState((s) => ({
                      ...s,
                      drawerMode: "create",
                    }));
                  }}
                >
                  <Plus className="mr-2 h-4 w-4" />
                  {t("settings:org.teams.create")}
                </Button>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {!capabilities.canManageTeams ? (
            <div className="mb-4 rounded-lg border border-dashed border-primary/40 bg-primary/5 p-4 text-left">
              <h3 className="text-sm font-semibold">
                {t("settings:org.upsell.teamsTitle")}
              </h3>
              <p className="mt-1 text-xs text-muted-foreground">
                {t("settings:org.upsell.teamsDescription")}
              </p>
              <Button
                variant="outline"
                size="sm"
                className="mt-3"
                onClick={() => openForOrg(currentOrgId)}
              >
                {t("settings:org.upsell.viewPlans")}
              </Button>
            </div>
          ) : (
            <>
              {teams.length > 0 ? (
                <div className="space-y-2">
                  {teams.map((team: Team) => (
                    <TeamRow
                      key={team.id}
                      team={team}
                      onSelect={(team) => {
                        setState((s) => ({
                          ...s,
                          selectedTeam: team,
                          drawerMode: "view",
                        }));
                      }}
                      onEdit={(team) => {
                        setState((s) => ({
                          ...s,
                          selectedTeam: team,
                          drawerMode: "edit",
                        }));
                      }}
                      onDelete={(team) => {
                        setState((s) => ({
                          ...s,
                          selectedTeam: team,
                          drawerMode: "",
                          isDeleteTeamDialogOpen: true,
                        }));
                      }}
                      canEdit={canManageTeams}
                      canDelete={canDeleteTeams}
                    />
                  ))}
                </div>
              ) : (
                <div className="rounded-lg border border-dashed bg-muted/20 px-6 py-10 text-center">
                  <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary">
                    <Users2 className="h-6 w-6" />
                  </div>
                  <h3 className="mt-4 text-base font-semibold">
                    {t("settings:org.teams.noTeamsTitle")}
                  </h3>
                  <p className="mx-auto mt-2 max-w-md text-sm text-muted-foreground">
                    {t("settings:org.teams.noTeamsDescription")}
                  </p>
                  {canManageTeams && (
                    <Button
                      size="sm"
                      className="mt-5"
                      onClick={() => {
                        setState((s) => ({
                          ...s,
                          drawerMode: "create",
                        }));
                      }}
                    >
                      <Plus className="mr-2 h-4 w-4" />
                      {t("settings:org.teams.create")}
                    </Button>
                  )}
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
};
