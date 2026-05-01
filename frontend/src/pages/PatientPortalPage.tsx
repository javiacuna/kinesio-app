import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useAuth } from "@/features/auth/AuthProvider";
import {
  cancelAppointment,
  createAppointment,
  listPatientAppointments,
} from "@/features/appointments/api";
import { listKinesiologists } from "@/features/kinesiologists/api";
import {
  type PatientAttachment,
  downloadPatientAttachment,
  listMyPatientAttachments,
  listMyPatientPlans,
} from "@/features/patients/detailApi";
import { formatLocalDateTime, formatLocalTime } from "@/shared/time/format";
import { addMinutesToHHmm, localDateTimeToUTC } from "@/shared/time/rfc3339";

function appointmentRange() {
  const from = new Date();
  from.setHours(0, 0, 0, 0);

  const to = new Date(from);
  to.setDate(to.getDate() + 180);
  to.setHours(23, 59, 59, 999);

  return {
    from: from.toISOString(),
    to: to.toISOString(),
  };
}

function todayISO() {
  const d = new Date();
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function isPastLocalDateTime(dateISO: string, timeHHmm: string) {
  if (!dateISO || !timeHHmm) return false;
  const [y, m, d] = dateISO.split("-").map(Number);
  const [hh, mm] = timeHHmm.split(":").map(Number);
  return new Date(y, m - 1, d, hh, mm, 0, 0).getTime() <= Date.now();
}

function isOverlapError(error: unknown) {
  return (error as any)?.status === 409 || (error as any)?.message === "overlap";
}

function isOutsideWorkingHours(startHHmm: string, durationMin: number, workStart?: string, workEnd?: string) {
  const end = addMinutesToHHmm(startHHmm, durationMin);
  const from = workStart || "08:00";
  const to = workEnd || "20:00";
  return startHHmm < from || end > to || end <= startHHmm;
}

export default function PatientPortalPage() {
  const { user } = useAuth();
  const range = useMemo(() => appointmentRange(), []);
  const [kinesiologistId, setKinesiologistId] = useState("");
  const [date, setDate] = useState(todayISO());
  const [startTime, setStartTime] = useState("09:00");
  const [durationMin, setDurationMin] = useState(45);
  const [notes, setNotes] = useState("");

  const appointmentsQ = useQuery({
    queryKey: ["appointments", "patient", range.from, range.to],
    queryFn: () => listPatientAppointments(range),
  });

  const plansQ = useQuery({
    queryKey: ["patients", "me", "plans"],
    queryFn: () => listMyPatientPlans(),
  });

  const attachmentsQ = useQuery({
    queryKey: ["patients", "me", "attachments"],
    queryFn: () => listMyPatientAttachments(),
  });

  const kinesiologistsQ = useQuery({
    queryKey: ["kinesiologists"],
    queryFn: () => listKinesiologists(),
  });

  const kinesiologists = kinesiologistsQ.data ?? [];
  const appointments = appointmentsQ.data ?? [];
  const plans = plansQ.data ?? [];
  const attachments = attachmentsQ.data ?? [];
  const selectedKinesiologist = kinesiologists.find((kinesiologist) => kinesiologist.id === kinesiologistId);
  const isCreateInPast = isPastLocalDateTime(date, startTime);
  const isCreateOutsideWorkingHours = selectedKinesiologist
    ? isOutsideWorkingHours(
        startTime,
        durationMin,
        selectedKinesiologist.work_start_time,
        selectedKinesiologist.work_end_time,
      )
    : false;

  useEffect(() => {
    if (!kinesiologistId && kinesiologists.length > 0) {
      setKinesiologistId(kinesiologists[0].id);
    }
  }, [kinesiologistId, kinesiologists]);

  const createM = useMutation({
    mutationFn: createAppointment,
    onSuccess: () => {
      setNotes("");
      appointmentsQ.refetch();
    },
  });

  const cancelM = useMutation({
    mutationFn: cancelAppointment,
    onSuccess: () => appointmentsQ.refetch(),
  });

  function create() {
    if (!kinesiologistId || isCreateInPast || isCreateOutsideWorkingHours) return;

    const endTime = addMinutesToHHmm(startTime, durationMin);
    createM.mutate({
      kinesiologist_id: kinesiologistId,
      start_at: localDateTimeToUTC(date, startTime),
      end_at: localDateTimeToUTC(date, endTime),
      notes: notes.trim() ? notes.trim() : undefined,
    });
  }

  async function openAttachment(attachment: PatientAttachment) {
    const blob = await downloadPatientAttachment(attachment);
    const url = URL.createObjectURL(blob);
    window.open(url, "_blank", "noopener,noreferrer");
    window.setTimeout(() => URL.revokeObjectURL(url), 60_000);
  }

  return (
    <main className="max-w-4xl mx-auto p-6 space-y-6">
      <header>
        <h1 className="text-2xl font-semibold">Portal paciente</h1>
        <p className="text-sm text-gray-600">{user?.email}</p>
      </header>

      <section className="bg-white border rounded-lg p-4 space-y-4">
        <h2 className="text-lg font-semibold">Nuevo turno</h2>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="md:col-span-2">
            <label className="text-sm font-medium">Kinesiólogo</label>
            <select
              className="mt-1 w-full border rounded-lg p-2"
              value={kinesiologistId}
              onChange={(e) => setKinesiologistId(e.target.value)}
            >
              <option value="">Seleccionar...</option>
              {kinesiologists.map((kinesiologist) => (
                <option key={kinesiologist.id} value={kinesiologist.id}>
                  {kinesiologist.last_name}, {kinesiologist.first_name}
                </option>
              ))}
            </select>
            {kinesiologistsQ.isError && (
              <p className="text-sm text-red-600 mt-1">No se pudieron cargar los kinesiólogos.</p>
            )}
            {selectedKinesiologist && (
              <p className="text-sm text-gray-600 mt-2">
                Horario de atención: {selectedKinesiologist.work_start_time} a {selectedKinesiologist.work_end_time}.
              </p>
            )}
          </div>

          <div>
            <label className="text-sm font-medium">Fecha</label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              type="date"
              min={todayISO()}
              value={date}
              onChange={(e) => setDate(e.target.value)}
            />
          </div>

          <div>
            <label className="text-sm font-medium">Hora</label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              type="time"
              value={startTime}
              onChange={(e) => setStartTime(e.target.value)}
            />
          </div>

          <div>
            <label className="text-sm font-medium">Duración</label>
            <select
              className="mt-1 w-full border rounded-lg p-2"
              value={durationMin}
              onChange={(e) => setDurationMin(Number(e.target.value))}
            >
              <option value={30}>30 min</option>
              <option value={45}>45 min</option>
              <option value={60}>60 min</option>
            </select>
          </div>

          <div>
            <label className="text-sm font-medium">Notas</label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
            />
          </div>
        </div>

        <button
          className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
          disabled={!kinesiologistId || isCreateInPast || isCreateOutsideWorkingHours || createM.isPending}
          onClick={create}
        >
          {createM.isPending ? "Reservando..." : "Reservar turno"}
        </button>

        {isCreateInPast && (
          <p className="text-sm text-red-600">No se pueden reservar turnos en horarios pasados.</p>
        )}

        {isCreateOutsideWorkingHours && selectedKinesiologist && (
          <p className="text-sm text-red-600">
            El turno debe estar dentro del horario de {selectedKinesiologist.work_start_time} a {selectedKinesiologist.work_end_time}.
          </p>
        )}

        {createM.isError && (
          <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700">
            {isOverlapError(createM.error)
              ? "Ese horario ya no está disponible para el kinesiólogo seleccionado."
              : `No se pudo reservar el turno: ${(createM.error as any)?.message}`}
          </div>
        )}
      </section>

      <section className="bg-white border rounded-lg p-4">
        <h2 className="text-lg font-semibold">Mis planes</h2>

        {plansQ.isLoading && (
          <p className="text-sm text-gray-600 mt-2">Cargando planes...</p>
        )}

        {plansQ.isError && (
          <p className="text-sm text-red-600 mt-2">No se pudieron cargar tus planes.</p>
        )}

        {plansQ.isSuccess && plans.length === 0 && (
          <p className="text-sm text-gray-600 mt-2">No tenés planes de ejercicios.</p>
        )}

        {plans.length > 0 && (
          <div className="mt-3 divide-y">
            {plans.map((plan) => (
              <article key={plan.id} className="py-3">
                <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
                  <div className="font-medium">
                    {plan.frequency === "daily" ? "Diario" : "Semanal"} · {plan.duration_weeks} semanas
                  </div>
                  <span className="text-xs px-2 py-1 rounded-md bg-gray-100 text-gray-700 w-fit">
                    {plan.status === "closed" ? "Cerrado" : "Activo"}
                  </span>
                </div>
                {plan.observations && (
                  <p className="text-sm text-gray-700 mt-2">{plan.observations}</p>
                )}
                <div className="mt-3 space-y-2">
                  {plan.items.map((item) => (
                    <div key={item.id} className="rounded-lg border p-3">
                      <div className="font-medium">{item.name}</div>
                      <div className="text-sm text-gray-600">
                        {item.estimated_minutes} min
                        {item.sets ? ` · ${item.sets} series` : ""}
                        {item.reps ? ` · ${item.reps} repeticiones` : ""}
                      </div>
                      {item.description && (
                        <p className="text-sm text-gray-700 mt-1">{item.description}</p>
                      )}
                    </div>
                  ))}
                </div>
              </article>
            ))}
          </div>
        )}
      </section>

      <section className="bg-white border rounded-lg p-4">
        <h2 className="text-lg font-semibold">Mis estudios</h2>

        {attachmentsQ.isLoading && (
          <p className="text-sm text-gray-600 mt-2">Cargando estudios...</p>
        )}

        {attachmentsQ.isError && (
          <p className="text-sm text-red-600 mt-2">No se pudieron cargar tus estudios.</p>
        )}

        {attachmentsQ.isSuccess && attachments.length === 0 && (
          <p className="text-sm text-gray-600 mt-2">No tenés estudios compartidos.</p>
        )}

        {attachments.length > 0 && (
          <div className="mt-3 divide-y">
            {attachments.map((attachment) => (
              <article key={attachment.id} className="py-3 flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <div className="font-medium">{attachment.file_name}</div>
                  <div className="text-sm text-gray-600">
                    {attachment.category || attachment.kind} · {Math.max(1, Math.round(attachment.size_bytes / 1024))} KB ·{" "}
                    {formatLocalDateTime(attachment.created_at)}
                  </div>
                  {attachment.notes && <p className="text-sm text-gray-700 mt-1">{attachment.notes}</p>}
                </div>
                <button
                  className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-100"
                  onClick={() => openAttachment(attachment)}
                >
                  Ver archivo
                </button>
              </article>
            ))}
          </div>
        )}
      </section>

      <section className="bg-white border rounded-lg p-4">
        <h2 className="text-lg font-semibold">Mis turnos</h2>

        {appointmentsQ.isLoading && (
          <p className="text-sm text-gray-600 mt-2">Cargando turnos...</p>
        )}

        {appointmentsQ.isError && (
          <p className="text-sm text-red-600 mt-2">
            No se pudieron cargar tus turnos.
          </p>
        )}

        {appointmentsQ.isSuccess && appointments.length === 0 && (
          <p className="text-sm text-gray-600 mt-2">No tenés turnos próximos agendados.</p>
        )}

        {appointments.length > 0 && (
          <div className="mt-3 divide-y">
            {appointments.map((appointment) => (
              <article key={appointment.id} className="py-3 flex items-start justify-between gap-4">
                <div>
                  <div className="font-medium">
                    {formatLocalDateTime(appointment.start_at)}
                  </div>
                  <div className="text-sm text-gray-600">
                    {formatLocalTime(appointment.start_at)} a {formatLocalTime(appointment.end_at)}
                  </div>
                  {appointment.notes && (
                    <div className="text-sm text-gray-600 mt-1">Notas: {appointment.notes}</div>
                  )}
                </div>

                <div className="flex items-center gap-2">
                  <span
                    className={
                      appointment.status === "cancelled"
                        ? "text-xs px-2 py-1 rounded-md bg-gray-100 text-gray-700"
                        : "text-xs px-2 py-1 rounded-md bg-blue-100 text-blue-700"
                    }
                  >
                    {appointment.status === "cancelled" ? "Cancelado" : "Programado"}
                  </span>

                  {appointment.status !== "cancelled" && (
                    <button
                      className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-100 disabled:opacity-50"
                      disabled={cancelM.isPending}
                      onClick={() => {
                        const reason = window.prompt("Motivo de cancelación (opcional):") ?? undefined;
                        cancelM.mutate({ id: appointment.id, reason });
                      }}
                    >
                      {cancelM.isPending ? "Cancelando..." : "Cancelar"}
                    </button>
                  )}
                </div>
              </article>
            ))}
          </div>
        )}

        {cancelM.isError && (
          <p className="text-sm text-red-600 mt-2">
            No se pudo cancelar el turno: {(cancelM.error as any)?.message}
          </p>
        )}
      </section>
    </main>
  );
}
