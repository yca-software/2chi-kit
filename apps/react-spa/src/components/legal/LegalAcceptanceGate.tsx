import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Button,
  Checkbox,
  Paragraph,
} from "@yca-software/design-system";
import { LEGAL_VERSION } from "@/constants";
import { useAcceptTermsMutation } from "@/api";
import { useUserState } from "@/states";

export const LegalAcceptanceGate = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const isUserProfileReady = useUserState((state) => state.isUserProfileReady);
  const userData = useUserState((state) => state.userData);
  const { t } = useTranslation();
  const [accepted, setAccepted] = useState(false);
  const acceptMutation = useAcceptTermsMutation();

  // Don't block until user is loaded and we can check their accepted version
  const currentVersion = userData.user?.termsVersion;
  const mustAccept =
    isUserProfileReady && !!userData.user && currentVersion !== LEGAL_VERSION;

  const handleAccept = () => {
    if (!accepted) return;
    acceptMutation.mutate({ termsVersion: LEGAL_VERSION });
  };

  return (
    <>
      {children}
      {mustAccept && (
        <Dialog open onOpenChange={() => {}}>
          <DialogContent
            className="max-w-lg"
            showCloseButton={false}
            onPointerDownOutside={(e) => e.preventDefault()}
            onEscapeKeyDown={(e) => e.preventDefault()}
          >
            <DialogHeader>
              <DialogTitle>
                {t("legalAcceptance.title", { ns: "common" })}
              </DialogTitle>
              <DialogDescription asChild>
                <Paragraph size="sm" className="text-muted-foreground mt-1">
                  {t("legalAcceptance.description", { ns: "common" })}
                </Paragraph>
              </DialogDescription>
            </DialogHeader>
            <div className="flex flex-col gap-4 py-4">
              <a
                href="/terms-of-service"
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm underline underline-offset-4 hover:text-primary"
              >
                {t("legalAcceptance.viewTermsLink", { ns: "common" })}
              </a>
              <label className="flex items-center gap-2 cursor-pointer">
                <Checkbox
                  checked={accepted}
                  onCheckedChange={(checked: boolean | "indeterminate") =>
                    setAccepted(checked === true)
                  }
                />
                <span className="text-sm">
                  {t("legalAcceptance.checkboxLabel", { ns: "common" })}
                </span>
              </label>
            </div>
            <DialogFooter>
              <Button
                onClick={handleAccept}
                disabled={!accepted || acceptMutation.isPending}
              >
                {acceptMutation.isPending
                  ? t("loading", { ns: "common" })
                  : t("legalAcceptance.acceptButton", { ns: "common" })}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </>
  );
};
