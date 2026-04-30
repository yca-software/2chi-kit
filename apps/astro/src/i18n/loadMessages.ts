import type { Locale } from "./config";
import en from "./locales/en.json";
import fr from "./locales/fr.json";

export type Messages = typeof en;

const catalogs: Record<Locale, Messages> = { en, fr };

export function loadMessages(locale: Locale): Messages {
	return catalogs[locale];
}
