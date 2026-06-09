import { Link, Outlet, useLocation, useNavigate } from "react-router-dom";
import { queryClient } from "./queryClient";
import { useAuth } from "@/features/auth/AuthProvider";
import { NotificationBell } from "@/features/notifications/NotificationBell";
import { LanguageSelect, useLanguage } from "@/shared/i18n/LanguageProvider";

export function AppShell() {
  const { user, logout } = useAuth();
  const { t } = useLanguage();
  const navigate = useNavigate();
  const location = useLocation();

  function closeSession() {
    logout();
    queryClient.clear();
    navigate("/login", { replace: true });
  }

  const canSeePatients = user?.role === "admin" || user?.role === "recepcionista";
  const canSeeDashboard = user?.role === "admin" || user?.role === "recepcionista" || user?.role === "kinesiologo";
  const canSeeAgenda = user?.role === "admin" || user?.role === "recepcionista" || user?.role === "kinesiologo";
  const canSeeMaterials = user?.role === "admin" || user?.role === "recepcionista" || user?.role === "kinesiologo";
  const canSeeFinance = user?.role === "admin" || user?.role === "recepcionista";
  const canSeeReports = user?.role === "admin" || user?.role === "recepcionista";
  const canSeeStaff = user?.role === "admin";
  const roleText =
    user?.role === "admin"
      ? t("role.admin")
      : user?.role === "recepcionista"
        ? t("role.receptionist")
        : user?.role === "kinesiologo"
          ? t("role.kinesiologist")
          : user?.role === "paciente"
            ? t("role.patient")
            : t("role.none");

  return (
    <div className="min-h-screen bg-gray-50">
      <aside className="bg-white border-b lg:fixed lg:inset-y-0 lg:left-0 lg:w-64 lg:border-b-0 lg:border-r">
        <div className="px-6 py-4 space-y-5">
          <div className="flex items-start gap-3">
            <div className="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-black text-sm font-semibold text-white shadow-sm">
              KA
            </div>
            <div>
              <p className="text-xs uppercase tracking-[0.18em] text-gray-500">{t("app.name")}</p>
              <p className="mt-1 text-sm text-gray-700">
                {user?.email} · {roleText}
              </p>
            </div>
          </div>

          <nav className="flex flex-wrap gap-2 lg:flex-col">
            {canSeeDashboard && (
              <Link
                className={navClass(location.pathname === "/dashboard")}
                to="/dashboard"
              >
                {t("nav.dashboard")}
              </Link>
            )}
            {canSeeAgenda && (
              <Link
                className={navClass(location.pathname === "/agenda")}
                to="/agenda"
              >
                {t("nav.agenda")}
              </Link>
            )}
            {user?.role === "paciente" && (
              <Link
                className={navClass(location.pathname === "/portal")}
                to="/portal"
              >
                {t("nav.portal")}
              </Link>
            )}
            {canSeePatients && (
              <Link
                className={navClass(location.pathname === "/patients")}
                to="/patients"
              >
                {t("nav.patients")}
              </Link>
            )}
            {canSeeMaterials && (
              <Link
                className={navClass(location.pathname === "/materials")}
                to="/materials"
              >
                {t("nav.materials")}
              </Link>
            )}
            {canSeeFinance && (
              <Link
                className={navClass(location.pathname === "/finance")}
                to="/finance"
              >
                {t("nav.finance")}
              </Link>
            )}
            {canSeeReports && (
              <Link
                className={navClass(location.pathname === "/reports")}
                to="/reports"
              >
                {t("nav.reports")}
              </Link>
            )}
            {canSeeStaff && (
              <Link
                className={navClass(location.pathname === "/staff")}
                to="/staff"
              >
                {t("nav.staff")}
              </Link>
            )}
            <Link
              className={navClass(location.pathname === "/profile")}
              to="/profile"
            >
              {t("nav.myProfile")}
            </Link>
          </nav>

          <div className="space-y-3">
            <NotificationBell />
            <LanguageSelect compact />
            <button
              type="button"
              className="w-full px-3 py-2 rounded-lg border text-sm text-gray-700 hover:bg-gray-100"
              onClick={closeSession}
            >
              {t("nav.signOut")}
            </button>
          </div>
        </div>
      </aside>

      <div className="lg:pl-64">
        <Outlet />
      </div>
    </div>
  );
}

function navClass(active: boolean) {
  return active
    ? "px-3 py-2 rounded-lg bg-black text-white text-sm shadow-sm"
    : "px-3 py-2 rounded-lg border text-sm text-gray-700 hover:bg-gray-100";
}
