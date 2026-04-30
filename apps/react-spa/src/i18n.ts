import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import { getPreferredLanguage } from "./helpers/language";

import enCommon from "./locales/en/common.json";

const resources = {
  en: { common: enCommon },
};

// Get initial language from localStorage or browser
const initialLanguage = getPreferredLanguage();

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    lng: initialLanguage,
    fallbackLng: "en",
    supportedLngs: ["en"],
    defaultNS: "common",
    ns: ["common"],
    interpolation: {
      escapeValue: false,
    },
    react: {
      useSuspense: false,
    },
    // Allow accessing translations from multiple namespaces
    compatibilityJSON: "v4",
  });

// Helper function to load a namespace dynamically
export const loadNamespace = async (namespace: string, lng?: string) => {
  const language = lng || i18n.language;

  if (i18n.hasResourceBundle(language, namespace)) {
    return;
  }

  try {
    const module = await import(`./locales/${language}/${namespace}.json`);
    i18n.addResourceBundle(language, namespace, module.default, true, true);
    if (!i18n.options.ns?.includes(namespace)) {
      i18n.options.ns = [...(i18n.options.ns || []), namespace];
    }
    await i18n.reloadResources(language, [namespace]);
  } catch (error) {
    if (language !== "en") {
      try {
        const module = await import(`./locales/en/${namespace}.json`);
        i18n.addResourceBundle(language, namespace, module.default, true, true);
        if (!i18n.options.ns?.includes(namespace)) {
          i18n.options.ns = [...(i18n.options.ns || []), namespace];
        }
        await i18n.reloadResources(language, [namespace]);
      } catch (fallbackError) {
        console.error(`Failed to load namespace ${namespace}:`, error);
      }
    } else {
      console.error(`Failed to load namespace ${namespace}:`, error);
    }
  }
};

export default i18n;
