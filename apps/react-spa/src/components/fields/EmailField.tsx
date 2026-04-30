import { useTranslation } from "react-i18next";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
} from "@yca-software/design-system";
import type { Control, FieldPath, FieldValues } from "react-hook-form";

type EmailFieldProps<T extends FieldValues> = {
  control: Control<T>;
  name: FieldPath<T>;
  label?: string;
  placeholder?: string;
  description?: string;
  autoComplete?: "email" | "username";
  className?: string;
};

export function EmailField<T extends FieldValues>({
  control,
  name,
  label,
  placeholder,
  description,
  autoComplete = "email",
  className,
}: EmailFieldProps<T>) {
  const { t } = useTranslation("common");
  const resolvedLabel = label ?? t("fields.email");
  const resolvedPlaceholder = placeholder ?? t("fields.emailPlaceholder");

  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem className={className}>
          <FormLabel>{resolvedLabel}</FormLabel>
          {description && (
            <p className="text-sm text-muted-foreground mt-0.5">
              {description}
            </p>
          )}
          <FormControl>
            <Input
              type="email"
              placeholder={resolvedPlaceholder}
              autoComplete={autoComplete}
              {...field}
            />
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}
