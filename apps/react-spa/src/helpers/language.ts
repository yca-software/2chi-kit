import { LANGUAGES, DEFAULT_LANGUAGE } from "@/constants";

const LANGUAGE_STORAGE_KEY = "user-language-preference";

export const getStoredLanguage = (): string | null => {
  if (typeof window === "undefined") return null;
  try {
    const stored = localStorage.getItem(LANGUAGE_STORAGE_KEY);
    if (stored && Object.keys(LANGUAGES).includes(stored)) {
      return stored;
    }
  } catch (error) {
    console.warn("Failed to read language from localStorage:", error);
  }
  return null;
};

export const setStoredLanguage = (language: string): void => {
  if (typeof window === "undefined") return;
  try {
    if (Object.keys(LANGUAGES).includes(language)) {
      localStorage.setItem(LANGUAGE_STORAGE_KEY, language);
    }
  } catch (error) {
    console.warn("Failed to save language to localStorage:", error);
  }
};

export const getBrowserLanguage = (): string | null => {
  if (typeof window === "undefined") return null;
  const browserLang = navigator.language.split("-")[0].toLowerCase();
  if (Object.keys(LANGUAGES).includes(browserLang)) {
    return browserLang;
  }
  return null;
};

export const getPreferredLanguage = (): string => {
  return getStoredLanguage() || getBrowserLanguage() || DEFAULT_LANGUAGE;
};
