import { useEffect, useState } from "react";
import {
  approvePatientSignup,
  listPendingPatientSignups,
  rejectPatientSignup,
} from "@/features/patientSignups/api";
import type { PendingPatientSignup } from "@/features/patientSignups/types";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

export default function PatientSignupsPage() {
  const { t } = useLanguage();
  const [items, setItems] = useState<PendingPatientSignup[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [approvingId, setApprovingId] = useState<string | null>(null);
  const [rejectingId, setRejectingId] = useState<string | null>(null);
  const [rejectReasons, setRejectReasons] = useState<Record<string, string>>({});

  useEffect(() => {
    load();
  }, []);

  async function load() {
    setIsLoading(true);
    setError("");
    try {
      const data = await listPendingPatientSignups("pending");
      setItems(data);
    } catch {
      setError(t("patientSignups.errorLoadFailed"));
    } finally {
      setIsLoading(false);
    }
  }

  async function approve(id: string) {
    setMessage("");
    setError("");
    setApprovingId(id);
    try {
      await approvePatientSignup(id);
      setItems((prev) => prev.filter((item) => item.id !== id));
      setMessage(t("patientSignups.approveSuccess"));
    } catch {
      setError(t("patientSignups.errorActionFailed"));
    } finally {
      setApprovingId(null);
    }
  }

  async function reject(id: string) {
    setMessage("");
    setError("");
    setRejectingId(id);
    try {
      await rejectPatientSignup(id, rejectReasons[id]);
      setItems((prev) => prev.filter((item) => item.id !== id));
      setMessage(t("patientSignups.rejectSuccess"));
    } catch {
      setError(t("patientSignups.errorActionFailed"));
    } finally {
      setRejectingId(null);
    }
  }

  return (
    <main>
      <div className="max-w-5xl mx-auto p-6 space-y-6">
        <header>
          <h1 className="text-2xl font-semibold">{t("patientSignups.title")}</h1>
          <p className="text-sm text-gray-600">{t("patientSignups.subtitle")}</p>
        </header>

        {(message || error) && (
          <div
            className={`rounded-lg border p-3 text-sm ${
              error ? "border-red-200 bg-red-50 text-red-700" : "border-green-200 bg-green-50 text-green-700"
            }`}
          >
            {error || message}
          </div>
        )}

        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          {isLoading ? (
            <p className="text-sm text-gray-600 py-3">{t("patientSignups.loading")}</p>
          ) : (
            <div className="divide-y">
              {items.length === 0 ? (
                <p className="text-sm text-gray-600 py-3">{t("patientSignups.empty")}</p>
              ) : (
                items.map((item) => (
                  <div key={item.id} className="py-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <div>
                      <div className="font-medium">
                        {item.first_name} {item.last_name}
                      </div>
                      <div className="text-sm text-gray-600">{item.email}</div>
                      <div className="text-xs text-gray-500">DNI: {item.dni}</div>
                    </div>
                    <div className="flex flex-col gap-2 sm:items-end">
                      <div className="flex items-center gap-2">
                        <span className="text-xs px-2 py-1 rounded-md bg-amber-100 text-amber-700">
                          {t("patientSignups.pending")}
                        </span>
                        <button
                          type="button"
                          className="px-3 py-1 rounded-lg border text-sm disabled:opacity-50"
                          disabled={approvingId === item.id || rejectingId === item.id}
                          onClick={() => approve(item.id)}
                        >
                          {approvingId === item.id ? t("patientSignups.approving") : t("patientSignups.approve")}
                        </button>
                        <button
                          type="button"
                          className="px-3 py-1 rounded-lg border text-sm disabled:opacity-50"
                          disabled={approvingId === item.id || rejectingId === item.id}
                          onClick={() => reject(item.id)}
                        >
                          {rejectingId === item.id ? t("patientSignups.rejecting") : t("patientSignups.reject")}
                        </button>
                      </div>
                      <input
                        className="border rounded-lg p-1 text-xs w-56"
                        placeholder={t("patientSignups.rejectReasonPlaceholder")}
                        value={rejectReasons[item.id] || ""}
                        onChange={(event) =>
                          setRejectReasons((prev) => ({ ...prev, [item.id]: event.target.value }))
                        }
                      />
                    </div>
                  </div>
                ))
              )}
            </div>
          )}
        </section>
      </div>
    </main>
  );
}
