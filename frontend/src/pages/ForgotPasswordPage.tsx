import { useState, type FormEvent } from "react";
import { Link, Navigate } from "react-router-dom";
import { requestPasswordReset } from "@/features/auth/api";
import { useAuth } from "@/features/auth/AuthProvider";
import { homePathForRole } from "@/features/auth/routing";

export default function ForgotPasswordPage() {
  const { isAuthenticated, user } = useAuth();
  const [email, setEmail] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSent, setIsSent] = useState(false);
  const [error, setError] = useState("");

  if (isAuthenticated) {
    return <Navigate to={homePathForRole(user?.role)} replace />;
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setIsSent(false);
    setIsSubmitting(true);

    try {
      await requestPasswordReset({ email });
      setIsSent(true);
    } catch {
      setError("No se pudo iniciar la recuperación.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="min-h-screen bg-gray-50 flex items-center justify-center px-6 py-10">
      <section className="w-full max-w-sm bg-white border rounded-lg p-6 shadow-sm">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold">Recuperar contraseña</h1>
          <p className="text-sm text-gray-600 mt-1">Ingresá tu email y te enviamos instrucciones.</p>
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

          {isSent && (
            <div className="border border-green-200 bg-green-50 rounded-lg p-3 text-sm text-green-700">
              Si existe una cuenta con ese email, enviamos las instrucciones.
            </div>
          )}

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
            {isSubmitting ? "Enviando..." : "Enviar instrucciones"}
          </button>
        </form>

        <div className="mt-4 text-center">
          <Link className="text-sm underline" to="/login">
            Volver al login
          </Link>
        </div>
      </section>
    </main>
  );
}
