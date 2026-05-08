import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { useAuth } from "@/features/auth/AuthProvider";
import { listKinesiologists } from "../features/kinesiologists/api";
import {
  cancelAppointment,
  createAppointment,
  listAppointmentsDay,
  updateAppointment,
} from "../features/appointments/api";
import { listPatients } from "@/features/patients/api";
import { addMinutesToHHmm, localDateTimeToUTC } from "../shared/time/rfc3339";
import { formatLocalTime } from "../shared/time/format";
import { PatientSearch } from "@/features/patients/components/PatientSearch";
import { AgendaGrid } from "@/features/appointments/components/AgendaGrid";
import type { Appointment } from "@/features/appointments/types";
import type { Patient } from "@/features/patients/types";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

function todayISO() {
  const d = new Date();
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function isOverlapError(error: unknown) {
  return (error as any)?.status === 409 || (error as any)?.message === "overlap";
}

function isInactivePatientError(error: unknown) {
  return (error as any)?.message === "patient_inactive";
}

function isPastLocalDateTime(dateISO: string, timeHHmm: string) {
  if (!dateISO || !timeHHmm) return false;
  const [y, m, d] = dateISO.split("-").map(Number);
  const [hh, mm] = timeHHmm.split(":").map(Number);
  return new Date(y, m - 1, d, hh, mm, 0, 0).getTime() <= Date.now();
}

function validationDetail(error: unknown, field: string) {
  return (error as any)?.body?.details?.[field] as string | undefined;
}

function isOutsideWorkingHours(startHHmm: string, durationMin: number, workStart?: string, workEnd?: string) {
  const start = startHHmm;
  const end = addMinutesToHHmm(startHHmm, durationMin);
  const from = workStart || "08:00";
  const to = workEnd || "20:00";
  return start < from || end > to || end <= start;
}

function localDateISO(iso: string) {
  const d = new Date(iso);
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function durationMinutes(startISO: string, endISO: string) {
  return Math.max(15, Math.round((new Date(endISO).getTime() - new Date(startISO).getTime()) / 60000));
}

function addDaysISO(dateISO: string, days: number) {
  const [year, month, day] = dateISO.split("-").map(Number);
  const date = new Date(year, month - 1, day);
  date.setDate(date.getDate() + days);
  const yyyy = date.getFullYear();
  const mm = String(date.getMonth() + 1).padStart(2, "0");
  const dd = String(date.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function weekDaysFor(dateISO: string) {
  const [year, month, day] = dateISO.split("-").map(Number);
  const date = new Date(year, month - 1, day);
  const mondayOffset = (date.getDay() + 6) % 7;
  date.setDate(date.getDate() - mondayOffset);
  const start = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
  return Array.from({ length: 7 }, (_, index) => addDaysISO(start, index));
}

function statusLabel(status: Appointment["status"], t: (key: string) => string) {
  return status === "cancelled" ? t("agenda.cancelled") : t("agenda.scheduled");
}

function patientName(patient?: Patient) {
  return patient ? `${patient.last_name}, ${patient.first_name}` : "Paciente";
}

function filterAppointments(
  appointments: Appointment[],
  status: "all" | Appointment["status"],
  patientId?: string,
) {
  return appointments.filter((appointment) => {
    if (status !== "all" && appointment.status !== status) return false;
    if (patientId && appointment.patient_id !== patientId) return false;
    return true;
  });
}

export default function AgendaPage() {
  const { user } = useAuth();
  const { t } = useLanguage();
  const isKinesiologist = user?.role === "kinesiologo";
  const canManageAppointments = user?.role === "admin" || user?.role === "recepcionista";
  // Agenda (listado)
  const [date, setDate] = useState(todayISO());
  const [kinesiologistId, setKinesiologistId] = useState("");
  const [viewMode, setViewMode] = useState<"day" | "week">("day");
  const [statusFilter, setStatusFilter] = useState<"all" | Appointment["status"]>("all");
  const [filterPatient, setFilterPatient] = useState<Patient | null>(null);

  // Crear turno (inputs UX)
  const [patientId, setPatientId] = useState("");
  const [apptDate, setApptDate] = useState(todayISO());
  const [startTime, setStartTime] = useState("09:00");
  const [durationMin, setDurationMin] = useState(45);
  const [notes, setNotes] = useState("");
  const [editingAppointment, setEditingAppointment] = useState<Appointment | null>(null);
  const [editDate, setEditDate] = useState(todayISO());
  const [editStartTime, setEditStartTime] = useState("09:00");
  const [editDurationMin, setEditDurationMin] = useState(45);
  const [editNotes, setEditNotes] = useState("");
  const [cancelAppointmentTarget, setCancelAppointmentTarget] = useState<Appointment | null>(null);
  const [cancelReason, setCancelReason] = useState("");

  useEffect(() => {
    const last = localStorage.getItem("last_patient_id");
    if (last && !patientId) setPatientId(last);
  }, [patientId]);

  useEffect(() => {
    setApptDate(date);
  }, [date]);

  const kinesioQ = useQuery({
    queryKey: ["kinesiologists"],
    queryFn: () => listKinesiologists(),
  });

  const patientsQ = useQuery({
    queryKey: ["patients", "agenda", "all"],
    queryFn: () => listPatients(500, true),
  });

  const kinesios = useMemo(() => kinesioQ.data ?? [], [kinesioQ.data]);
  const patients = patientsQ.data ?? [];
  const patientById = useMemo(
    () => new Map(patients.map((patient) => [patient.id, patient])),
    [patients],
  );
  const selectedKinesiologist = useMemo(
    () => kinesios.find((k) => k.id === kinesiologistId),
    [kinesiologistId, kinesios],
  );
  const weekDays = useMemo(() => weekDaysFor(date), [date]);
  const weekStart = weekDays[0] ?? date;
  const weekEnd = weekDays[6] ?? date;

  useEffect(() => {
    if (!isKinesiologist || !user?.email || kinesios.length === 0) return;

    const ownProfile =
      kinesios.find((k) => k.email.toLowerCase() === user.email.toLowerCase()) ?? kinesios[0];

    if (ownProfile && ownProfile.id !== kinesiologistId) {
      setKinesiologistId(ownProfile.id);
    }
  }, [isKinesiologist, kinesiologistId, kinesios, user?.email]);

  const canLoadAgenda = Boolean(kinesiologistId);

  const agendaQ = useQuery({
    queryKey: ["appointments", "day", date, kinesiologistId],
    queryFn: () => listAppointmentsDay({ date, kinesiologist_id: kinesiologistId }),
    enabled: canLoadAgenda,
  });

  const weekAgendaQ = useQuery({
    queryKey: ["appointments", "week", weekStart, kinesiologistId],
    queryFn: async () => {
      const results = await Promise.all(
        weekDays.map((day) => listAppointmentsDay({ date: day, kinesiologist_id: kinesiologistId })),
      );
      return weekDays.map((day, index) => ({ day, appointments: results[index] ?? [] }));
    },
    enabled: canLoadAgenda && viewMode === "week",
  });

  const createM = useMutation({
    mutationFn: createAppointment,
    onSuccess: () => agendaQ.refetch(),
    onError: (error) => {
      if (isInactivePatientError(error)) {
        localStorage.removeItem("last_patient_id");
        setPatientId("");
      }
    },
  });

  const cancelM = useMutation({
    mutationFn: cancelAppointment,
    onSuccess: () => {
      setCancelAppointmentTarget(null);
      setCancelReason("");
      agendaQ.refetch();
      weekAgendaQ.refetch();
    },
  });

  const rescheduleM = useMutation({
    mutationFn: (args: { id: string; start_at: string; end_at: string; notes?: string }) =>
      updateAppointment({ id: args.id, start_at: args.start_at, end_at: args.end_at, notes: args.notes }),
    onSuccess: () => {
      setEditingAppointment(null);
      agendaQ.refetch();
      weekAgendaQ.refetch();
    },
  });


  function create() {
    if (isCreateInPast) return;

    const endTime = addMinutesToHHmm(startTime, durationMin);

    createM.mutate({
      patient_id: patientId.trim(),
      kinesiologist_id: kinesiologistId,
      start_at: localDateTimeToUTC(apptDate, startTime),
      end_at: localDateTimeToUTC(apptDate, endTime),
      notes: notes.trim() ? notes.trim() : undefined,
    });
  }

  function openEditAppointment(appt: Appointment) {
    setEditingAppointment(appt);
    setEditDate(localDateISO(appt.start_at));
    setEditStartTime(formatLocalTime(appt.start_at));
    setEditDurationMin(durationMinutes(appt.start_at, appt.end_at));
    setEditNotes(appt.notes ?? "");
    rescheduleM.reset();
  }

  function submitEditAppointment() {
    if (!editingAppointment || isEditInPast || isEditOutsideWorkingHours) return;

    const endTime = addMinutesToHHmm(editStartTime, editDurationMin);
    rescheduleM.mutate({
      id: editingAppointment.id,
      start_at: localDateTimeToUTC(editDate, editStartTime),
      end_at: localDateTimeToUTC(editDate, endTime),
      notes: editNotes.trim() ? editNotes.trim() : undefined,
    });
  }

  function openCancelAppointment(appt: Appointment) {
    setCancelAppointmentTarget(appt);
    setCancelReason("");
    cancelM.reset();
  }

  function submitCancelAppointment() {
    if (!cancelAppointmentTarget) return;
    cancelM.mutate({
      id: cancelAppointmentTarget.id,
      reason: cancelReason.trim() || undefined,
    });
  }

  const createErr: any = createM.error;
  const hasCreateOverlap = isOverlapError(createM.error);
  const hasCreateInactivePatient = isInactivePatientError(createM.error);
  const hasRescheduleOverlap = isOverlapError(rescheduleM.error);
  const isCreateInPast = isPastLocalDateTime(apptDate, startTime);
  const isCreateOutsideWorkingHours = selectedKinesiologist
    ? isOutsideWorkingHours(
        startTime,
        durationMin,
        selectedKinesiologist.work_start_time,
        selectedKinesiologist.work_end_time,
      )
    : false;
  const createValidationMessage = validationDetail(createM.error, "start_at");
  const rescheduleValidationMessage = validationDetail(rescheduleM.error, "start_at");
  const isEditInPast = isPastLocalDateTime(editDate, editStartTime);
  const isEditOutsideWorkingHours = selectedKinesiologist
    ? isOutsideWorkingHours(
        editStartTime,
        editDurationMin,
        selectedKinesiologist.work_start_time,
        selectedKinesiologist.work_end_time,
      )
    : false;
  const dayAppointments = useMemo(
    () => filterAppointments(agendaQ.data ?? [], statusFilter, filterPatient?.id),
    [agendaQ.data, filterPatient?.id, statusFilter],
  );
  const weekAppointments = useMemo(
    () =>
      (weekAgendaQ.data ?? []).map((group) => ({
        ...group,
        appointments: filterAppointments(group.appointments, statusFilter, filterPatient?.id),
      })),
    [filterPatient?.id, statusFilter, weekAgendaQ.data],
  );

  return (
    <main>
      <div className="max-w-4xl mx-auto p-6 space-y-6">
        <header className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold">{t("agenda.title")}</h1>
            <p className="text-sm text-gray-600">
              {isKinesiologist ? t("agenda.myDay") : t("agenda.dailySubtitle")}
            </p>
          </div>
          {(user?.role === "admin" || user?.role === "recepcionista") && (
            <Link className="text-sm underline" to="/patients">
              {t("agenda.goPatients")}
            </Link>
          )}
        </header>

        {/* Filtros agenda */}
        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
            <div>
              <label className="text-sm font-medium">{t("agenda.date")}</label>
              <input
                className="mt-1 w-full border rounded-lg p-2"
                type="date"
                value={date}
                onChange={(e) => setDate(e.target.value)}
              />
            </div>

            <div>
              <label className="text-sm font-medium">{t("agenda.view")}</label>
              <select
                className="mt-1 w-full border rounded-lg p-2"
                value={viewMode}
                onChange={(event) => setViewMode(event.target.value as "day" | "week")}
              >
                <option value="day">{t("agenda.day")}</option>
                <option value="week">{t("agenda.week")}</option>
              </select>
            </div>

            <div className="md:col-span-2">
              <label className="text-sm font-medium">{t("agenda.kinesiologist")}</label>
              {isKinesiologist ? (
                <div className="mt-1 w-full border rounded-lg p-2 bg-gray-50 text-gray-700">
                  {selectedKinesiologist
                    ? `${selectedKinesiologist.last_name}, ${selectedKinesiologist.first_name}`
                    : kinesioQ.isLoading
                      ? t("dashboard.loadingAppointments")
                      : t("agenda.noOwnKinesiologistProfile")}
                </div>
              ) : (
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={kinesiologistId}
                  onChange={(e) => setKinesiologistId(e.target.value)}
                >
                  <option value="">{t("agenda.select")}</option>
                  {kinesios.map((k) => (
                    <option key={k.id} value={k.id}>
                      {k.last_name}, {k.first_name}
                    </option>
                  ))}
                </select>
              )}

              {kinesioQ.isError && (
                <p className="text-sm text-red-600 mt-1">
                  Error: {String(kinesioQ.error?.message)}
                </p>
              )}
              {selectedKinesiologist && (
                <p className="text-sm text-gray-600 mt-2">
                  {selectedKinesiologist.work_start_time} - {selectedKinesiologist.work_end_time}
                </p>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-[180px_1fr_auto] gap-3 items-end">
            <div>
              <label className="text-sm font-medium">{t("agenda.status")}</label>
              <select
                className="mt-1 w-full border rounded-lg p-2"
                value={statusFilter}
                onChange={(event) => setStatusFilter(event.target.value as "all" | Appointment["status"])}
              >
                <option value="all">{t("agenda.statusAll")}</option>
                <option value="scheduled">{t("agenda.statusScheduled")}</option>
                <option value="cancelled">{t("agenda.statusCancelled")}</option>
              </select>
            </div>

            <div>
              <PatientSearch
                valuePatientId={filterPatient?.id ?? ""}
                onSelect={setFilterPatient}
                placeholder={t("agenda.filterPatient")}
              />
            </div>

            <button
              type="button"
              className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50"
              onClick={() => {
                setStatusFilter("all");
                setFilterPatient(null);
              }}
            >
              {t("agenda.clearFilters")}
            </button>
          </div>
        </section>

        {/* Crear turno */}
        {canManageAppointments && (
        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <h2 className="text-lg font-semibold">{t("agenda.create")}</h2>

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
                {t("agenda.tipPatient")}
              </p>
            </div>

            <div>
              <label className="text-sm font-medium">{t("agenda.appointmentDate")}</label>
              <input
                className="mt-1 w-full border rounded-lg p-2"
                type="date"
                min={todayISO()}
                value={apptDate}
                onChange={(e) => setApptDate(e.target.value)}
              />
            </div>

            <div>
              <label className="text-sm font-medium">{t("agenda.startTime")}</label>
              <input
                className="mt-1 w-full border rounded-lg p-2"
                type="time"
                value={startTime}
                onChange={(e) => setStartTime(e.target.value)}
              />
            </div>

            <div>
              <label className="text-sm font-medium">{t("agenda.duration")}</label>
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
              <label className="text-sm font-medium">{t("agenda.notes")}</label>
              <input
                className="mt-1 w-full border rounded-lg p-2"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
              />
            </div>
          </div>

          <button
            className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
            disabled={!patientId.trim() || !kinesiologistId || isCreateInPast || isCreateOutsideWorkingHours || createM.isPending}
            onClick={create}
          >
            {createM.isPending ? t("agenda.creating") : t("agenda.create")}
          </button>

          {isCreateInPast && (
            <p className="text-sm text-red-600">
              {t("agenda.appointmentPast")}
            </p>
          )}

          {isCreateOutsideWorkingHours && selectedKinesiologist && (
            <p className="text-sm text-red-600">
              {t("agenda.appointmentOutOfHours")} {selectedKinesiologist.work_start_time} - {selectedKinesiologist.work_end_time}.
            </p>
          )}

          {createM.isError && (
            <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700">
              {hasCreateOverlap ? (
                <>
                  <div className="font-medium">Overlap detected</div>
                  <div>The kinesiologist already has an active appointment at that time.</div>
                </>
              ) : hasCreateInactivePatient ? (
                <>
                  <div className="font-medium">{t("agenda.patientArchived")}</div>
                  <div>{t("agenda.patientArchivedDetail")}</div>
                </>
              ) : (
                <>
                  <div className="font-medium">{t("agenda.error")}</div>
                  <div>{createValidationMessage ?? String(createErr?.message)}</div>
                </>
              )}
            </div>
          )}
        </section>
        )}

        {/* Agenda */}
        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <div className="flex items-center justify-between gap-3">
            <h2 className="text-lg font-semibold">
              {viewMode === "week" ? t("agenda.weekAppointments") : t("agenda.todayAppointments")}
            </h2>
            <div className="text-sm text-gray-600">
              {viewMode === "week" ? `${weekStart} a ${weekEnd}` : date}
            </div>
          </div>

          {!canLoadAgenda && (
            <p className="text-sm text-gray-600">
              {isKinesiologist
                ? t("agenda.noKinesiologistProfile")
                : t("agenda.kinesiologistMissing")}
            </p>
          )}

          {agendaQ.isLoading && <p className="text-sm text-gray-600">{t("dashboard.loadingAppointments")}</p>}
          {agendaQ.isError && <p className="text-sm text-red-600">{t("agenda.error")}: {String(agendaQ.error?.message)}</p>}

          {viewMode === "day" && agendaQ.data && (
            <div className="space-y-4">
              <AgendaGrid
                date={date}
                appointments={dayAppointments}
                canManageAppointments={canManageAppointments}
                workStartTime={selectedKinesiologist?.work_start_time}
                workEndTime={selectedKinesiologist?.work_end_time}
                getPatientName={(id) => patientName(patientById.get(id))}
                onPickSlot={(hhmm) => {
                  setApptDate(date);
                  setStartTime(hhmm);
                  setDurationMin(45);
                  // opcional: llevar al formulario
                  window.scrollTo({ top: 0, behavior: "smooth" });
                }}
                onCancel={(appt) => {
                  openCancelAppointment(appt);
                }}
                onReschedule={(appt) => {
                  openEditAppointment(appt);
                }}
              />

              <div className="divide-y">
                {dayAppointments.length === 0 ? (
                  <p className="text-sm text-gray-600 py-2">{t("agenda.noAppointments")}</p>
                ) : (
                  dayAppointments.map((a) => (
                    <div key={a.id} className="py-3 flex items-start justify-between gap-4">
                      <div>
                        <div className="font-medium">
                          {formatLocalTime(a.start_at)} → {formatLocalTime(a.end_at)}
                        </div>
                        <div className="text-sm text-gray-600">
                          {t("agenda.patient")}:{" "}
                          <Link className="underline" to={`/patients/${a.patient_id}`}>
                            {patientName(patientById.get(a.patient_id))}
                          </Link>
                        </div>
                        <div className="text-sm text-gray-600">{t("agenda.status")}: {statusLabel(a.status, t)}</div>
                        {a.status === "cancelled" && a.cancelled_reason && (
                          <div className="text-sm text-gray-600">{t("agenda.cancelReason")}: {a.cancelled_reason}</div>
                        )}
                        {a.notes && <div className="text-sm text-gray-600">{t("agenda.notes")}: {a.notes}</div>}
                      </div>

                      <div className="flex items-center gap-3">
                        {canManageAppointments && (
                          <div className="text-xs text-gray-500 font-mono">{a.id}</div>
                        )}

                        {canManageAppointments && a.status !== "cancelled" ? (
                          <button
                            className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-100 disabled:opacity-50"
                            disabled={cancelM.isPending || rescheduleM.isPending}
                            onClick={() => {
                              openCancelAppointment(a);
                            }}
                          >
                            {cancelM.isPending ? "..." : t("agenda.cancel")}
                          </button>
                        ) : a.status === "cancelled" ? (
                          <span className="text-xs px-2 py-1 rounded-md bg-gray-100 text-gray-700">
                            {t("agenda.cancelled")}
                          </span>
                        ) : (
                          <span className="text-xs px-2 py-1 rounded-md bg-blue-100 text-blue-700">
                            {t("agenda.scheduled")}
                          </span>
                        )}

                        {canManageAppointments && (
                          <button
                            className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-100 disabled:opacity-50"
                            disabled={rescheduleM.isPending || cancelM.isPending}
                            onClick={() => {
                              openEditAppointment(a);
                            }}
                          >
                            {rescheduleM.isPending ? "..." : t("agenda.reschedule")}
                          </button>
                        )}
                      </div>
                    </div>
                  ))
                )}

                {cancelM.isError && (
                  <p className="text-sm text-red-600">
                    {t("agenda.cancelError")}: {String((cancelM.error as any)?.message)}
                  </p>
                )}

                {rescheduleM.isError && (
                  <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700 mb-2">
                    {hasRescheduleOverlap ? (
                      <>
                        <div className="font-medium">Solapamiento al reprogramar</div>
                        <div>El kinesiólogo ya tiene un turno activo en ese horario.</div>
                      </>
                    ) : (
                      <>
                        <div className="font-medium">Error al reprogramar</div>
                        <div>{rescheduleValidationMessage ?? String((rescheduleM.error as any)?.message)}</div>
                      </>
                    )}
                  </div>
                )}
              </div>
            </div>
          )}

          {viewMode === "week" && (
            <div className="space-y-3">
              {weekAgendaQ.isLoading && <p className="text-sm text-gray-600">{t("dashboard.loadingAppointments")}</p>}
              {weekAgendaQ.isError && <p className="text-sm text-red-600">{t("agenda.error")}: {String(weekAgendaQ.error?.message)}</p>}
              {weekAppointments.map((group) => (
                <section key={group.day} className="border rounded-lg p-3">
                  <div className="font-medium">{group.day}</div>
                  {group.appointments.length === 0 ? (
                    <p className="text-sm text-gray-600 mt-2">{t("agenda.noAppointments")}</p>
                  ) : (
                    <div className="divide-y mt-2">
                      {group.appointments.map((appointment) => (
                        <div key={appointment.id} className="py-2 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
                          <div>
                            <div className="font-medium">
                              {formatLocalTime(appointment.start_at)} → {formatLocalTime(appointment.end_at)}
                            </div>
                            <div className="text-sm text-gray-600">
                              {patientName(patientById.get(appointment.patient_id))} · {statusLabel(appointment.status, t)}
                            </div>
                            {appointment.status === "cancelled" && appointment.cancelled_reason && (
                              <div className="text-sm text-gray-600">{t("agenda.cancelReason")}: {appointment.cancelled_reason}</div>
                            )}
                            {appointment.notes && <div className="text-sm text-gray-600">{t("agenda.notes")}: {appointment.notes}</div>}
                          </div>
                          {canManageAppointments && appointment.status !== "cancelled" && (
                            <div className="flex flex-wrap gap-2">
                              <button
                                type="button"
                                className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                                onClick={() => openEditAppointment(appointment)}
                              >
                                {t("agenda.reschedule")}
                              </button>
                              <button
                                type="button"
                                className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                                onClick={() => openCancelAppointment(appointment)}
                              >
                                {t("agenda.cancel")}
                              </button>
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  )}
                </section>
              ))}
            </div>
          )}
        </section>
      </div>

      {editingAppointment && (
        <div className="fixed inset-0 z-50 bg-black/40 flex items-center justify-center p-4">
          <section className="bg-white rounded-xl shadow-xl p-4 w-full max-w-xl space-y-4">
            <div className="flex items-start justify-between gap-4">
              <div>
                <h2 className="text-lg font-semibold">{t("agenda.reschedule")}</h2>
                <p className="text-sm text-gray-600">
                  {t("agenda.patient")}: <span className="font-mono">{editingAppointment.patient_id}</span>
                </p>
              </div>
              <button
                type="button"
                className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                onClick={() => setEditingAppointment(null)}
              >
                {t("agenda.close")}
              </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div>
                <label className="text-sm font-medium">{t("agenda.date")}</label>
                <input
                  className="mt-1 w-full border rounded-lg p-2"
                  type="date"
                  min={todayISO()}
                  value={editDate}
                  onChange={(event) => setEditDate(event.target.value)}
                />
              </div>

              <div>
                <label className="text-sm font-medium">{t("agenda.startTime")}</label>
                <input
                  className="mt-1 w-full border rounded-lg p-2"
                  type="time"
                  value={editStartTime}
                  onChange={(event) => setEditStartTime(event.target.value)}
                />
              </div>

              <div>
                <label className="text-sm font-medium">{t("agenda.duration")}</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={editDurationMin}
                  onChange={(event) => setEditDurationMin(Number(event.target.value))}
                >
                  <option value={30}>30 min</option>
                  <option value={45}>45 min</option>
                  <option value={60}>60 min</option>
                  <option value={90}>90 min</option>
                </select>
              </div>

              <div>
                <label className="text-sm font-medium">{t("agenda.notes")}</label>
                <input
                  className="mt-1 w-full border rounded-lg p-2"
                  value={editNotes}
                  onChange={(event) => setEditNotes(event.target.value)}
                />
              </div>
            </div>

            {selectedKinesiologist && (
              <p className="text-sm text-gray-600">
                {selectedKinesiologist.work_start_time} - {selectedKinesiologist.work_end_time}
              </p>
            )}

            {isEditInPast && (
              <p className="text-sm text-red-600">{t("agenda.appointmentPast")}</p>
            )}

            {isEditOutsideWorkingHours && selectedKinesiologist && (
              <p className="text-sm text-red-600">
                {t("agenda.appointmentOutOfHours")} {selectedKinesiologist.work_start_time} - {selectedKinesiologist.work_end_time}.
              </p>
            )}

            {rescheduleM.isError && (
              <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700">
                {hasRescheduleOverlap ? (
                  <>
                    <div className="font-medium">Solapamiento al reprogramar</div>
                    <div>El kinesiólogo ya tiene un turno activo en ese horario.</div>
                  </>
                ) : (
                  <>
                    <div className="font-medium">Error al modificar</div>
                    <div>{rescheduleValidationMessage ?? String((rescheduleM.error as any)?.message)}</div>
                  </>
                )}
              </div>
            )}

            <div className="flex justify-end gap-2">
              <button
                type="button"
                className="px-4 py-2 rounded-lg border hover:bg-gray-50"
                onClick={() => setEditingAppointment(null)}
              >
                {t("common.cancel")}
              </button>
              <button
                type="button"
                className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
                disabled={isEditInPast || isEditOutsideWorkingHours || rescheduleM.isPending}
                onClick={submitEditAppointment}
              >
                {rescheduleM.isPending ? t("common.saving") : t("common.saveChanges")}
              </button>
            </div>
          </section>
        </div>
      )}

      {cancelAppointmentTarget && (
        <div className="fixed inset-0 z-50 bg-black/40 flex items-center justify-center p-4">
          <section className="bg-white rounded-xl shadow-xl p-4 w-full max-w-lg space-y-4">
            <div className="flex items-start justify-between gap-4">
              <div>
                <h2 className="text-lg font-semibold">{t("agenda.cancelAppointment")}</h2>
                <p className="text-sm text-gray-600">
                  {formatLocalTime(cancelAppointmentTarget.start_at)} a {formatLocalTime(cancelAppointmentTarget.end_at)} ·{" "}
                  {patientName(patientById.get(cancelAppointmentTarget.patient_id))}
                </p>
              </div>
              <button
                type="button"
                className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                onClick={() => setCancelAppointmentTarget(null)}
              >
                {t("agenda.close")}
              </button>
            </div>

            <div>
              <label className="text-sm font-medium">{t("agenda.cancelReason")}</label>
              <textarea
                className="mt-1 w-full border rounded-lg p-2 min-h-24"
                value={cancelReason}
                onChange={(event) => setCancelReason(event.target.value)}
              />
            </div>

            {cancelM.isError && (
              <p className="text-sm text-red-600">
                {t("agenda.cancelError")}: {String((cancelM.error as any)?.message)}
              </p>
            )}

            <div className="flex justify-end gap-2">
              <button
                type="button"
                className="px-4 py-2 rounded-lg border hover:bg-gray-50"
                onClick={() => setCancelAppointmentTarget(null)}
              >
                {t("common.back")}
              </button>
              <button
                type="button"
                className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
                disabled={cancelM.isPending}
                onClick={submitCancelAppointment}
              >
                {cancelM.isPending ? "..." : t("agenda.confirmCancel")}
              </button>
            </div>
          </section>
        </div>
      )}
    </main>
  );
}
