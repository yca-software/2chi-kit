import { Globe, Check } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  Button,
  getFlagEmoji,
} from "@yca-software/design-system";
import { LANGUAGES } from "@/constants";
import {
  getStoredLanguage,
  setStoredLanguage,
  getPreferredLanguage,
} from "@/helpers";
import { useUpdateLanguageMutation } from "@/api";
import { useUserState } from "@/states";
import { useShallow } from "zustand/shallow";
import { useQueryClient } from "@tanstack/react-query";
import { USER_QUERY_KEYS } from "@/constants/queryKeys";
import { useTranslation } from "react-i18next";

interface LanguageSelectorProps {
  variant?: "default" | "ghost" | "outline";
  size?: "default" | "sm" | "icon";
  showLabel?: boolean;
}

export const LanguageSelector = ({
  variant = "ghost",
  size = "icon",
}: LanguageSelectorProps) => {
  const { i18n } = useTranslation();
  const queryClient = useQueryClient();
  const { userData, tokens } = useUserState(
    useShallow((state) => ({
      userData: state.userData,
      tokens: state.tokens,
    })),
  );

  const userLanguage = userData.user?.language;
  const storedLanguage = getStoredLanguage();
  const currentLanguage =
    userLanguage || storedLanguage || getPreferredLanguage();
  const isAuthenticated = !!tokens.refresh;

  const updateLanguageMutation = useUpdateLanguageMutation({
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [USER_QUERY_KEYS.CURRENT] });
    },
    onError: () => {},
  });

  const handleLanguageChange = async (lang: string) => {
    setStoredLanguage(lang);
    await i18n.changeLanguage(lang);
    const namespaces = i18n.options.ns || [];
    await i18n.reloadResources(lang, namespaces);
    if (isAuthenticated && lang !== userLanguage) {
      updateLanguageMutation.mutate({ language: lang });
    }
  };

  const availableLanguages = LANGUAGES;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant={variant}
          size={size}
          className="cursor-pointer"
          aria-label="Select language"
        >
          <Globe className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {Object.entries(availableLanguages).map(([code, name]) => (
          <DropdownMenuItem
            key={code}
            onClick={() => handleLanguageChange(code)}
            className="flex items-center justify-between gap-2"
          >
            <span className="flex items-center gap-2">
              <span>{getFlagEmoji(code)}</span>
              <span>{name}</span>
            </span>
            {currentLanguage === code && <Check className="h-4 w-4" />}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
};
