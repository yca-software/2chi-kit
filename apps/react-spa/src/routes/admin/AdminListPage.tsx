import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Input,
  cn,
} from "@yca-software/design-system";
import { Loader2, Search } from "lucide-react";

export interface AdminListPageColumn<T> {
  key: string;
  header: string;
  render: (item: T) => React.ReactNode;
  /** Column class for alignment/width */
  className?: string;
}

export interface AdminListPageProps<T> {
  title: string;
  description: string;
  /** Optional actions (e.g. "Create" button) rendered next to the page title. */
  headerActions?: React.ReactNode;
  /** Card header title (e.g. "All Users") */
  cardTitle: string;
  searchPlaceholder: string;
  emptyMessage: string;
  columns: AdminListPageColumn<T>[];
  items: T[];
  isLoading: boolean;
  hasNextPage: boolean;
  isFetchingNextPage: boolean;
  loadMoreRef: React.RefObject<HTMLTableRowElement | null>;
  search: string;
  onSearchChange: (value: string) => void;
  onRowClick: (item: T) => void;
  getRowKey: (item: T) => string;
}

/**
 * Reusable admin list page: search + card + table + infinite scroll.
 * Mobile-friendly: responsive table, touch targets.
 */
export function AdminListPage<T>({
  title,
  description,
  headerActions,
  cardTitle,
  searchPlaceholder,
  emptyMessage,
  columns,
  items,
  isLoading,
  hasNextPage,
  isFetchingNextPage,
  loadMoreRef,
  search,
  onSearchChange,
  onRowClick,
  getRowKey,
}: AdminListPageProps<T>) {
  return (
    <div className="space-y-4 sm:space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold sm:text-3xl">{title}</h1>
          <p className="mt-1 text-sm text-muted-foreground sm:text-base">
            {description}
          </p>
        </div>
        {headerActions}
      </div>

      <Card>
        <CardHeader className="gap-4">
          <div className="space-y-2">
            <CardTitle>{cardTitle}</CardTitle>
            <CardDescription>{searchPlaceholder}</CardDescription>
          </div>
          <div className="relative max-w-full sm:max-w-sm">
            <Search
              className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
              aria-hidden
            />
            <Input
              placeholder={searchPlaceholder}
              value={search}
              onChange={(e) => onSearchChange(e.target.value)}
              className="pl-9"
              aria-label={searchPlaceholder}
            />
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div
              className="flex justify-center py-12"
              role="status"
              aria-live="polite"
            >
              <Loader2
                className="h-8 w-8 animate-spin text-primary"
                aria-hidden
              />
              <span className="sr-only">Loading...</span>
            </div>
          ) : items.length === 0 ? (
            <p className="py-12 text-center text-muted-foreground">
              {emptyMessage}
            </p>
          ) : (
            <div className="overflow-x-auto -mx-4 sm:mx-0 rounded-md border">
              <table className="w-full min-w-[400px] text-sm">
                <thead>
                  <tr className="border-b bg-muted/50">
                    {columns.map((col) => (
                      <th
                        key={col.key}
                        className={cn(
                          "px-4 py-3 text-left font-medium",
                          col.className,
                        )}
                      >
                        {col.header}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {items.map((item) => (
                    <tr
                      key={getRowKey(item)}
                      role="button"
                      tabIndex={0}
                      onClick={() => onRowClick(item)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter" || e.key === " ") {
                          e.preventDefault();
                          onRowClick(item);
                        }
                      }}
                      className="border-b last:border-0 hover:bg-muted/30 cursor-pointer transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset"
                    >
                      {columns.map((col) => (
                        <td
                          key={col.key}
                          className={cn(
                            "px-4 py-3",
                            col.className ?? "text-muted-foreground",
                          )}
                        >
                          {col.render(item)}
                        </td>
                      ))}
                    </tr>
                  ))}
                  {hasNextPage && (
                    <tr ref={loadMoreRef}>
                      <td
                        colSpan={columns.length}
                        className="px-4 py-4 text-center text-muted-foreground"
                      >
                        {isFetchingNextPage ? (
                          <Loader2
                            className="inline-block h-5 w-5 animate-spin"
                            aria-hidden
                          />
                        ) : null}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
