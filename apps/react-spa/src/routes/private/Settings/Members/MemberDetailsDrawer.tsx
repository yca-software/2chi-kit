import { useEffect, useState } from "react";
import { useTranslationNamespace } from "@/helpers/hooks/useTranslationNamespace";
import {
  Button,
  Select,
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@yca-software/design-system";
import { User, Trash2 } from "lucide-react";
import type { OrganizationMemberWithUser, Role } from "@/types/organization";
import { useUpdateMemberRoleMutation } from "@/api";
import { MutationError } from "@/types";
import { toast } from "sonner";

interface MemberDetailsDrawerProps {
  currentOrgId: string;
  open: boolean;
  onClose: () => void;
  selectedMember: OrganizationMemberWithUser | null;
  roles: Role[];
  canUpdateRole: boolean;
  canRemove: boolean;
  isSelf: boolean;
  onDeleteClick: (member: OrganizationMemberWithUser) => void;
  onMemberUpdated: (member: OrganizationMemberWithUser) => void;
}

export function MemberDetailsDrawer({
  open,
  onClose,
  selectedMember,
  roles,
  canUpdateRole,
  canRemove,
  isSelf,
  onDeleteClick,
  onMemberUpdated,
  currentOrgId,
}: MemberDetailsDrawerProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const [selectedRoleId, setSelectedRoleId] = useState<string>(
    selectedMember?.roleId ? String(selectedMember.roleId) : "",
  );

  // Keep selectedRoleId in sync if selectedMember changes
  useEffect(() => {
    setSelectedRoleId(
      selectedMember?.roleId ? String(selectedMember.roleId) : "",
    );
  }, [selectedMember?.roleId, selectedMember?.id]);

  const updateRoleMutation = useUpdateMemberRoleMutation(
    currentOrgId,
    selectedMember?.id ?? "",
    {
      onSuccess: (data) => onMemberUpdated(data as OrganizationMemberWithUser),
      onError: (err: MutationError) => {
        toast.error(err.error?.message ?? t("common:defaultError"));
      },
    },
  );

  if (!selectedMember) {
    return null;
  }

  return (
    <Sheet open={open} onOpenChange={onClose}>
      <SheetContent className="sm:max-w-lg">
        <SheetHeader>
          <SheetTitle>
            {selectedMember.userFirstName} {selectedMember.userLastName}
            {isSelf && (
              <span className="ml-2 text-xs text-muted-foreground">
                ({t("common:you")})
              </span>
            )}
          </SheetTitle>
          <SheetDescription>
            {t("settings:org.members.detailsDescription")}
          </SheetDescription>
        </SheetHeader>

        <div className="mt-1 flex flex-col gap-4">
          {canRemove && !isSelf && (
            <div className="flex items-center justify-end">
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="text-destructive hover:text-destructive"
                onClick={() => onDeleteClick(selectedMember)}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                {t("common:remove")}
              </Button>
            </div>
          )}

          <div className="mt-1 flex items-center gap-3">
            <div className="bg-primary/10 flex h-10 w-10 items-center justify-center rounded-full">
              <User className="h-5 w-5 text-primary" />
            </div>
            <div>
              <p className="font-medium">
                {selectedMember.userFirstName} {selectedMember.userLastName}
              </p>
              <p className="text-sm text-muted-foreground">
                {selectedMember.userEmail}
              </p>
            </div>
          </div>

          <div>
            <p className="text-xs font-semibold uppercase text-muted-foreground">
              {t("settings:org.members.role")}
            </p>
            <div className="mt-2">
              <Select
                value={selectedRoleId}
                onValueChange={setSelectedRoleId}
                options={roles.map((r) => ({ value: r.id, label: r.name }))}
                placeholder={t("settings:org.members.selectRole")}
                disabled={!canUpdateRole || isSelf}
                aria-label={t("settings:org.members.role")}
              />
            </div>
          </div>

          {canUpdateRole && !isSelf && (
            <div className="flex justify-end">
              <Button
                type="button"
                size="sm"
                disabled={updateRoleMutation.isPending}
                onClick={() => {
                  if (selectedRoleId) {
                    updateRoleMutation.mutate({ roleId: selectedRoleId });
                  }
                }}
              >
                {t("common:save")}
              </Button>
            </div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}
