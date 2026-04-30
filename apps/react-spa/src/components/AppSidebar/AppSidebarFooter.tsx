import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  cn,
} from "@yca-software/design-system";
import { LogOut, Settings } from "lucide-react";
import { useTranslationNamespace, getInitials } from "@/helpers";
import type { User } from "@/types";

export interface AppSidebarFooterProps {
  collapsed: boolean;
  user: User | null;
  onOpenUserSettings: () => void;
  onSignOut: () => void;
}

export function AppSidebarFooter({
  collapsed,
  user,
  onOpenUserSettings,
  onSignOut,
}: AppSidebarFooterProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  return (
    <div
      className={cn(
        "shrink-0 border-t space-y-1 transition-[padding] duration-200",
        collapsed ? "p-2" : "p-4",
      )}
    >
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          {collapsed ? (
            <button
              type="button"
              className="flex w-full cursor-pointer items-center justify-center rounded-lg p-1 transition-colors hover:bg-accent"
              aria-label={user?.email ?? "User menu"}
            >
              <Avatar className="h-8 w-8 rounded-full">
                {user?.avatarURL ? (
                  <AvatarImage
                    src={user.avatarURL}
                    alt={`${user.firstName} ${user.lastName}`}
                  />
                ) : null}
                <AvatarFallback className="rounded-full bg-muted text-muted-foreground text-xs font-medium">
                  {getInitials(user?.firstName, user?.lastName)}
                </AvatarFallback>
              </Avatar>
            </button>
          ) : (
            <button
              type="button"
              className="w-full flex cursor-pointer items-center gap-3 rounded-lg px-3 py-2 transition-colors hover:bg-accent"
            >
              <Avatar className="h-8 w-8 shrink-0 rounded-full">
                {user?.avatarURL ? (
                  <AvatarImage
                    src={user.avatarURL}
                    alt={`${user.firstName} ${user.lastName}`}
                  />
                ) : null}
                <AvatarFallback className="rounded-full bg-muted text-muted-foreground text-xs font-medium">
                  {getInitials(user?.firstName, user?.lastName)}
                </AvatarFallback>
              </Avatar>
              <div className="flex-1 min-w-0 text-left">
                <p className="truncate text-sm font-medium">
                  {user?.firstName} {user?.lastName}
                </p>
                <p className="truncate text-xs text-muted-foreground">
                  {user?.email}
                </p>
              </div>
            </button>
          )}
        </DropdownMenuTrigger>
        <DropdownMenuContent
          align={collapsed ? "center" : "start"}
          side={collapsed ? "right" : "top"}
          className="w-56"
        >
          <DropdownMenuItem onClick={onOpenUserSettings}>
            <Settings className="mr-2 h-4 w-4" />
            {t("settings:user.title")}
          </DropdownMenuItem>
          <DropdownMenuItem onClick={onSignOut}>
            <LogOut className="mr-2 h-4 w-4" />
            {t("settings:nav.signOut")}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
