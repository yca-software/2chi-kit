import { Info } from "lucide-react";
import {
  Checkbox,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Tooltip,
} from "@yca-software/design-system";
import type { Control, FieldPath, FieldValues } from "react-hook-form";

export type PermissionAction = "read" | "write" | "delete";

export interface PermissionOption {
  key: string;
  action: PermissionAction;
  labelKey: string;
}

export interface PermissionGroup {
  contextKey: string;
  labelKey: string;
  descriptionKey: string;
  permissions: PermissionOption[];
}

type PermissionsFieldProps<T extends FieldValues> = {
  control: Control<T>;
  name: FieldPath<T>;
  label: string;
  groups: PermissionGroup[];
  t: (key: string) => string;
  disabled?: boolean;
};

export function PermissionsField<T extends FieldValues>({
  control,
  name,
  label,
  groups,
  t,
  disabled = false,
}: PermissionsFieldProps<T>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => {
        const selected = (field.value as string[]) ?? [];

        const togglePermission = (key: string, group: PermissionOption[]) => {
          const readKey = getReadKeyForGroup(group);
          const opt = group.find((p) => p.key === key);
          if (!opt) return;

          const next = selected.includes(key)
            ? selected.filter((k) => k !== key)
            : [...selected, key];

          if (opt.action === "write" || opt.action === "delete") {
            if (next.includes(key) && readKey && !next.includes(readKey)) {
              next.push(readKey);
            }
          }
          if (opt.action === "delete") {
            const writeKey = getWriteKeyForGroup(group);
            if (next.includes(key) && writeKey && !next.includes(writeKey)) {
              next.push(writeKey);
            }
          }

          field.onChange(next);
        };

        return (
          <FormItem>
            <FormLabel>{label}</FormLabel>
            <div className="space-y-4">
              {groups.map((group) => {
                const readKey = getReadKeyForGroup(group.permissions);
                const writeKey = getWriteKeyForGroup(group.permissions);
                const readDisabled = readKey
                  ? hasWriteOrDelete(group.permissions, selected)
                  : false;
                const writeDisabled = writeKey
                  ? hasDelete(group.permissions, selected)
                  : false;
                const readChecked =
                  readKey !== null && (selected.includes(readKey) || readDisabled);
                const writeChecked =
                  writeKey !== null &&
                  (selected.includes(writeKey) || writeDisabled);

                return (
                  <div key={group.contextKey} className="space-y-2">
                    <div className="flex items-center gap-1.5">
                      <p className="text-sm font-medium">{t(group.labelKey)}</p>
                      <Tooltip
                        content={t(group.descriptionKey)}
                        side="top"
                        align="start"
                      >
                        <button
                          type="button"
                          className="text-muted-foreground hover:text-foreground cursor-help inline-flex rounded p-0.5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                          aria-label={t(group.descriptionKey)}
                        >
                          <Info className="size-3.5" aria-hidden />
                        </button>
                      </Tooltip>
                    </div>
                    <div className="flex flex-wrap gap-4 pl-2">
                      {group.permissions.map((perm) => {
                        const isRead = perm.action === "read";
                        const isWrite = perm.action === "write";
                        const checked = isRead
                          ? readChecked
                          : isWrite
                            ? writeChecked
                            : selected.includes(perm.key);
                        const permissionDisabled =
                          (isRead && readDisabled) || (isWrite && writeDisabled);
                        const globallyDisabled = disabled || permissionDisabled;

                        return (
                          <div key={perm.key} className="flex items-center space-x-2">
                            <Checkbox
                              id={perm.key}
                              checked={isRead ? readChecked : checked}
                              disabled={globallyDisabled}
                              onCheckedChange={() =>
                                !globallyDisabled &&
                                togglePermission(perm.key, group.permissions)
                              }
                            />
                            <label
                              htmlFor={perm.key}
                              className="text-sm font-normal leading-none cursor-pointer peer-disabled:opacity-70"
                            >
                              {t(perm.labelKey)}
                            </label>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </div>
            <FormMessage />
          </FormItem>
        );
      }}
    />
  );
}

function getReadKeyForGroup(permissions: PermissionOption[]): string | null {
  const readOpt = permissions.find((p) => p.action === "read");
  return readOpt?.key ?? null;
}

function getWriteKeyForGroup(permissions: PermissionOption[]): string | null {
  const writeOpt = permissions.find((p) => p.action === "write");
  return writeOpt?.key ?? null;
}

function hasWriteOrDelete(
  permissions: PermissionOption[],
  selected: string[],
): boolean {
  return permissions.some(
    (p) =>
      (p.action === "write" || p.action === "delete") &&
      selected.includes(p.key),
  );
}

function hasDelete(permissions: PermissionOption[], selected: string[]): boolean {
  return permissions.some(
    (p) => p.action === "delete" && selected.includes(p.key),
  );
}
