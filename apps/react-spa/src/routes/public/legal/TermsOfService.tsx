import { Helmet } from "react-helmet-async";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@yca-software/design-system";
import { useTranslationNamespace } from "@/helpers";
import { LEGAL_EFFECTIVE_DATE, LEGAL_VERSION } from "@/constants/legal";

function formatEffectiveDate(iso: string) {
  try {
    return new Date(iso + "T00:00:00").toLocaleDateString(undefined, {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  } catch {
    return iso;
  }
}

export const TermsOfService = () => {
  const { t, isLoading } = useTranslationNamespace(["legal"]);

  if (isLoading) return null;

  return (
    <>
      <Helmet>
        <title>{t("legal:termsOfService.title")}</title>
      </Helmet>
      <Card>
        <CardHeader>
          <CardTitle className="text-3xl">
            {t("legal:termsOfService.title")}
          </CardTitle>
          <p className="text-sm text-muted-foreground">
            {t("legal:termsOfService.lastUpdated")}{" "}
            {formatEffectiveDate(LEGAL_EFFECTIVE_DATE)}
            {" · "}
            {t("legal:termsOfService.version")} {LEGAL_VERSION}
          </p>
        </CardHeader>
        <CardContent className="prose prose-gray dark:prose-invert max-w-none">
          <section className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:termsOfService.acceptance.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:termsOfService.acceptance.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:termsOfService.useLicense.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:termsOfService.useLicense.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:termsOfService.userAccount.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:termsOfService.userAccount.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:termsOfService.prohibitedUses.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:termsOfService.prohibitedUses.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:termsOfService.termination.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:termsOfService.termination.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:termsOfService.disclaimer.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:termsOfService.disclaimer.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:termsOfService.privacyIntro.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:termsOfService.privacyIntro.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:termsOfService.informationWeCollect.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:termsOfService.informationWeCollect.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:termsOfService.howWeUse.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:termsOfService.howWeUse.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:termsOfService.dataSharing.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:termsOfService.dataSharing.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:termsOfService.dataSecurity.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:termsOfService.dataSecurity.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:termsOfService.yourRights.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:termsOfService.yourRights.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:termsOfService.cookies.title")}
              </h2>
              <p className="text-muted-foreground mt-2">
                {t("legal:termsOfService.cookies.description")}{" "}
                <a
                  href="/cookie-policy"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="underline underline-offset-4 hover:text-primary"
                >
                  {t("legal:termsOfService.cookies.cookiePolicyLink")}
                </a>
                .
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:termsOfService.contact.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:termsOfService.contact.description")}
              </p>
            </div>
          </section>
        </CardContent>
      </Card>
    </>
  );
};
