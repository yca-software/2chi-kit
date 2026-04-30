import { ThemeToggle } from "./ThemeToggle";
import { LanguageSelector } from "./LanguageSelector";
import { Logo } from "./Logo";
import { cn } from "@yca-software/design-system";

interface SiteHeaderProps {
  className?: string;
  showLanguageSelector?: boolean;
}

export const SiteHeader = ({
  className,
  showLanguageSelector = true,
}: SiteHeaderProps) => {
  return (
    <header
      className={cn(
        "flex min-w-0 w-full items-center justify-between gap-2 p-3 min-[400px]:gap-4 min-[400px]:p-4 md:p-6",
        className,
      )}
    >
      <div className="flex items-center">
        <Logo />
      </div>
      <div className="flex items-center gap-3">
        <ThemeToggle />
        {showLanguageSelector && (
          <LanguageSelector variant="ghost" size="icon" />
        )}
      </div>
    </header>
  );
};
