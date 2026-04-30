import { useTranslationNamespace } from "@/helpers";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Button,
  Form,
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetFooter,
} from "@yca-software/design-system";
import { EmailField, RoleSelectField } from "@/components";
import { Loader2 } from "lucide-react";
import type { Role, MutationError } from "@/types";
import { useCreateInvitationMutation } from "@/api";
import { toast } from "sonner";

export interface InviteFormData {
  email: string;
  roleId: string;
}

interface InviteDrawerProps {
  open: boolean;
  onClose: () => void;
  roles: Role[];
  currentOrgId: string;
}

export function InviteDrawer({
  open,
  onClose,
  roles,
  currentOrgId,
}: InviteDrawerProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);

  const inviteSchema = z.object({
    email: z
      .string()
      .toLowerCase()
      .trim()
      .email({ message: t("settings:org.validation.emailInvalid") }),
    roleId: z.string().min(1, t("settings:org.validation.roleRequired")),
  });

  const form = useForm<InviteFormData>({
    resolver: zodResolver(inviteSchema),
    defaultValues: { email: "", roleId: "" },
  });

  const createInvitationMutation = useCreateInvitationMutation(currentOrgId, {
    onSuccess: () => {
      form.reset();
      onClose();
    },
    onError: (err: MutationError) => {
      toast.error(err.error?.message ?? t("common:defaultError"));
    },
  });

  return (
    <Sheet open={open} onOpenChange={onClose}>
      <SheetContent className="sm:max-w-lg">
        <SheetHeader>
          <SheetTitle>{t("settings:org.members.inviteMember")}</SheetTitle>
          <SheetDescription>
            {t("settings:org.members.inviteDescription")}
          </SheetDescription>
        </SheetHeader>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit((data) => {
              createInvitationMutation.mutate({
                email: data.email,
                roleId: data.roleId,
              });
            })}
            className="flex flex-col gap-6"
          >
            <EmailField
              control={form.control}
              name="email"
              autoComplete="email"
            />

            <RoleSelectField
              control={form.control}
              name="roleId"
              options={roles.map((r) => ({ value: r.id, label: r.name }))}
            />

            <SheetFooter>
              <Button
                type="submit"
                disabled={createInvitationMutation.isPending}
              >
                {createInvitationMutation.isPending && (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                )}
                {t("common:invite")}
              </Button>
            </SheetFooter>
          </form>
        </Form>
      </SheetContent>
    </Sheet>
  );
}
