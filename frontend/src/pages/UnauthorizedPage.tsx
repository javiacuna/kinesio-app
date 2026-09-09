import { useEffect, useState } from "react";
import { Navigate, useNavigate } from "react-router-dom";
import { queryClient } from "@/app/queryClient";
import { useAuth } from "@/features/auth/AuthProvider";
import { getMySignupStatus } from "@/features/patientSignups/api";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

export default function UnauthorizedPage() {
  const { user, isLoading, isAuthenticated, logout } = useAuth();
  const { t } = useLanguage();
  const navigate = useNavigate();
  const [signupStatus, setSignupStatus] = useState<"loading" | "none" | "pending" | "approved" | "rejected">(
    "loading",
  );

  useEffect(() => {
    if (!isAuthenticated || user?.role) return;
    let active = true;
    getMySignupStatus()
      .then((res) => {
        if (active) setSignupStatus(res.status);
      })
      .catch(() => {
        if (active) setSignupStatus("none");
      });
    return () => {
      active = false;
    };
  }, [isAuthenticated, user?.role]);

  if (isLoading) {
    return (
      <main className="min-h-screen bg-gray-50 flex items-center justify-center p-6">
        <p className="text-sm text-gray-600">{t("common.loadingSession")}</p>
      </main>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  function closeSession() {
    logout();
    queryClient.clear();
    navigate("/login", { replace: true });
  }

  const isPendingSignup = !user?.role && signupStatus === "pending";

  return (
    <main className="min-h-screen bg-gray-50 flex items-center justify-center px-6 py-10">
      <section className="w-full max-w-md bg-white border rounded-lg p-6 shadow-sm space-y-4">
        <div>
          <h1 className="text-2xl font-semibold">
            {isPendingSignup ? t("unauthorized.pendingTitle") : t("unauthorized.title")}
          </h1>
          <p className="text-sm text-gray-600 mt-2">
            {isPendingSignup ? t("unauthorized.pendingSubtitle") : t("unauthorized.subtitle")}
          </p>
        </div>

        <div className="rounded-lg border bg-gray-50 p-3 text-sm text-gray-700">
          <div>
            Email: <span className="font-mono">{user?.email}</span>
          </div>
          <div>
            {t("unauthorized.role")}: <span className="font-mono">{user?.role || t("role.none")}</span>
          </div>
        </div>

        <button
          type="button"
          className="w-full px-4 py-2 rounded-lg bg-black text-white"
          onClick={closeSession}
        >
          {t("nav.signOut")}
        </button>
      </section>
    </main>
  );
}
