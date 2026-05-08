import { Navigate } from "react-router-dom";
import { useAuth } from "@/features/auth/AuthProvider";
import { homePathForRole } from "@/features/auth/routing";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

export function RoleRedirect() {
  const { user, isLoading, isAuthenticated } = useAuth();
  const { t } = useLanguage();

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-6">
        <p className="text-sm text-gray-600">{t("common.loadingSession")}</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <Navigate to={homePathForRole(user?.role)} replace />;
}
