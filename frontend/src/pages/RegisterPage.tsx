import { useState, type FormEvent } from "react";
import { Link, Navigate } from "react-router-dom";
import { registerPatientAccount } from "@/features/patientSignups/api";
import { useAuth } from "@/features/auth/AuthProvider";
import { homePathForRole } from "@/features/auth/routing";
import { LanguageSelect, useLanguage } from "@/shared/i18n/LanguageProvider";
import { RequiredLabel } from "@/shared/ui/RequiredLabel";

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export default function RegisterPage() {
  const { isAuthenticated, user } = useAuth();
  const { t } = useLanguage();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [dni, setDni] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [error, setError] = useState("");
  const [successMessage, setSuccessMessage] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (isAuthenticated) {
    return <Navigate to={homePathForRole(user?.role)} replace />;
  }

  function validate() {
    const errors: Record<string, string> = {};
    if (!email || !emailPattern.test(email)) {
      errors.email = t("patients.errorEmailInvalid");
    }
    if (password.length < 6) {
      errors.password = t("auth.registerFailed");
    }
    if (!dni || !/^\d+$/.test(dni)) {
      errors.dni = t("patients.errorDni");
    }
    if (!firstName.trim()) {
      errors.first_name = t("patients.errorFirstName");
    }
    if (!lastName.trim()) {
      errors.last_name = t("patients.errorLastName");
    }
    return errors;
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setSuccessMessage("");

    const errors = validate();
    setFieldErrors(errors);
    if (Object.keys(errors).length > 0) {
      return;
    }

    setIsSubmitting(true);
    try {
      const result = await registerPatientAccount({
        email,
        password,
        dni,
        first_name: firstName,
        last_name: lastName,
      });
      setSuccessMessage(
        result.status === "approved" ? t("auth.registerSuccessApproved") : t("auth.registerSuccessPending"),
      );
    } catch (err) {
      const message = (err as Error)?.message;
      if (message === "email_already_registered") {
        setError(t("auth.emailAlreadyRegistered"));
      } else if (message === "validation_error") {
        const details = (err as { body?: { details?: Record<string, string> } })?.body?.details;
        if (details) setFieldErrors(details);
        setError(t("auth.registerFailed"));
      } else {
        setError(t("auth.registerFailed"));
      }
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
          <h1 className="text-2xl font-semibold">{t("auth.register")}</h1>
          <p className="text-sm text-gray-600 mt-1">{t("auth.registerHelp")}</p>
        </div>

        {successMessage ? (
          <div className="border border-green-200 bg-green-50 rounded-lg p-3 text-sm text-green-700">
            {successMessage}
          </div>
        ) : (
          <form className="space-y-4" onSubmit={submit} noValidate>
            <div>
              <label className="text-sm font-medium" htmlFor="email">
                <RequiredLabel required>{t("auth.email")}</RequiredLabel>
              </label>
              <input
                id="email"
                className="mt-1 w-full border rounded-lg p-2"
                type="email"
                autoComplete="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
              />
              {fieldErrors.email && <p className="text-xs text-red-600 mt-1">{fieldErrors.email}</p>}
            </div>

            <div>
              <label className="text-sm font-medium" htmlFor="password">
                <RequiredLabel required>{t("auth.password")}</RequiredLabel>
              </label>
              <input
                id="password"
                className="mt-1 w-full border rounded-lg p-2"
                type="password"
                autoComplete="new-password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
              />
              {fieldErrors.password && <p className="text-xs text-red-600 mt-1">{fieldErrors.password}</p>}
            </div>

            <div>
              <label className="text-sm font-medium" htmlFor="dni">
                <RequiredLabel required>{t("auth.dni")}</RequiredLabel>
              </label>
              <input
                id="dni"
                className="mt-1 w-full border rounded-lg p-2"
                type="text"
                inputMode="numeric"
                value={dni}
                onChange={(event) => setDni(event.target.value)}
              />
              {fieldErrors.dni && <p className="text-xs text-red-600 mt-1">{fieldErrors.dni}</p>}
            </div>

            <div>
              <label className="text-sm font-medium" htmlFor="first_name">
                <RequiredLabel required>{t("auth.firstName")}</RequiredLabel>
              </label>
              <input
                id="first_name"
                className="mt-1 w-full border rounded-lg p-2"
                type="text"
                value={firstName}
                onChange={(event) => setFirstName(event.target.value)}
              />
              {fieldErrors.first_name && <p className="text-xs text-red-600 mt-1">{fieldErrors.first_name}</p>}
            </div>

            <div>
              <label className="text-sm font-medium" htmlFor="last_name">
                <RequiredLabel required>{t("auth.lastName")}</RequiredLabel>
              </label>
              <input
                id="last_name"
                className="mt-1 w-full border rounded-lg p-2"
                type="text"
                value={lastName}
                onChange={(event) => setLastName(event.target.value)}
              />
              {fieldErrors.last_name && <p className="text-xs text-red-600 mt-1">{fieldErrors.last_name}</p>}
            </div>

            {error && (
              <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700">{error}</div>
            )}

            <button
              className="w-full px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
              disabled={isSubmitting}
              type="submit"
            >
              {isSubmitting ? t("auth.registerSubmitting") : t("auth.registerSubmit")}
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
