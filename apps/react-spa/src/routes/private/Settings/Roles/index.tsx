import { Navigate, useParams } from "react-router";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Button,
} from "@yca-software/design-system";
import { Loader2, Plus, Shield } from "lucide-react";
import { useGetOrganizationQuery, useListRolesQuery } from "@/api";
import {
  getSubscriptionCapabilities,
  useTranslationNamespace,
} from "@/helpers";
import { usePricingModalStore } from "@/states";
import { useShallow } from "zustand/shallow";
import { RoleRow } from "./RoleRow";
import { DeleteRoleDialog } from "./DeleteRoleDialog";
import { RoleDetailsDrawer } from "./RoleDetailsDrawer";
import {
  useSettingsPermissions,
  SETTINGS_SECTIONS,
} from "../useSettingsPermissions";
import { Role } from "@/types";
import { useState } from "react";
import { CreateRoleDrawer } from "./CreateRoleDrawer";

type RolesUIState = {
  selectedRole: Role | null;
  drawerMode: string; // "view", "edit", "create"
  isDeleteRoleDialogOpen: boolean;
};

export const RolesSettings = () => {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const { orgId } = useParams<{ orgId: string }>();
  const { openForOrg } = usePricingModalStore(
    useShallow((state) => ({
      openForOrg: state.openForOrg,
    })),
  );

  const [state, setState] = useState<RolesUIState>({
    selectedRole: null,
    drawerMode: "",
    isDeleteRoleDialogOpen: false,
  });

  const permissions = useSettingsPermissions();

  const { data: organization } = useGetOrganizationQuery(orgId || "");
  const capabilities = getSubscriptionCapabilities(organization, null);
  const canManageRoles =
    permissions.role.canWrite && capabilities.canManageRoles;
  const canDeleteRoles =
    permissions.role.canDelete && capabilities.canManageRoles;

  const currentOrgId = orgId || "";

  const { data: rolesData, isLoading } = useListRolesQuery(currentOrgId);
  const roles = rolesData ?? [];

  if (!permissions.role.canRead) {
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
      <CreateRoleDrawer
        open={state.drawerMode === "create"}
        onClose={() => {
          setState((s) => ({ ...s, drawerMode: "" }));
        }}
        currentOrgId={currentOrgId}
      />

      <DeleteRoleDialog
        selectedRole={state.selectedRole}
        currentOrgId={currentOrgId}
        open={state.isDeleteRoleDialogOpen && !!state.selectedRole}
        onClose={() =>
          setState((s) => ({
            ...s,
            isDeleteRoleDialogOpen: false,
          }))
        }
        onSuccess={() => {
          setState((s) => ({
            ...s,
            isDeleteRoleDialogOpen: false,
            drawerMode: "",
            selectedRole: null,
          }));
        }}
      />

      <RoleDetailsDrawer
        currentOrgId={currentOrgId}
        open={
          ["view", "edit"].includes(state.drawerMode) && !!state.selectedRole
        }
        onClose={() =>
          setState((s) => ({ ...s, drawerMode: "", selectedRole: null }))
        }
        onEditClose={() => setState((s) => ({ ...s, drawerMode: "view" }))}
        selectedRole={state.selectedRole}
        onEditClick={() =>
          setState((s) => ({
            ...s,
            drawerMode: "edit",
            selectedRole: s.selectedRole,
          }))
        }
        onDeleteClick={() =>
          setState((s) => ({
            ...s,
            selectedRole: s.selectedRole,
            isDeleteRoleDialogOpen: true,
          }))
        }
        onRoleUpdated={(role) =>
          setState((s) => ({ ...s, selectedRole: role, drawerMode: "view" }))
        }
        mode={state.drawerMode}
        canEdit={canManageRoles}
        canDelete={canDeleteRoles}
      />

      <Card>
        <CardHeader>
          <div className="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="min-w-0">
              <CardTitle>{t("settings:org.roles.title")}</CardTitle>
              <CardDescription>
                {t("settings:org.roles.description")}
              </CardDescription>
            </div>
            <div className="flex shrink-0 flex-wrap items-center gap-2">
              {canManageRoles && (
                <Button
                  onClick={() =>
                    setState((s) => ({ ...s, drawerMode: "create" }))
                  }
                >
                  <Plus className="mr-2 h-4 w-4" />
                  {t("settings:org.roles.createRole")}
                </Button>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {!capabilities.canManageRoles ? (
            <div className="mb-4 rounded-lg border border-dashed border-primary/40 bg-primary/5 p-4 text-left">
              <h3 className="text-sm font-semibold">
                {t("settings:org.upsell.rolesTitle")}
              </h3>
              <p className="mt-1 text-xs text-muted-foreground">
                {t("settings:org.upsell.rolesDescription")}
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
              {roles.length > 0 ? (
                <div className="space-y-2">
                  {roles.map((role) => (
                    <RoleRow
                      key={role.id}
                      role={role}
                      onSelect={(role) =>
                        setState((s) => ({
                          ...s,
                          selectedRole: role,
                          drawerMode: "view",
                        }))
                      }
                      onEdit={() =>
                        setState((s) => ({
                          ...s,
                          drawerMode: "edit",
                          selectedRole: role,
                        }))
                      }
                      onDelete={(role) =>
                        setState((s) => ({
                          ...s,
                          selectedRole: role,
                          isDeleteRoleDialogOpen: true,
                        }))
                      }
                      canEdit={canManageRoles}
                      canDelete={canDeleteRoles}
                    />
                  ))}
                </div>
              ) : (
                <div className="rounded-lg border border-dashed bg-muted/20 px-6 py-10 text-center">
                  <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary">
                    <Shield className="h-6 w-6" />
                  </div>
                  <h3 className="mt-4 text-base font-semibold">
                    {t("settings:org.roles.noRolesTitle")}
                  </h3>
                  <p className="mx-auto mt-2 max-w-md text-sm text-muted-foreground">
                    {t("settings:org.roles.noRolesDescription")}
                  </p>
                  {canManageRoles && (
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
                      {t("settings:org.roles.createRole")}
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
