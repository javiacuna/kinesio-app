import { useEffect, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { inviteUserAccess } from "@/features/auth/adminApi";
import { useAuth } from "@/features/auth/AuthProvider";
import { archivePatient, createPatient, listPatients, searchPatients, updatePatient } from "@/features/patients/api";
import type { Patient } from "@/features/patients/types";
import { uploadPatientAttachment } from "@/features/patients/detailApi";
import { listFinanciers, type Financier } from "@/features/finance/api";
import { RequiredLabel } from "@/shared/ui/RequiredLabel";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export default function PatientsPage() {
  const { user } = useAuth();
  const { t } = useLanguage();
  const [dni, setDni] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [financierId, setFinancierId] = useState("");
  const [financierMemberNumber, setFinancierMemberNumber] = useState("");
  const [financiers, setFinanciers] = useState<Financier[]>([]);
  const [initialFiles, setInitialFiles] = useState<File[]>([]);

  const [created, setCreated] = useState<Patient | null>(null);
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});
  const [patients, setPatients] = useState<Patient[]>([]);
  const [search, setSearch] = useState("");
  const [includeInactive, setIncludeInactive] = useState(false);
  const [error, setError] = useState<string>("");
  const [inviteMessage, setInviteMessage] = useState("");
  const [isInviting, setIsInviting] = useState(false);
  const [invitingPatientId, setInvitingPatientId] = useState<string | null>(null);
  const [togglingPatientId, setTogglingPatientId] = useState<string | null>(null);
  const [isLoadingPatients, setIsLoadingPatients] = useState(false);
  const [isCreatingPatient, setIsCreatingPatient] = useState(false);

  const selectedFinancier = financiers.find((f) => f.id === financierId);
  const selectedFinancierIsParticular = !selectedFinancier || selectedFinancier.kind === "particular";

  useEffect(() => {
    refreshPatients();
    listFinanciers({ includeInactive: false })
      .then(setFinanciers)
      .catch(() => setFinanciers([]));
  }, []);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      refreshPatients(search);
    }, 250);

    return () => window.clearTimeout(timer);
  }, [search, includeInactive]);

  async function refreshPatients(query = search) {
    setIsLoadingPatients(true);
    setError("");
    try {
      const normalizedQuery = query.trim();
      const res = normalizedQuery
        ? await searchPatients(normalizedQuery, 50, includeInactive)
        : await listPatients(50, includeInactive);
      setPatients(res);
    } catch (e: any) {
      setError(e?.message ?? t("patients.errorLoadFailed"));
    } finally {
      setIsLoadingPatients(false);
    }
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (isCreatingPatient) return;
    setError("");
    setInviteMessage("");
    setCreated(null);
    const validation = validatePatientForm();
    setFormErrors(validation);
    if (Object.keys(validation).length > 0) return;

    setIsCreatingPatient(true);
    try {
      const res = await createPatient({
        dni,
        first_name: firstName,
        last_name: lastName,
        email,
        financier_id: financierId || null,
        financier_member_number: selectedFinancierIsParticular ? null : financierMemberNumber.trim() || null,
      });

      setError("");
      setCreated(res);
      localStorage.setItem("last_patient_id", res.id);
      setDni("");
      setFirstName("");
      setLastName("");
      setEmail("");
      setFinancierId("");
      setFinancierMemberNumber("");
      if (initialFiles.length > 0) {
        await Promise.all(
          initialFiles.map((file) =>
            uploadPatientAttachment(res.id, file, {
              category: file.type.startsWith("image/") ? "foto" : file.type.startsWith("video/") ? "video" : "documento",
              patient_visible: false,
            }),
          ),
        );
        setInitialFiles([]);
      }
      await refreshPatients("");
    } catch (e: any) {
      setCreated(null);
      setError(e?.message ?? t("patients.error"));
    } finally {
      setIsCreatingPatient(false);
    }
  }

  function validatePatientForm() {
    const validation: Record<string, string> = {};
    if (!dni.trim()) validation.dni = t("patients.errorDni");
    if (!firstName.trim()) validation.firstName = t("patients.errorFirstName");
    if (!lastName.trim()) validation.lastName = t("patients.errorLastName");
    if (!email.trim()) {
      validation.email = t("patients.errorEmailRequired");
    } else if (!emailPattern.test(email.trim())) {
      validation.email = t("patients.errorEmailInvalid");
    }
    return validation;
  }

  async function invitePatient() {
    if (!created) return;

    setError("");
    setInviteMessage("");
    setIsInviting(true);
    try {
      await inviteUserAccess({ email: created.email, role: "paciente" });
      setInviteMessage(`${t("patients.inviteSentToPrefix")} ${created.email}.`);
    } catch (e: any) {
      setError(e?.message ?? t("patients.errorInviteFailed"));
    } finally {
      setIsInviting(false);
    }
  }

  async function invitePatientFromList(patient: Patient) {
    setError("");
    setInviteMessage("");
    setInvitingPatientId(patient.id);
    try {
      await inviteUserAccess({ email: patient.email, role: "paciente" });
      setInviteMessage(`${t("patients.inviteSentToPrefix")} ${patient.email}.`);
    } catch (e: any) {
      setError(e?.message ?? t("patients.errorInviteFailed"));
    } finally {
      setInvitingPatientId(null);
    }
  }

  function selectForAgenda(patient: Patient) {
    if (!patient.active) {
      setError(t("patients.errorArchivedInAgenda"));
      return;
    }
    localStorage.setItem("last_patient_id", patient.id);
    setCreated(patient);
    setInviteMessage(`${patient.last_name}, ${patient.first_name} ${t("patients.selectedForAgendaSuffix")}`);
  }

  async function togglePatientActive(patient: Patient) {
    setError("");
    setInviteMessage("");
    setTogglingPatientId(patient.id);

    try {
      if (patient.active) {
        await archivePatient(patient.id);
        setInviteMessage(`${patient.last_name}, ${patient.first_name} ${t("patients.archivedNoticeSuffix")}`);
      } else {
        const updated = await updatePatient({
          id: patient.id,
          dni: patient.dni,
          first_name: patient.first_name,
          last_name: patient.last_name,
          email: patient.email,
          phone: patient.phone ?? null,
          active: true,
        });
        setInviteMessage(`${updated.last_name}, ${updated.first_name} ${t("patients.reactivatedNoticeSuffix")}`);
      }
      await refreshPatients();
    } catch (e: any) {
      setError(e?.message ?? t("patients.errorToggleActiveFailed"));
    } finally {
      setTogglingPatientId(null);
    }
  }

  return (
    <main>
      <div className="max-w-5xl mx-auto p-6 space-y-6">
        <header className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold">{t("patients.title")}</h1>
            <p className="text-sm text-gray-600">{t("patients.subtitle")}</p>
          </div>
          <Link className="text-sm underline" to="/agenda">{t("patients.goToAgenda")}</Link>
        </header>

        <form className="bg-white rounded-xl shadow p-4 space-y-3" onSubmit={submit} noValidate>
          <p className="text-xs text-gray-500">{t("patients.requiredFieldsNote")}</p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label className="text-sm font-medium"><RequiredLabel required>{t("patients.dni")}</RequiredLabel></label>
              <input className="mt-1 w-full border rounded-lg p-2" value={dni} onChange={(e) => setDni(e.target.value)} required />
              {formErrors.dni && <p className="text-xs text-red-600 mt-1">{formErrors.dni}</p>}
            </div>
            <div>
              <label className="text-sm font-medium"><RequiredLabel required>{t("patients.email")}</RequiredLabel></label>
              <input className="mt-1 w-full border rounded-lg p-2" type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
              {formErrors.email && <p className="text-xs text-red-600 mt-1">{formErrors.email}</p>}
            </div>
            <div>
              <label className="text-sm font-medium"><RequiredLabel required>{t("patients.firstName")}</RequiredLabel></label>
              <input className="mt-1 w-full border rounded-lg p-2" value={firstName} onChange={(e) => setFirstName(e.target.value)} required />
              {formErrors.firstName && <p className="text-xs text-red-600 mt-1">{formErrors.firstName}</p>}
            </div>
            <div>
              <label className="text-sm font-medium"><RequiredLabel required>{t("patients.lastName")}</RequiredLabel></label>
              <input className="mt-1 w-full border rounded-lg p-2" value={lastName} onChange={(e) => setLastName(e.target.value)} required />
              {formErrors.lastName && <p className="text-xs text-red-600 mt-1">{formErrors.lastName}</p>}
            </div>
            <div>
              <label className="text-sm font-medium">{t("patients.financier")}</label>
              <select
                className="mt-1 w-full border rounded-lg p-2"
                value={financierId}
                onChange={(e) => {
                  const id = e.target.value;
                  setFinancierId(id);
                  const next = financiers.find((f) => f.id === id);
                  if (!next || next.kind === "particular") setFinancierMemberNumber("");
                }}
              >
                <option value="">{t("patients.noFinancier")}</option>
                {financiers.map((financier) => (
                  <option key={financier.id} value={financier.id}>
                    {financier.name}
                  </option>
                ))}
              </select>
            </div>
            {!selectedFinancierIsParticular && (
              <div>
                <label className="text-sm font-medium">{t("patients.memberNumber")}</label>
                <input
                  className="mt-1 w-full border rounded-lg p-2"
                  value={financierMemberNumber}
                  onChange={(e) => setFinancierMemberNumber(e.target.value)}
                />
              </div>
            )}
            <div className="md:col-span-2">
              <label className="text-sm font-medium">{t("patients.initialFiles")}</label>
              <input
                className="mt-1 w-full border rounded-lg p-2"
                type="file"
                multiple
                accept="image/*,video/*,application/pdf"
                onChange={(event) => setInitialFiles(Array.from(event.target.files ?? []))}
              />
              {initialFiles.length > 0 && (
                <p className="text-xs text-gray-500 mt-1">
                  {initialFiles.length} {t("patients.filesWillUpload")}
                </p>
              )}
            </div>
          </div>

          <button className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50" type="submit" disabled={isCreatingPatient}>
            {isCreatingPatient ? t("common.saving") : t("patients.createPatient")}
          </button>

          {error && <p className="text-sm text-red-600">{t("patients.error")}: {error}</p>}

          {created && (
            <div className="mt-2 border rounded-lg p-3 bg-gray-50">
              <div className="font-medium">{t("patients.patientCreated")}</div>
              <div className="text-sm text-gray-700">ID: <span className="font-mono">{created.id}</span></div>
              <div className="text-sm text-gray-700">{created.last_name}, {created.first_name} — {created.dni}</div>
              <p className="text-xs text-gray-500 mt-1">{t("patients.savedAsLastPatient")}</p>
              {user?.role === "admin" && (
                <button
                  type="button"
                  className="mt-3 px-3 py-2 rounded-lg border text-sm hover:bg-white disabled:opacity-50"
                  disabled={isInviting}
                  onClick={invitePatient}
                >
                  {isInviting ? t("patients.sending") : t("patients.sendPortalAccess")}
                </button>
              )}
              {inviteMessage && <p className="text-sm text-green-700 mt-2">{inviteMessage}</p>}
            </div>
          )}
        </form>

        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
            <div className="flex-1">
              <h2 className="text-lg font-semibold">{t("patients.listTitle")}</h2>
              <label className="text-sm font-medium mt-3 block">{t("patients.search")}</label>
              <input
                className="mt-1 w-full border rounded-lg p-2"
                placeholder={t("patients.searchPlaceholder")}
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
              <label className="flex items-center gap-2 text-sm mt-3">
                <input
                  type="checkbox"
                  checked={includeInactive}
                  onChange={(e) => setIncludeInactive(e.target.checked)}
                />
                {t("patients.showArchived")}
              </label>
            </div>
            <button
              type="button"
              className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50"
              onClick={() => refreshPatients()}
            >
              {t("patients.refresh")}
            </button>
          </div>

          {isLoadingPatients && <p className="text-sm text-gray-600">{t("patients.loading")}</p>}

          <div className="divide-y">
            {!isLoadingPatients && patients.length === 0 ? (
              <p className="text-sm text-gray-600 py-3">{t("patients.empty")}</p>
            ) : (
              patients.map((patient) => (
                <div key={patient.id} className="py-3 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                  <div>
                    <div className="font-medium">{patient.last_name}, {patient.first_name}</div>
                    <div className="text-sm text-gray-600">{t("patients.dni")}: {patient.dni}</div>
                    <div className="text-sm text-gray-600">{patient.email}</div>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <span className={`text-xs px-2 py-1 rounded-md self-center ${patient.active ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-700"}`}>
                      {patient.active ? t("patients.active") : t("patients.archived")}
                    </span>
                    <button
                      type="button"
                      className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                      disabled={!patient.active}
                      onClick={() => selectForAgenda(patient)}
                    >
                      {t("patients.useInAgenda")}
                    </button>
                    <Link
                      className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                      to={`/patients/${patient.id}`}
                    >
                      {t("patients.viewDetail")}
                    </Link>
                    {user?.role === "admin" && (
                      <button
                        type="button"
                        className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50 disabled:opacity-50"
                        disabled={invitingPatientId === patient.id}
                        onClick={() => invitePatientFromList(patient)}
                      >
                        {invitingPatientId === patient.id ? t("patients.sending") : t("patients.sendAccess")}
                      </button>
                    )}
                    <button
                      type="button"
                      className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50 disabled:opacity-50"
                      disabled={togglingPatientId === patient.id}
                      onClick={() => togglePatientActive(patient)}
                    >
                      {togglingPatientId === patient.id
                        ? t("patients.saving")
                        : patient.active
                          ? t("patients.archive")
                          : t("patients.reactivate")}
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </section>
      </div>
    </main>
  );
}
