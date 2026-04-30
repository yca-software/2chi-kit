import { useTranslationNamespace } from "@/helpers/hooks/useTranslationNamespace";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Button,
} from "@yca-software/design-system";
import { Pencil } from "lucide-react";
import type { Organization } from "@/types/organization";

interface GeneralInfoDisplayProps {
  organization: Organization | null | undefined;
  onEdit: () => void;
  canEdit?: boolean;
}

export function GeneralInfoDisplay({ organization, onEdit, canEdit = true }: GeneralInfoDisplayProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  return (
    <Card>
      <CardHeader>
        <div className="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="min-w-0">
            <CardTitle>{t("settings:org.general.title")}</CardTitle>
            <CardDescription>
              {t("settings:org.general.description")}
            </CardDescription>
          </div>
          {canEdit && (
            <div className="shrink-0">
              <Button variant="outline" size="sm" onClick={onEdit}>
                <Pencil className="mr-2 h-4 w-4" />
                {t("edit", { ns: "common" })}
              </Button>
            </div>
          )}
        </div>
      </CardHeader>
      <CardContent>
        {organization ? (
          <dl className="space-y-4">
            <div>
              <dt className="text-sm font-medium text-muted-foreground">
                {t("settings:org.general.name")}
              </dt>
              <dd className="mt-1 text-sm">{organization.name}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">
                {t("settings:org.general.address")}
              </dt>
              <dd className="mt-1 text-sm">{organization.address}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-muted-foreground">
                {t("settings:org.general.city")}
              </dt>
              <dd className="mt-1 text-sm">{organization.city}</dd>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  {t("settings:org.general.zip")}
                </dt>
                <dd className="mt-1 text-sm">{organization.zip}</dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-muted-foreground">
                  {t("settings:org.general.country")}
                </dt>
                <dd className="mt-1 text-sm">{organization.country}</dd>
              </div>
            </div>
          </dl>
        ) : null}
      </CardContent>
    </Card>
  );
}
