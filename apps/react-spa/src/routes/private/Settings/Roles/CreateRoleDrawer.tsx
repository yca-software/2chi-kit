import { useTranslationNamespace } from "@/helpers";
import { FormDrawer } from "@yca-software/design-system";
import { RoleForm } from "./RoleForm";
import { useMemo } from "react";
import z from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useCreateRoleMutation } from "@/api";
import { MutationError } from "@/types";
import { toast } from "sonner";

interface CreateRoleDrawerProps {
  open: boolean;
  onClose: () => void;
  currentOrgId: string;
}

export type RoleFormData = {
  name: string;
  description: string;
  permissions: string[];
};

export function CreateRoleDrawer({
  open,
  onClose,
  currentOrgId,
}: CreateRoleDrawerProps) {
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
      name: "",
      description: "",
      permissions: [],
    },
  });

  const createMutation = useCreateRoleMutation(currentOrgId, {
    onSuccess: () => {
      onClose();
      form.reset();
    },
    onError: (err: MutationError) => {
      toast.error(err.error?.message ?? t("common:defaultError"));
    },
  });

  return (
    <FormDrawer
      open={open}
      onOpenChange={onClose}
      title={t("settings:org.roles.createRole")}
      description={t("settings:org.roles.createRoleDescription")}
    >
      <RoleForm
        mode="create"
        form={form}
        onSubmit={createMutation.mutate}
        isPending={createMutation.isPending}
      />
    </FormDrawer>
  );
}
