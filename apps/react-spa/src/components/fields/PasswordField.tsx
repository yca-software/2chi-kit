import type { ReactNode } from "react";
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

type PasswordFieldProps<T extends FieldValues> = {
  control: Control<T>;
  name: FieldPath<T>;
  /** Override label (default: common.fields.password) */
  label?: string;
  /** Override placeholder (default: common.fields.passwordPlaceholder) */
  placeholder?: string;
  autoComplete?: "current-password" | "new-password";
  /** Optional node to show on the right of the label (e.g. "Forgot password?" link) */
  rightLabel?: ReactNode;
  className?: string;
};

export function PasswordField<T extends FieldValues>({
  control,
  name,
  label,
  placeholder,
  autoComplete = "current-password",
  rightLabel,
  className,
}: PasswordFieldProps<T>) {
  const { t } = useTranslation("common");
  const resolvedLabel = label ?? t("fields.password");
  const resolvedPlaceholder = placeholder ?? t("fields.passwordPlaceholder");

  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem className={className}>
          <div className="flex items-center justify-between">
            <FormLabel>{resolvedLabel}</FormLabel>
            {rightLabel}
          </div>
          <FormControl>
            <Input
              type="password"
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
