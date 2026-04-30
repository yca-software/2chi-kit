import * as React from "react";
import { useTranslation } from "react-i18next";
import { Button, H3, Paragraph } from "@yca-software/design-system";
import { useCookieConsent } from "@/states/cookieConsent";

export const CookieBanner = () => {
  const { consent, setCookieConsent } = useCookieConsent();
  const { t } = useTranslation();
  const [isVisible, setIsVisible] = React.useState(!consent);

  React.useEffect(() => {
    setIsVisible(!consent);
  }, [consent]);

  if (!isVisible) return null;

  const handleAcceptAll = () => {
    setCookieConsent(new Date().toISOString(), true);
    setIsVisible(false);
  };

  const handleAcceptNecessary = () => {
    setCookieConsent(new Date().toISOString(), false);
    setIsVisible(false);
  };

  const policyLinks = [
    { to: "/terms-of-service", label: t("cookieBanner.termsOfService") },
    { to: "/cookie-policy", label: t("cookieBanner.cookiePolicy") },
  ];

  return (
    <div className="fixed bottom-0 left-0 right-0 z-[100] border-t border-border bg-background shadow-2xl md:bg-background/95 md:backdrop-blur-sm">
      <div className="mx-auto max-w-7xl px-4 py-4 sm:px-6 lg:px-8">
        <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div className="flex-1 space-y-2">
            <H3 className="text-base sm:text-lg">{t("cookieBanner.title")}</H3>
            <Paragraph size="sm">{t("cookieBanner.description")}</Paragraph>
            <div className="flex flex-wrap items-center gap-4 text-sm">
              {policyLinks.map((link, index) => (
                <span key={link.to} className="flex items-center gap-4">
                  <a
                    href={link.to}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="underline underline-offset-4 hover:text-primary"
                  >
                    {link.label}
                  </a>
                  {index < policyLinks.length - 1 && (
                    <span className="text-muted-foreground">•</span>
                  )}
                </span>
              ))}
            </div>
          </div>
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
            <Button
              variant="secondary"
              onClick={handleAcceptNecessary}
              className="w-full whitespace-nowrap sm:w-auto"
            >
              {t("cookieBanner.necessaryOnly")}
            </Button>
            <Button
              variant="default"
              onClick={handleAcceptAll}
              className="w-full whitespace-nowrap sm:w-auto"
            >
              {t("cookieBanner.acceptAll")}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};
