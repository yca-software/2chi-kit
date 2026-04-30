import { useMemo, useState } from "react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
  Button,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@yca-software/design-system";
import { Monitor, Globe, Smartphone, Loader2, Trash2 } from "lucide-react";
import {
  useListActiveRefreshTokensQuery,
  useRevokeRefreshTokenMutation,
  useRevokeAllRefreshTokensMutation,
} from "@/api";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { REFRESH_TOKEN_QUERY_KEYS } from "@/constants";
import { useTranslationNamespace } from "@/helpers";

interface SessionsSectionProps {
  userId: string;
}

export function SessionsSection({ userId }: SessionsSectionProps) {
  void userId; // reserved for future use (e.g. scoping sessions)
  const { t } = useTranslationNamespace(["settings", "common"]);
  const queryClient = useQueryClient();
  const [showSessions, setShowSessions] = useState(false);
  const [tokenToRevoke, setTokenToRevoke] = useState<string | null>(null);

  const { data: sessionsData, isLoading: sessionsLoading } =
    useListActiveRefreshTokensQuery(true);
  const sessions = sessionsData ?? [];

  // Treat the most recent session (by createdAt) as the current one and hide revoke for it
  const currentSessionId = useMemo(() => {
    if (sessions.length === 0) return null;
    const latest = sessions.reduce((a, b) =>
      new Date(b.createdAt).getTime() > new Date(a.createdAt).getTime() ? b : a,
    );
    return latest.id;
  }, [sessions]);

  const revokeMutation = useRevokeRefreshTokenMutation({
    onSuccess: () => {
      toast.success(t("settings:user.sessionRevoked"));
      setTokenToRevoke(null);
      queryClient.invalidateQueries({
        queryKey: [REFRESH_TOKEN_QUERY_KEYS.ACTIVE],
      });
    },
    onError: () => {
      toast.error(t("common:defaultError"));
      setTokenToRevoke(null);
    },
  });

  const revokeAllMutation = useRevokeAllRefreshTokensMutation({
    onSuccess: () => {
      toast.success(t("settings:user.allSessionsRevoked"));
      queryClient.invalidateQueries({
        queryKey: [REFRESH_TOKEN_QUERY_KEYS.ACTIVE],
      });
    },
    onError: () => {
      toast.error(t("common:defaultError"));
    },
  });

  const getDeviceIcon = (userAgent: string) => {
    if (userAgent.includes("Mobile")) {
      return <Smartphone className="h-4 w-4" />;
    }
    return <Monitor className="h-4 w-4" />;
  };

  const getBrowserInfo = (userAgent: string) => {
    if (userAgent.includes("Chrome")) return "Chrome";
    if (userAgent.includes("Firefox")) return "Firefox";
    if (userAgent.includes("Safari")) return "Safari";
    if (userAgent.includes("Edge")) return "Edge";
    return "Unknown";
  };

  return (
    <div className="mt-4 space-y-2">
      <Button
        variant="outline"
        size="sm"
        onClick={() => setShowSessions(!showSessions)}
        className="w-full justify-start"
      >
        <Monitor className="mr-2 h-4 w-4" />
        {t("settings:user.activeSessions")} ({sessions.length})
      </Button>

      {showSessions && (
        <div className="mt-4 space-y-4">
          {sessionsLoading ? (
            <div className="flex items-center justify-center py-4">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : sessions.length === 0 ? (
            <p className="text-sm text-muted-foreground text-center py-4">
              {t("settings:user.noActiveSessions")}
            </p>
          ) : (
            <>
              <div className="flex items-center justify-between">
                <p className="text-sm text-muted-foreground">
                  {t("settings:user.activeSessionsDescription")}
                </p>
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button variant="outline" size="sm">
                      <Trash2 className="h-4 w-4 mr-2" />
                      {t("settings:user.revokeAll")}
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>
                        {t("settings:user.revokeAllSessions")}
                      </AlertDialogTitle>
                      <AlertDialogDescription>
                        {t("settings:user.revokeAllSessionsDescription")}
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>
                        {t("common:cancel")}
                      </AlertDialogCancel>
                      <AlertDialogAction
                        onClick={() => revokeAllMutation.mutate()}
                        className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                      >
                        {t("settings:user.revokeAll")}
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("settings:user.device")}</TableHead>
                    <TableHead>{t("settings:user.browser")}</TableHead>
                    <TableHead>{t("settings:user.location")}</TableHead>
                    <TableHead>{t("settings:user.lastActive")}</TableHead>
                    <TableHead className="text-right">
                      {t("settings:user.actions")}
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sessions.map((session) => (
                    <TableRow key={session.id}>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          {getDeviceIcon(session.userAgent)}
                          <span className="text-sm">
                            {session.userAgent.includes("Mobile")
                              ? t("settings:user.mobile")
                              : t("settings:user.desktop")}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Globe className="h-4 w-4 text-muted-foreground" />
                          <span className="text-sm">
                            {getBrowserInfo(session.userAgent)}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-muted-foreground">
                          {session.ip}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-muted-foreground">
                          {new Date(session.createdAt).toLocaleString()}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        {session.id === currentSessionId ? (
                          <span className="text-xs text-muted-foreground">
                            {t("settings:user.currentSession")}
                          </span>
                        ) : (
                          <AlertDialog>
                            <AlertDialogTrigger asChild>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => setTokenToRevoke(session.id)}
                              >
                                <Trash2 className="h-4 w-4" />
                              </Button>
                            </AlertDialogTrigger>
                            <AlertDialogContent>
                              <AlertDialogHeader>
                                <AlertDialogTitle>
                                  {t("settings:user.revokeSession")}
                                </AlertDialogTitle>
                                <AlertDialogDescription>
                                  {t("settings:user.revokeSessionDescription")}
                                </AlertDialogDescription>
                              </AlertDialogHeader>
                              <AlertDialogFooter>
                                <AlertDialogCancel
                                  onClick={() => setTokenToRevoke(null)}
                                >
                                  {t("common:cancel")}
                                </AlertDialogCancel>
                                <AlertDialogAction
                                  onClick={() => {
                                    if (tokenToRevoke) {
                                      revokeMutation.mutate(tokenToRevoke);
                                    }
                                  }}
                                  className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                                >
                                  {t("settings:user.revoke")}
                                </AlertDialogAction>
                              </AlertDialogFooter>
                            </AlertDialogContent>
                          </AlertDialog>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </>
          )}
        </div>
      )}
    </div>
  );
}
