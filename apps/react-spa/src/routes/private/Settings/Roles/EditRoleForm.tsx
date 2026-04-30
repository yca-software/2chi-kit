import { useTranslationNamespace } from "@/helpers";
import { MutationError, Role } from "@/types";
import { useMemo } from "react";
import { useForm } from "react-hook-form";
import z from "zod";
import { RoleForm, RoleFormData } from "./RoleForm";
import { zodResolver } from "@hookform/resolvers/zod";
import { useUpdateRoleMutation } from "@/api";
import { toast } from "sonner";

export interface EditRoleFormProps {
  currentOrgId: string;
  selectedRole: Role;
  onRoleUpdated: (role: Role) => void;
  onClose: () => void;
}

export function EditRoleForm({
  currentOrgId,
  selectedRole,
  onRoleUpdated,
  onClose,
}: EditRoleFormProps) {
  const { t, isLoading: tLoading } = useTranslationNamespace([
    "settings",
    "common",
  ]);

  const roleSchema = useMemo(() => {
    if (tLoading) {
      return z.object({
        name: z.string(),
        description: z.string(),
        permissions: z.array(z.string()),
      });
    }
    return z.object({
      name: z.string().min(1, t("settings:org.validation.nameRequired")),
      description: z.string(),
      permissions: z
        .array(z.string())
        .min(1, t("settings:org.validation.permissionsRequired")),
    });
  }, [t, tLoading]);

  const form = useForm<RoleFormData>({
    resolver: zodResolver(roleSchema),
    defaultValues: {
      name: selectedRole.name,
      description: selectedRole.description ?? "",
      permissions: selectedRole.permissions ?? [],
    },
  });

  const updateMutation = useUpdateRoleMutation(currentOrgId, selectedRole.id, {
    onSuccess: (updatedRole: Role) => {
      form.reset({
        name: updatedRole.name,
        description: updatedRole.description,
        permissions: updatedRole.permissions,
      });
      onRoleUpdated(updatedRole);
    },
    onError: (err: MutationError) => {
      toast.error(err.error?.message ?? t("common:defaultError"));
    },
  });

  return (
    <RoleForm
      mode="edit"
      form={form}
      onSubmit={(data) =>
        updateMutation.mutate({
          name: data.name,
          description: data.description,
          permissions: data.permissions,
        })
      }
      isPending={updateMutation.isPending}
      onCancel={onClose}
    />
  );
}
