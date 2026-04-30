import { useState } from "react";
import { Link, useParams } from "react-router";
import { useNavigate } from "react-router";
import { Helmet } from "react-helmet-async";
import { useTranslationNamespace } from "@/helpers";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Button,
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogCancel,
  ConfirmDialog,
} from "@yca-software/design-system";
import { Loader2, UserCircle, LogIn, Copy, Trash2 } from "lucide-react";
import {
  useAdminUserDetailQuery,
  useAdminImpersonateUserMutation,
  useAdminDeleteUserMutation,
} from "@/api";
import { buildImpersonateScript } from "@/helpers";
import { AdminDetailPage } from "../AdminDetailPage";
import { DetailFieldList } from "../DetailFieldList";

function formatDate(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

export function AdminUserDetail() {
  const { t } = useTranslationNamespace(["admin"]);
  const navigate = useNavigate();
  const { userId } = useParams<{ userId: string }>();
  const [impersonateDialogOpen, setImpersonateDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [impersonateScript, setImpersonateScript] = useState<string | null>(
    null,
  );
  const [copied, setCopied] = useState(false);

  const { data, isLoading, isError } = useAdminUserDetailQuery(userId);
  const impersonateMutation = useAdminImpersonateUserMutation({
    onSuccess: (tokens) => {
      setImpersonateScript(
        buildImpersonateScript(tokens.accessToken, tokens.refreshToken),
      );
      setImpersonateDialogOpen(true);
    },
    onError: () => {},
  });
  const deleteMutation = useAdminDeleteUserMutation({
    onSuccess: () => {
      navigate("/admin/users");
    },
  });

  const handleCopyScript = async () => {
    if (!impersonateScript) return;
    await navigator.clipboard.writeText(impersonateScript);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const fullName = data
    ? [data.user.firstName, data.user.lastName].filter(Boolean).join(" ") ||
      data.user.email
    : "";

  const headerActions = data && (
    <div className="flex items-center gap-2">
      <Button
        onClick={() => impersonateMutation.mutate(userId!)}
        disabled={impersonateMutation.isPending}
        aria-label={t("admin:users.impersonate")}
      >
        {impersonateMutation.isPending ? (
          <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden />
        ) : (
          <LogIn className="mr-2 h-4 w-4" aria-hidden />
        )}
        {t("admin:users.impersonate")}
      </Button>
      <Button
        variant="destructive"
        onClick={() => setDeleteDialogOpen(true)}
        disabled={deleteMutation.isPending}
        aria-label={t("admin:users.delete.button")}
      >
        <Trash2 className="mr-2 h-4 w-4" aria-hidden />
        {t("admin:users.delete.button")}
      </Button>
    </div>
  );

  return (
    <>
      {data && (
        <Helmet>
          <title>
            {fullName} – {t("admin:users.details")}
          </title>
        </Helmet>
      )}
      <AdminDetailPage
        backHref="/admin/users"
        backLabel={t("admin:users.backToList")}
        isLoading={isLoading || !userId}
        isError={!!isError || (!isLoading && !!userId && !data)}
        notFoundMessage={t("admin:users.notFound")}
        headerActions={headerActions}
      >
        {data && (
          <>
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <UserCircle className="h-5 w-5" aria-hidden />
                  {t("admin:users.details")}
                </CardTitle>
                <CardDescription>{fullName}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-6">
                <DetailFieldList
                  fields={[
                    { label: t("admin:users.name"), value: fullName },
                    { label: t("admin:users.email"), value: data.user.email },
                    {
                      label: t("admin:users.joinedOn", {
                        date: formatDate(data.user.createdAt),
                      }),
                      value: formatDate(data.user.createdAt),
                    },
                    ...(data.adminAccess
                      ? [
                          {
                            label: t("admin:users.status"),
                            value: t("admin:users.admin"),
                          },
                        ]
                      : []),
                  ]}
                />

                {data.roles && data.roles.length > 0 && (
                  <div>
                    <h3 className="mb-2 text-sm font-medium">
                      {t("admin:users.organizations")}
                    </h3>
                    <ul className="space-y-2 rounded-md border p-3">
                      {data.roles.map((r) => (
                        <li
                          key={`${r.organizationId}-${r.roleId}`}
                          className="flex flex-wrap items-center justify-between gap-2 text-sm"
                        >
                          <Link
                            to={`/admin/organizations/${r.organizationId}`}
                            className="font-medium text-primary underline-offset-4 hover:underline"
                          >
                            {r.organizationName}
                          </Link>
                          <span className="text-muted-foreground">
                            {r.roleName}
                          </span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </CardContent>
            </Card>

            <AlertDialog
              open={impersonateDialogOpen}
              onOpenChange={setImpersonateDialogOpen}
            >
              <AlertDialogContent className="max-h-[90vh] flex flex-col max-w-2xl">
                <AlertDialogHeader>
                  <AlertDialogTitle>
                    {t("admin:users.impersonateDialogTitle")}
                  </AlertDialogTitle>
                  <AlertDialogDescription asChild>
                    <div className="space-y-4 pt-2 overflow-y-auto">
                      <p className="text-sm text-muted-foreground">
                        {t("admin:users.impersonateSessionNote")}
                      </p>
                      <ol className="list-decimal list-inside space-y-2 text-sm text-muted-foreground">
                        <li>{t("admin:users.impersonateStep1")}</li>
                        <li>
                          {t("admin:users.impersonateStep2", {
                            url: window.location.origin,
                          })}
                        </li>
                        <li>{t("admin:users.impersonateStep3")}</li>
                        <li>{t("admin:users.impersonateStep4")}</li>
                        <li>{t("admin:users.impersonateStep5")}</li>
                      </ol>
                      {impersonateScript && (
                        <div className="space-y-2">
                          <p className="text-sm font-medium">
                            {t("admin:users.impersonateScriptLabel")}
                          </p>
                          <pre className="max-h-48 overflow-auto break-all whitespace-pre-wrap rounded-md border bg-muted/50 p-3 font-mono text-xs">
                            {impersonateScript}
                          </pre>
                          <Button
                            type="button"
                            variant="secondary"
                            size="sm"
                            onClick={handleCopyScript}
                          >
                            <Copy className="mr-2 h-4 w-4" aria-hidden />
                            {copied
                              ? t("common:copied")
                              : t("admin:users.impersonateCopyScript")}
                          </Button>
                        </div>
                      )}
                    </div>
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel
                    onClick={() => setImpersonateDialogOpen(false)}
                  >
                    {t("common:close")}
                  </AlertDialogCancel>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
            <ConfirmDialog
              open={deleteDialogOpen}
              onOpenChange={setDeleteDialogOpen}
              title={t("common:confirm")}
              description={t("admin:users.delete.confirmDescription")}
              cancelLabel={t("common:cancel")}
              confirmLabel={
                deleteMutation.isPending
                  ? t("common:deleting")
                  : t("common:delete")
              }
              variant="destructive"
              isPending={deleteMutation.isPending}
              onConfirm={() => {
                if (!userId) return;
                deleteMutation.mutate(userId);
              }}
              closeOnOutsideClick
            />
          </>
        )}
      </AdminDetailPage>
    </>
  );
}
