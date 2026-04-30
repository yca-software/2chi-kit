import {
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Select,
} from "@yca-software/design-system";
import type { Control, FieldPath, FieldValues } from "react-hook-form";

type ExpiresAtOption = {
  value: string;
  label: string;
};

type ExpiresAtFieldProps<T extends FieldValues> = {
  control: Control<T>;
  name: FieldPath<T>;
  label: string;
  placeholder?: string;
  ariaLabel?: string;
  options: ExpiresAtOption[];
};

export function ExpiresAtField<T extends FieldValues>({
  control,
  name,
  label,
  placeholder,
  ariaLabel,
  options,
}: ExpiresAtFieldProps<T>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem>
          <FormLabel>{label}</FormLabel>
          <Select
            value={(field.value as string) ?? ""}
            onValueChange={(next) => {
              field.onChange(next);
            }}
            options={options}
            placeholder={placeholder}
            aria-label={ariaLabel ?? label}
          />
          <FormMessage />
        </FormItem>
      )}
    />
  );
}
