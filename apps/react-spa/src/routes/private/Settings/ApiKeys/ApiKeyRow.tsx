import { DateTime } from "luxon";
import { Key, Pencil, Trash2 } from "lucide-react";
import { useTranslationNamespace } from "@/helpers";
import { EntityRow } from "@/components";
import type { ApiKey } from "@/types";

interface ApiKeyRowProps {
  apiKey: ApiKey;
  onSelect: (apiKey: ApiKey) => void;
  onEdit: (apiKey: ApiKey) => void;
  onDelete: (apiKey: ApiKey) => void;
  canEdit?: boolean;
  canDelete?: boolean;
}

export function ApiKeyRow({
  apiKey,
  onSelect,
  onEdit,
  onDelete,
  canEdit = true,
  canDelete = true,
}: ApiKeyRowProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  const expiresLabel = DateTime.fromISO(apiKey.expiresAt).toLocaleString(
    DateTime.DATE_MED,
  );

  const actions = [
    ...(canEdit
      ? [
          {
            label: t("common:edit"),
            icon: Pencil,
            onClick: () => onEdit(apiKey),
          },
        ]
      : []),
    ...(canDelete
      ? [
          {
            label: t("common:delete"),
            icon: Trash2,
            onClick: () => onDelete(apiKey),
            variant: "destructive" as const,
          },
        ]
      : []),
  ];

  return (
    <EntityRow
      icon={Key}
      title={apiKey.name}
      subtitle={t("settings:org.subscription.expiresAt", {
        date: expiresLabel,
      })}
      onClick={() => onSelect(apiKey)}
      actions={actions}
    >
      <p className="mt-2 font-mono text-sm text-muted-foreground">
        {apiKey.keyPrefix}...
      </p>
    </EntityRow>
  );
}
