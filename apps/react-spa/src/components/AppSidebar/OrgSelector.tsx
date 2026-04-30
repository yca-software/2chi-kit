import { useEffect, useMemo, useState } from "react";
import { useUserState } from "@/states";
import { useShallow } from "zustand/shallow";
import { Building2, ChevronsUpDown, Check, Plus } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  Button,
  cn,
} from "@yca-software/design-system";
import { useNavigate, useLocation } from "react-router";
import {
  resolveDefaultOrganizationId,
  useTranslationNamespace,
} from "@/helpers";
import { Onboarding } from "@/components";

interface OrgSelectorProps {
  variant?: "default" | "compact" | "iconOnly";
}

export const OrgSelector = ({ variant = "default" }: OrgSelectorProps) => {
  const { t, isLoading: tLoading } = useTranslationNamespace([
    "settings",
    "dashboard",
    "common",
  ]);
  const navigate = useNavigate();
  const location = useLocation();
  const [createOrgDialogOpen, setCreateOrgDialogOpen] = useState(false);
  const { selectedOrgId, setSelectedOrgId, userData } = useUserState(
    useShallow((state) => ({
      selectedOrgId: state.selectedOrgId,
      setSelectedOrgId: state.setSelectedOrgId,
      userData: state.userData,
    })),
  );

  // Get organizations from user state (populated by Root.tsx)
  const organizations = userData.roles || [];

  const selectedOrg = useMemo(() => {
    return organizations.find((o) => o.organizationId === selectedOrgId);
  }, [organizations, selectedOrgId]);

  useEffect(() => {
    if (organizations.length > 0) {
      const resolved = resolveDefaultOrganizationId(
        organizations,
        selectedOrgId,
      );
      if (resolved && resolved !== selectedOrgId) {
        setSelectedOrgId(resolved);
      }
    }
  }, [organizations, selectedOrgId, setSelectedOrgId]);

  const handleOrgChange = (orgId: string) => {
    setSelectedOrgId(orgId);
    // Update URL if we're on dashboard or settings
    if (location.pathname.startsWith("/dashboard")) {
      navigate(`/dashboard/${orgId}`, { replace: true });
    } else if (location.pathname.startsWith("/settings")) {
      navigate(`/settings/${orgId}`, { replace: true });
    }
  };

  if (tLoading) {
    return null;
  }

  if (organizations.length === 0) {
    return (
      <>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setCreateOrgDialogOpen(true)}
        >
          <Plus className="mr-2 h-4 w-4" />
          {t("settings:orgSelector.createOrg")}
        </Button>
        <CreateOrgDialog
          open={createOrgDialogOpen}
          onOpenChange={setCreateOrgDialogOpen}
          t={t}
        />
      </>
    );
  }

  const triggerHeight = variant === "compact" ? "h-9" : "h-10";
  const isIconOnly = variant === "iconOnly";

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <button
            type="button"
            className={cn(
              "flex min-w-0 cursor-pointer items-center gap-2 rounded-md border bg-background text-sm font-normal shadow-xs outline-none transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 dark:border-input dark:bg-input/30 dark:hover:bg-input/50",
              triggerHeight,
              isIconOnly ? "h-9 w-9 justify-center p-0" : "pl-1.5 pr-3",
              variant === "compact" && !isIconOnly && "w-auto max-w-[180px]",
              variant === "default" && "w-full max-w-[240px]",
            )}
          >
            <div className="flex shrink-0 items-center justify-center rounded-sm bg-primary/10 p-1.5">
              <Building2 className="h-4 w-4 text-primary" />
            </div>
            {!isIconOnly && (
              <>
                <span
                  className={cn(
                    "min-w-0 flex-1 truncate text-left",
                    selectedOrg?.organizationName
                      ? "font-medium"
                      : "text-muted-foreground",
                  )}
                >
                  {selectedOrg?.organizationName ||
                    t("settings:orgSelector.selectOrg")}
                </span>
                <ChevronsUpDown className="h-4 w-4 shrink-0 text-muted-foreground" />
              </>
            )}
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="w-[240px]">
          {organizations.map((org) => (
            <DropdownMenuItem
              key={org.organizationId}
              onClick={() => handleOrgChange(org.organizationId)}
              className="flex items-center justify-between cursor-pointer"
            >
              <div className="flex items-center gap-2">
                <Building2 className="h-4 w-4 text-muted-foreground" />
                <span>{org.organizationName}</span>
              </div>
              {org.organizationId === selectedOrgId && (
                <Check className="h-4 w-4" />
              )}
            </DropdownMenuItem>
          ))}
          <div className="my-1 h-px bg-border" />
          <DropdownMenuItem
            onClick={() => setCreateOrgDialogOpen(true)}
            className="flex items-center gap-2 cursor-pointer"
          >
            <Plus className="h-4 w-4" />
            {t("settings:orgSelector.createOrg")}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <CreateOrgDialog
        open={createOrgDialogOpen}
        onOpenChange={setCreateOrgDialogOpen}
        t={t}
      />
    </>
  );
};

interface CreateOrgDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  t: (key: string, options?: { ns?: string }) => string;
}

function CreateOrgDialog({ open, onOpenChange, t }: CreateOrgDialogProps) {
  const handleComplete = () => {
    onOpenChange(false);
  };

  return (
    <Onboarding
      open={open}
      onOpenChange={onOpenChange}
      title={t("dashboard:onboarding.createOrgTitle")}
      description={t("dashboard:onboarding.createOrgDescription")}
      navigateToDashboard={false}
      onComplete={handleComplete}
    />
  );
}
