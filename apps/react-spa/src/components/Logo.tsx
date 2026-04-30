import { cn } from "@yca-software/design-system";
import { Link } from "react-router";

interface LogoProps {
  className?: string;
  showText?: boolean;
  size?: "sm" | "md" | "lg";
}

export const Logo = ({
  className,
  showText = true,
  size = "md",
}: LogoProps) => {
  const sizeClasses = {
    sm: "h-6 w-6",
    md: "h-8 w-8",
    lg: "h-10 w-10",
  };

  const textSizeClasses = {
    sm: "text-base",
    md: "text-lg",
    lg: "text-xl",
  };

  return (
    <Link to="/" className={cn("flex items-center gap-2", className)}>
      <div
        className={cn(
          "flex items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold",
          sizeClasses[size],
          textSizeClasses[size],
        )}
      >
        EA
      </div>
      {showText && (
        <span className={cn("font-semibold", textSizeClasses[size])}>
          Example App
        </span>
      )}
    </Link>
  );
};
