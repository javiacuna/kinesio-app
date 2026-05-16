import { useMemo } from "react";
import type { Appointment } from "../types";
import { generateSlots } from "@/shared/time/slots";
import { hhmmToMinutes, toLocalHHmm } from "@/shared/time/local";
import { formatLocalTime } from "@/shared/time/format";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

type Props = {
  date: string; // YYYY-MM-DD (solo para mostrar / precarga)
  appointments: Appointment[];
  canManageAppointments?: boolean;
  workStartTime?: string;
  workEndTime?: string;
  getPatientName?: (patientId: string) => string;

  onPickSlot: (hhmm: string) => void;

  onCancel: (appt: Appointment) => void;
  onReschedule: (appt: Appointment) => void;
};

export function AgendaGrid({
  date,
  appointments,
  canManageAppointments = true,
  workStartTime = "08:00",
  workEndTime = "20:00",
  onPickSlot,
  onCancel,
  onReschedule,
  getPatientName = (patientId) => patientId,
}: Props) {
  const { t } = useLanguage();
  const slots = useMemo(
    () => generateSlots(workStartTime || "08:00", workEndTime || "20:00", 15),
    [workEndTime, workStartTime],
  );

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
    <section className="border rounded-lg p-3 space-y-3">
      <div className="flex items-baseline justify-between">
        <h2 className="text-lg font-semibold">{t("agenda.grid")}</h2>
        <div className="text-sm text-gray-600">
          {date} · {workStartTime} a {workEndTime}
        </div>
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
                <div className="text-sm text-gray-700">{t("agenda.occupied")}</div>
              </div>
            );
          }

          // Turno que empieza en este slot
          if (appt) {
            const cancelled = appt.status === "cancelled";
            const completed = appt.status === "completed";
            return (
              <div
                key={hhmm}
                className={`flex items-center justify-between border rounded-lg p-2 ${
                  cancelled ? "bg-gray-50" : completed ? "bg-green-50" : "bg-blue-50"
                }`}
              >
                <div className="flex flex-col">
                  <div className="text-sm font-mono">
                    {formatLocalTime(appt.start_at)} → {formatLocalTime(appt.end_at)}
                  </div>
                  <div className="text-sm text-gray-800">
                    {t("agenda.patient")}: {getPatientName(appt.patient_id)}
                  </div>
                  {appt.notes && <div className="text-xs text-gray-600">{t("agenda.notes")}: {appt.notes}</div>}
                  {cancelled && appt.cancelled_reason && (
                    <div className="text-xs text-gray-600">{t("agenda.cancelReason")}: {appt.cancelled_reason}</div>
                  )}
                  {completed && (
                    <div className="text-xs text-green-700">Movimiento financiero generado.</div>
                  )}
                </div>

                <div className="flex items-center gap-2">
                  {cancelled ? (
                    <span className="text-xs px-2 py-1 rounded-md bg-gray-100 text-gray-700">
                      {t("agenda.cancelled")}
                    </span>
                  ) : completed ? (
                    <span className="text-xs px-2 py-1 rounded-md bg-green-100 text-green-700">
                      Realizado
                    </span>
                  ) : canManageAppointments ? (
                    <>
                      <button
                        type="button"
                        className="px-3 py-1 rounded-lg border text-sm hover:bg-white"
                        onClick={() => onReschedule(appt)}
                      >
                        {t("agenda.reschedule")}
                      </button>
                      <button
                        type="button"
                        className="px-3 py-1 rounded-lg border text-sm hover:bg-white"
                        onClick={() => onCancel(appt)}
                      >
                        {t("agenda.cancel")}
                      </button>
                    </>
                  ) : (
                    <span className="text-xs px-2 py-1 rounded-md bg-blue-100 text-blue-700">
                      {t("agenda.scheduled")}
                    </span>
                  )}
                </div>
              </div>
            );
          }

          // Slot libre
          return (
            <div key={hhmm} className="flex items-center justify-between border rounded-lg p-2">
              <div className="text-sm font-mono">{hhmm}</div>
              {canManageAppointments ? (
                <button
                  type="button"
                  className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                  onClick={() => onPickSlot(hhmm)}
                >
                  {t("agenda.create")}
                </button>
              ) : (
                <span className="text-sm text-gray-600">{t("agenda.free")}</span>
              )}
            </div>
          );
        })}
      </div>
    </section>
  );
}
