import { Helmet } from "react-helmet-async";
import { Link } from "react-router";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@yca-software/design-system";
import { Users, Building2 } from "lucide-react";
import { useTranslationNamespace } from "@/helpers";

export function AdminDashboard() {
  const { t } = useTranslationNamespace(["admin"]);

  const quickActions = [
    {
      title: t("admin:dashboard.manageUsers"),
      icon: Users,
      href: "/admin/users",
      color: "text-blue-500",
    },
    {
      title: t("admin:dashboard.manageOrganizations"),
      icon: Building2,
      href: "/admin/organizations",
      color: "text-emerald-500",
    },
  ];

  return (
    <>
      <Helmet>
        <title>{t("admin:dashboard.title")}</title>
      </Helmet>
      <div className="space-y-8">
        <div>
          <h1 className="text-3xl font-bold">{t("admin:dashboard.title")}</h1>
          <p className="text-muted-foreground mt-1">
            {t("admin:dashboard.description")}
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>{t("admin:dashboard.quickActions")}</CardTitle>
            <CardDescription>
              {t("admin:dashboard.quickActionsDescription")}
            </CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4 md:grid-cols-3">
            {quickActions.map((action) => (
              <Link
                key={action.href}
                to={action.href}
                className="flex items-center gap-3 p-4 rounded-lg bg-accent/50 hover:bg-accent transition-colors cursor-pointer"
              >
                <action.icon className={`h-5 w-5 ${action.color}`} />
                <span>{action.title}</span>
              </Link>
            ))}
          </CardContent>
        </Card>
      </div>
    </>
  );
}
