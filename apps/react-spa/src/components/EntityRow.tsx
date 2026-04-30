import type { LucideIcon } from "lucide-react";
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  Tooltip,
  cn,
} from "@yca-software/design-system";
import { MoreVertical } from "lucide-react";

export interface EntityRowAction {
  label: string;
  icon: LucideIcon;
  onClick: () => void;
  variant?: "default" | "destructive";
}

export interface EntityRowProps {
  icon: LucideIcon;
  title: React.ReactNode;
  /** Tooltip for title (when title is truncated). Omit when title is a ReactNode. */
  titleTooltip?: string;
  description?: string;
  /** Secondary line below title (e.g. created date, email) */
  subtitle?: string;
  onClick?: () => void;
  actions?: EntityRowAction[];
  /** Muted styling for archived items */
  archived?: boolean;
  /** Extra content below main block (e.g. key prefix, metadata) */
  children?: React.ReactNode;
  className?: string;
}

/**
 * Reusable row for settings entity lists (roles, teams, members, API keys).
 * Consistent layout: icon + title + optional description/subtitle + actions dropdown.
 * Mobile-friendly: touch targets, responsive truncation.
 */
export function EntityRow({
  icon: Icon,
  title,
  titleTooltip = typeof title === "string" ? title : undefined,
  description,
  subtitle,
  onClick,
  actions = [],
  archived = false,
  children,
  className,
}: EntityRowProps) {
  const hasActions = actions.length > 0;
  const isClickable = !!onClick;

  return (
    <div
      role={isClickable ? "button" : undefined}
      tabIndex={isClickable ? 0 : undefined}
      onClick={onClick}
      onKeyDown={
        isClickable
          ? (e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                onClick?.();
              }
            }
          : undefined
      }
      className={cn(
        "flex flex-col gap-2 rounded-lg border p-4 transition-colors",
        archived ? "bg-muted/30" : "hover:bg-muted/40",
        isClickable &&
          "cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
        className,
      )}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="flex min-w-0 flex-1 items-center gap-3">
          <div
            className={cn(
              "flex h-10 w-10 shrink-0 items-center justify-center rounded-lg",
              archived ? "bg-muted" : "bg-primary/10",
            )}
          >
            <Icon
              className={cn(
                "h-5 w-5",
                archived ? "text-muted-foreground" : "text-primary",
              )}
              aria-hidden
            />
          </div>
          <div className="min-w-0 flex-1">
            {titleTooltip ? (
              <Tooltip content={titleTooltip} side="top" align="start">
                <p className="truncate font-medium max-w-full sm:max-w-md">
                  {title}
                </p>
              </Tooltip>
            ) : (
              <p className="truncate font-medium max-w-full sm:max-w-md">
                {title}
              </p>
            )}
            {subtitle && (
              <p className="mt-0.5 truncate text-xs text-muted-foreground max-w-full sm:max-w-md">
                {subtitle}
              </p>
            )}
          </div>
        </div>
        {hasActions && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="shrink-0 -mr-1"
                onClick={(e) => e.stopPropagation()}
                aria-label="Open menu"
              >
                <MoreVertical className="h-4 w-4" aria-hidden />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              align="end"
              sideOffset={4}
              onClick={(e) => e.stopPropagation()}
            >
              {actions.map((action) => {
                const ActionIcon = action.icon;
                return (
                  <DropdownMenuItem
                    key={action.label}
                    onClick={action.onClick}
                    className={
                      action.variant === "destructive"
                        ? "text-destructive focus:text-destructive"
                        : undefined
                    }
                  >
                    <ActionIcon className="mr-2 h-4 w-4" aria-hidden />
                    {action.label}
                  </DropdownMenuItem>
                );
              })}
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </div>
      {description && (
        <Tooltip content={description} side="top" align="start">
          <p className="text-sm text-muted-foreground overflow-hidden line-clamp-2 [display:-webkit-box] [-webkit-line-clamp:2] [-webkit-box-orient:vertical]">
            {description}
          </p>
        </Tooltip>
      )}
      {children}
    </div>
  );
}
