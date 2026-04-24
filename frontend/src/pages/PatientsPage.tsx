import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { inviteUserAccess } from "@/features/auth/adminApi";
import { useAuth } from "@/features/auth/AuthProvider";
import { archivePatient, createPatient, listPatients, searchPatients, updatePatient } from "@/features/patients/api";
import type { Patient } from "@/features/patients/types";

export default function PatientsPage() {
  const { user } = useAuth();
  const [dni, setDni] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");

  const [created, setCreated] = useState<Patient | null>(null);
  const [patients, setPatients] = useState<Patient[]>([]);
  const [search, setSearch] = useState("");
  const [includeInactive, setIncludeInactive] = useState(false);
  const [error, setError] = useState<string>("");
  const [inviteMessage, setInviteMessage] = useState("");
  const [isInviting, setIsInviting] = useState(false);
  const [invitingPatientId, setInvitingPatientId] = useState<string | null>(null);
  const [togglingPatientId, setTogglingPatientId] = useState<string | null>(null);
  const [isLoadingPatients, setIsLoadingPatients] = useState(false);

  useEffect(() => {
    refreshPatients();
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
      setError(e?.message ?? "No se pudieron cargar los pacientes.");
    } finally {
      setIsLoadingPatients(false);
    }
  }

  async function submit() {
    setError("");
    setInviteMessage("");
    setCreated(null);

    try {
      const res = await createPatient({
        dni,
        first_name: firstName,
        last_name: lastName,
        email,
      });

      setCreated(res);
      localStorage.setItem("last_patient_id", res.id);
      setDni("");
      setFirstName("");
      setLastName("");
      setEmail("");
      await refreshPatients("");
    } catch (e: any) {
      setError(e?.message ?? "Error");
    }
  }

  async function invitePatient() {
    if (!created) return;

    setError("");
    setInviteMessage("");
    setIsInviting(true);
    try {
      await inviteUserAccess({ email: created.email, role: "paciente" });
      setInviteMessage(`Invitación enviada a ${created.email}.`);
    } catch (e: any) {
      setError(e?.message ?? "No se pudo enviar la invitación.");
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
      setInviteMessage(`Invitación enviada a ${patient.email}.`);
    } catch (e: any) {
      setError(e?.message ?? "No se pudo enviar la invitación.");
    } finally {
      setInvitingPatientId(null);
    }
  }

  function selectForAgenda(patient: Patient) {
    if (!patient.active) {
      setError("No se puede usar en Agenda un paciente archivado.");
      return;
    }
    localStorage.setItem("last_patient_id", patient.id);
    setCreated(patient);
    setInviteMessage(`${patient.last_name}, ${patient.first_name} quedó seleccionado para Agenda.`);
  }

  async function togglePatientActive(patient: Patient) {
    setError("");
    setInviteMessage("");
    setTogglingPatientId(patient.id);

    try {
      if (patient.active) {
        await archivePatient(patient.id);
        setInviteMessage(`${patient.last_name}, ${patient.first_name} fue archivado.`);
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
        setInviteMessage(`${updated.last_name}, ${updated.first_name} fue reactivado.`);
      }
      await refreshPatients();
    } catch (e: any) {
      setError(e?.message ?? "No se pudo actualizar el estado del paciente.");
    } finally {
      setTogglingPatientId(null);
    }
  }

  return (
    <main>
      <div className="max-w-5xl mx-auto p-6 space-y-6">
        <header className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold">Pacientes</h1>
            <p className="text-sm text-gray-600">Alta, búsqueda y listado de pacientes.</p>
          </div>
          <Link className="text-sm underline" to="/agenda">Ir a Agenda</Link>
        </header>

        <div className="bg-white rounded-xl shadow p-4 space-y-3">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label className="text-sm font-medium">DNI</label>
              <input className="mt-1 w-full border rounded-lg p-2" value={dni} onChange={(e) => setDni(e.target.value)} />
            </div>
            <div>
              <label className="text-sm font-medium">Email</label>
              <input className="mt-1 w-full border rounded-lg p-2" value={email} onChange={(e) => setEmail(e.target.value)} />
            </div>
            <div>
              <label className="text-sm font-medium">Nombre</label>
              <input className="mt-1 w-full border rounded-lg p-2" value={firstName} onChange={(e) => setFirstName(e.target.value)} />
            </div>
            <div>
              <label className="text-sm font-medium">Apellido</label>
              <input className="mt-1 w-full border rounded-lg p-2" value={lastName} onChange={(e) => setLastName(e.target.value)} />
            </div>
          </div>

          <button className="px-4 py-2 rounded-lg bg-black text-white" onClick={submit}>
            Crear paciente
          </button>

          {error && <p className="text-sm text-red-600">Error: {error}</p>}

          {created && (
            <div className="mt-2 border rounded-lg p-3 bg-gray-50">
              <div className="font-medium">Paciente creado</div>
              <div className="text-sm text-gray-700">ID: <span className="font-mono">{created.id}</span></div>
              <div className="text-sm text-gray-700">{created.last_name}, {created.first_name} — {created.dni}</div>
              <p className="text-xs text-gray-500 mt-1">Se guardó como last_patient_id para usar en Agenda.</p>
              {user?.role === "admin" && (
                <button
                  type="button"
                  className="mt-3 px-3 py-2 rounded-lg border text-sm hover:bg-white disabled:opacity-50"
                  disabled={isInviting}
                  onClick={invitePatient}
                >
                  {isInviting ? "Enviando..." : "Enviar acceso al portal"}
                </button>
              )}
              {inviteMessage && <p className="text-sm text-green-700 mt-2">{inviteMessage}</p>}
            </div>
          )}
        </div>

        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
            <div className="flex-1">
              <h2 className="text-lg font-semibold">Listado de pacientes</h2>
              <label className="text-sm font-medium mt-3 block">Buscar</label>
              <input
                className="mt-1 w-full border rounded-lg p-2"
                placeholder="DNI, email, nombre o apellido"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
              <label className="flex items-center gap-2 text-sm mt-3">
                <input
                  type="checkbox"
                  checked={includeInactive}
                  onChange={(e) => setIncludeInactive(e.target.checked)}
                />
                Mostrar pacientes archivados
              </label>
            </div>
            <button
              type="button"
              className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50"
              onClick={() => refreshPatients()}
            >
              Actualizar
            </button>
          </div>

          {isLoadingPatients && <p className="text-sm text-gray-600">Cargando pacientes...</p>}

          <div className="divide-y">
            {!isLoadingPatients && patients.length === 0 ? (
              <p className="text-sm text-gray-600 py-3">No hay pacientes para mostrar.</p>
            ) : (
              patients.map((patient) => (
                <div key={patient.id} className="py-3 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                  <div>
                    <div className="font-medium">{patient.last_name}, {patient.first_name}</div>
                    <div className="text-sm text-gray-600">DNI: {patient.dni}</div>
                    <div className="text-sm text-gray-600">{patient.email}</div>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <span className={`text-xs px-2 py-1 rounded-md self-center ${patient.active ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-700"}`}>
                      {patient.active ? "Activo" : "Archivado"}
                    </span>
                    <button
                      type="button"
                      className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                      disabled={!patient.active}
                      onClick={() => selectForAgenda(patient)}
                    >
                      Usar en Agenda
                    </button>
                    {user?.role === "admin" && (
                      <button
                        type="button"
                        className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50 disabled:opacity-50"
                        disabled={invitingPatientId === patient.id}
                        onClick={() => invitePatientFromList(patient)}
                      >
                        {invitingPatientId === patient.id ? "Enviando..." : "Enviar acceso"}
                      </button>
                    )}
                    <button
                      type="button"
                      className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50 disabled:opacity-50"
                      disabled={togglingPatientId === patient.id}
                      onClick={() => togglePatientActive(patient)}
                    >
                      {togglingPatientId === patient.id
                        ? "Guardando..."
                        : patient.active
                          ? "Archivar"
                          : "Reactivar"}
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
