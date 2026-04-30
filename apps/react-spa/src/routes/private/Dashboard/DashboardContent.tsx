import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@yca-software/design-system";
import { useTranslationNamespace } from "@/helpers";

interface DashboardContentProps {
  organizationName?: string;
}

export const DashboardContent = ({
  organizationName,
}: DashboardContentProps) => {
  const { t } = useTranslationNamespace(["dashboard", "common"]);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">
          {t("dashboard:title")}
        </h1>
        <p className="text-muted-foreground">
          {organizationName
            ? t("dashboard:welcomeToOrg", { orgName: organizationName })
            : t("dashboard:description")}
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t("dashboard:gettingStarted.title")}</CardTitle>
          <CardDescription>
            {t("dashboard:gettingStarted.description")}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            {t("dashboard:gettingStarted.content")}
          </p>
        </CardContent>
      </Card>
    </div>
  );
};
