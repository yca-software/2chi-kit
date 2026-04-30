import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@yca-software/design-system";
import type { Control, FieldPath, FieldValues } from "react-hook-form";
import type { UseFormReturn } from "react-hook-form";
import { LocationInput } from "@/components";

type AddressFieldProps<T extends FieldValues> = {
  control: Control<T>;
  name: FieldPath<T>;
  form: UseFormReturn<T>;
  label: string;
  placeholder?: string;
  displayValue?: string;
  className?: string;
};

export function AddressField<T extends FieldValues>({
  control,
  name,
  form,
  label,
  placeholder,
  displayValue,
  className,
}: AddressFieldProps<T>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem className={className}>
          <FormLabel>{label}</FormLabel>
          <FormControl>
            <LocationInput
              value={field.value}
              displayValue={displayValue}
              onChange={(placeId) => field.onChange(placeId ?? "")}
              onError={(err) => {
                form.setError(name, { type: "manual", message: err });
              }}
              placeholder={placeholder ?? label}
            />
          </FormControl>
          <FormMessage />
        </FormItem>
      )}
    />
  );
}
