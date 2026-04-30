import { Button, cn } from "@yca-software/design-system";
import { X } from "lucide-react";

export interface AppSidebarHeaderProps {
  collapsed: boolean;
  onClose?: () => void;
}

export function AppSidebarHeader({
  collapsed,
  onClose,
}: AppSidebarHeaderProps) {
  return (
    <div
      className={cn(
        "relative flex h-16 shrink-0 items-center border-b transition-[padding] duration-200",
        collapsed ? "justify-center px-2" : "justify-between px-4",
      )}
    >
      <div
        className={cn(
          "flex items-center gap-2 min-w-0",
          collapsed && "justify-center",
        )}
      >
        <div className="bg-primary text-primary-foreground flex h-8 w-8 shrink-0 items-center justify-center rounded-lg font-bold">
          Y
        </div>
        {!collapsed && (
          <span className="truncate font-semibold">YCA Software</span>
        )}
      </div>
      {onClose && (
        <Button
          variant="ghost"
          size="icon"
          className="lg:hidden shrink-0"
          onClick={onClose}
        >
          <X className="h-5 w-5" />
        </Button>
      )}
    </div>
  );
}
