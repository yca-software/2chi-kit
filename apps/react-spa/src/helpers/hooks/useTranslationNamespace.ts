import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { loadNamespace } from "@/i18n";

export function useTranslationNamespace(namespaces: string[] | string) {
  const namespaceArray = Array.isArray(namespaces) ? namespaces : [namespaces];
  const { t, i18n } = useTranslation(["common", ...namespaceArray]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadNamespaces = async () => {
      const language = i18n.language || i18n.options.fallbackLng || "en";

      const namespacesToLoad = namespaceArray.filter(
        (ns) => !i18n.hasResourceBundle(language as string, ns),
      );

      if (namespacesToLoad.length === 0) {
        setIsLoading(false);
        return;
      }

      setIsLoading(true);
      try {
        await Promise.all(
          namespacesToLoad.map((ns) => loadNamespace(ns, language as string)),
        );
      } catch (error) {
        console.error("Failed to load namespaces:", error);
      } finally {
        setIsLoading(false);
      }
    };

    loadNamespaces();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [i18n.language, namespaceArray.join(",")]);

  return { t, isLoading, i18n };
}
