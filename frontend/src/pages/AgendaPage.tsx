import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { listKinesiologists } from "../features/kinesiologists/api";
import { createAppointment, listAppointmentsDay, updateAppointment } from "../features/appointments/api";
import { addMinutesToHHmm, localDateTimeToUTC } from "../shared/time/rfc3339";
import { formatLocalTime } from "../shared/time/format";
import { PatientSearch } from "@/features/patients/components/PatientSearch";
import { AgendaGrid } from "@/features/appointments/components/AgendaGrid";

function todayISO() {
  const d = new Date();
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

export default function AgendaPage() {
  // Agenda (listado)
  const [date, setDate] = useState(todayISO());
  const [kinesiologistId, setKinesiologistId] = useState("");

  // Crear turno (inputs UX)
  const [patientId, setPatientId] = useState("");
  const [apptDate, setApptDate] = useState(todayISO());
  const [startTime, setStartTime] = useState("09:00");
  const [durationMin, setDurationMin] = useState(45);
  const [notes, setNotes] = useState("");

  useEffect(() => {
    const last = localStorage.getItem("last_patient_id");
    if (last && !patientId) setPatientId(last);
  }, [patientId]);

  useEffect(() => {
    setApptDate(date);
  }, [date]);

  const kinesioQ = useQuery({
    queryKey: ["kinesiologists"],
    queryFn: listKinesiologists,
  });

  const kinesios = useMemo(() => kinesioQ.data ?? [], [kinesioQ.data]);

  const canLoadAgenda = Boolean(kinesiologistId);

  const agendaQ = useQuery({
    queryKey: ["appointments", "day", date, kinesiologistId],
    queryFn: () => listAppointmentsDay({ date, kinesiologist_id: kinesiologistId }),
    enabled: canLoadAgenda,
  });

  const createM = useMutation({
    mutationFn: createAppointment,
    onSuccess: () => agendaQ.refetch(),
  });

  const cancelM = useMutation({
    mutationFn: (args: { id: string; reason?: string }) =>
      updateAppointment({
        id: args.id,
        status: "cancelled",
        cancelled_reason: args.reason ?? "Cancelado desde la agenda",
      }),
    onSuccess: () => agendaQ.refetch(),
  });

  const rescheduleM = useMutation({
    mutationFn: (args: { id: string; start_at: string; end_at: string }) =>
      updateAppointment({ id: args.id, start_at: args.start_at, end_at: args.end_at }),
    onSuccess: () => agendaQ.refetch(),
  });


  function create() {
    const endTime = addMinutesToHHmm(startTime, durationMin);

    createM.mutate({
      patient_id: patientId.trim(),
      kinesiologist_id: kinesiologistId,
      start_at: localDateTimeToUTC(apptDate, startTime),
      end_at: localDateTimeToUTC(apptDate, endTime),
      notes: notes.trim() ? notes.trim() : undefined,
    });
  }

  const createErr: any = createM.error;
  const isOverlap = createErr?.status === 409;

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto p-6 space-y-6">
        <header className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold">Agenda</h1>
            <p className="text-sm text-gray-600">Agenda diaria y creación de turnos.</p>
          </div>
          <a className="text-sm underline" href="/patients">Ir a Pacientes</a>
        </header>

        {/* Filtros agenda */}
        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div>
              <label className="text-sm font-medium">Fecha (para ver agenda)</label>
              <input
                className="mt-1 w-full border rounded-lg p-2"
                type="date"
                value={date}
                onChange={(e) => setDate(e.target.value)}
              />
            </div>

            <div className="md:col-span-2">
              <label className="text-sm font-medium">Kinesiólogo</label>
              <select
                className="mt-1 w-full border rounded-lg p-2"
                value={kinesiologistId}
                onChange={(e) => setKinesiologistId(e.target.value)}
              >
                <option value="">Seleccionar…</option>
                {kinesios.map((k) => (
                  <option key={k.id} value={k.id}>
                    {k.last_name}, {k.first_name}
                  </option>
                ))}
              </select>

              {kinesioQ.isError && (
                <p className="text-sm text-red-600 mt-1">
                  Error: {String(kinesioQ.error?.message)}
                </p>
              )}
            </div>
          </div>
        </section>

        {/* Crear turno */}
        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <h2 className="text-lg font-semibold">Crear turno</h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div className="md:col-span-2">
              <PatientSearch
                valuePatientId={patientId}
                onSelect={(p) => {
                  setPatientId(p.id);
                  localStorage.setItem("last_patient_id", p.id);
                }}
              />
              <p className="text-xs text-gray-500 mt-2">
                Tip: si no existe, crealo en /patients y después buscá por DNI o email.
              </p>
            </div>

            <div>
              <label className="text-sm font-medium">Fecha del turno</label>
              <input
                className="mt-1 w-full border rounded-lg p-2"
                type="date"
                value={apptDate}
                onChange={(e) => setApptDate(e.target.value)}
              />
            </div>

            <div>
              <label className="text-sm font-medium">Hora inicio</label>
              <input
                className="mt-1 w-full border rounded-lg p-2"
                type="time"
                value={startTime}
                onChange={(e) => setStartTime(e.target.value)}
              />
            </div>

            <div>
              <label className="text-sm font-medium">Duración (min)</label>
              <input
                className="mt-1 w-full border rounded-lg p-2"
                type="number"
                min={15}
                step={15}
                value={durationMin}
                onChange={(e) => setDurationMin(Number(e.target.value))}
              />

              <div className="flex gap-2 mt-2">
                {[30, 45, 60].map((m) => (
                  <button
                    key={m}
                    type="button"
                    className="px-3 py-1 rounded-lg border bg-white text-sm hover:bg-gray-100"
                    onClick={() => setDurationMin(m)}
                  >
                    {m} min
                  </button>
                ))}
              </div>
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
            disabled={!patientId.trim() || !kinesiologistId || createM.isPending}
            onClick={create}
          >
            {createM.isPending ? "Creando…" : "Crear turno"}
          </button>

          {createM.isError && (
            <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700">
              {isOverlap ? (
                <>
                  <div className="font-medium">Solapamiento detectado</div>
                  <div>El kinesiólogo ya tiene un turno en ese horario. Elegí otro horario.</div>
                </>
              ) : (
                <>
                  <div className="font-medium">Error</div>
                  <div>{String(createErr?.message)}</div>
                </>
              )}
            </div>
          )}
        </section>

        {/* Agenda del día */}
        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <h2 className="text-lg font-semibold">Turnos del día</h2>

          {!canLoadAgenda && <p className="text-sm text-gray-600">Seleccioná un kinesiólogo.</p>}

          {agendaQ.isLoading && <p className="text-sm text-gray-600">Cargando…</p>}
          {agendaQ.isError && <p className="text-sm text-red-600">Error: {String(agendaQ.error?.message)}</p>}

          {agendaQ.data && (
            <div className="space-y-4">
              <AgendaGrid
                date={date}
                appointments={agendaQ.data}
                onPickSlot={(hhmm) => {
                  setApptDate(date);
                  setStartTime(hhmm);
                  setDurationMin(45);
                  // opcional: llevar al formulario
                  window.scrollTo({ top: 0, behavior: "smooth" });
                }}
                onCancel={(appt) => {
                  const reason = window.prompt("Motivo de cancelación (opcional):") ?? undefined;
                  cancelM.mutate({ id: appt.id, reason });
                }}
                onReschedule={(appt) => {
                  const newDate = window.prompt("Nueva fecha (YYYY-MM-DD):", date);
                  if (!newDate) return;

                  const newStart = window.prompt("Nueva hora inicio (HH:MM):", "09:00");
                  if (!newStart) return;

                  const dur = window.prompt("Duración en minutos:", "45");
                  if (!dur) return;

                  const durationMin = Number(dur);
                  if (!Number.isFinite(durationMin) || durationMin <= 0) return;

                  const endTime = addMinutesToHHmm(newStart, durationMin);

                  rescheduleM.mutate({
                    id: appt.id,
                    start_at: localDateTimeToUTC(newDate, newStart),
                    end_at: localDateTimeToUTC(newDate, endTime),
                  });
                }}
              />

              {/* Tu listado actual (lo dejás por ahora) */}
              <div className="divide-y">
                {agendaQ.data.length === 0 ? (
                  <p className="text-sm text-gray-600 py-2">No hay turnos.</p>
                ) : (
                  agendaQ.data.map((a) => (
                    <div key={a.id} className="py-3 flex items-start justify-between gap-4">
                      <div>
                        <div className="font-medium">
                          {formatLocalTime(a.start_at)} → {formatLocalTime(a.end_at)}
                        </div>
                        <div className="text-sm text-gray-600">Paciente: {a.patient_id}</div>
                        <div className="text-sm text-gray-600">Estado: {a.status}</div>
                        {a.notes && <div className="text-sm text-gray-600">Notas: {a.notes}</div>}
                      </div>

                      <div className="flex items-center gap-3">
                        <div className="text-xs text-gray-500 font-mono">{a.id}</div>

                        {a.status !== "cancelled" ? (
                          <button
                            className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-100 disabled:opacity-50"
                            disabled={cancelM.isPending || rescheduleM.isPending}
                            onClick={() => {
                              const reason = window.prompt("Motivo de cancelación (opcional):") ?? undefined;
                              cancelM.mutate({ id: a.id, reason });
                            }}
                          >
                            {cancelM.isPending ? "Cancelando…" : "Cancelar"}
                          </button>
                        ) : (
                          <span className="text-xs px-2 py-1 rounded-md bg-gray-100 text-gray-700">
                            Cancelado
                          </span>
                        )}

                        <button
                          className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-100 disabled:opacity-50"
                          disabled={rescheduleM.isPending || cancelM.isPending}
                          onClick={() => {
                            const newDate = window.prompt("Nueva fecha (YYYY-MM-DD):", date);
                            if (!newDate) return;

                            const newStart = window.prompt("Nueva hora inicio (HH:MM):", "09:00");
                            if (!newStart) return;

                            const dur = window.prompt("Duración en minutos:", "45");
                            if (!dur) return;

                            const durationMin = Number(dur);
                            if (!Number.isFinite(durationMin) || durationMin <= 0) return;

                            const endTime = addMinutesToHHmm(newStart, durationMin);

                            rescheduleM.mutate({
                              id: a.id,
                              start_at: localDateTimeToUTC(newDate, newStart),
                              end_at: localDateTimeToUTC(newDate, endTime),
                            });
                          }}
                        >
                          {rescheduleM.isPending ? "Reprogramando…" : "Reprogramar"}
                        </button>
                      </div>
                    </div>
                  ))
                )}

                {cancelM.isError && (
                  <p className="text-sm text-red-600">
                    Error al cancelar: {String((cancelM.error as any)?.message)}
                  </p>
                )}

                {rescheduleM.isError && (
                  <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700 mb-2">
                    {(rescheduleM.error as any)?.status === 409 ? (
                      <>
                        <div className="font-medium">Solapamiento al reprogramar</div>
                        <div>Ese horario se superpone con otro turno del kinesiólogo.</div>
                      </>
                    ) : (
                      <>
                        <div className="font-medium">Error al reprogramar</div>
                        <div>{String((rescheduleM.error as any)?.message)}</div>
                      </>
                    )}
                  </div>
                )}
              </div>
            </div>
          )}
        </section>
      </div>
    </div>
  );
}

