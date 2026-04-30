import { SiteHeader } from "@/components/SiteHeader";
import { Outlet } from "react-router";

export const LegalLayout = () => {
  return (
    <div className="min-h-screen bg-background">
      <SiteHeader />
      <main className="mx-auto min-w-0 max-w-4xl px-3 pb-12 pt-6 min-[400px]:px-4 sm:px-6 lg:px-8">
        <Outlet />
      </main>
    </div>
  );
};
