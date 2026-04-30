import { Users2, Pencil, Trash2 } from "lucide-react";
import { useTranslationNamespace } from "@/helpers";
import { EntityRow } from "@/components";
import type { Team } from "@/types";

interface TeamRowProps {
  team: Team;
  onEdit: (team: Team) => void;
  onSelect: (team: Team) => void;
  onDelete: (team: Team) => void;
  canEdit?: boolean;
  canDelete?: boolean;
}

export function TeamRow({
  team,
  onEdit,
  onSelect,
  onDelete,
  canEdit = true,
  canDelete = true,
}: TeamRowProps) {
  const { t } = useTranslationNamespace(["common"]);

  const actions = [
    ...(canEdit
      ? [
          {
            label: t("common:edit"),
            icon: Pencil,
            onClick: () => onEdit(team),
          },
        ]
      : []),
    ...(canDelete
      ? [
          {
            label: t("common:delete"),
            icon: Trash2,
            onClick: () => onDelete(team),
            variant: "destructive" as const,
          },
        ]
      : []),
  ];

  return (
    <EntityRow
      icon={Users2}
      title={team.name}
      description={team.description}
      onClick={() => onSelect(team)}
      actions={actions}
    />
  );
}
