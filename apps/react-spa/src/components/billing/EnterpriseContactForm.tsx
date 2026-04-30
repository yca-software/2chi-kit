import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Button,
  Form,
  Alert,
  AlertDescription,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  Textarea,
} from "@yca-software/design-system";
import { Loader2 } from "lucide-react";
import { useTranslationNamespace } from "@/helpers";
import { useSupportMutation } from "@/api";

type EnterpriseFormData = { message: string };

interface EnterpriseContactFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  organizationId?: string;
  organizationName?: string;
  onSuccess?: () => void;
}

export function EnterpriseContactForm({
  open,
  onOpenChange,
  organizationId,
  organizationName,
  onSuccess,
}: EnterpriseContactFormProps) {
  const { t } = useTranslationNamespace(["pricing", "common"]);
  const enterpriseSchema = z.object({
    message: z.string().min(1, t("pricing:enterpriseForm.messageRequired")),
  });
  const form = useForm<EnterpriseFormData>({
    resolver: zodResolver(enterpriseSchema),
    defaultValues: { message: "" },
  });
  const mutation = useSupportMutation({
    onSuccess: () => {
      form.reset({ message: "" });
      onOpenChange(false);
      onSuccess?.();
    },
  });

  const onSubmit = (data: EnterpriseFormData) => {
    mutation.mutate({
      message: data.message,
      subject: `Enterprise billing request from ${organizationName} (${organizationId})`,
    });
  };

  const success = mutation.isSuccess;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("pricing:enterpriseForm.title")}</DialogTitle>
          <DialogDescription>
            {t("pricing:enterpriseForm.description")}
          </DialogDescription>
        </DialogHeader>
        {success ? (
          <p className="py-4 text-sm text-muted-foreground">
            {t("pricing:enterpriseForm.successMessage")}
          </p>
        ) : (
          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(onSubmit)}
              className="flex flex-col gap-4"
            >
              {mutation.error && (
                <Alert variant="destructive">
                  <AlertDescription>
                    {mutation.error.error?.message ??
                      t("defaultError", { ns: "common" })}
                  </AlertDescription>
                </Alert>
              )}
              <div className="space-y-2">
                <label className="text-sm font-medium leading-none">
                  {t("pricing:enterpriseForm.messageLabel")}
                </label>
                <Textarea
                  {...form.register("message")}
                  placeholder={t("pricing:enterpriseForm.messagePlaceholder")}
                  className="min-h-[100px]"
                />
                {form.formState.errors.message && (
                  <p className="text-sm text-destructive">
                    {form.formState.errors.message.message}
                  </p>
                )}
              </div>
              <DialogFooter>
                <Button type="submit" disabled={mutation.isPending}>
                  {mutation.isPending ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      {t("common:loading")}
                    </>
                  ) : (
                    t("pricing:enterpriseForm.submitButton")
                  )}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        )}
      </DialogContent>
    </Dialog>
  );
}
