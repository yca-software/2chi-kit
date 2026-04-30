import { useTranslationNamespace } from "@/helpers";
import { ApiKey } from "@/types";
import {
  Button,
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@yca-software/design-system";
import { Pencil, Trash2 } from "lucide-react";
import { DateTime } from "luxon";
import { ROLE_PERMISSION_GROUPS } from "../Roles/rolePermissionGroups";
import { EditApiKeyForm } from "./EditApiKeyForm";

export interface ApiKeyDetailsDrawerProps {
  currentOrgId: string;
  open: boolean;
  onClose: () => void;
  onEditClose: () => void;
  selectedApiKey: ApiKey | null;
  onEditClick: (apiKey: ApiKey) => void;
  onDeleteClick: (apiKey: ApiKey) => void;
  onApiKeyUpdated: (apiKey: ApiKey) => void;
  mode: string;
  canEdit: boolean;
  canDelete: boolean;
}

export function ApiKeyDetailsDrawer({
  currentOrgId,
  open,
  onClose,
  onEditClose,
  selectedApiKey,
  onEditClick,
  onDeleteClick,
  onApiKeyUpdated,
  mode,
  canEdit,
  canDelete,
}: ApiKeyDetailsDrawerProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  const permissions = selectedApiKey ? selectedApiKey.permissions : [];
  const groupedSelections = ROLE_PERMISSION_GROUPS.map((group) => {
    const selected = group.permissions.filter((perm) =>
      permissions.includes(perm.key),
    );
    return { group, selected };
  }).filter(({ selected }) => selected.length > 0);

  if (!selectedApiKey) {
    return null;
  }

  return (
    <Sheet open={open} onOpenChange={onClose}>
      <SheetContent className="sm:max-w-lg">
        <SheetHeader>
          <SheetTitle>{selectedApiKey.name}</SheetTitle>
          <SheetDescription>
            {t("settings:org.apiKeys.detailsDescription")}
          </SheetDescription>
        </SheetHeader>

        {mode === "view" ? (
          <div className="mt-1 flex flex-col gap-4">
            {(canEdit || canDelete) && (
              <div className="flex items-center justify-end gap-2">
                {canEdit && (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      onEditClick(selectedApiKey);
                    }}
                  >
                    <Pencil className="mr-2 h-4 w-4" />
                    {t("common:edit")}
                  </Button>
                )}
                {canDelete && (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="text-destructive hover:text-destructive"
                    onClick={() => onDeleteClick(selectedApiKey)}
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    {t("common:delete")}
                  </Button>
                )}
              </div>
            )}

            <div>
              <p className="text-xs font-semibold uppercase text-muted-foreground">
                {t("settings:org.apiKeys.name")}
              </p>
              <p className="mt-1 text-sm">{selectedApiKey.name}</p>
            </div>

            <div>
              <p className="text-xs font-semibold uppercase text-muted-foreground">
                {t("settings:org.apiKeys.prefix")}
              </p>
              <p className="mt-1 font-mono text-sm text-muted-foreground">
                {selectedApiKey.keyPrefix}...
              </p>
            </div>

            <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
              <div>
                <p className="text-xs font-semibold uppercase text-muted-foreground">
                  {t("settings:org.apiKeys.createdAt")}
                </p>
                <p className="mt-1">
                  {DateTime.fromISO(selectedApiKey.createdAt).toLocaleString(
                    DateTime.DATETIME_MED,
                  )}
                </p>
              </div>
              <div>
                <p className="text-xs font-semibold uppercase text-muted-foreground">
                  {t("settings:org.subscription.expiresAt", { date: "" })}
                </p>
                <p className="mt-1">
                  {DateTime.fromISO(selectedApiKey.expiresAt).toLocaleString(
                    DateTime.DATETIME_MED,
                  )}
                </p>
              </div>
            </div>

            <div>
              <p className="text-xs font-semibold uppercase text-muted-foreground">
                {t("settings:org.apiKeys.permissions")}
              </p>
              {groupedSelections.length === 0 ? (
                <p className="mt-1 text-sm text-muted-foreground">
                  {t("settings:org.roles.noPermissions")}
                </p>
              ) : (
                <div className="mt-2 space-y-3">
                  {groupedSelections.map(({ group, selected }) => (
                    <div key={group.contextKey}>
                      <p className="text-xs font-semibold text-muted-foreground">
                        {t(group.labelKey)}
                      </p>
                      <div className="mt-1 flex flex-wrap gap-2">
                        {selected.map((perm) => (
                          <span
                            key={perm.key}
                            className="rounded-full bg-muted px-2 py-1 text-xs text-muted-foreground"
                          >
                            {t(perm.labelKey)}
                          </span>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        ) : (
          <div className="mt-6">
            <EditApiKeyForm
              currentOrgId={currentOrgId}
              selectedApiKey={selectedApiKey}
              onApiKeyUpdated={onApiKeyUpdated}
              onClose={onEditClose}
            />
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}
