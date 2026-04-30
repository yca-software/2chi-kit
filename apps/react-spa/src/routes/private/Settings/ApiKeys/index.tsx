import { Navigate, useParams } from "react-router";
import {
  useTranslationNamespace,
  getSubscriptionCapabilities,
} from "@/helpers";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Button,
} from "@yca-software/design-system";
import { Plus, Loader2, KeyRound } from "lucide-react";
import { useGetOrganizationQuery, useListApiKeysQuery } from "@/api";
import { usePricingModalStore } from "@/states";
import { useShallow } from "zustand/shallow";
import { ApiKeyRow } from "./ApiKeyRow";
import { CreateApiKeyDrawer } from "./CreateApiKeyDrawer";
import {
  useSettingsPermissions,
  SETTINGS_SECTIONS,
} from "../useSettingsPermissions";
import { ApiKey } from "@/types";
import { useState } from "react";
import { ApiKeyDetailsDrawer } from "./ApiKeyDetailsDrawer";
import { KeyCreatedDialog } from "./KeyCreatedDialog";
import { RevokeApiKeyDialog } from "./RevokeApiKeyDialog";

type ApiKeysUIState = {
  selectedApiKey: ApiKey | null;
  drawerMode: string; // "view", "edit", "create"
  isDeleteApiKeyDialogOpen: boolean;
  keyCreatedDialog: {
    open: boolean;
    keyValue: string;
  };
};

export const ApiKeysSettings = () => {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const { orgId } = useParams<{ orgId: string }>();
  const { openForOrg } = usePricingModalStore(
    useShallow((state) => ({
      openForOrg: state.openForOrg,
    })),
  );

  const [state, setState] = useState<ApiKeysUIState>({
    selectedApiKey: null,
    drawerMode: "",
    isDeleteApiKeyDialogOpen: false,
    keyCreatedDialog: {
      open: false,
      keyValue: "",
    },
  });

  const permissions = useSettingsPermissions();

  const { data: organization } = useGetOrganizationQuery(orgId || "");
  const capabilities = getSubscriptionCapabilities(organization, null);
  const canManageApiKeys =
    permissions.apiKey.canWrite && capabilities.canUseApiKeys;
  const canDeleteApiKeys =
    permissions.apiKey.canDelete && capabilities.canUseApiKeys;

  const currentOrgId = orgId || "";

  const { data: apiKeysData, isLoading } = useListApiKeysQuery(currentOrgId);
  const apiKeys = apiKeysData ?? [];

  if (!permissions.apiKey.canRead) {
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
      <CreateApiKeyDrawer
        open={state.drawerMode === "create"}
        onClose={() => {
          setState((s) => ({
            ...s,
            drawerMode: "",
          }));
        }}
        onKeyCreated={(keyValue) => {
          setState((s) => ({
            ...s,
            drawerMode: "",
            keyCreatedDialog: {
              open: true,
              keyValue,
            },
          }));
        }}
        currentOrgId={currentOrgId}
      />

      <KeyCreatedDialog
        open={state.keyCreatedDialog.open}
        keyValue={state.keyCreatedDialog.keyValue}
        onClose={() =>
          setState((s) => ({
            ...s,
            keyCreatedDialog: {
              open: false,
              keyValue: "",
            },
          }))
        }
      />

      <RevokeApiKeyDialog
        currentOrgId={currentOrgId}
        open={state.isDeleteApiKeyDialogOpen}
        selectedApiKey={state.selectedApiKey}
        onClose={() =>
          setState((s) => ({
            ...s,
            isDeleteApiKeyDialogOpen: false,
          }))
        }
        onSuccess={() => {
          setState((s) => ({
            ...s,
            isDeleteApiKeyDialogOpen: false,
            drawerMode: "",
            selectedApiKey: null,
          }));
        }}
      />

      <ApiKeyDetailsDrawer
        currentOrgId={currentOrgId}
        open={
          ["view", "edit"].includes(state.drawerMode) && !!state.selectedApiKey
        }
        onClose={() => {
          setState((s) => ({
            ...s,
            selectedApiKey: null,
            drawerMode: "",
          }));
        }}
        onEditClose={() => {
          setState((s) => ({
            ...s,
            drawerMode: "view",
          }));
        }}
        selectedApiKey={state.selectedApiKey}
        onEditClick={(apiKey) => {
          setState((s) => ({
            ...s,
            selectedApiKey: apiKey,
            drawerMode: "edit",
          }));
        }}
        onDeleteClick={(apiKey) => {
          setState((s) => ({
            ...s,
            selectedApiKey: apiKey,
            isDeleteApiKeyDialogOpen: true,
          }));
        }}
        onApiKeyUpdated={(apiKey) => {
          setState((s) => ({
            ...s,
            selectedApiKey: apiKey,
            drawerMode: "view",
          }));
        }}
        mode={state.drawerMode}
        canEdit={canManageApiKeys}
        canDelete={canDeleteApiKeys}
      />

      <Card>
        <CardHeader>
          <div className="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="min-w-0">
              <CardTitle>{t("settings:org.apiKeys.title")}</CardTitle>
              <CardDescription>
                {t("settings:org.apiKeys.description")}
              </CardDescription>
            </div>
            <div className="flex shrink-0 flex-wrap items-center gap-2">
              {canManageApiKeys && (
                <Button
                  size="sm"
                  onClick={() =>
                    setState((s) => ({ ...s, drawerMode: "create" }))
                  }
                >
                  <Plus className="mr-2 h-4 w-4" />
                  {t("common:create")}
                </Button>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {!capabilities.canUseApiKeys ? (
            <div className="mb-4 rounded-lg border border-dashed border-primary/40 bg-primary/5 p-4 text-left">
              <h3 className="text-sm font-semibold">
                {t("settings:org.upsell.apiKeysTitle")}
              </h3>
              <p className="mt-1 text-xs text-muted-foreground">
                {t("settings:org.upsell.apiKeysDescription")}
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
              {apiKeys.length > 0 ? (
                <div className="space-y-2">
                  {apiKeys.map((apiKey) => (
                    <ApiKeyRow
                      key={apiKey.id}
                      apiKey={apiKey}
                      onSelect={(apiKey) =>
                        setState((s) => ({
                          ...s,
                          selectedApiKey: apiKey,
                          drawerMode: "view",
                        }))
                      }
                      onEdit={(apiKey) =>
                        setState((s) => ({
                          ...s,
                          selectedApiKey: apiKey,
                          drawerMode: "edit",
                        }))
                      }
                      onDelete={(apiKey) =>
                        setState((s) => ({
                          ...s,
                          selectedApiKey: apiKey,
                          isDeleteApiKeyDialogOpen: true,
                        }))
                      }
                      canEdit={canManageApiKeys}
                      canDelete={canDeleteApiKeys}
                    />
                  ))}
                </div>
              ) : (
                <div className="rounded-lg border border-dashed bg-muted/20 px-6 py-10 text-center">
                  <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary">
                    <KeyRound className="h-6 w-6" />
                  </div>
                  <h3 className="mt-4 text-base font-semibold">
                    {t("settings:org.apiKeys.noApiKeysTitle")}
                  </h3>
                  <p className="mx-auto mt-2 max-w-md text-sm text-muted-foreground">
                    {t("settings:org.apiKeys.noApiKeysDescription")}
                  </p>
                  {canManageApiKeys && (
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
                      {t("settings:org.apiKeys.createApiKey")}
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
