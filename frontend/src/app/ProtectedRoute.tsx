import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "@/features/auth/AuthProvider";
import type { AuthRole } from "@/features/auth/types";
import { homePathForRole } from "@/features/auth/routing";

type ProtectedRouteProps = {
  roles?: AuthRole[];
};

export function ProtectedRoute({ roles }: ProtectedRouteProps) {
  const { user, isLoading, isAuthenticated } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-6">
        <p className="text-sm text-gray-600">Validando sesión...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  if (roles?.length && user?.role !== "admin" && !roles.includes(user?.role as AuthRole)) {
    const fallbackPath = homePathForRole(user?.role);
    return <Navigate to={fallbackPath === location.pathname ? "/unauthorized" : fallbackPath} replace />;
  }

  return <Outlet />;
}
