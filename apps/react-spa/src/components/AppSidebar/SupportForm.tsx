import { useMemo, useState } from "react";
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
import { NameField } from "@/components";

type SupportFormData = {
  subject: string;
  message: string;
};

interface SupportFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function SupportForm({ open, onOpenChange }: SupportFormProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const supportSchema = useMemo(
    () =>
      z.object({
        subject: z
          .string()
          .min(1, t("settings:support.validation.subjectRequired")),
        message: z
          .string()
          .min(1, t("settings:support.validation.messageRequired")),
      }),
    [t],
  );
  const mutation = useSupportMutation({
    onSuccess: () => {
      setSuccess(true);
      setTimeout(() => {
        onOpenChange(false);
        setSuccess(false);
        form.reset();
      }, 5000);
    },
  });

  const [success, setSuccess] = useState(false);
  const form = useForm<SupportFormData>({
    resolver: zodResolver(supportSchema),
    defaultValues: { subject: "", message: "" },
  });

  const onSubmit = (data: SupportFormData) => {
    mutation.mutate({
      subject: data.subject,
      message: data.message,
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("settings:support.title")}</DialogTitle>
          <DialogDescription>
            {t("settings:support.description")}
          </DialogDescription>
        </DialogHeader>
        {success ? (
          <p className="py-4 text-sm text-muted-foreground">
            {t("settings:support.successMessage")}
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
              <NameField
                control={form.control}
                name="subject"
                label={t("settings:support.subjectLabel")}
                placeholder={t("settings:support.subjectPlaceholder")}
              />
              <div className="space-y-2">
                <label className="text-sm font-medium leading-none">
                  {t("settings:support.messageLabel")}
                </label>
                <Textarea
                  {...form.register("message")}
                  placeholder={t("settings:support.messagePlaceholder")}
                  className="min-h-[120px]"
                />
                {form.formState.errors.message && (
                  <p className="text-sm text-destructive">
                    {form.formState.errors.message.message}
                  </p>
                )}
              </div>
              <DialogFooter>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => onOpenChange(false)}
                >
                  {t("common:cancel")}
                </Button>
                <Button type="submit" disabled={mutation.isPending}>
                  {mutation.isPending ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      {t("common:loading")}
                    </>
                  ) : (
                    t("settings:support.submitButton")
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
