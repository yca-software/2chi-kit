import { useTranslation } from "react-i18next";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  cn,
} from "@yca-software/design-system";
import { ChevronDown } from "lucide-react";
import type { Control, FieldPath, FieldValues } from "react-hook-form";

export type RoleOption = { value: string; label: string };

type RoleSelectFieldProps<T extends FieldValues> = {
  control: Control<T>;
  name: FieldPath<T>;
  options: RoleOption[];
  className?: string;
};

export function RoleSelectField<T extends FieldValues>({
  control,
  name,
  options,
  className,
}: RoleSelectFieldProps<T>) {
  const { t } = useTranslation("settings");
  const label = t("org.members.selectRole");

  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => {
        const selectedOption = options.find((opt) => opt.value === field.value);
        const displayValue = selectedOption?.label ?? label;

        return (
          <FormItem className={className}>
            <FormLabel>{label}</FormLabel>
            <FormControl>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button
                    type="button"
                    role="combobox"
                    aria-label={label}
                    onBlur={field.onBlur}
                    className={cn(
                      "flex h-9 w-full cursor-pointer items-center justify-between rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs outline-none transition-colors",
                      "hover:bg-accent/50 focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
                      "disabled:pointer-events-none disabled:opacity-50",
                      "aria-invalid:ring-destructive/20 aria-invalid:border-destructive",
                      !field.value && "text-muted-foreground",
                    )}
                  >
                    <span className="min-w-0 truncate text-left">
                      {displayValue}
                    </span>
                    <ChevronDown className="ml-2 h-4 w-4 shrink-0 text-muted-foreground" />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="start"
                  sideOffset={0}
                  className={cn(
                    "min-w-(--radix-dropdown-menu-trigger-width) z-101",
                  )}
                >
                  {options.map((opt) => (
                    <DropdownMenuItem
                      key={opt.value}
                      onClick={() => field.onChange(opt.value)}
                    >
                      {opt.label}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </FormControl>
            <FormMessage />
          </FormItem>
        );
      }}
    />
  );
}
