import { useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { inviteUserAccess } from "@/features/auth/adminApi";
import { useAuth } from "@/features/auth/AuthProvider";
import { listPatientAppointments } from "@/features/appointments/api";
import { listKinesiologists } from "@/features/kinesiologists/api";
import { formatLocalDateTime, formatLocalTime } from "@/shared/time/format";
import { archivePatient, getPatient, updatePatient } from "@/features/patients/api";
import {
  createPatientEvolution,
  listPatientEvolutions,
  listPatientMaterialLoans,
  listPatientPlans,
} from "@/features/patients/detailApi";

function historyRange() {
  const from = new Date();
  from.setFullYear(from.getFullYear() - 1);
  from.setHours(0, 0, 0, 0);

  const to = new Date();
  to.setFullYear(to.getFullYear() + 1);
  to.setHours(23, 59, 59, 999);

  return { from: from.toISOString(), to: to.toISOString() };
}

export default function PatientDetailPage() {
  const { patientId = "" } = useParams();
  const { user } = useAuth();
  const canManagePatient = user?.role === "admin" || user?.role === "recepcionista";
  const canCreateEvolution = user?.role === "admin" || user?.role === "kinesiologo";
  const backPath = user?.role === "kinesiologo" ? "/agenda" : "/patients";
  const range = useMemo(() => historyRange(), []);
  const [evolutionKinesiologistId, setEvolutionKinesiologistId] = useState("");
  const [evolutionAppointmentId, setEvolutionAppointmentId] = useState("");
  const [painLevel, setPainLevel] = useState("");
  const [evolutionNotes, setEvolutionNotes] = useState("");

  const patientQ = useQuery({
    queryKey: ["patients", "detail", patientId],
    queryFn: () => getPatient(patientId),
    enabled: Boolean(patientId),
  });

  const appointmentsQ = useQuery({
    queryKey: ["appointments", "patient-detail", patientId, range.from, range.to],
    queryFn: () => listPatientAppointments({ patient_id: patientId, from: range.from, to: range.to }),
    enabled: Boolean(patientId),
  });

  const plansQ = useQuery({
    queryKey: ["patients", "plans", patientId],
    queryFn: () => listPatientPlans(patientId),
    enabled: Boolean(patientId),
  });

  const evolutionsQ = useQuery({
    queryKey: ["patients", "evolutions", patientId],
    queryFn: () => listPatientEvolutions(patientId),
    enabled: Boolean(patientId),
  });

  const loansQ = useQuery({
    queryKey: ["patients", "material-loans", patientId],
    queryFn: () => listPatientMaterialLoans(patientId),
    enabled: Boolean(patientId),
  });

  const kinesiologistsQ = useQuery({
    queryKey: ["kinesiologists", "patient-detail"],
    queryFn: () => listKinesiologists(),
    enabled: canCreateEvolution,
  });

  const archiveM = useMutation({
    mutationFn: archivePatient,
    onSuccess: () => patientQ.refetch(),
  });

  const reactivateM = useMutation({
    mutationFn: async () => {
      const patient = patientQ.data;
      if (!patient) return;
      await updatePatient({
        id: patient.id,
        dni: patient.dni,
        first_name: patient.first_name,
        last_name: patient.last_name,
        email: patient.email,
        phone: patient.phone ?? null,
        active: true,
      });
    },
    onSuccess: () => patientQ.refetch(),
  });

  const inviteM = useMutation({
    mutationFn: (email: string) => inviteUserAccess({ email, role: "paciente" }),
  });

  const createEvolutionM = useMutation({
    mutationFn: () =>
      createPatientEvolution(patientId, {
        kinesiologist_id: evolutionKinesiologistId,
        appointment_id: evolutionAppointmentId || null,
        pain_level: painLevel === "" ? null : Number(painLevel),
        notes: evolutionNotes,
      }),
    onSuccess: () => {
      setEvolutionAppointmentId("");
      setPainLevel("");
      setEvolutionNotes("");
      evolutionsQ.refetch();
    },
  });

  const patient = patientQ.data;
  const appointments = appointmentsQ.data ?? [];
  const plans = plansQ.data ?? [];
  const evolutions = evolutionsQ.data ?? [];
  const loans = loansQ.data ?? [];
  const kinesiologists = kinesiologistsQ.data ?? [];
  const upcomingScheduledAppointments = appointments.filter(
    (appointment) => appointment.status !== "cancelled",
  );
  const timeline = useMemo(
    () =>
      [
        ...appointments.map((appointment) => ({
          id: `appointment:${appointment.id}`,
          at: appointment.start_at,
          kind: "Turno",
          title:
            appointment.status === "cancelled"
              ? "Turno cancelado"
              : "Turno programado",
          detail: `${formatLocalTime(appointment.start_at)} a ${formatLocalTime(appointment.end_at)}${appointment.notes ? ` · ${appointment.notes}` : ""}`,
        })),
        ...evolutions.map((evolution) => ({
          id: `evolution:${evolution.id}`,
          at: evolution.created_at,
          kind: "Evolución",
          title: evolution.pain_level != null ? `Dolor ${evolution.pain_level}/10` : "Evolución clínica",
          detail: evolution.notes,
        })),
        ...plans.map((plan) => ({
          id: `plan:${plan.id}`,
          at: plan.created_at,
          kind: "Plan",
          title: `${plan.frequency} · ${plan.duration_weeks} semanas`,
          detail: `${plan.items.length} ejercicios · estado ${plan.status}`,
        })),
        ...loans.map((loan) => ({
          id: `loan:${loan.id}`,
          at: loan.loaned_at,
          kind: "Material",
          title: loan.returned_at ? "Material devuelto" : "Material prestado",
          detail: `Cantidad ${loan.qty}${loan.notes ? ` · ${loan.notes}` : ""}`,
        })),
      ].sort((a, b) => new Date(b.at).getTime() - new Date(a.at).getTime()),
    [appointments, evolutions, loans, plans],
  );

  useEffect(() => {
    if (evolutionKinesiologistId || kinesiologists.length === 0) return;

    const ownProfile =
      user?.email
        ? kinesiologists.find((kinesiologist) => kinesiologist.email.toLowerCase() === user.email.toLowerCase())
        : undefined;
    setEvolutionKinesiologistId((ownProfile ?? kinesiologists[0]).id);
  }, [evolutionKinesiologistId, kinesiologists, user?.email]);

  return (
    <main>
      <div className="max-w-5xl mx-auto p-6 space-y-6">
        <header className="flex items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold">Detalle paciente</h1>
            <p className="text-sm text-gray-600">
              {patient ? `${patient.last_name}, ${patient.first_name}` : "Cargando ficha..."}
            </p>
          </div>
          <Link className="text-sm underline" to={backPath}>
            Volver
          </Link>
        </header>

        {patientQ.isError && (
          <div className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            No se pudo cargar el paciente.
          </div>
        )}

        {patient && (
          <section className="bg-white rounded-xl shadow p-4 space-y-4">
            <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
              <div>
                <h2 className="text-lg font-semibold">Ficha</h2>
                <div className="mt-2 grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-1 text-sm text-gray-700">
                  <div>DNI: <span className="font-medium">{patient.dni}</span></div>
                  <div>Email: <span className="font-medium">{patient.email}</span></div>
                  <div>Teléfono: <span className="font-medium">{patient.phone || "-"}</span></div>
                  <div>Estado: <span className="font-medium">{patient.active ? "Activo" : "Archivado"}</span></div>
                  {patient.birth_date && <div>Fecha nacimiento: <span className="font-medium">{patient.birth_date}</span></div>}
                </div>
                {patient.clinical_notes && (
                  <p className="text-sm text-gray-700 mt-3">{patient.clinical_notes}</p>
                )}
              </div>

              {canManagePatient && (
                <div className="flex flex-wrap gap-2">
                  <Link
                    className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50"
                    to="/agenda"
                    onClick={() => localStorage.setItem("last_patient_id", patient.id)}
                  >
                    Usar en Agenda
                  </Link>
                  {user?.role === "admin" && (
                    <button
                      type="button"
                      className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50 disabled:opacity-50"
                      disabled={inviteM.isPending}
                      onClick={() => inviteM.mutate(patient.email)}
                    >
                      {inviteM.isPending ? "Enviando..." : "Enviar acceso"}
                    </button>
                  )}
                  <button
                    type="button"
                    className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50 disabled:opacity-50"
                    disabled={archiveM.isPending || reactivateM.isPending}
                    onClick={() => {
                      if (patient.active) archiveM.mutate(patient.id);
                      else reactivateM.mutate();
                    }}
                  >
                    {patient.active ? "Archivar" : "Reactivar"}
                  </button>
                </div>
              )}
            </div>

            {inviteM.isSuccess && <p className="text-sm text-green-700">Invitación enviada.</p>}
            {(archiveM.isError || reactivateM.isError || inviteM.isError) && (
              <p className="text-sm text-red-600">No se pudo completar la acción.</p>
            )}
          </section>
        )}

        {canCreateEvolution && (
          <section className="bg-white rounded-xl shadow p-4 space-y-4">
            <div>
              <h2 className="text-lg font-semibold">Registrar evolución</h2>
              <p className="text-sm text-gray-600">Nueva nota clínica asociada a este paciente.</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div className="md:col-span-2">
                <label className="text-sm font-medium">Kinesiólogo</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={evolutionKinesiologistId}
                  onChange={(event) => setEvolutionKinesiologistId(event.target.value)}
                >
                  <option value="">Seleccionar...</option>
                  {kinesiologists.map((kinesiologist) => (
                    <option key={kinesiologist.id} value={kinesiologist.id}>
                      {kinesiologist.last_name}, {kinesiologist.first_name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-sm font-medium">Dolor</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={painLevel}
                  onChange={(event) => setPainLevel(event.target.value)}
                >
                  <option value="">Sin registrar</option>
                  {Array.from({ length: 11 }, (_, value) => (
                    <option key={value} value={value}>
                      {value}/10
                    </option>
                  ))}
                </select>
              </div>

              <div className="md:col-span-3">
                <label className="text-sm font-medium">Turno asociado</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={evolutionAppointmentId}
                  onChange={(event) => setEvolutionAppointmentId(event.target.value)}
                >
                  <option value="">Sin asociar</option>
                  {upcomingScheduledAppointments.map((appointment) => (
                    <option key={appointment.id} value={appointment.id}>
                      {formatLocalDateTime(appointment.start_at)}
                    </option>
                  ))}
                </select>
              </div>

              <div className="md:col-span-3">
                <label className="text-sm font-medium">Notas</label>
                <textarea
                  className="mt-1 w-full border rounded-lg p-2 min-h-28"
                  value={evolutionNotes}
                  onChange={(event) => setEvolutionNotes(event.target.value)}
                />
              </div>
            </div>

            <button
              type="button"
              className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
              disabled={!evolutionKinesiologistId || !evolutionNotes.trim() || createEvolutionM.isPending}
              onClick={() => createEvolutionM.mutate()}
            >
              {createEvolutionM.isPending ? "Guardando..." : "Guardar evolución"}
            </button>

            {createEvolutionM.isError && (
              <p className="text-sm text-red-600">No se pudo guardar la evolución.</p>
            )}
            {createEvolutionM.isSuccess && (
              <p className="text-sm text-green-700">Evolución registrada.</p>
            )}
          </section>
        )}

        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <h2 className="text-lg font-semibold">Timeline clínico</h2>
          {timeline.length === 0 ? (
            <p className="text-sm text-gray-600">Sin actividad registrada.</p>
          ) : (
            <div className="divide-y">
              {timeline.map((item) => (
                <div key={item.id} className="py-3 grid grid-cols-1 md:grid-cols-[160px_120px_1fr] gap-2">
                  <div className="text-sm text-gray-600">{formatLocalDateTime(item.at)}</div>
                  <div>
                    <span className="text-xs px-2 py-1 rounded-md bg-gray-100 text-gray-700">
                      {item.kind}
                    </span>
                  </div>
                  <div>
                    <div className="font-medium">{item.title}</div>
                    <div className="text-sm text-gray-600">{item.detail}</div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <h2 className="text-lg font-semibold">Turnos</h2>
          {appointmentsQ.isLoading && <p className="text-sm text-gray-600">Cargando turnos...</p>}
          {appointmentsQ.isError && <p className="text-sm text-red-600">No se pudieron cargar los turnos.</p>}
          {!appointmentsQ.isLoading && appointments.length === 0 ? (
            <p className="text-sm text-gray-600">No hay turnos en el rango consultado.</p>
          ) : (
            <div className="divide-y">
              {appointments.map((appointment) => (
                <div key={appointment.id} className="py-3 flex items-center justify-between gap-3">
                  <div>
                    <div className="font-medium">{formatLocalDateTime(appointment.start_at)}</div>
                    <div className="text-sm text-gray-600">
                      {formatLocalTime(appointment.start_at)} a {formatLocalTime(appointment.end_at)}
                    </div>
                    {appointment.notes && <div className="text-sm text-gray-600">Notas: {appointment.notes}</div>}
                  </div>
                  <span className={`text-xs px-2 py-1 rounded-md ${appointment.status === "cancelled" ? "bg-gray-100 text-gray-700" : "bg-blue-100 text-blue-700"}`}>
                    {appointment.status === "cancelled" ? "Cancelado" : "Programado"}
                  </span>
                </div>
              ))}
            </div>
          )}
        </section>

        <section className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Panel title="Evoluciones" isLoading={evolutionsQ.isLoading} isError={evolutionsQ.isError} empty={evolutions.length === 0}>
            {evolutions.map((evolution) => (
              <div key={evolution.id} className="py-3 border-b last:border-b-0">
                <div className="font-medium">{formatLocalDateTime(evolution.created_at)}</div>
                {evolution.pain_level != null && (
                  <div className="text-sm text-gray-600">Dolor: {evolution.pain_level}/10</div>
                )}
                <p className="text-sm text-gray-700 mt-1">{evolution.notes}</p>
              </div>
            ))}
          </Panel>

          <Panel title="Planes" isLoading={plansQ.isLoading} isError={plansQ.isError} empty={plans.length === 0}>
            {plans.map((plan) => (
              <div key={plan.id} className="py-3 border-b last:border-b-0">
                <div className="font-medium">{plan.frequency} · {plan.duration_weeks} semanas</div>
                <div className="text-sm text-gray-600">Estado: {plan.status}</div>
                <div className="text-sm text-gray-600">Ejercicios: {plan.items.length}</div>
                {plan.observations && <p className="text-sm text-gray-700 mt-1">{plan.observations}</p>}
              </div>
            ))}
          </Panel>

          <Panel title="Materiales" isLoading={loansQ.isLoading} isError={loansQ.isError} empty={loans.length === 0}>
            {loans.map((loan) => (
              <div key={loan.id} className="py-3 border-b last:border-b-0">
                <div className="font-medium">Cantidad: {loan.qty}</div>
                <div className="text-sm text-gray-600">Prestado: {formatLocalDateTime(loan.loaned_at)}</div>
                <div className="text-sm text-gray-600">
                  {loan.returned_at ? `Devuelto: ${formatLocalDateTime(loan.returned_at)}` : "Pendiente de devolución"}
                </div>
                {loan.notes && <p className="text-sm text-gray-700 mt-1">{loan.notes}</p>}
              </div>
            ))}
          </Panel>
        </section>
      </div>
    </main>
  );
}

function Panel({
  title,
  isLoading,
  isError,
  empty,
  children,
}: {
  title: string;
  isLoading: boolean;
  isError: boolean;
  empty: boolean;
  children: ReactNode;
}) {
  return (
    <section className="bg-white rounded-xl shadow p-4">
      <h2 className="text-lg font-semibold">{title}</h2>
      {isLoading && <p className="text-sm text-gray-600 mt-2">Cargando...</p>}
      {isError && <p className="text-sm text-red-600 mt-2">No se pudo cargar la información.</p>}
      {!isLoading && !isError && empty ? (
        <p className="text-sm text-gray-600 mt-2">Sin registros.</p>
      ) : (
        <div className="mt-2">{children}</div>
      )}
    </section>
  );
}
