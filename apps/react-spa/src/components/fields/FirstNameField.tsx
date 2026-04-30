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

type FirstNameFieldProps<T extends FieldValues> = {
  control: Control<T>;
  name: FieldPath<T>;
  placeholder?: string;
  className?: string;
};

export function FirstNameField<T extends FieldValues>({
  control,
  name,
  placeholder,
  className,
}: FirstNameFieldProps<T>) {
  const { t } = useTranslation("common");

  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem className={className}>
          <FormLabel>{t("fields.firstName")}</FormLabel>
          <FormControl>
            <Input
              placeholder={placeholder ?? t("fields.firstNamePlaceholder")}
              autoComplete="given-name"
              {...field}
            />
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}
