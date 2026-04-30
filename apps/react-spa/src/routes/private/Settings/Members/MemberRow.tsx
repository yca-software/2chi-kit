import { User, Trash2 } from "lucide-react";
import { useTranslationNamespace } from "@/helpers/hooks/useTranslationNamespace";
import { EntityRow } from "@/components/EntityRow";
import type { OrganizationMemberWithUser } from "@/types/organization";

interface MemberRowProps {
  member: OrganizationMemberWithUser;
  roleName: string;
  isSelf: boolean;
  onSelect: (memberId: string) => void;
  canRemove: boolean;
  onRemoveClick: (memberId: string) => void;
}

export function MemberRow({
  member,
  roleName,
  isSelf,
  onSelect,
  canRemove,
  onRemoveClick,
}: MemberRowProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const fullName = `${member.userFirstName ?? ""} ${member.userLastName ?? ""}`.trim() || member.userEmail;
  const showMenu = !isSelf && canRemove;

  const actions = showMenu
    ? [
        {
          label: t("common:remove"),
          icon: Trash2,
          onClick: () => onRemoveClick(member.id),
          variant: "destructive" as const,
        },
      ]
    : [];

  return (
    <EntityRow
      icon={User}
      title={
        <>
          {fullName}
          {isSelf && (
            <span className="ml-2 text-xs text-muted-foreground">
              ({t("common:you")})
            </span>
          )}
        </>
      }
      titleTooltip={fullName}
      subtitle={member.userEmail}
      onClick={() => onSelect(member.id)}
      actions={actions}
    >
      <div className="mt-2 flex items-center gap-2">
        <span className="inline-flex max-w-40 items-center justify-center truncate rounded-full bg-muted px-2.5 py-0.5 text-xs font-medium text-muted-foreground">
          {roleName || t("settings:org.members.selectRole")}
        </span>
      </div>
    </EntityRow>
  );
}
