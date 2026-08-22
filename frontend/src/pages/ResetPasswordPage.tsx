import { useState, type FormEvent } from "react";
import { Link, Navigate, useSearchParams } from "react-router-dom";
import { confirmPasswordReset } from "@/features/auth/api";
import { useAuth } from "@/features/auth/AuthProvider";
import { homePathForRole } from "@/features/auth/routing";
import { LanguageSelect, useLanguage } from "@/shared/i18n/LanguageProvider";
import { RequiredLabel } from "@/shared/ui/RequiredLabel";
import { PasswordInput } from "@/shared/ui/PasswordInput";
import { passwordPolicyPattern } from "@/shared/validation/password";

export default function ResetPasswordPage() {
  const { isAuthenticated, user } = useAuth();
  const { t } = useLanguage();
  const [searchParams] = useSearchParams();
  const oobCode = searchParams.get("oobCode") ?? "";
  const [newPassword, setNewPassword] = useState("");
  const [repeatPassword, setRepeatPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDone, setIsDone] = useState(false);
  const [error, setError] = useState("");

  if (isAuthenticated) {
    return <Navigate to={homePathForRole(user?.role)} replace />;
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");

    if (!passwordPolicyPattern.test(newPassword)) {
      setError(t("resetPassword.passwordPolicy"));
      return;
    }
    if (newPassword !== repeatPassword) {
      setError(t("resetPassword.passwordMismatch"));
      return;
    }

    setIsSubmitting(true);
    try {
      await confirmPasswordReset({ oob_code: oobCode, new_password: newPassword });
      setIsDone(true);
    } catch (err) {
      const message = (err as Error)?.message;
      setError(
        message === "validation_error"
          ? t("resetPassword.passwordPolicy")
          : message === "invalid_or_expired_code"
            ? t("resetPassword.invalidCode")
            : t("resetPassword.failed"),
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="min-h-screen bg-gray-50 flex items-center justify-center px-6 py-10">
      <section className="w-full max-w-sm bg-white border rounded-lg p-6 shadow-sm">
        <div className="mb-6">
          <div className="mb-4">
            <LanguageSelect />
          </div>
          <h1 className="text-2xl font-semibold">{t("resetPassword.title")}</h1>
          <p className="text-sm text-gray-600 mt-1">{t("resetPassword.help")}</p>
        </div>

        {!oobCode ? (
          <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700">
            {t("resetPassword.invalidCode")}
          </div>
        ) : isDone ? (
          <div className="border border-green-200 bg-green-50 rounded-lg p-3 text-sm text-green-700">
            {t("resetPassword.done")}
          </div>
        ) : (
          <form className="space-y-4" onSubmit={submit}>
            <div>
              <label className="text-sm font-medium" htmlFor="new-password">
                <RequiredLabel required>{t("resetPassword.newPassword")}</RequiredLabel>
              </label>
              <PasswordInput
                id="new-password"
                className="mt-1 w-full border rounded-lg p-2"
                autoComplete="new-password"
                value={newPassword}
                onChange={(event) => setNewPassword(event.target.value)}
                minLength={8}
                required
              />
            </div>

            <div>
              <label className="text-sm font-medium" htmlFor="repeat-password">
                <RequiredLabel required>{t("resetPassword.repeatPassword")}</RequiredLabel>
              </label>
              <PasswordInput
                id="repeat-password"
                className="mt-1 w-full border rounded-lg p-2"
                autoComplete="new-password"
                value={repeatPassword}
                onChange={(event) => setRepeatPassword(event.target.value)}
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
              {isSubmitting ? t("resetPassword.submitting") : t("resetPassword.submit")}
            </button>
          </form>
        )}

        <div className="mt-4 text-center">
          <Link className="text-sm underline" to="/login">
            {t("auth.recoveryBack")}
          </Link>
        </div>
      </section>
    </main>
  );
}
