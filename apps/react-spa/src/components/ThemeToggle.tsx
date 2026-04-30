import { ThemeToggle as SharedThemeToggle } from "@yca-software/design-system";
import { useThemeStore } from "@/states/theme";

export function ThemeToggle() {
  return <SharedThemeToggle useThemeStore={useThemeStore} />;
}
