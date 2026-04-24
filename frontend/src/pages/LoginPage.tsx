import { useState, type FormEvent } from "react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "@/features/auth/AuthProvider";
import { homePathForRole } from "@/features/auth/routing";

export default function LoginPage() {
  const { loginWithEmail, isAuthenticated, user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (isAuthenticated) {
    return <Navigate to={homePathForRole(user?.role)} replace />;
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setIsSubmitting(true);

    try {
      const me = await loginWithEmail({ email, password });
      const from = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname;
      navigate(from && from !== "/login" ? from : homePathForRole(me.role), { replace: true });
    } catch (err) {
      const message = (err as Error)?.message;
      setError(message === "invalid_credentials" ? "Email o contraseña incorrectos." : "No se pudo iniciar sesión.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="min-h-screen bg-gray-50 flex items-center justify-center px-6 py-10">
      <section className="w-full max-w-sm bg-white border rounded-lg p-6 shadow-sm">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold">Ingresar</h1>
          <p className="text-sm text-gray-600 mt-1">Ingresá con tu email y contraseña.</p>
        </div>

        <form className="space-y-4" onSubmit={submit}>
          <div>
            <label className="text-sm font-medium" htmlFor="email">
              Email
            </label>
            <input
              id="email"
              className="mt-1 w-full border rounded-lg p-2"
              type="email"
              autoComplete="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              required
            />
          </div>

          <div>
            <label className="text-sm font-medium" htmlFor="password">
              Contraseña
            </label>
            <input
              id="password"
              className="mt-1 w-full border rounded-lg p-2"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              required
            />
          </div>

          {error && (
            <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700">
              {error}
            </div>
          )}

          <button
            className="w-full px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
            disabled={isSubmitting}
            type="submit"
          >
            {isSubmitting ? "Ingresando..." : "Ingresar"}
          </button>
        </form>
      </section>
    </main>
  );
}
