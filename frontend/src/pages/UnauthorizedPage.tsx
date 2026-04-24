import { Navigate, useNavigate } from "react-router-dom";
import { queryClient } from "@/app/queryClient";
import { useAuth } from "@/features/auth/AuthProvider";

export default function UnauthorizedPage() {
  const { user, isLoading, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();

  if (isLoading) {
    return (
      <main className="min-h-screen bg-gray-50 flex items-center justify-center p-6">
        <p className="text-sm text-gray-600">Validando sesión...</p>
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

  return (
    <main className="min-h-screen bg-gray-50 flex items-center justify-center px-6 py-10">
      <section className="w-full max-w-md bg-white border rounded-lg p-6 shadow-sm space-y-4">
        <div>
          <h1 className="text-2xl font-semibold">Acceso no configurado</h1>
          <p className="text-sm text-gray-600 mt-2">
            Tu usuario inició sesión, pero no tiene un rol asignado para usar la app.
          </p>
        </div>

        <div className="rounded-lg border bg-gray-50 p-3 text-sm text-gray-700">
          <div>
            Email: <span className="font-mono">{user?.email}</span>
          </div>
          <div>
            Rol: <span className="font-mono">{user?.role || "sin rol"}</span>
          </div>
        </div>

        <button
          type="button"
          className="w-full px-4 py-2 rounded-lg bg-black text-white"
          onClick={closeSession}
        >
          Cerrar sesión
        </button>
      </section>
    </main>
  );
}
