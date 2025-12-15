import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { listKinesiologists } from "../features/kinesiologists/api";
import { createAppointment, listAppointmentsDay } from "../features/appointments/api";

function todayISO() {
  const d = new Date();
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

export default function AgendaPage() {
  const [date, setDate] = useState(todayISO());
  const [kinesiologistId, setKinesiologistId] = useState("");
  const [patientId, setPatientId] = useState("");
  const [startAt, setStartAt] = useState("2025-12-15T14:00:00Z");
  const [endAt, setEndAt] = useState("2025-12-15T14:45:00Z");
  const [notes, setNotes] = useState("");

  useEffect(() => {
    const last = localStorage.getItem("last_patient_id");
    if (last && !patientId) setPatientId(last);
  }, [patientId]);

  const kinesioQ = useQuery({
    queryKey: ["kinesiologists"],
    queryFn: listKinesiologists,
  });

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

  const kinesios = useMemo(() => kinesioQ.data ?? [], [kinesioQ.data]);

  function create() {
    createM.mutate({
      patient_id: patientId.trim(),
      kinesiologist_id: kinesiologistId,
      start_at: startAt.trim(),
      end_at: endAt.trim(),
      notes: notes.trim() ? notes.trim() : undefined,
    });
  }

  const createError = (createM.error as any)?.message;
  const createStatus = (createM.error as any)?.status;

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

        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div>
              <label className="text-sm font-medium">Fecha</label>
              <input className="mt-1 w-full border rounded-lg p-2" type="date" value={date} onChange={(e) => setDate(e.target.value)} />
            </div>

            <div className="md:col-span-2">
              <label className="text-sm font-medium">Kinesiólogo</label>
              <select className="mt-1 w-full border rounded-lg p-2" value={kinesiologistId} onChange={(e) => setKinesiologistId(e.target.value)}>
                <option value="">Seleccionar…</option>
                {kinesios.map((k) => (
                  <option key={k.id} value={k.id}>
                    {k.last_name}, {k.first_name}
                  </option>
                ))}
              </select>
              {kinesioQ.isError && <p className="text-sm text-red-600 mt-1">Error: {String(kinesioQ.error?.message)}</p>}
            </div>
          </div>
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <h2 className="text-lg font-semibold">Crear turno</h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label className="text-sm font-medium">patient_id</label>
              <input className="mt-1 w-full border rounded-lg p-2" value={patientId} onChange={(e) => setPatientId(e.target.value)} />
              <p className="text-xs text-gray-500 mt-1">Tip: creá un paciente en /patients y se guarda automáticamente.</p>
            </div>
            <div>
              <label className="text-sm font-medium">Notas</label>
              <input className="mt-1 w-full border rounded-lg p-2" value={notes} onChange={(e) => setNotes(e.target.value)} />
            </div>
            <div>
              <label className="text-sm font-medium">start_at (RFC3339)</label>
              <input className="mt-1 w-full border rounded-lg p-2" value={startAt} onChange={(e) => setStartAt(e.target.value)} />
            </div>
            <div>
              <label className="text-sm font-medium">end_at (RFC3339)</label>
              <input className="mt-1 w-full border rounded-lg p-2" value={endAt} onChange={(e) => setEndAt(e.target.value)} />
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
            <p className="text-sm text-red-600">
              Error: {createError}
              {createStatus === 409 ? " (solapamiento)" : ""}
            </p>
          )}
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <h2 className="text-lg font-semibold">Turnos del día</h2>

          {!canLoadAgenda && <p className="text-sm text-gray-600">Seleccioná un kinesiólogo.</p>}

          {agendaQ.isLoading && <p className="text-sm text-gray-600">Cargando…</p>}
          {agendaQ.isError && <p className="text-sm text-red-600">Error: {String(agendaQ.error?.message)}</p>}

          {agendaQ.data && (
            <div className="divide-y">
              {agendaQ.data.length === 0 ? (
                <p className="text-sm text-gray-600 py-2">No hay turnos.</p>
              ) : (
                agendaQ.data.map((a) => (
                  <div key={a.id} className="py-3 flex items-start justify-between gap-4">
                    <div>
                      <div className="font-medium">{a.start_at} → {a.end_at}</div>
                      <div className="text-sm text-gray-600">Paciente: {a.patient_id}</div>
                      <div className="text-sm text-gray-600">Estado: {a.status}</div>
                      {a.notes && <div className="text-sm text-gray-600">Notas: {a.notes}</div>}
                    </div>
                    <div className="text-xs text-gray-500 font-mono">{a.id}</div>
                  </div>
                ))
              )}
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
