import { cn } from "@yca-software/design-system";

export interface DetailField {
  label: string;
  value: React.ReactNode;
  /** Span 2 columns on sm+ */
  span?: 1 | 2;
}

export interface DetailFieldListProps {
  fields: DetailField[];
  className?: string;
}

/**
 * Definition list for detail views. Responsive grid layout.
 */
export function DetailFieldList({ fields, className }: DetailFieldListProps) {
  return (
    <dl className={cn("grid gap-4 sm:grid-cols-2", className)}>
      {fields.map((field, i) => (
        <div
          key={i}
          className={cn("space-y-1", field.span === 2 && "sm:col-span-2")}
        >
          <dt className="text-sm font-medium text-muted-foreground">
            {field.label}
          </dt>
          <dd className="text-sm">{field.value}</dd>
        </div>
      ))}
    </dl>
  );
}
