import { Shield, Pencil, Trash2 } from "lucide-react";
import { useTranslationNamespace } from "@/helpers";
import { EntityRow } from "@/components";
import type { Role } from "@/types";

interface RoleRowProps {
  role: Role;
  onSelect: (role: Role) => void;
  onEdit: (role: Role) => void;
  onDelete: (role: Role) => void;
  canEdit?: boolean;
  canDelete?: boolean;
}

export function RoleRow({
  role,
  onSelect,
  onEdit,
  onDelete,
  canEdit: canEditPermission = true,
  canDelete: canDeletePermission = true,
}: RoleRowProps) {
  const { t } = useTranslationNamespace(["common"]);
  const canEdit = !role.locked && canEditPermission;
  const canDelete = !role.locked && canDeletePermission;

  const actions = [
    ...(canEdit
      ? [
          {
            label: t("common:edit"),
            icon: Pencil,
            onClick: () => onEdit(role),
          },
        ]
      : []),
    ...(canDelete
      ? [
          {
            label: t("common:delete"),
            icon: Trash2,
            onClick: () => onDelete(role),
            variant: "destructive" as const,
          },
        ]
      : []),
  ];

  return (
    <EntityRow
      icon={Shield}
      title={role.name}
      description={role.description}
      onClick={() => onSelect(role)}
      actions={actions}
    />
  );
}
