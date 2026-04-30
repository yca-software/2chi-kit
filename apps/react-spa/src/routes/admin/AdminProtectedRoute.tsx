import { Navigate, Outlet } from "react-router";
import { useUserState } from "@/states";
import { Loader2 } from "lucide-react";

export function AdminProtectedRoute() {
  const isUserProfileReady = useUserState((state) => state.isUserProfileReady);
  const userData = useUserState((state) => state.userData);

  if (!isUserProfileReady) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!userData.user) {
    return <Navigate to="/" replace />;
  }

  if (!userData.admin) {
    return <Navigate to="/dashboard" replace />;
  }

  return <Outlet />;
}
