import { Link, Outlet, useLocation, useNavigate } from "react-router-dom";
import { queryClient } from "./queryClient";
import { useAuth } from "@/features/auth/AuthProvider";

export function AppShell() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  function closeSession() {
    logout();
    queryClient.clear();
    navigate("/login", { replace: true });
  }

  const canSeePatients = user?.role === "admin" || user?.role === "recepcionista";
  const canSeeAgenda = user?.role === "admin" || user?.role === "recepcionista" || user?.role === "kinesiologo";
  const canSeeStaff = user?.role === "admin";

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="border-b bg-white">
        <div className="max-w-5xl mx-auto px-6 py-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="text-xs uppercase text-gray-500">Kinesio App</p>
            <p className="text-sm text-gray-700">
              {user?.email} · {user?.role || "sin rol"}
            </p>
          </div>

          <nav className="flex flex-wrap items-center gap-2">
            {canSeeAgenda && (
              <Link
                className={navClass(location.pathname === "/agenda")}
                to="/agenda"
              >
                Agenda
              </Link>
            )}
            {user?.role === "paciente" && (
              <Link
                className={navClass(location.pathname === "/portal")}
                to="/portal"
              >
                Portal
              </Link>
            )}
            {canSeePatients && (
              <Link
                className={navClass(location.pathname === "/patients")}
                to="/patients"
              >
                Pacientes
              </Link>
            )}
            {canSeeStaff && (
              <Link
                className={navClass(location.pathname === "/staff")}
                to="/staff"
              >
                Equipo
              </Link>
            )}
            <button
              type="button"
              className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-100"
              onClick={closeSession}
            >
              Cerrar sesión
            </button>
          </nav>
        </div>
      </header>

      <Outlet />
    </div>
  );
}

function navClass(active: boolean) {
  return active
    ? "px-3 py-2 rounded-lg bg-black text-white text-sm"
    : "px-3 py-2 rounded-lg border text-sm hover:bg-gray-100";
}
