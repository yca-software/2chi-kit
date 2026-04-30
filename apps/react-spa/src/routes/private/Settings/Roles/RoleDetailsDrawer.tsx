import { useTranslationNamespace } from "@/helpers";
import { Role } from "@/types";
import { Pencil, Trash2, Lock } from "lucide-react";
import { ROLE_PERMISSION_GROUPS } from "./rolePermissionGroups";
import {
  Button,
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@yca-software/design-system";
import { EditRoleForm } from "./EditRoleForm";

interface RoleDetailsDrawerProps {
  currentOrgId: string;
  open: boolean;
  onClose: () => void;
  onEditClose: () => void;
  selectedRole: Role | null;
  onEditClick: (role: Role) => void;
  onDeleteClick: (role: Role) => void;
  onRoleUpdated: (role: Role) => void;
  mode: string;
  canEdit: boolean;
  canDelete: boolean;
}

export function RoleDetailsDrawer({
  currentOrgId,
  open,
  onClose,
  onEditClose,
  selectedRole,
  onEditClick,
  onDeleteClick,
  onRoleUpdated,
  mode,
  canEdit,
  canDelete,
}: RoleDetailsDrawerProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  if (!selectedRole) {
    return null;
  }

  const permissions = selectedRole.permissions ?? [];
  const groupedSelections = ROLE_PERMISSION_GROUPS.map((group) => {
    const selected = group.permissions.filter((perm) =>
      permissions.includes(perm.key),
    );
    return { group, selected };
  }).filter(({ selected }) => selected.length > 0);

  return (
    <Sheet open={open} onOpenChange={onClose}>
      <SheetContent className="sm:max-w-lg">
        <SheetHeader>
          <div className="flex items-center gap-2">
            <SheetTitle>{selectedRole.name}</SheetTitle>
            {selectedRole.locked && (
              <span className="inline-flex items-center gap-1 rounded-full bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground">
                <Lock className="h-3 w-3" />
                {t("settings:org.roles.systemDefaultTag")}
              </span>
            )}
          </div>
          <SheetDescription>
            {t("settings:org.roles.detailsDescription")}
          </SheetDescription>
        </SheetHeader>

        {mode === "view" ? (
          <div className="mt-1 flex flex-col gap-4">
            {!selectedRole.locked && (canEdit || canDelete) && (
              <div className="flex items-center justify-end gap-2">
                {canEdit && (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      onEditClick(selectedRole);
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
                    onClick={() => onDeleteClick(selectedRole)}
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    {t("common:delete")}
                  </Button>
                )}
              </div>
            )}

            {selectedRole.description && (
              <div className="mt-1">
                <p className="text-xs font-semibold uppercase text-muted-foreground">
                  {t("settings:org.roles.descriptionLabel")}
                </p>
                <p className="mt-1 text-sm">{selectedRole.description}</p>
              </div>
            )}

            <div>
              <p className="text-xs font-semibold uppercase text-muted-foreground">
                {t("settings:org.roles.permissions")}
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
            <EditRoleForm
              currentOrgId={currentOrgId}
              selectedRole={selectedRole}
              onRoleUpdated={onRoleUpdated}
              onClose={onEditClose}
            />
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}
