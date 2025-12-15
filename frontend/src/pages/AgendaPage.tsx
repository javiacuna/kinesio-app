import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { listKinesiologists } from "@/features/kinesiologists/api";
import { listAppointmentsDay } from "@/features/appointments/api";

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

  const kinesioQ = useQuery({
    queryKey: ["kinesiologists"],
    queryFn: listKinesiologists,
  });

  const agendaQ = useQuery({
    queryKey: ["appointments", date, kinesiologistId],
    queryFn: () =>
      listAppointmentsDay({
        date,
        kinesiologist_id: kinesiologistId,
      }),
    enabled: Boolean(kinesiologistId),
  });

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto p-6 space-y-6">
        <h1 className="text-2xl font-semibold">Agenda</h1>

        <div className="bg-white p-4 rounded-xl shadow space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="text-sm font-medium">Fecha</label>
              <input
                type="date"
                className="mt-1 w-full border rounded-lg p-2"
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
                {kinesioQ.data?.map((k) => (
                  <option key={k.id} value={k.id}>
                    {k.last_name}, {k.first_name}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        <div className="bg-white p-4 rounded-xl shadow">
          <h2 className="text-lg font-semibold mb-3">Turnos del día</h2>

          {agendaQ.isLoading && <p>Cargando…</p>}
          {agendaQ.isError && (
            <p className="text-red-600">
              Error: {String(agendaQ.error?.message)}
            </p>
          )}

          {agendaQ.data && (
            <div className="divide-y">
              {agendaQ.data.length === 0 ? (
                <p className="text-gray-500 py-2">No hay turnos.</p>
              ) : (
                agendaQ.data.map((a) => (
                  <div key={a.id} className="py-3 flex justify-between">
                    <div>
                      <div className="font-medium">
                        {a.start_at} → {a.end_at}
                      </div>
                      <div className="text-sm text-gray-600">
                        Paciente: {a.patient_id}
                      </div>
                    </div>
                    <div className="text-sm">{a.status}</div>
                  </div>
                ))
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
