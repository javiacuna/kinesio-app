import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useAuth } from "@/features/auth/AuthProvider";
import {
  cancelAppointment,
  createAppointment,
  listPatientAppointments,
} from "@/features/appointments/api";
import type { Appointment } from "@/features/appointments/types";
import { listKinesiologists } from "@/features/kinesiologists/api";
import type { Kinesiologist } from "@/features/kinesiologists/types";
import {
  type PatientAttachment,
  type PatientEvolution,
  downloadPatientAttachment,
  listMyPatientAttachments,
  listMyPatientEvolutions,
  listMyPatientPlans,
} from "@/features/patients/detailApi";
import { formatLocalDateTime, formatLocalTime } from "@/shared/time/format";
import { addMinutesToHHmm, localDateTimeToUTC } from "@/shared/time/rfc3339";

const defaultWorkDays = [1, 2, 3, 4, 5];
const workDayLabels: Record<number, string> = {
  1: "Lun",
  2: "Mar",
  3: "Mié",
  4: "Jue",
  5: "Vie",
  6: "Sáb",
  7: "Dom",
};

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

function isOutsideWorkingSchedule(dateISO: string, startHHmm: string, durationMin: number, kinesiologist?: Kinesiologist) {
  if (!kinesiologist) return false;
  if (isOutsideWorkingHours(startHHmm, durationMin, kinesiologist.work_start_time, kinesiologist.work_end_time)) {
    return true;
  }

  const days = kinesiologist.work_days?.length ? kinesiologist.work_days : defaultWorkDays;
  const [year, month, day] = dateISO.split("-").map(Number);
  const date = new Date(year, month - 1, day);
  const isoDay = date.getDay() === 0 ? 7 : date.getDay();
  return !days.includes(isoDay);
}

function workDaysLabel(days?: number[]) {
  const selected = days?.length ? days : defaultWorkDays;
  return selected.map((day) => workDayLabels[day]).filter(Boolean).join(", ");
}

export default function PatientPortalPage() {
  const { user } = useAuth();
  const range = useMemo(() => appointmentRange(), []);
  const [kinesiologistId, setKinesiologistId] = useState("");
  const [date, setDate] = useState(todayISO());
  const [startTime, setStartTime] = useState("09:00");
  const [durationMin, setDurationMin] = useState(45);
  const [notes, setNotes] = useState("");
  const [cancelTarget, setCancelTarget] = useState<Appointment | null>(null);
  const [cancelReason, setCancelReason] = useState("");

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

  const evolutionsQ = useQuery({
    queryKey: ["patients", "me", "evolutions"],
    queryFn: () => listMyPatientEvolutions(),
  });

  const kinesiologistsQ = useQuery({
    queryKey: ["kinesiologists"],
    queryFn: () => listKinesiologists(),
  });

  const kinesiologists = kinesiologistsQ.data ?? [];
  const appointments = appointmentsQ.data ?? [];
  const plans = plansQ.data ?? [];
  const attachments = attachmentsQ.data ?? [];
  const evolutions = evolutionsQ.data ?? [];
  const selectedKinesiologist = kinesiologists.find((kinesiologist) => kinesiologist.id === kinesiologistId);
  const isCreateInPast = isPastLocalDateTime(date, startTime);
  const isCreateOutsideWorkingSchedule = isOutsideWorkingSchedule(date, startTime, durationMin, selectedKinesiologist);

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
    onSuccess: () => {
      setCancelTarget(null);
      setCancelReason("");
      appointmentsQ.refetch();
    },
  });

  function create() {
    if (!kinesiologistId || isCreateInPast || isCreateOutsideWorkingSchedule) return;

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
                Horario de atención: {selectedKinesiologist.work_start_time} a {selectedKinesiologist.work_end_time} · {workDaysLabel(selectedKinesiologist.work_days)}.
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
          disabled={!kinesiologistId || isCreateInPast || isCreateOutsideWorkingSchedule || createM.isPending}
          onClick={create}
        >
          {createM.isPending ? "Reservando..." : "Reservar turno"}
        </button>

        {isCreateInPast && (
          <p className="text-sm text-red-600">No se pueden reservar turnos en horarios pasados.</p>
        )}

        {isCreateOutsideWorkingSchedule && selectedKinesiologist && (
          <p className="text-sm text-red-600">
            El turno debe estar dentro del horario de {selectedKinesiologist.work_start_time} a {selectedKinesiologist.work_end_time} y en sus días laborales: {workDaysLabel(selectedKinesiologist.work_days)}.
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
        <h2 className="text-lg font-semibold">Mi evolución clínica</h2>

        {evolutionsQ.isLoading && (
          <p className="text-sm text-gray-600 mt-2">Cargando evolución...</p>
        )}

        {evolutionsQ.isError && (
          <p className="text-sm text-red-600 mt-2">No se pudo cargar tu evolución clínica.</p>
        )}

        {evolutionsQ.isSuccess && evolutions.length === 0 && (
          <p className="text-sm text-gray-600 mt-2">Todavía no hay evoluciones compartidas.</p>
        )}

        {evolutions.length > 0 && (
          <div className="mt-3 space-y-4">
            <PatientEvolutionChart evolutions={evolutions} />
            <div className="divide-y">
              {evolutions.map((evolution) => (
                <article key={evolution.id} className="py-3">
                  <div className="font-medium">{formatLocalDateTime(evolution.created_at)}</div>
                  <EvolutionMetricLine evolution={evolution} />
                  <p className="text-sm text-gray-700 mt-2">{evolution.notes}</p>
                </article>
              ))}
            </div>
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
              <article key={appointment.id} className="py-3 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <div className="font-medium">
                    {formatLocalDateTime(appointment.start_at)}
                  </div>
                  <div className="text-sm text-gray-600">
                    {formatLocalTime(appointment.start_at)} a {formatLocalTime(appointment.end_at)}
                  </div>
                  <div className="text-sm text-gray-600">
                    {appointment.modality === "virtual" ? "Videollamada" : "Presencial"}
                    {appointment.modality === "virtual" && appointment.video_call_url && (
                      <>
                        {" · "}
                        <a className="underline text-blue-700" href={appointment.video_call_url} target="_blank" rel="noreferrer">
                          Abrir videollamada
                        </a>
                      </>
                    )}
                  </div>
                  {appointment.notes && (
                    <div className="text-sm text-gray-600 mt-1">Notas: {appointment.notes}</div>
                  )}
                  {appointment.status === "cancelled" && appointment.cancelled_reason && (
                    <div className="text-sm text-gray-600 mt-1">Motivo: {appointment.cancelled_reason}</div>
                  )}
                </div>

                <div className="flex flex-wrap items-center gap-2">
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
                        setCancelTarget(appointment);
                        setCancelReason("");
                      }}
                    >
                      Cancelar
                    </button>
                  )}
                </div>
              </article>
            ))}
          </div>
        )}

        {cancelTarget && (
          <div className="mt-4 border rounded-lg p-3 space-y-3 bg-gray-50">
            <div>
              <h3 className="font-medium">Cancelar turno</h3>
              <p className="text-sm text-gray-600">{formatLocalDateTime(cancelTarget.start_at)}</p>
            </div>
            <div>
              <label className="text-sm font-medium">Motivo</label>
              <textarea
                className="mt-1 w-full border rounded-lg p-2 min-h-20 bg-white"
                value={cancelReason}
                onChange={(event) => setCancelReason(event.target.value)}
                placeholder="Opcional"
              />
            </div>
            <div className="flex flex-col gap-2 sm:flex-row">
              <button
                type="button"
                className="px-3 py-2 rounded-lg bg-black text-white text-sm disabled:opacity-50"
                disabled={cancelM.isPending}
                onClick={() => cancelM.mutate({ id: cancelTarget.id, reason: cancelReason.trim() || undefined })}
              >
                {cancelM.isPending ? "Cancelando..." : "Confirmar cancelación"}
              </button>
              <button
                type="button"
                className="px-3 py-2 rounded-lg border text-sm hover:bg-white"
                disabled={cancelM.isPending}
                onClick={() => {
                  setCancelTarget(null);
                  setCancelReason("");
                }}
              >
                Volver
              </button>
            </div>
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

type EvolutionMetricKey = "pain_level" | "mobility_score" | "strength_score" | "functional_score";

const evolutionMetricDefs: Array<{
  key: EvolutionMetricKey;
  label: string;
  color: string;
  max: number;
}> = [
  { key: "pain_level", label: "Dolor", color: "#dc2626", max: 10 },
  { key: "mobility_score", label: "Movilidad", color: "#2563eb", max: 100 },
  { key: "strength_score", label: "Fuerza", color: "#16a34a", max: 100 },
  { key: "functional_score", label: "Funcionalidad", color: "#9333ea", max: 100 },
];

function PatientEvolutionChart({ evolutions }: { evolutions: PatientEvolution[] }) {
  const points = [...evolutions]
    .filter((evolution) => evolutionMetricDefs.some((metric) => evolution[metric.key] != null))
    .sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());

  if (points.length === 0) {
    return <p className="text-sm text-gray-600">No hay métricas clínicas para graficar.</p>;
  }

  const width = 700;
  const height = 240;
  const padding = { top: 18, right: 24, bottom: 48, left: 42 };
  const chartWidth = width - padding.left - padding.right;
  const chartHeight = height - padding.top - padding.bottom;
  const xFor = (index: number) => padding.left + (points.length === 1 ? chartWidth / 2 : (chartWidth * index) / (points.length - 1));
  const yFor = (value: number) => padding.top + chartHeight - (chartHeight * value) / 100;

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap gap-3">
        {evolutionMetricDefs.map((metric) => (
          <div key={metric.key} className="flex items-center gap-2 text-sm text-gray-700">
            <span className="h-3 w-3 rounded-full" style={{ backgroundColor: metric.color }} />
            {metric.label}
          </div>
        ))}
      </div>

      <div className="overflow-x-auto">
        <svg className="min-w-[620px] w-full" viewBox={`0 0 ${width} ${height}`} role="img" aria-label="Gráfico de evolución clínica">
          {[0, 25, 50, 75, 100].map((value) => (
            <g key={value}>
              <line x1={padding.left} x2={width - padding.right} y1={yFor(value)} y2={yFor(value)} stroke="#e5e7eb" />
              <text x={padding.left - 9} y={yFor(value) + 4} textAnchor="end" className="fill-gray-500 text-[11px]">
                {value}
              </text>
            </g>
          ))}

          {evolutionMetricDefs.map((metric) => {
            const metricPoints = points
              .map((evolution, index) => {
                const rawValue = evolution[metric.key];
                if (rawValue == null) return null;
                return {
                  x: xFor(index),
                  y: yFor(metric.key === "pain_level" ? rawValue * 10 : rawValue),
                  label: `${metric.label}: ${rawValue}/${metric.max}`,
                };
              })
              .filter((point): point is { x: number; y: number; label: string } => point != null);

            return (
              <g key={metric.key}>
                {metricPoints.length > 1 && (
                  <polyline
                    fill="none"
                    stroke={metric.color}
                    strokeWidth="3"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    points={metricPoints.map((point) => `${point.x},${point.y}`).join(" ")}
                  />
                )}
                {metricPoints.map((point) => (
                  <circle key={`${metric.key}-${point.x}-${point.y}`} cx={point.x} cy={point.y} r="4" fill={metric.color}>
                    <title>{point.label}</title>
                  </circle>
                ))}
              </g>
            );
          })}

          {points.map((evolution, index) => (
            <text key={evolution.id} x={xFor(index)} y={height - 18} textAnchor="middle" className="fill-gray-600 text-[11px]">
              {formatLocalDateTime(evolution.created_at)}
            </text>
          ))}
        </svg>
      </div>
    </div>
  );
}

function EvolutionMetricLine({ evolution }: { evolution: PatientEvolution }) {
  const values = evolutionMetricDefs
    .map((metric) => {
      const value = evolution[metric.key];
      return value == null ? null : `${metric.label}: ${value}/${metric.max}`;
    })
    .filter((value): value is string => value != null);

  if (values.length === 0) return null;

  return (
    <div className="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-sm text-gray-600">
      {values.map((value) => (
        <span key={value}>{value}</span>
      ))}
    </div>
  );
}
