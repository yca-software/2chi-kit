import { useTranslationNamespace } from "@/helpers";
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  Tooltip,
} from "@yca-software/design-system";
import { Mail, MoreVertical } from "lucide-react";
import type { Invitation, Role, MutationError } from "@/types";
import { useRevokeInvitationMutation } from "@/api";
import { toast } from "sonner";

interface PendingInvitationsProps {
  currentOrgId: string;
  invitations: Invitation[];
  roles: Role[];
  canRemove: boolean;
}

export function PendingInvitations({
  currentOrgId,
  invitations,
  roles,
  canRemove,
}: PendingInvitationsProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  const revokeInvitationMutation = useRevokeInvitationMutation(currentOrgId, {
    onError: (err: MutationError) => {
      toast.error(err.error?.message ?? t("common:defaultError"));
    },
  });

  if (invitations.length === 0) return null;

  return (
    <div className="mt-6 pt-6 border-t">
      <h4 className="mb-3 text-sm font-medium">
        {t("settings:org.members.pendingInvitations")}
      </h4>
      <div className="space-y-2">
        {invitations.map((inv) => {
          const roleName = roles.find((r) => r.id === inv.roleId)?.name ?? "";
          return (
            <div
              key={inv.id}
              className="flex items-start justify-between gap-2 rounded-lg border p-3"
            >
              <div className="flex items-center gap-3 min-w-0">
                <div className="bg-primary/10 flex h-8 w-8 shrink-0 items-center justify-center rounded-full">
                  <Mail className="h-4 w-4 text-primary" />
                </div>
                <div className="min-w-0">
                  <Tooltip content={inv.email} side="top" align="start">
                    <p className="text-sm font-medium truncate max-w-xs sm:max-w-md">
                      {inv.email}
                    </p>
                  </Tooltip>
                  {roleName && (
                    <Tooltip content={roleName} side="top" align="start">
                      <p className="mt-0.5 text-xs text-muted-foreground truncate max-w-xs sm:max-w-md">
                        {roleName}
                      </p>
                    </Tooltip>
                  )}
                </div>
              </div>
              {canRemove && (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="shrink-0"
                      disabled={revokeInvitationMutation.isPending}
                    >
                      <MoreVertical className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>

                  <DropdownMenuContent align="end" sideOffset={4}>
                    <DropdownMenuItem
                      onClick={() => {
                        revokeInvitationMutation.mutate(inv.id);
                      }}
                      disabled={revokeInvitationMutation.isPending}
                    >
                      {t("common:revoke")}
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
