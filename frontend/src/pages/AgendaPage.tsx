import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { useAuth } from "@/features/auth/AuthProvider";
import { listKinesiologists } from "../features/kinesiologists/api";
import {
  cancelAppointment,
  createAppointment,
  createAppointmentPackage,
  generateAppointmentVideoCall,
  listAppointmentsDay,
  updateAppointment,
  updateAppointmentPackage,
} from "../features/appointments/api";
import { listPatients } from "@/features/patients/api";
import { completeAppointment, listFinanciers } from "@/features/finance/api";
import { addMinutesToHHmm, localDateTimeToUTC } from "../shared/time/rfc3339";
import { formatLocalTime } from "../shared/time/format";
import { PatientSearch } from "@/features/patients/components/PatientSearch";
import { AgendaGrid } from "@/features/appointments/components/AgendaGrid";
import type { Appointment } from "@/features/appointments/types";
import type { Patient } from "@/features/patients/types";
import type { Kinesiologist } from "@/features/kinesiologists/types";
import { useLanguage } from "@/shared/i18n/LanguageProvider";
import { RequiredLabel } from "@/shared/ui/RequiredLabel";

const defaultWorkDays = [1, 2, 3, 4, 5];
const workDayLabelKeys: Record<number, string> = {
  1: "day.mon",
  2: "day.tue",
  3: "day.wed",
  4: "day.thu",
  5: "day.fri",
  6: "day.sat",
  7: "day.sun",
};

function todayISO() {
  const d = new Date();
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function currentRoundedTime() {
  const d = new Date();
  const roundedMinutes = Math.ceil(d.getMinutes() / 15) * 15;
  d.setMinutes(roundedMinutes, 0, 0);
  const hh = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");
  return `${hh}:${mm}`;
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

function workDayLabel(day: number, t: (key: string) => string) {
  const key = workDayLabelKeys[day];
  return key ? t(key) : String(day);
}

function workDaysLabel(days: number[] | undefined, t: (key: string) => string) {
  const selected = days?.length ? days : defaultWorkDays;
  return selected.map((day) => workDayLabel(day, t)).filter(Boolean).join(", ");
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
  if (status === "cancelled") return t("agenda.cancelled");
  if (status === "completed") return t("agenda.appointmentDone");
  return t("agenda.scheduled");
}

function patientName(patient: Patient | undefined, t: (key: string) => string) {
  return patient ? `${patient.last_name}, ${patient.first_name}` : t("agenda.patient");
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
  const canManageAppointments = user?.role === "admin" || user?.role === "recepcionista" || user?.role === "kinesiologo";
  // Agenda (listado)
  const [date, setDate] = useState(todayISO());
  const [kinesiologistId, setKinesiologistId] = useState("");
  const [viewMode, setViewMode] = useState<"day" | "week">("week");
  const [statusFilter, setStatusFilter] = useState<"all" | Appointment["status"]>("all");
  const [filterPatient, setFilterPatient] = useState<Patient | null>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [filterPatientSearchResetKey, setFilterPatientSearchResetKey] = useState(0);
  const [createPatientSearchResetKey, setCreatePatientSearchResetKey] = useState(0);

  // Crear turno (inputs UX)
  const [patientId, setPatientId] = useState("");
  const [createMode, setCreateMode] = useState<"single" | "package">("single");
  const [apptDate, setApptDate] = useState(todayISO());
  const [startTime, setStartTime] = useState(currentRoundedTime());
  const [durationMin, setDurationMin] = useState(45);
  const [practiceId, setPracticeId] = useState("");
  const [financierId, setFinancierId] = useState("");
  const [modality, setModality] = useState<"in_person" | "virtual">("in_person");
  const [videoLinkMode, setVideoLinkMode] = useState<"manual" | "auto">("manual");
  const [videoCallUrl, setVideoCallUrl] = useState("");
  const [sessionsCount, setSessionsCount] = useState(8);
  const [notes, setNotes] = useState("");
  const [editingAppointment, setEditingAppointment] = useState<Appointment | null>(null);
  const [editingPackageId, setEditingPackageId] = useState<string | null>(null);
  const [editPackageDate, setEditPackageDate] = useState(todayISO());
  const [editPackageDays, setEditPackageDays] = useState<number[]>(defaultWorkDays);
  const [editDate, setEditDate] = useState(todayISO());
  const [editStartTime, setEditStartTime] = useState("09:00");
  const [editDurationMin, setEditDurationMin] = useState(45);
  const [editPracticeId, setEditPracticeId] = useState("");
  const [editFinancierId, setEditFinancierId] = useState("");
  const [editModality, setEditModality] = useState<"in_person" | "virtual">("in_person");
  const [editVideoCallUrl, setEditVideoCallUrl] = useState("");
  const [editNotes, setEditNotes] = useState("");
  const [cancelAppointmentTarget, setCancelAppointmentTarget] = useState<Appointment | null>(null);
  const [cancelReason, setCancelReason] = useState("");
  const [completeTarget, setCompleteTarget] = useState<Appointment | null>(null);
  const [completePracticeId, setCompletePracticeId] = useState("");
  const [completeFinancierId, setCompleteFinancierId] = useState("");

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

  const financiersQ = useQuery({
    queryKey: ["financiers", "agenda"],
    queryFn: () => listFinanciers(),
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
  const selectedPatient = patientById.get(patientId);
  const financiers = financiersQ.data ?? [];
  const weekDays = useMemo(() => weekDaysFor(date), [date]);
  const weekStart = weekDays[0] ?? date;
  const weekEnd = weekDays[6] ?? date;

  const clearCreatePatient = () => {
    setPatientId("");
    setCreatePatientSearchResetKey((value) => value + 1);
  };

  const resetAgendaFilters = () => {
    setDate(todayISO());
    setViewMode("week");
    setStatusFilter("all");
    setFilterPatient(null);
    setFilterPatientSearchResetKey((value) => value + 1);
  };

  useEffect(() => {
    if (!isKinesiologist || !user?.email || kinesios.length === 0) return;

    // No caemos a kinesios[0]: si el email no matchea ningun perfil, mostramos
    // el aviso de "perfil no encontrado" en vez de mostrarle la agenda de otro profesional.
    const ownProfile = kinesios.find((k) => k.email.toLowerCase() === user.email.toLowerCase());

    if (ownProfile && ownProfile.id !== kinesiologistId) {
      setKinesiologistId(ownProfile.id);
    }
  }, [isKinesiologist, kinesiologistId, kinesios, user?.email]);

  useEffect(() => {
    const practices = selectedKinesiologist?.practices?.filter((practice) => practice.active) ?? [];
    if (practices.length > 0 && !practices.some((practice) => practice.id === practiceId)) {
      setPracticeId(practices[0].id);
    }
  }, [practiceId, selectedKinesiologist]);

  useEffect(() => {
    if (financiers.length > 0 && !financiers.some((financier) => financier.id === financierId)) {
      setFinancierId(financiers[0].id);
    }
  }, [financierId, financiers]);

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

  const generateVideoM = useMutation({
    mutationFn: generateAppointmentVideoCall,
    onSuccess: (updated) => {
      if (editingAppointment?.id === updated.id) {
        setEditingAppointment(updated);
        setEditModality(updated.modality ?? "virtual");
        setEditVideoCallUrl(updated.video_call_url ?? "");
      }
      agendaQ.refetch();
      weekAgendaQ.refetch();
    },
    onSettled: () => {
      agendaQ.refetch();
      weekAgendaQ.refetch();
    },
  });

  const createM = useMutation({
    mutationFn: createAppointment,
    onSuccess: (created, variables) => {
      if (variables.modality === "virtual" && videoLinkMode === "auto" && !created.video_call_url) {
        generateVideoM.mutate(created.id);
        return;
      }
      agendaQ.refetch();
    },
    onError: (error) => {
      if (isInactivePatientError(error)) {
        clearCreatePatient();
      }
    },
  });

  const createPackageM = useMutation({
    mutationFn: createAppointmentPackage,
    onSuccess: () => {
      agendaQ.refetch();
      weekAgendaQ.refetch();
    },
    onError: (error) => {
      if (isInactivePatientError(error)) {
        clearCreatePatient();
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
    mutationFn: (args: {
      id: string;
      start_at: string;
      end_at: string;
      modality: "in_person" | "virtual";
      video_call_url?: string;
      notes?: string;
    }) =>
      updateAppointment({
        id: args.id,
        start_at: args.start_at,
        end_at: args.end_at,
        modality: args.modality,
        video_call_url: args.video_call_url,
        notes: args.notes,
      }),
    onSuccess: () => {
      setEditingAppointment(null);
      agendaQ.refetch();
      weekAgendaQ.refetch();
    },
  });

  const packageUpdateM = useMutation({
    mutationFn: (args: {
      id: string;
      start_date: string;
      start_time: string;
      duration_min: number;
      practice_id?: string;
      financier_id?: string;
      work_days: number[];
      notes?: string;
    }) =>
      updateAppointmentPackage({
        id: args.id,
        start_date: args.start_date,
        start_time: args.start_time,
        duration_min: args.duration_min,
        practice_id: args.practice_id,
        financier_id: args.financier_id,
        work_days: args.work_days,
        notes: args.notes,
      }),
    onSuccess: () => {
      setEditingPackageId(null);
      agendaQ.refetch();
      weekAgendaQ.refetch();
    },
  });

  const completeM = useMutation({
    mutationFn: completeAppointment,
    onSuccess: () => {
      setCompleteTarget(null);
      agendaQ.refetch();
      weekAgendaQ.refetch();
    },
  });


  function create() {
    if (isCreateInPast) return;

    const endTime = addMinutesToHHmm(startTime, durationMin);

    if (createMode === "package") {
      createPackageM.mutate({
        patient_id: patientId.trim(),
        kinesiologist_id: kinesiologistId,
        practice_id: practiceId || undefined,
        financier_id: financierId || undefined,
        start_date: apptDate,
        start_time: startTime,
        duration_min: durationMin,
        sessions_count: sessionsCount,
        weekdays_only: true,
        notes: notes.trim() ? notes.trim() : undefined,
      });
      return;
    }

    createM.mutate({
      patient_id: patientId.trim(),
      kinesiologist_id: kinesiologistId,
      practice_id: practiceId || undefined,
      financier_id: financierId || undefined,
      start_at: localDateTimeToUTC(apptDate, startTime),
      end_at: localDateTimeToUTC(apptDate, endTime),
      modality,
      video_call_url:
        modality === "virtual" && videoLinkMode === "manual" && videoCallUrl.trim()
          ? videoCallUrl.trim()
          : undefined,
      notes: notes.trim() ? notes.trim() : undefined,
    });
  }

  function openEditAppointment(appt: Appointment) {
    setEditingAppointment(appt);
    setEditingPackageId(null);
    setEditDate(localDateISO(appt.start_at));
    setEditStartTime(formatLocalTime(appt.start_at));
    setEditDurationMin(durationMinutes(appt.start_at, appt.end_at));
    setEditPracticeId(appt.practice_id ?? practiceId);
    setEditFinancierId(appt.financier_id ?? financierId);
    setEditModality(appt.modality ?? "in_person");
    setEditVideoCallUrl(appt.video_call_url ?? "");
    setEditNotes(appt.notes ?? "");
    rescheduleM.reset();
  }

  function openEditPackage(appt: Appointment) {
    if (!appt.package_id) return;
    setEditingPackageId(appt.package_id);
    setEditingAppointment(null);
    setEditStartTime(formatLocalTime(appt.start_at));
    setEditDurationMin(durationMinutes(appt.start_at, appt.end_at));
    setEditNotes(appt.notes ?? "");
    setEditPackageDate(localDateISO(appt.start_at));
    setEditPackageDays(selectedKinesiologist?.work_days?.length ? selectedKinesiologist.work_days : defaultWorkDays);
    packageUpdateM.reset();
  }

  function submitEditAppointment() {
    if (!editingAppointment || isEditInPast || isEditOutsideWorkingHours) return;

    const endTime = addMinutesToHHmm(editStartTime, editDurationMin);
    rescheduleM.mutate({
      id: editingAppointment.id,
      start_at: localDateTimeToUTC(editDate, editStartTime),
      end_at: localDateTimeToUTC(editDate, endTime),
      modality: editModality,
      video_call_url: editModality === "virtual" ? editVideoCallUrl.trim() : undefined,
      notes: editNotes.trim() ? editNotes.trim() : undefined,
    });
  }

  function submitEditPackage() {
    if (!editingPackageId || isPackageEditOutsideWorkingHours || !editPracticeId || !editFinancierId) return;
    packageUpdateM.mutate({
      id: editingPackageId,
      start_date: editPackageDate,
      start_time: editStartTime,
      duration_min: editDurationMin,
      practice_id: editPracticeId,
      financier_id: editFinancierId,
      work_days: editPackageDays,
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

  function openCompleteAppointment(appt: Appointment) {
    setCompleteTarget(appt);
    setCompletePracticeId(appt.practice_id ?? practiceId);
    setCompleteFinancierId(appt.financier_id ?? financierId);
    completeM.reset();
  }

  function submitCompleteAppointment() {
    if (!completeTarget || !completePracticeId || !completeFinancierId) return;
    completeM.mutate({
      appointment_id: completeTarget.id,
      practice_id: completePracticeId,
      financier_id: completeFinancierId,
    });
  }

  const activeCreateError = createMode === "package" ? createPackageM.error : createM.error;
  const createErr: any = activeCreateError;
  const videoGenerateErr: any = generateVideoM.error;
  const hasCreateOverlap = isOverlapError(activeCreateError);
  const hasCreateInactivePatient = isInactivePatientError(activeCreateError);
  const hasRescheduleOverlap = isOverlapError(rescheduleM.error);
  const hasPackageUpdateOverlap = isOverlapError(packageUpdateM.error);
  const isCreateInPast = isPastLocalDateTime(apptDate, startTime);
  const isCreateOutsideWorkingHours = isOutsideWorkingSchedule(
    apptDate,
    startTime,
    durationMin,
    selectedKinesiologist,
  );
  const createValidationMessage =
    validationDetail(activeCreateError, "start_at") ??
    validationDetail(activeCreateError, "start_date") ??
    validationDetail(activeCreateError, "start_time") ??
    validationDetail(activeCreateError, "session");
  const rescheduleValidationMessage = validationDetail(rescheduleM.error, "start_at");
  const packageUpdateValidationMessage =
    validationDetail(packageUpdateM.error, "start_time") ??
    validationDetail(packageUpdateM.error, "start_date") ??
    validationDetail(packageUpdateM.error, "work_days") ??
    validationDetail(packageUpdateM.error, "session");
  const isEditInPast = isPastLocalDateTime(editDate, editStartTime);
  const isEditOutsideWorkingHours = isOutsideWorkingSchedule(
    editDate,
    editStartTime,
    editDurationMin,
    selectedKinesiologist,
  );
  const isPackageEditOutsideWorkingHours = selectedKinesiologist
    ? isOutsideWorkingHours(
        editStartTime,
        editDurationMin,
        selectedKinesiologist.work_start_time,
        selectedKinesiologist.work_end_time,
      )
    : false;
  const isPackageEditDaysInvalid = editPackageDays.length === 0;
  const isGeneratingVideoCall = generateVideoM.isPending;
  const isCreating = createM.isPending || createPackageM.isPending || isGeneratingVideoCall;
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
      <div className="max-w-7xl mx-auto p-6 space-y-6">
        <header className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold">{t("agenda.title")}</h1>
            <p className="text-sm text-gray-600">
              {isKinesiologist ? t("agenda.myDay") : t("agenda.dailySubtitle")}
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            {canManageAppointments && (
              <button
                type="button"
                className="rounded-lg bg-black px-4 py-2 text-sm text-white hover:bg-gray-800"
                onClick={() => {
                  if (showCreateForm) {
                    setShowCreateForm(false);
                    return;
                  }
                  clearCreatePatient();
                  setStartTime(currentRoundedTime());
                  setShowCreateForm(true);
                }}
              >
                {showCreateForm ? t("agenda.hideForm") : t("agenda.newAppointment")}
              </button>
            )}
            {(user?.role === "admin" || user?.role === "recepcionista" || user?.role === "kinesiologo") && (
              <Link className="text-sm underline" to="/patients">
                {t("agenda.goPatients")}
              </Link>
            )}
          </div>
        </header>

        {/* Filtros agenda */}
        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div className="flex flex-col gap-1 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <h2 className="text-lg font-semibold">{t("agenda.scheduleView")}</h2>
              <p className="text-sm text-gray-600">
                {t("agenda.scheduleViewHelp")}
              </p>
            </div>
            <button
              type="button"
              className="w-fit px-3 py-2 rounded-lg border text-sm hover:bg-gray-50"
              onClick={resetAgendaFilters}
            >
              {t("agenda.clearFilters")}
            </button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-[180px_160px_minmax(260px,1fr)] gap-3">
            <div>
              <label className="text-sm font-medium"><RequiredLabel required>{t("agenda.date")}</RequiredLabel></label>
              <input
                className="mt-1 w-full border rounded-lg p-2"
                type="date"
                value={date}
                onChange={(e) => setDate(e.target.value)}
              />
            </div>

            <div>
              <label className="text-sm font-medium"><RequiredLabel required>{t("agenda.view")}</RequiredLabel></label>
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
              <label className="text-sm font-medium"><RequiredLabel required>{t("agenda.kinesiologist")}</RequiredLabel></label>
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
                  {selectedKinesiologist.work_start_time} - {selectedKinesiologist.work_end_time} · {workDaysLabel(selectedKinesiologist.work_days, t)}
                </p>
              )}
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-[180px_1fr] gap-3 items-end">
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
                <option value="completed">{t("agenda.statusCompleted")}</option>
              </select>
            </div>

            <div>
              <PatientSearch
                valuePatientId={filterPatient?.id ?? ""}
                onSelect={setFilterPatient}
                label={t("agenda.filterAgenda")}
                placeholder={t("agenda.filterPatient")}
                showSelected={false}
                resetKey={filterPatientSearchResetKey}
              />
            </div>
          </div>

          {(statusFilter !== "all" || filterPatient) && (
            <div className="flex flex-wrap gap-2 text-xs">
              {statusFilter !== "all" && (
                <span className="rounded-full bg-gray-100 px-3 py-1 text-gray-700">
                  {t("agenda.status")}: {statusLabel(statusFilter, t)}
                </span>
              )}
              {filterPatient && (
                <span className="rounded-full bg-gray-100 px-3 py-1 text-gray-700">
                  {t("agenda.patient")}: {patientName(filterPatient, t)}
                </span>
              )}
            </div>
          )}
        </section>

        {/* Crear turno */}
        {canManageAppointments && showCreateForm && (
          <section className="bg-white rounded-xl shadow p-4 space-y-4">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
              <div>
                <h2 className="text-lg font-semibold">{t("agenda.create")}</h2>
                <p className="text-sm text-gray-600">{t("agenda.createSummary")}</p>
              </div>

              <div className="flex flex-wrap items-center gap-2">
                <div className="inline-flex w-fit rounded-xl border bg-gray-50 p-1">
                  <button
                    type="button"
                    className={`px-3 py-2 rounded-lg text-sm ${createMode === "single" ? "bg-black text-white shadow-sm" : "hover:bg-white"}`}
                    onClick={() => setCreateMode("single")}
                  >
                    {t("agenda.singleAppointment")}
                  </button>
                  <button
                    type="button"
                    className={`px-3 py-2 rounded-lg text-sm ${createMode === "package" ? "bg-black text-white shadow-sm" : "hover:bg-white"}`}
                    onClick={() => setCreateMode("package")}
                  >
                    {t("agenda.appointmentPackage")}
                  </button>
                </div>
                <button
                  type="button"
                  className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50"
                  onClick={() => setShowCreateForm(false)}
                >
                  {t("agenda.hide")}
                </button>
              </div>
            </div>

            <div className="grid grid-cols-1 xl:grid-cols-[1fr_1fr] gap-4">
              <div className="rounded-xl border bg-gray-50/60 p-4 space-y-3">
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <div className="text-sm font-semibold">1. {t("agenda.patient")}</div>
                    <p className="text-xs text-gray-500">{t("agenda.searchPatientHelp")}</p>
                  </div>
                  {selectedPatient && (
                    <span className="rounded-full bg-green-100 px-3 py-1 text-xs text-green-700">
                      {t("agenda.selected")}
                    </span>
                  )}
                </div>

                <PatientSearch
                  valuePatientId={patientId}
                  label={null}
                  showSelected={false}
                  resetKey={createPatientSearchResetKey}
                  onSelect={(p) => {
                    setPatientId(p.id);
                  }}
                />

                {selectedPatient ? (
                  <div className="rounded-lg border bg-white p-3">
                    <div className="font-medium">{patientName(selectedPatient, t)}</div>
                    <div className="text-sm text-gray-600">
                      DNI {selectedPatient.dni} · {selectedPatient.email}
                    </div>
                  </div>
                ) : (
                  <p className="text-sm text-gray-500">{t("agenda.noPatientSelected")}</p>
                )}
              </div>

              <div className="rounded-xl border bg-gray-50/60 p-4 space-y-3">
                <div>
                  <div className="text-sm font-semibold">2. {t("agenda.stepDateTime")}</div>
                  {selectedKinesiologist && (
                    <p className="text-xs text-gray-500">
                      {t("agenda.attends")} {selectedKinesiologist.work_start_time} - {selectedKinesiologist.work_end_time} · {workDaysLabel(selectedKinesiologist.work_days, t)}
                    </p>
                  )}
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div>
                    <label className="text-sm font-medium"><RequiredLabel required>{t("agenda.appointmentDate")}</RequiredLabel></label>
                    <input
                      className="mt-1 w-full border rounded-lg p-2 bg-white"
                      type="date"
                      min={todayISO()}
                      value={apptDate}
                      onChange={(e) => setApptDate(e.target.value)}
                    />
                  </div>

                  <div>
                    <label className="text-sm font-medium"><RequiredLabel required>{t("agenda.startTime")}</RequiredLabel></label>
                    <input
                      className="mt-1 w-full border rounded-lg p-2 bg-white"
                      type="time"
                      value={startTime}
                      onChange={(e) => setStartTime(e.target.value)}
                    />
                  </div>
                </div>

                <div>
                  <label className="text-sm font-medium"><RequiredLabel required>{t("agenda.duration")}</RequiredLabel></label>
                  <div className="mt-2 flex flex-wrap gap-2">
                    {[30, 45, 60].map((m) => (
                      <button
                        key={m}
                        type="button"
                        className={`px-3 py-2 rounded-lg border text-sm ${durationMin === m ? "bg-black text-white" : "bg-white hover:bg-gray-100"}`}
                        onClick={() => setDurationMin(m)}
                      >
                        {m} min
                      </button>
                    ))}
                    <input
                      className="w-28 border rounded-lg p-2 bg-white"
                      type="number"
                      min={15}
                      step={15}
                      value={durationMin}
                      onChange={(e) => setDurationMin(Number(e.target.value))}
                      aria-label={t("agenda.duration")}
                    />
                  </div>
                </div>
              </div>

              <div className="xl:col-span-2 rounded-xl border p-4 space-y-4">
                <div>
                  <div className="text-sm font-semibold">3. {t("agenda.stepDetails")}</div>
                  <p className="text-xs text-gray-500">{t("agenda.detailsHelp")}</p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  {createMode === "single" && (
                    <div className="md:col-span-2 space-y-3">
                      <label className="text-sm font-medium"><RequiredLabel required>{t("agenda.modality")}</RequiredLabel></label>
                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                        <button
                          type="button"
                          className={`rounded-xl border p-3 text-left ${modality === "in_person" ? "border-black bg-gray-50" : "bg-white hover:bg-gray-50"}`}
                          onClick={() => {
                            setModality("in_person");
                            setVideoCallUrl("");
                            setVideoLinkMode("manual");
                          }}
                        >
                          <div className="font-medium">{t("agenda.inPerson")}</div>
                          <div className="text-xs text-gray-500">{t("agenda.inPersonHelp")}</div>
                        </button>
                        <button
                          type="button"
                          className={`rounded-xl border p-3 text-left ${modality === "virtual" ? "border-black bg-blue-50" : "bg-white hover:bg-gray-50"}`}
                          onClick={() => setModality("virtual")}
                        >
                          <div className="font-medium">{t("agenda.virtual")}</div>
                          <div className="text-xs text-gray-500">{t("agenda.virtualHelp")}</div>
                        </button>
                      </div>

                      {modality === "virtual" && (
                        <div className="rounded-xl border border-blue-100 bg-blue-50/60 p-3 space-y-3">
                          <div className="flex flex-wrap gap-2">
                            <button
                              type="button"
                              className={`px-3 py-1 rounded-lg border text-sm ${videoLinkMode === "manual" ? "bg-black text-white" : "bg-white hover:bg-gray-50"}`}
                              onClick={() => setVideoLinkMode("manual")}
                            >
                              {t("agenda.videoCallManual")}
                            </button>
                            <button
                              type="button"
                              className={`px-3 py-1 rounded-lg border text-sm ${videoLinkMode === "auto" ? "bg-black text-white" : "bg-white hover:bg-gray-50"}`}
                              onClick={() => setVideoLinkMode("auto")}
                            >
                              {t("agenda.videoCallAuto")}
                            </button>
                          </div>

                          {videoLinkMode === "manual" ? (
                            <div>
                              <label className="text-sm font-medium">{t("agenda.videoCallUrl")}</label>
                              <input
                                className="mt-1 w-full border rounded-lg p-2 bg-white"
                                type="url"
                                placeholder="https://meet.google.com/..."
                                value={videoCallUrl}
                                onChange={(event) => setVideoCallUrl(event.target.value)}
                              />
                              <p className="text-xs text-gray-500 mt-1">{t("agenda.videoCallUrlHelp")}</p>
                            </div>
                          ) : (
                            <p className="text-sm text-gray-700">{t("agenda.videoCallAutoHelp")}</p>
                          )}
                        </div>
                      )}
                    </div>
                  )}

                  <div>
                    <label className="text-sm font-medium">{t("agenda.practice")}</label>
                    <select
                      className="mt-1 w-full border rounded-lg p-2"
                      value={practiceId}
                      onChange={(event) => setPracticeId(event.target.value)}
                    >
                      <option value="">{t("agenda.select")}</option>
                      {(selectedKinesiologist?.practices ?? []).filter((practice) => practice.active).map((practice) => (
                        <option key={practice.id} value={practice.id}>
                          {practice.name}
                        </option>
                      ))}
                    </select>
                    {selectedKinesiologist && selectedKinesiologist.practices.length === 0 && (
                      <p className="text-xs text-amber-700 mt-1">{t("agenda.noPracticesAssigned")}</p>
                    )}
                  </div>

                  <div>
                    <label className="text-sm font-medium">{t("agenda.funder")}</label>
                    <select
                      className="mt-1 w-full border rounded-lg p-2"
                      value={financierId}
                      onChange={(event) => setFinancierId(event.target.value)}
                    >
                      <option value="">{t("agenda.select")}</option>
                      {financiers.map((financier) => (
                        <option key={financier.id} value={financier.id}>
                          {financier.name}
                        </option>
                      ))}
                    </select>
                  </div>

                  {createMode === "package" && (
                    <div>
                      <label className="text-sm font-medium"><RequiredLabel required>{t("agenda.sessionsCount")}</RequiredLabel></label>
                      <input
                        className="mt-1 w-full border rounded-lg p-2"
                        type="number"
                        min={1}
                        max={80}
                        value={sessionsCount}
                        onChange={(event) => setSessionsCount(Number(event.target.value))}
                      />
                      <p className="text-xs text-gray-500 mt-2">
                        {t("agenda.weekdaysOnly")}
                      </p>
                    </div>
                  )}

                  <div className={createMode === "package" ? "" : "md:col-span-2"}>
                    <label className="text-sm font-medium">{t("agenda.notes")}</label>
                    <input
                      className="mt-1 w-full border rounded-lg p-2"
                      value={notes}
                      onChange={(e) => setNotes(e.target.value)}
                    />
                  </div>
                </div>
              </div>
            </div>

            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div className="text-sm text-gray-600">
                {selectedPatient ? patientName(selectedPatient, t) : t("agenda.noPatientSelected")}
                {" · "}
                {apptDate} {startTime}
                {" · "}
                {createMode === "single" && modality === "virtual" ? t("agenda.virtual") : t("agenda.inPerson")}
              </div>
              <button
                className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
                disabled={
                  !patientId.trim() ||
                  !kinesiologistId ||
                  !practiceId ||
                  !financierId ||
                  isCreateInPast ||
                  isCreateOutsideWorkingHours ||
                  (createMode === "package" && sessionsCount <= 0) ||
                  isCreating
                }
                onClick={create}
              >
                {isCreating
                  ? isGeneratingVideoCall
                    ? t("agenda.generatingVideoCall")
                    : t("agenda.creating")
                  : createMode === "package"
                    ? t("agenda.createPackage")
                    : t("agenda.create")}
              </button>
            </div>

          {isCreateInPast && (
            <p className="text-sm text-red-600">
              {t("agenda.appointmentPast")}
            </p>
          )}

          {isCreateOutsideWorkingHours && selectedKinesiologist && (
            <p className="text-sm text-red-600">
              {t("agenda.appointmentOutOfHours")} {selectedKinesiologist.work_start_time} - {selectedKinesiologist.work_end_time} · {workDaysLabel(selectedKinesiologist.work_days, t)}.
            </p>
          )}

          {(createM.isError || createPackageM.isError) && (
            <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700">
              {hasCreateOverlap ? (
                <>
                  <div className="font-medium">{t("agenda.overlapTitle")}</div>
                  <div>{t("agenda.overlapDetail")}</div>
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

          {generateVideoM.isError && (
            <div className="border border-amber-200 bg-amber-50 rounded-lg p-3 text-sm text-amber-800">
              <div className="font-medium">{t("agenda.videoCallGenerateError")}</div>
              <div>
                {videoGenerateErr?.message === "video_provider_not_configured"
                  ? t("agenda.videoProviderNotConfigured")
                  : t("agenda.videoProviderError")}
              </div>
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
                getPatientName={(id) => patientName(patientById.get(id), t)}
                onPickSlot={(hhmm) => {
                  clearCreatePatient();
                  setApptDate(date);
                  setStartTime(hhmm);
                  setDurationMin(45);
                  setShowCreateForm(true);
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
                            {patientName(patientById.get(a.patient_id), t)}
                          </Link>
                        </div>
                        <div className="text-sm text-gray-600">{t("agenda.status")}: {statusLabel(a.status, t)}</div>
                        {a.status === "completed" && (
                          <div className="text-sm text-green-700">{t("agenda.financeMovementGenerated")}</div>
                        )}
                        {a.status === "cancelled" && a.cancelled_reason && (
                          <div className="text-sm text-gray-600">{t("agenda.cancelReason")}: {a.cancelled_reason}</div>
                        )}
                        <div className="text-sm text-gray-600">
                          {t("agenda.modality")}: {a.modality === "virtual" ? t("agenda.virtual") : t("agenda.inPerson")}
                          {a.modality === "virtual" && a.video_call_url && (
                            <>
                              {" · "}
                              <a className="underline text-blue-700" href={a.video_call_url} target="_blank" rel="noreferrer">
                                {t("agenda.openVideoCall")}
                              </a>
                            </>
                          )}
                        </div>
                        {a.package_id && (
                          <div className="text-sm text-gray-600">
                            {t("agenda.packageSession")} {a.package_session_number ?? "-"}
                          </div>
                        )}
                        {a.notes && <div className="text-sm text-gray-600">{t("agenda.notes")}: {a.notes}</div>}
                      </div>

                      <div className="flex items-center gap-3">
                        {canManageAppointments && (
                          <div className="text-xs text-gray-500 font-mono">{a.id}</div>
                        )}

                        {canManageAppointments && a.status === "scheduled" && (
                          <button
                            className="px-3 py-1 rounded-lg border text-sm hover:bg-green-50 disabled:opacity-50"
                            disabled={completeM.isPending}
                            onClick={() => openCompleteAppointment(a)}
                          >
                            {completeM.isPending ? "..." : t("agenda.appointmentDone")}
                          </button>
                        )}

                        {canManageAppointments && a.status === "scheduled" ? (
                          <button
                            className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-100 disabled:opacity-50"
                            disabled={cancelM.isPending || rescheduleM.isPending || completeM.isPending}
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
                        ) : a.status === "completed" ? (
                          <span className="text-xs px-2 py-1 rounded-md bg-green-100 text-green-700">
                            {t("agenda.appointmentDone")}
                          </span>
                        ) : (
                          <span className="text-xs px-2 py-1 rounded-md bg-blue-100 text-blue-700">
                            {t("agenda.scheduled")}
                          </span>
                        )}

                        {canManageAppointments && a.status === "scheduled" && (
                          <button
                            className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-100 disabled:opacity-50"
                            disabled={rescheduleM.isPending || cancelM.isPending || completeM.isPending}
                            onClick={() => {
                              openEditAppointment(a);
                            }}
                          >
                            {rescheduleM.isPending ? "..." : t("agenda.reschedule")}
                          </button>
                        )}

                        {canManageAppointments && a.package_id && a.status === "scheduled" && (
                          <button
                            className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-100 disabled:opacity-50"
                            disabled={packageUpdateM.isPending}
                            onClick={() => {
                              openEditPackage(a);
                            }}
                          >
                            {packageUpdateM.isPending ? "..." : t("agenda.editPackage")}
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
                        <div className="font-medium">{t("agenda.rescheduleOverlap")}</div>
                        <div>{t("agenda.overlapDetail")}</div>
                      </>
                    ) : (
                      <>
                        <div className="font-medium">{t("agenda.rescheduleError")}</div>
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
              {!weekAgendaQ.isLoading && !weekAgendaQ.isError && (
                <WeeklyCalendar
                  groups={weekAppointments}
                  canManageAppointments={canManageAppointments}
                  getPatientName={(patientId) => patientName(patientById.get(patientId), t)}
                  onPickDay={(day) => {
                    clearCreatePatient();
                    setDate(day);
                    setApptDate(day);
                    setShowCreateForm(true);
                    window.scrollTo({ top: 0, behavior: "smooth" });
                  }}
                  onComplete={openCompleteAppointment}
                  onEditPackage={openEditPackage}
                  onReschedule={openEditAppointment}
                  onCancel={openCancelAppointment}
                  t={t}
                />
              )}
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

              <div>
                <label className="text-sm font-medium">{t("agenda.modality")}</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={editModality}
                  onChange={(event) => {
                    const next = event.target.value as "in_person" | "virtual";
                    setEditModality(next);
                    if (next === "in_person") setEditVideoCallUrl("");
                  }}
                >
                  <option value="in_person">{t("agenda.inPerson")}</option>
                  <option value="virtual">{t("agenda.virtual")}</option>
                </select>
              </div>

              {editModality === "virtual" && (
                <div className="space-y-2">
                  <label className="text-sm font-medium">{t("agenda.videoCallUrl")}</label>
                  <input
                    className="mt-1 w-full border rounded-lg p-2"
                    type="url"
                    placeholder="https://meet.google.com/..."
                    value={editVideoCallUrl}
                    onChange={(event) => setEditVideoCallUrl(event.target.value)}
                  />
                  <div className="flex flex-wrap items-center gap-2">
                    <button
                      type="button"
                      className="px-3 py-1 rounded-lg border bg-white text-sm hover:bg-gray-50 disabled:opacity-50"
                      disabled={generateVideoM.isPending}
                      onClick={() => {
                        if (editingAppointment) generateVideoM.mutate(editingAppointment.id);
                      }}
                    >
                      {generateVideoM.isPending ? t("agenda.generatingVideoCall") : t("agenda.generateVideoCall")}
                    </button>
                    {editVideoCallUrl && (
                      <span className="text-xs text-green-700">
                        {t("agenda.videoCallReady")}
                      </span>
                    )}
                  </div>
                </div>
              )}
            </div>

            {selectedKinesiologist && (
              <p className="text-sm text-gray-600">
                {selectedKinesiologist.work_start_time} - {selectedKinesiologist.work_end_time} · {workDaysLabel(selectedKinesiologist.work_days, t)}
              </p>
            )}

            {isEditInPast && (
              <p className="text-sm text-red-600">{t("agenda.appointmentPast")}</p>
            )}

            {isEditOutsideWorkingHours && selectedKinesiologist && (
              <p className="text-sm text-red-600">
                {t("agenda.appointmentOutOfHours")} {selectedKinesiologist.work_start_time} - {selectedKinesiologist.work_end_time} · {workDaysLabel(selectedKinesiologist.work_days, t)}.
              </p>
            )}

            {rescheduleM.isError && (
              <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700">
                {hasRescheduleOverlap ? (
                  <>
                    <div className="font-medium">{t("agenda.rescheduleOverlap")}</div>
                    <div>{t("agenda.overlapDetail")}</div>
                  </>
                ) : (
                  <>
                    <div className="font-medium">{t("agenda.updateError")}</div>
                    <div>{rescheduleValidationMessage ?? String((rescheduleM.error as any)?.message)}</div>
                  </>
                )}
              </div>
            )}

            {generateVideoM.isError && (
              <div className="border border-amber-200 bg-amber-50 rounded-lg p-3 text-sm text-amber-800">
                <div className="font-medium">{t("agenda.videoCallGenerateError")}</div>
                <div>
                  {videoGenerateErr?.message === "video_provider_not_configured"
                    ? t("agenda.videoProviderNotConfigured")
                    : t("agenda.videoProviderError")}
                </div>
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

      {editingPackageId && (
        <div className="fixed inset-0 z-50 bg-black/40 flex items-center justify-center p-4">
          <section className="bg-white rounded-xl shadow-xl p-4 w-full max-w-xl space-y-4">
            <div className="flex items-start justify-between gap-4">
              <div>
                <h2 className="text-lg font-semibold">{t("agenda.editPackage")}</h2>
                <p className="text-sm text-gray-600">
                  {t("agenda.editPackageDetail")}
                </p>
              </div>
              <button
                type="button"
                className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                onClick={() => setEditingPackageId(null)}
              >
                {t("agenda.close")}
              </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div>
                <label className="text-sm font-medium">{t("agenda.packageStartDate")}</label>
                <input
                  className="mt-1 w-full border rounded-lg p-2"
                  type="date"
                  min={todayISO()}
                  value={editPackageDate}
                  onChange={(event) => setEditPackageDate(event.target.value)}
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
                <label className="text-sm font-medium">{t("agenda.practice")}</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={editPracticeId}
                  onChange={(event) => setEditPracticeId(event.target.value)}
                >
                  <option value="">{t("agenda.select")}</option>
                  {(selectedKinesiologist?.practices ?? []).filter((practice) => practice.active).map((practice) => (
                    <option key={practice.id} value={practice.id}>
                      {practice.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-sm font-medium">{t("agenda.funder")}</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={editFinancierId}
                  onChange={(event) => setEditFinancierId(event.target.value)}
                >
                  <option value="">{t("agenda.select")}</option>
                  {financiers.map((financier) => (
                    <option key={financier.id} value={financier.id}>
                      {financier.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="md:col-span-2">
                <label className="text-sm font-medium">{t("agenda.notes")}</label>
                <input
                  className="mt-1 w-full border rounded-lg p-2"
                  value={editNotes}
                  onChange={(event) => setEditNotes(event.target.value)}
                />
              </div>

              <div className="md:col-span-2">
                <label className="text-sm font-medium">{t("agenda.packageDays")}</label>
                <div className="mt-2 flex flex-wrap gap-2">
                  {(selectedKinesiologist?.work_days?.length ? selectedKinesiologist.work_days : defaultWorkDays).map((day) => {
                    const checked = editPackageDays.includes(day);
                    return (
                      <label
                        key={day}
                        className={`px-3 py-2 rounded-lg border text-sm cursor-pointer ${checked ? "bg-black text-white" : "bg-white hover:bg-gray-50"}`}
                      >
                        <input
                          className="sr-only"
                          type="checkbox"
                          checked={checked}
                          onChange={(event) => {
                            const next = event.target.checked
                              ? [...editPackageDays, day]
                              : editPackageDays.filter((value) => value !== day);
                            setEditPackageDays(next.sort((a, b) => a - b));
                          }}
                        />
                        {workDayLabel(day, t)}
                      </label>
                    );
                  })}
                </div>
              </div>
            </div>

            {selectedKinesiologist && (
              <p className="text-sm text-gray-600">
                {selectedKinesiologist.work_start_time} - {selectedKinesiologist.work_end_time} · {workDaysLabel(selectedKinesiologist.work_days, t)}
              </p>
            )}

            {isPackageEditOutsideWorkingHours && selectedKinesiologist && (
              <p className="text-sm text-red-600">
                {t("agenda.appointmentOutOfHours")} {selectedKinesiologist.work_start_time} - {selectedKinesiologist.work_end_time}.
              </p>
            )}

            {isPackageEditDaysInvalid && (
              <p className="text-sm text-red-600">{t("agenda.packageDaysRequired")}</p>
            )}

            {(!editPracticeId || !editFinancierId) && (
              <p className="text-sm text-red-600">{t("agenda.packageMissingBilling")}</p>
            )}

            {packageUpdateM.isError && (
              <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700">
                {hasPackageUpdateOverlap ? (
                  <>
                    <div className="font-medium">{t("agenda.overlapTitle")}</div>
                    <div>{t("agenda.overlapDetail")}</div>
                  </>
                ) : (
                  <>
                    <div className="font-medium">{t("agenda.packageUpdateError")}</div>
                    <div>{packageUpdateValidationMessage ?? String((packageUpdateM.error as any)?.message)}</div>
                  </>
                )}
              </div>
            )}

            <div className="flex justify-end gap-2">
              <button
                type="button"
                className="px-4 py-2 rounded-lg border hover:bg-gray-50"
                onClick={() => setEditingPackageId(null)}
              >
                {t("common.cancel")}
              </button>
              <button
                type="button"
                className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
                disabled={isPackageEditOutsideWorkingHours || isPackageEditDaysInvalid || !editPracticeId || !editFinancierId || packageUpdateM.isPending}
                onClick={submitEditPackage}
              >
                {packageUpdateM.isPending ? t("common.saving") : t("common.saveChanges")}
              </button>
            </div>
          </section>
        </div>
      )}

      {completeTarget && (
        <div className="fixed inset-0 z-50 bg-black/40 flex items-center justify-center p-4">
          <section className="bg-white rounded-xl shadow-xl p-4 w-full max-w-lg space-y-4">
            <div className="flex items-start justify-between gap-4">
              <div>
                <h2 className="text-lg font-semibold">{t("agenda.appointmentDone")}</h2>
                <p className="text-sm text-gray-600">
                  {formatLocalTime(completeTarget.start_at)} a {formatLocalTime(completeTarget.end_at)} ·{" "}
                  {patientName(patientById.get(completeTarget.patient_id), t)}
                </p>
              </div>
              <button
                type="button"
                className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                onClick={() => setCompleteTarget(null)}
              >
                {t("agenda.close")}
              </button>
            </div>

            <div className="grid grid-cols-1 gap-3">
              <div>
                <label className="text-sm font-medium">{t("agenda.billablePractice")}</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={completePracticeId}
                  onChange={(event) => setCompletePracticeId(event.target.value)}
                >
                  <option value="">{t("agenda.select")}</option>
                  {(selectedKinesiologist?.practices ?? []).filter((practice) => practice.active).map((practice) => (
                    <option key={practice.id} value={practice.id}>
                      {practice.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-sm font-medium">{t("agenda.funder")}</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={completeFinancierId}
                  onChange={(event) => setCompleteFinancierId(event.target.value)}
                >
                  <option value="">{t("agenda.select")}</option>
                  {financiers.map((financier) => (
                    <option key={financier.id} value={financier.id}>
                      {financier.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {completeM.isError && (
              <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700">
                {(completeM.error as any)?.message === "tariff_not_found" ? (
                  <>
                    <div className="font-medium">{t("agenda.tariffMissing")}</div>
                    <div>{t("agenda.billingValueMissing")}</div>
                  </>
                ) : (completeM.error as any)?.message === "financial_movement_already_generated" ? (
                  <>
                    <div className="font-medium">{t("agenda.financeMovementGenerated")}</div>
                    <div>{t("agenda.financeMovementAlreadyGenerated")}</div>
                  </>
                ) : (
                  <>
                    <div className="font-medium">{t("agenda.financeMovementNotCreated")}</div>
                    <div>{String((completeM.error as any)?.message)}</div>
                  </>
                )}
              </div>
            )}

            <div className="flex justify-end gap-2">
              <button
                type="button"
                className="px-4 py-2 rounded-lg border hover:bg-gray-50"
                onClick={() => setCompleteTarget(null)}
              >
                {t("common.back")}
              </button>
              <button
                type="button"
                className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
                disabled={!completePracticeId || !completeFinancierId || completeM.isPending}
                onClick={submitCompleteAppointment}
              >
                {completeM.isPending ? t("agenda.financeMovementGenerating") : t("agenda.financeMovementGenerate")}
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
                  {patientName(patientById.get(cancelAppointmentTarget.patient_id), t)}
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

function WeeklyCalendar({
  groups,
  canManageAppointments,
  getPatientName,
  onPickDay,
  onComplete,
  onEditPackage,
  onReschedule,
  onCancel,
  t,
}: {
  groups: Array<{ day: string; appointments: Appointment[] }>;
  canManageAppointments: boolean;
  getPatientName: (patientId: string) => string;
  onPickDay: (day: string) => void;
  onComplete: (appointment: Appointment) => void;
  onEditPackage: (appointment: Appointment) => void;
  onReschedule: (appointment: Appointment) => void;
  onCancel: (appointment: Appointment) => void;
  t: (key: string) => string;
}) {
  return (
    <>
      <div className="hidden lg:grid lg:grid-cols-7 gap-3">
        {groups.map((group) => (
          <section key={group.day} className="border rounded-lg min-h-72 bg-gray-50 overflow-hidden">
            <button
              type="button"
              className="w-full text-left px-3 py-2 bg-white border-b hover:bg-gray-50"
              onClick={() => onPickDay(group.day)}
            >
              <div className="text-sm font-semibold">{weekDayTitle(group.day, t)}</div>
              <div className="text-xs text-gray-600">{group.day}</div>
            </button>

            <div className="p-2 space-y-2">
              {group.appointments.length === 0 ? (
                <button
                  type="button"
                  className="w-full min-h-40 rounded-lg border border-dashed bg-white text-sm text-gray-500 hover:bg-gray-50"
                  onClick={() => onPickDay(group.day)}
                >
                  {t("agenda.noAppointments")}
                </button>
              ) : (
                group.appointments.map((appointment) => (
                  <article
                    key={appointment.id}
                    className={`rounded-lg border p-2 text-sm ${appointmentCardClass(appointment.status)}`}
                  >
                    <div className="font-medium">
                      {formatLocalTime(appointment.start_at)} → {formatLocalTime(appointment.end_at)}
                    </div>
                    <div className="text-gray-700">{getPatientName(appointment.patient_id)}</div>
                    <div className="text-xs text-gray-600">{statusLabel(appointment.status, t)}</div>
                    <div className="text-xs text-gray-600">
                      {appointment.modality === "virtual" ? t("agenda.virtual") : t("agenda.inPerson")}
                      {appointment.modality === "virtual" && appointment.video_call_url && (
                        <>
                          {" · "}
                          <a
                            className="underline text-blue-700"
                            href={appointment.video_call_url}
                            target="_blank"
                            rel="noreferrer"
                          >
                            {t("agenda.openVideoCall")}
                          </a>
                        </>
                      )}
                    </div>
                    {appointment.package_id && (
                      <div className="text-xs text-gray-600">
                        {t("agenda.packageSession")} {appointment.package_session_number ?? "-"}
                      </div>
                    )}
                    {appointment.status === "cancelled" && appointment.cancelled_reason && (
                      <div className="text-xs text-gray-600">{t("agenda.cancelReason")}: {appointment.cancelled_reason}</div>
                    )}
                    {appointment.notes && <div className="text-xs text-gray-600">{appointment.notes}</div>}

                    {canManageAppointments && appointment.status === "scheduled" && (
                      <div className="mt-2 flex flex-wrap gap-1">
                        <button
                          type="button"
                          className="px-2 py-1 rounded-md border bg-white text-xs hover:bg-green-50"
                          onClick={() => onComplete(appointment)}
                        >
                          {t("agenda.appointmentDone")}
                        </button>
                        <button
                          type="button"
                          className="px-2 py-1 rounded-md border bg-white text-xs hover:bg-gray-50"
                          onClick={() => onReschedule(appointment)}
                        >
                          {t("agenda.reschedule")}
                        </button>
                        {appointment.package_id && (
                          <button
                            type="button"
                            className="px-2 py-1 rounded-md border bg-white text-xs hover:bg-gray-50"
                            onClick={() => onEditPackage(appointment)}
                          >
                            {t("agenda.editPackage")}
                          </button>
                        )}
                        <button
                          type="button"
                          className="px-2 py-1 rounded-md border bg-white text-xs hover:bg-gray-50"
                          onClick={() => onCancel(appointment)}
                        >
                          {t("agenda.cancel")}
                        </button>
                      </div>
                    )}
                  </article>
                ))
              )}
            </div>
          </section>
        ))}
      </div>

      <div className="space-y-3 lg:hidden">
        {groups.map((group) => (
          <section key={group.day} className="border rounded-lg p-3">
            <button type="button" className="text-left" onClick={() => onPickDay(group.day)}>
              <div className="font-medium">{weekDayTitle(group.day, t)}</div>
              <div className="text-sm text-gray-600">{group.day}</div>
            </button>
            {group.appointments.length === 0 ? (
              <p className="text-sm text-gray-600 mt-2">{t("agenda.noAppointments")}</p>
            ) : (
              <div className="divide-y mt-2">
                {group.appointments.map((appointment) => (
                  <div key={appointment.id} className="py-2 flex flex-col gap-2">
                    <div>
                      <div className="font-medium">
                        {formatLocalTime(appointment.start_at)} → {formatLocalTime(appointment.end_at)}
                      </div>
                      <div className="text-sm text-gray-600">
                        {getPatientName(appointment.patient_id)} · {statusLabel(appointment.status, t)}
                      </div>
                      <div className="text-sm text-gray-600">
                        {appointment.modality === "virtual" ? t("agenda.virtual") : t("agenda.inPerson")}
                        {appointment.modality === "virtual" && appointment.video_call_url && (
                          <>
                            {" · "}
                            <a
                              className="underline text-blue-700"
                              href={appointment.video_call_url}
                              target="_blank"
                              rel="noreferrer"
                            >
                              {t("agenda.openVideoCall")}
                            </a>
                          </>
                        )}
                      </div>
                      {appointment.status === "completed" && (
                        <div className="text-sm text-green-700">{t("agenda.financeMovementGenerated")}</div>
                      )}
                      {appointment.status === "cancelled" && appointment.cancelled_reason && (
                        <div className="text-sm text-gray-600">{t("agenda.cancelReason")}: {appointment.cancelled_reason}</div>
                      )}
                      {appointment.package_id && (
                        <div className="text-sm text-gray-600">
                          {t("agenda.packageSession")} {appointment.package_session_number ?? "-"}
                        </div>
                      )}
                      {appointment.notes && <div className="text-sm text-gray-600">{t("agenda.notes")}: {appointment.notes}</div>}
                    </div>
                    {canManageAppointments && appointment.status === "scheduled" && (
                      <div className="flex flex-wrap gap-2">
                        <button
                          type="button"
                          className="px-3 py-1 rounded-lg border text-sm hover:bg-green-50"
                          onClick={() => onComplete(appointment)}
                        >
                          {t("agenda.appointmentDone")}
                        </button>
                        {appointment.package_id && (
                          <button
                            type="button"
                            className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                            onClick={() => onEditPackage(appointment)}
                          >
                            {t("agenda.editPackage")}
                          </button>
                        )}
                        <button
                          type="button"
                          className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                          onClick={() => onReschedule(appointment)}
                        >
                          {t("agenda.reschedule")}
                        </button>
                        <button
                          type="button"
                          className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                          onClick={() => onCancel(appointment)}
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
    </>
  );
}

function weekDayTitle(dayISO: string, t: (key: string) => string) {
  const [year, month, day] = dayISO.split("-").map(Number);
  const date = new Date(year, month - 1, day);
  const isoDay = date.getDay() === 0 ? 7 : date.getDay();
  return workDayLabel(isoDay, t) ?? dayISO;
}

function appointmentCardClass(status: Appointment["status"]) {
  if (status === "cancelled") return "bg-gray-100 border-gray-200";
  if (status === "completed") return "bg-green-50 border-green-200";
  return "bg-blue-50 border-blue-200";
}
