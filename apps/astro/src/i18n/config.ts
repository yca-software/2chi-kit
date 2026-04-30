/** Locales shipped with the template — add a folder under `src/pages/<locale>/` and a catalog file when extending. */
export const locales = ["en", "fr"] as const;

export type Locale = (typeof locales)[number];

export const defaultLocale: Locale = "en";

export function isLocale(value: string | undefined): value is Locale {
	return value === "en" || value === "fr";
}
