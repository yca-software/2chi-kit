import { Helmet } from "react-helmet-async";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@yca-software/design-system";
import { useTranslationNamespace } from "@/helpers";
import { COOKIE_POLICY_EFFECTIVE_DATE } from "@/constants";

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

export const CookiePolicy = () => {
  const { t, isLoading } = useTranslationNamespace(["legal"]);

  if (isLoading) return null;

  return (
    <>
      <Helmet>
        <title>{t("legal:cookiePolicy.title")}</title>
      </Helmet>
      <Card>
        <CardHeader>
          <CardTitle className="text-3xl">
            {t("legal:cookiePolicy.title")}
          </CardTitle>
          <p className="text-sm text-muted-foreground">
            {t("legal:cookiePolicy.lastUpdated")}{" "}
            {formatEffectiveDate(COOKIE_POLICY_EFFECTIVE_DATE)}
          </p>
        </CardHeader>
        <CardContent className="prose prose-gray dark:prose-invert max-w-none">
          <section className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:cookiePolicy.whatAreCookies.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:cookiePolicy.whatAreCookies.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:cookiePolicy.howWeUseCookies.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:cookiePolicy.howWeUseCookies.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:cookiePolicy.typesOfCookies.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:cookiePolicy.typesOfCookies.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:cookiePolicy.managingCookies.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:cookiePolicy.managingCookies.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:cookiePolicy.updates.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:cookiePolicy.updates.description")}
              </p>
            </div>

            <div>
              <h2 className="text-xl font-semibold">
                {t("legal:cookiePolicy.contact.title")}
              </h2>
              <p className="text-muted-foreground mt-2 whitespace-pre-line">
                {t("legal:cookiePolicy.contact.description")}
              </p>
            </div>
          </section>
        </CardContent>
      </Card>
    </>
  );
};
