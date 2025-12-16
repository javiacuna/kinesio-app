import { useMemo } from "react";
import type { Appointment } from "../types";
import { generateSlots } from "@/shared/time/slots";
import { hhmmToMinutes, toLocalHHmm } from "@/shared/time/local";
import { formatLocalTime } from "@/shared/time/format";

type Props = {
  date: string; // YYYY-MM-DD (solo para mostrar / precarga)
  appointments: Appointment[];

  onPickSlot: (hhmm: string) => void;

  onCancel: (appt: Appointment) => void;
  onReschedule: (appt: Appointment) => void;
};

export function AgendaGrid({ date, appointments, onPickSlot, onCancel, onReschedule }: Props) {
  const slots = useMemo(() => generateSlots("08:00", "20:00", 15), []);

  // Indexar turnos por inicio local HH:MM (para encontrar rápido)
  const byStart = useMemo(() => {
    const m = new Map<string, Appointment>();
    for (const a of appointments) {
      m.set(toLocalHHmm(a.start_at), a);
    }
    return m;
  }, [appointments]);

  // Para marcar ocupación de slots intermedios (si hay turnos de 45/60 etc.)
  const occupied = useMemo(() => {
    const set = new Set<string>();
    for (const a of appointments) {
      if (a.status === "cancelled") continue;

      const start = hhmmToMinutes(toLocalHHmm(a.start_at));
      const end = hhmmToMinutes(toLocalHHmm(a.end_at));

      for (let t = start; t < end; t += 15) {
        const hhmm = `${String(Math.floor(t / 60)).padStart(2, "0")}:${String(t % 60).padStart(2, "0")}`;
        set.add(hhmm);
      }
    }
    return set;
  }, [appointments]);

  return (
    <section className="bg-white rounded-xl shadow p-4 space-y-3">
      <div className="flex items-baseline justify-between">
        <h2 className="text-lg font-semibold">Grilla horaria</h2>
        <div className="text-sm text-gray-600">{date}</div>
      </div>

      <div className="grid grid-cols-1 gap-2">
        {slots.map((hhmm) => {
          const appt = byStart.get(hhmm);
          const isOcc = occupied.has(hhmm);

          // Slot ocupado por un turno que no empieza justo en este slot:
          // lo mostramos como bloque “ocupado” sin acciones.
          if (!appt && isOcc) {
            return (
              <div
                key={hhmm}
                className="flex items-center justify-between border rounded-lg p-2 bg-gray-50"
              >
                <div className="text-sm font-mono text-gray-700">{hhmm}</div>
                <div className="text-sm text-gray-700">Ocupado</div>
              </div>
            );
          }

          // Turno que empieza en este slot
          if (appt) {
            const cancelled = appt.status === "cancelled";
            return (
              <div
                key={hhmm}
                className={`flex items-center justify-between border rounded-lg p-2 ${
                  cancelled ? "bg-gray-50" : "bg-blue-50"
                }`}
              >
                <div className="flex flex-col">
                  <div className="text-sm font-mono">
                    {formatLocalTime(appt.start_at)} → {formatLocalTime(appt.end_at)}
                  </div>
                  <div className="text-sm text-gray-800">
                    Paciente: <span className="font-mono">{appt.patient_id}</span>
                  </div>
                  {appt.notes && <div className="text-xs text-gray-600">Notas: {appt.notes}</div>}
                </div>

                <div className="flex items-center gap-2">
                  {cancelled ? (
                    <span className="text-xs px-2 py-1 rounded-md bg-gray-100 text-gray-700">
                      Cancelado
                    </span>
                  ) : (
                    <>
                      <button
                        type="button"
                        className="px-3 py-1 rounded-lg border text-sm hover:bg-white"
                        onClick={() => onReschedule(appt)}
                      >
                        Reprogramar
                      </button>
                      <button
                        type="button"
                        className="px-3 py-1 rounded-lg border text-sm hover:bg-white"
                        onClick={() => onCancel(appt)}
                      >
                        Cancelar
                      </button>
                    </>
                  )}
                </div>
              </div>
            );
          }

          // Slot libre
          return (
            <div key={hhmm} className="flex items-center justify-between border rounded-lg p-2">
              <div className="text-sm font-mono">{hhmm}</div>
              <button
                type="button"
                className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                onClick={() => onPickSlot(hhmm)}
              >
                Crear acá
              </button>
            </div>
          );
        })}
      </div>
    </section>
  );
}
