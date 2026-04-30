import {
  Avatar,
  AvatarFallback,
  Button,
  Card,
  CardContent,
} from "@yca-software/design-system";
import { Mail, CheckCircle2, XCircle, Loader2 } from "lucide-react";
import type { User } from "@/types";
import { useTranslationNamespace, getInitials } from "@/helpers";
import { useResendVerificationEmailMutation } from "@/api";
import { toast } from "sonner";

interface ProfileDisplayProps {
  user: User | null;
}

export function ProfileDisplay({ user }: ProfileDisplayProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  const isEmailVerified = user?.emailVerifiedAt != null;

  const resendVerificationMutation = useResendVerificationEmailMutation({
    onSuccess: () => {
      toast.success(t("settings:user.verificationEmailSent"));
    },
    onError: (error) => {
      if (error?.error?.errorCode === "CONFLICT_CONFLICTING_DATA") {
        toast.info(t("settings:user.emailAlreadyVerified"));
      } else {
        toast.error(t("common:defaultError"));
      }
    },
  });

  return (
    <Card className="border-0 shadow-none bg-muted/30">
      <CardContent className="p-5">
        <div className="flex items-center gap-4">
          <Avatar className="h-16 w-16 rounded-xl border-2 border-background shadow-sm">
            <AvatarFallback className="rounded-xl bg-primary/10 text-primary text-lg font-semibold">
              {getInitials(user?.firstName, user?.lastName)}
            </AvatarFallback>
          </Avatar>
          <div className="flex-1 space-y-1">
            <h3 className="text-lg font-semibold leading-none">
              {user
                ? `${user.firstName} ${user.lastName}`
                : t("settings:nav.user")}
            </h3>
            <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
              <Mail className="h-3.5 w-3.5" />
              <span>{user?.email}</span>
            </div>
            <div className="flex items-center gap-2 mt-2">
              {isEmailVerified ? (
                <div className="flex items-center gap-1.5 text-xs text-green-600 dark:text-green-400">
                  <CheckCircle2 className="h-3.5 w-3.5" />
                  <span>{t("settings:user.emailVerified")}</span>
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <div className="flex items-center gap-1.5 text-xs text-amber-600 dark:text-amber-400">
                    <XCircle className="h-3.5 w-3.5" />
                    <span>{t("settings:user.emailNotVerified")}</span>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => resendVerificationMutation.mutate()}
                    disabled={resendVerificationMutation.isPending}
                    className="h-7 text-xs cursor-pointer"
                  >
                    {resendVerificationMutation.isPending ? (
                      <>
                        <Loader2 className="h-3 w-3 mr-1.5 animate-spin" />
                        {t("common:sending")}
                      </>
                    ) : (
                      t("settings:user.resendVerificationEmail")
                    )}
                  </Button>
                </div>
              )}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
