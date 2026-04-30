import { useTranslationNamespace } from "@/helpers/hooks/useTranslationNamespace";
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogCancel,
  Button,
  Input,
} from "@yca-software/design-system";
import { useState } from "react";

interface KeyCreatedDialogProps {
  open: boolean;
  keyValue: string;
  onClose: () => void;
}

export function KeyCreatedDialog({
  open,
  keyValue,
  onClose,
}: KeyCreatedDialogProps) {
  const { t } = useTranslationNamespace(["settings", "common"]);
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(keyValue);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <AlertDialog open={open} onOpenChange={onClose}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            {t("settings:org.apiKeys.keyCreatedTitle")}
          </AlertDialogTitle>
          <AlertDialogDescription>
            {t("settings:org.apiKeys.keyCreated")}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <div className="space-y-4 pt-2">
          <div className="flex gap-2">
            <Input value={keyValue} readOnly className="font-mono text-sm" />
            <Button onClick={handleCopy} variant="secondary">
              {copied ? t("common:copied") : t("settings:org.apiKeys.copyKey")}
            </Button>
          </div>
        </div>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={onClose}>
            {t("common:close")}
          </AlertDialogCancel>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
