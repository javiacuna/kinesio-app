import { useState, type FormEvent } from "react";
import { changePassword } from "@/features/auth/api";
import { useAuth } from "@/features/auth/AuthProvider";
import { saveAuthTokens } from "@/features/auth/session";

function roleLabel(role?: string) {
  switch (role) {
    case "admin":
      return "Administrador";
    case "recepcionista":
      return "Recepción";
    case "kinesiologo":
      return "Kinesiólogo";
    case "paciente":
      return "Paciente";
    default:
      return "Sin rol";
  }
}

export default function ProfilePage() {
  const { user } = useAuth();
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [repeatPassword, setRepeatPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [success, setSuccess] = useState("");
  const [error, setError] = useState("");

  async function submitPassword(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSuccess("");
    setError("");

    if (newPassword.length < 6) {
      setError("La nueva contraseña debe tener al menos 6 caracteres.");
      return;
    }
    if (newPassword !== repeatPassword) {
      setError("Las contraseñas nuevas no coinciden.");
      return;
    }

    setIsSubmitting(true);
    try {
      const res = await changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      });
      saveAuthTokens(res.id_token, res.refresh_token);
      setCurrentPassword("");
      setNewPassword("");
      setRepeatPassword("");
      setSuccess("Contraseña actualizada.");
    } catch (err) {
      const message = (err as Error)?.message;
      setError(
        message === "invalid_current_password"
          ? "La contraseña actual no es correcta."
          : "No se pudo cambiar la contraseña.",
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="max-w-3xl mx-auto p-6 space-y-6">
      <header>
        <h1 className="text-2xl font-semibold">Mi perfil</h1>
        <p className="text-sm text-gray-600">Datos de sesión y seguridad.</p>
      </header>

      <section className="bg-white rounded-xl shadow p-4 space-y-3">
        <h2 className="text-lg font-semibold">Cuenta</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div>
            <div className="text-sm text-gray-600">Email</div>
            <div className="font-medium break-all">{user?.email}</div>
          </div>
          <div>
            <div className="text-sm text-gray-600">Rol</div>
            <div className="font-medium">{roleLabel(user?.role)}</div>
          </div>
        </div>
      </section>

      <section className="bg-white rounded-xl shadow p-4 space-y-4">
        <h2 className="text-lg font-semibold">Cambiar contraseña</h2>

        <form className="space-y-4" onSubmit={submitPassword}>
          <div>
            <label className="text-sm font-medium" htmlFor="current-password">
              Contraseña actual
            </label>
            <input
              id="current-password"
              className="mt-1 w-full border rounded-lg p-2"
              type="password"
              autoComplete="current-password"
              value={currentPassword}
              onChange={(event) => setCurrentPassword(event.target.value)}
              required
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label className="text-sm font-medium" htmlFor="new-password">
                Nueva contraseña
              </label>
              <input
                id="new-password"
                className="mt-1 w-full border rounded-lg p-2"
                type="password"
                autoComplete="new-password"
                value={newPassword}
                onChange={(event) => setNewPassword(event.target.value)}
                required
              />
            </div>
            <div>
              <label className="text-sm font-medium" htmlFor="repeat-password">
                Repetir nueva contraseña
              </label>
              <input
                id="repeat-password"
                className="mt-1 w-full border rounded-lg p-2"
                type="password"
                autoComplete="new-password"
                value={repeatPassword}
                onChange={(event) => setRepeatPassword(event.target.value)}
                required
              />
            </div>
          </div>

          {success && (
            <div className="border border-green-200 bg-green-50 rounded-lg p-3 text-sm text-green-700">
              {success}
            </div>
          )}

          {error && (
            <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700">
              {error}
            </div>
          )}

          <button
            type="submit"
            className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
            disabled={isSubmitting}
          >
            {isSubmitting ? "Guardando..." : "Guardar contraseña"}
          </button>
        </form>
      </section>
    </main>
  );
}
