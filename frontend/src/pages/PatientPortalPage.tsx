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
  type PatientCheckIn,
  type PatientEvolution,
  createMyPatientCheckIn,
  downloadPatientAttachment,
  listMyPatientCheckIns,
  listMyPatientAttachments,
  listMyPatientEvolutions,
  listMyPatientPlans,
} from "@/features/patients/detailApi";
import { formatLocalDateTime, formatLocalTime } from "@/shared/time/format";
import { addMinutesToHHmm, localDateTimeToUTC } from "@/shared/time/rfc3339";
import { getYouTubeEmbedUrl } from "@/shared/media/youtube";
import { VideoPreviewModal, type VideoPreview } from "@/shared/ui/VideoPreviewModal";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

type Translate = (key: string) => string;

const defaultWorkDays = [1, 2, 3, 4, 5];
const workDayKeys: Record<number, string> = {
  1: "day.mon",
  2: "day.tue",
  3: "day.wed",
  4: "day.thu",
  5: "day.fri",
  6: "day.sat",
  7: "day.sun",
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

function workDaysLabel(days: number[] | undefined, t: Translate) {
  const selected = days?.length ? days : defaultWorkDays;
  return selected.map((day) => t(workDayKeys[day])).filter(Boolean).join(", ");
}

export default function PatientPortalPage() {
  const { user } = useAuth();
  const { t } = useLanguage();
  const range = useMemo(() => appointmentRange(), []);
  const [kinesiologistId, setKinesiologistId] = useState("");
  const [date, setDate] = useState(todayISO());
  const [startTime, setStartTime] = useState("09:00");
  const [durationMin, setDurationMin] = useState(45);
  const [notes, setNotes] = useState("");
  const [newAppointmentTouched, setNewAppointmentTouched] = useState(false);
  const [cancelTarget, setCancelTarget] = useState<Appointment | null>(null);
  const [cancelReason, setCancelReason] = useState("");
  const [checkInPainLevel, setCheckInPainLevel] = useState("");
  const [checkInMobilityScore, setCheckInMobilityScore] = useState("");
  const [checkInStrengthScore, setCheckInStrengthScore] = useState("");
  const [checkInFunctionalScore, setCheckInFunctionalScore] = useState("");
  const [checkInNotes, setCheckInNotes] = useState("");
  const [videoPreview, setVideoPreview] = useState<VideoPreview | null>(null);

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

  const checkInsQ = useQuery({
    queryKey: ["patients", "me", "check-ins"],
    queryFn: () => listMyPatientCheckIns(),
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
  const checkIns = checkInsQ.data ?? [];
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

  const createCheckInM = useMutation({
    mutationFn: () =>
      createMyPatientCheckIn({
        pain_level: checkInPainLevel === "" ? null : Number(checkInPainLevel),
        mobility_score: checkInMobilityScore === "" ? null : Number(checkInMobilityScore),
        strength_score: checkInStrengthScore === "" ? null : Number(checkInStrengthScore),
        functional_score: checkInFunctionalScore === "" ? null : Number(checkInFunctionalScore),
        notes: checkInNotes.trim(),
      }),
    onSuccess: () => {
      setCheckInPainLevel("");
      setCheckInMobilityScore("");
      setCheckInStrengthScore("");
      setCheckInFunctionalScore("");
      setCheckInNotes("");
      checkInsQ.refetch();
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
        <h1 className="text-2xl font-semibold">{t("portal.title")}</h1>
        <p className="text-sm text-gray-600">{user?.email}</p>
      </header>

      <section className="bg-white border rounded-lg p-4 space-y-4">
        <h2 className="text-lg font-semibold">{t("portal.newAppointment")}</h2>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div className="md:col-span-2">
            <label className="text-sm font-medium">{t("portal.kinesiologist")}</label>
            <select
              className="mt-1 w-full border rounded-lg p-2"
              value={kinesiologistId}
              onChange={(e) => {
                setKinesiologistId(e.target.value);
                setNewAppointmentTouched(true);
              }}
            >
              <option value="">{t("portal.select")}</option>
              {kinesiologists.map((kinesiologist) => (
                <option key={kinesiologist.id} value={kinesiologist.id}>
                  {kinesiologist.last_name}, {kinesiologist.first_name}
                </option>
              ))}
            </select>
            {kinesiologistsQ.isError && (
              <p className="text-sm text-red-600 mt-1">{t("portal.errorLoadKinesiologists")}</p>
            )}
            {selectedKinesiologist && (
              <p className="text-sm text-gray-600 mt-2">
                {t("portal.workingHours")}: {selectedKinesiologist.work_start_time} {t("portal.errorOutsideScheduleMiddle")} {selectedKinesiologist.work_end_time} · {workDaysLabel(selectedKinesiologist.work_days, t)}.
              </p>
            )}
          </div>

          <div>
            <label className="text-sm font-medium">{t("portal.date")}</label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              type="date"
              min={todayISO()}
              value={date}
              onChange={(e) => {
                setDate(e.target.value);
                setNewAppointmentTouched(true);
              }}
            />
          </div>

          <div>
            <label className="text-sm font-medium">{t("portal.time")}</label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              type="time"
              value={startTime}
              onChange={(e) => {
                setStartTime(e.target.value);
                setNewAppointmentTouched(true);
              }}
            />
          </div>

          <div>
            <label className="text-sm font-medium">{t("portal.duration")}</label>
            <select
              className="mt-1 w-full border rounded-lg p-2"
              value={durationMin}
              onChange={(e) => {
                setDurationMin(Number(e.target.value));
                setNewAppointmentTouched(true);
              }}
            >
              <option value={30}>30 {t("portal.min")}</option>
              <option value={45}>45 {t("portal.min")}</option>
              <option value={60}>60 {t("portal.min")}</option>
            </select>
          </div>

          <div>
            <label className="text-sm font-medium">{t("portal.notes")}</label>
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
          {createM.isPending ? t("portal.booking") : t("portal.bookAppointment")}
        </button>

        {newAppointmentTouched && isCreateInPast && (
          <p className="text-sm text-red-600">{t("portal.errorPastTime")}</p>
        )}

        {newAppointmentTouched && isCreateOutsideWorkingSchedule && selectedKinesiologist && (
          <p className="text-sm text-red-600">
            {t("portal.errorOutsideSchedulePrefix")} {selectedKinesiologist.work_start_time} {t("portal.errorOutsideScheduleMiddle")} {selectedKinesiologist.work_end_time} {t("portal.errorOutsideScheduleSuffix")} {workDaysLabel(selectedKinesiologist.work_days, t)}.
          </p>
        )}

        {createM.isError && (
          <div className="border border-red-200 bg-red-50 rounded-lg p-3 text-sm text-red-700">
            {isOverlapError(createM.error)
              ? t("portal.errorOverlap")
              : `${t("portal.errorBookFailedPrefix")} ${(createM.error as any)?.message}`}
          </div>
        )}
      </section>

      <section className="bg-white border rounded-lg p-4 space-y-4">
        <div>
          <h2 className="text-lg font-semibold">{t("portal.checkInTitle")}</h2>
          <p className="text-sm text-gray-600">{t("portal.checkInSubtitle")}</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
          <div>
            <label className="text-sm font-medium">{t("portal.pain")}</label>
            <select
              className="mt-1 w-full border rounded-lg p-2"
              value={checkInPainLevel}
              onChange={(event) => setCheckInPainLevel(event.target.value)}
            >
              <option value="">{t("portal.notRegistered")}</option>
              {Array.from({ length: 11 }, (_, value) => (
                <option key={value} value={value}>
                  {value}/10
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="text-sm font-medium">{t("portal.mobility")}</label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              type="number"
              min={0}
              max={100}
              placeholder={t("portal.rangePlaceholder")}
              value={checkInMobilityScore}
              onChange={(event) => setCheckInMobilityScore(event.target.value)}
            />
          </div>

          <div>
            <label className="text-sm font-medium">{t("portal.strength")}</label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              type="number"
              min={0}
              max={100}
              placeholder={t("portal.rangePlaceholder")}
              value={checkInStrengthScore}
              onChange={(event) => setCheckInStrengthScore(event.target.value)}
            />
          </div>

          <div>
            <label className="text-sm font-medium">{t("portal.functional")}</label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              type="number"
              min={0}
              max={100}
              placeholder={t("portal.rangePlaceholder")}
              value={checkInFunctionalScore}
              onChange={(event) => setCheckInFunctionalScore(event.target.value)}
            />
          </div>

          <div className="md:col-span-4">
            <label className="text-sm font-medium">{t("portal.comment")}</label>
            <textarea
              className="mt-1 w-full border rounded-lg p-2 min-h-24"
              value={checkInNotes}
              onChange={(event) => setCheckInNotes(event.target.value)}
              placeholder={t("portal.checkInNotesPlaceholder")}
            />
          </div>
        </div>

        <button
          type="button"
          className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
          disabled={!checkInNotes.trim() || createCheckInM.isPending}
          onClick={() => createCheckInM.mutate()}
        >
          {createCheckInM.isPending ? t("portal.sending") : t("portal.sendCheckIn")}
        </button>

        {createCheckInM.isError && (
          <p className="text-sm text-red-600">{t("portal.errorCheckInFailed")}</p>
        )}
        {createCheckInM.isSuccess && (
          <p className="text-sm text-green-700">{t("portal.checkInSent")}</p>
        )}

        <div className="border-t pt-4">
          <h3 className="font-medium">{t("portal.myCheckIns")}</h3>
          {checkInsQ.isLoading && <p className="text-sm text-gray-600 mt-2">{t("portal.loadingCheckIns")}</p>}
          {checkInsQ.isError && <p className="text-sm text-red-600 mt-2">{t("portal.errorLoadCheckIns")}</p>}
          {!checkInsQ.isLoading && !checkInsQ.isError && checkIns.length === 0 && (
            <p className="text-sm text-gray-600 mt-2">{t("portal.noCheckIns")}</p>
          )}
          {checkIns.length > 0 && (
            <div className="mt-3 divide-y">
              {checkIns.map((checkIn) => (
                <PatientCheckInLine key={checkIn.id} checkIn={checkIn} />
              ))}
            </div>
          )}
        </div>
      </section>

      <section className="bg-white border rounded-lg p-4">
        <h2 className="text-lg font-semibold">{t("portal.myEvolution")}</h2>

        {evolutionsQ.isLoading && (
          <p className="text-sm text-gray-600 mt-2">{t("portal.loadingEvolution")}</p>
        )}

        {evolutionsQ.isError && (
          <p className="text-sm text-red-600 mt-2">{t("portal.errorLoadEvolution")}</p>
        )}

        {evolutionsQ.isSuccess && evolutions.length === 0 && (
          <p className="text-sm text-gray-600 mt-2">{t("portal.noEvolution")}</p>
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
        <h2 className="text-lg font-semibold">{t("portal.myPlans")}</h2>

        {plansQ.isLoading && (
          <p className="text-sm text-gray-600 mt-2">{t("portal.loadingPlans")}</p>
        )}

        {plansQ.isError && (
          <p className="text-sm text-red-600 mt-2">{t("portal.errorLoadPlans")}</p>
        )}

        {plansQ.isSuccess && plans.length === 0 && (
          <p className="text-sm text-gray-600 mt-2">{t("portal.noPlans")}</p>
        )}

        {plans.length > 0 && (
          <div className="mt-3 divide-y">
            {plans.map((plan) => (
              <article key={plan.id} className="py-3">
                <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
                  <div className="font-medium">
                    {plan.frequency === "daily" ? t("portal.daily") : t("portal.weekly")} · {plan.duration_weeks} {t("portal.weeks")}
                  </div>
                  <span className="text-xs px-2 py-1 rounded-md bg-gray-100 text-gray-700 w-fit">
                    {plan.status === "closed" ? t("portal.closed") : t("portal.active")}
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
                        {item.estimated_minutes} {t("portal.min")}
                        {item.sets ? ` · ${item.sets} ${t("portal.sets")}` : ""}
                        {item.reps ? ` · ${item.reps} ${t("portal.reps")}` : ""}
                      </div>
                      {item.description && (
                        <p className="text-sm text-gray-700 mt-1">{item.description}</p>
                      )}
                      {(item.video_url || item.guide_url) && (
                        <div className="mt-2 flex flex-wrap gap-2">
                          {item.video_url && (
                            getYouTubeEmbedUrl(item.video_url) ? (
                              <button
                                type="button"
                                className="text-sm underline text-blue-700"
                                onClick={() => setVideoPreview({ title: item.name, url: item.video_url! })}
                              >
                                {t("portal.watchVideo")}
                              </button>
                            ) : (
                              <a className="text-sm underline text-blue-700" href={item.video_url} target="_blank" rel="noreferrer">
                                {t("portal.watchVideo")}
                              </a>
                            )
                          )}
                          {item.guide_url && (
                            <a className="text-sm underline text-blue-700" href={item.guide_url} target="_blank" rel="noreferrer">
                              {t("portal.watchGuide")}
                            </a>
                          )}
                        </div>
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
        <h2 className="text-lg font-semibold">{t("portal.myStudies")}</h2>

        {attachmentsQ.isLoading && (
          <p className="text-sm text-gray-600 mt-2">{t("portal.loadingStudies")}</p>
        )}

        {attachmentsQ.isError && (
          <p className="text-sm text-red-600 mt-2">{t("portal.errorLoadStudies")}</p>
        )}

        {attachmentsQ.isSuccess && attachments.length === 0 && (
          <p className="text-sm text-gray-600 mt-2">{t("portal.noStudies")}</p>
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
                  {t("portal.viewFile")}
                </button>
              </article>
            ))}
          </div>
        )}
      </section>

      <section className="bg-white border rounded-lg p-4">
        <h2 className="text-lg font-semibold">{t("portal.myAppointments")}</h2>

        {appointmentsQ.isLoading && (
          <p className="text-sm text-gray-600 mt-2">{t("portal.loadingAppointments")}</p>
        )}

        {appointmentsQ.isError && (
          <p className="text-sm text-red-600 mt-2">
            {t("portal.errorLoadAppointments")}
          </p>
        )}

        {appointmentsQ.isSuccess && appointments.length === 0 && (
          <p className="text-sm text-gray-600 mt-2">{t("portal.noUpcomingAppointments")}</p>
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
                    {appointment.modality === "virtual" ? t("portal.videocall") : t("portal.inPerson")}
                    {appointment.modality === "virtual" && appointment.video_call_url && (
                      <>
                        {" · "}
                        <a className="underline text-blue-700" href={appointment.video_call_url} target="_blank" rel="noreferrer">
                          {t("portal.openVideocall")}
                        </a>
                      </>
                    )}
                  </div>
                  {appointment.notes && (
                    <div className="text-sm text-gray-600 mt-1">{t("portal.notes")}: {appointment.notes}</div>
                  )}
                  {appointment.status === "cancelled" && appointment.cancelled_reason && (
                    <div className="text-sm text-gray-600 mt-1">{t("portal.reasonLabel")}: {appointment.cancelled_reason}</div>
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
                    {appointment.status === "cancelled" ? t("portal.cancelled") : t("portal.scheduled")}
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
                      {t("common.cancel")}
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
              <h3 className="font-medium">{t("portal.cancelAppointmentTitle")}</h3>
              <p className="text-sm text-gray-600">{formatLocalDateTime(cancelTarget.start_at)}</p>
            </div>
            <div>
              <label className="text-sm font-medium">{t("portal.reasonLabel")}</label>
              <textarea
                className="mt-1 w-full border rounded-lg p-2 min-h-20 bg-white"
                value={cancelReason}
                onChange={(event) => setCancelReason(event.target.value)}
                placeholder={t("portal.optional")}
              />
            </div>
            <div className="flex flex-col gap-2 sm:flex-row">
              <button
                type="button"
                className="px-3 py-2 rounded-lg bg-black text-white text-sm disabled:opacity-50"
                disabled={cancelM.isPending}
                onClick={() => cancelM.mutate({ id: cancelTarget.id, reason: cancelReason.trim() || undefined })}
              >
                {cancelM.isPending ? t("portal.cancelling") : t("portal.confirmCancel")}
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
                {t("common.back")}
              </button>
            </div>
          </div>
        )}

        {cancelM.isError && (
          <p className="text-sm text-red-600 mt-2">
            {t("portal.errorCancelFailedPrefix")} {(cancelM.error as any)?.message}
          </p>
        )}
      </section>

      <VideoPreviewModal preview={videoPreview} onClose={() => setVideoPreview(null)} />
    </main>
  );
}

type EvolutionMetricKey = "pain_level" | "mobility_score" | "strength_score" | "functional_score";

function getEvolutionMetricDefs(t: Translate): Array<{
  key: EvolutionMetricKey;
  label: string;
  color: string;
  max: number;
}> {
  return [
    { key: "pain_level", label: t("portal.pain"), color: "#dc2626", max: 10 },
    { key: "mobility_score", label: t("portal.mobility"), color: "#2563eb", max: 100 },
    { key: "strength_score", label: t("portal.strength"), color: "#16a34a", max: 100 },
    { key: "functional_score", label: t("portal.functional"), color: "#9333ea", max: 100 },
  ];
}

function PatientEvolutionChart({ evolutions }: { evolutions: PatientEvolution[] }) {
  const { t } = useLanguage();
  const evolutionMetricDefs = getEvolutionMetricDefs(t);
  const points = [...evolutions]
    .filter((evolution) => evolutionMetricDefs.some((metric) => evolution[metric.key] != null))
    .sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());

  if (points.length === 0) {
    return <p className="text-sm text-gray-600">{t("portal.noMetricsToChart")}</p>;
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
        <svg className="min-w-[620px] w-full" viewBox={`0 0 ${width} ${height}`} role="img" aria-label={t("portal.evolutionChartAriaLabel")}>
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

          {points.map((evolution, index) => {
            const anchor =
              points.length === 1 ? "middle" : index === 0 ? "start" : index === points.length - 1 ? "end" : "middle";
            return (
              <text key={evolution.id} x={xFor(index)} y={height - 18} textAnchor={anchor} className="fill-gray-600 text-[11px]">
                {formatLocalDateTime(evolution.created_at)}
              </text>
            );
          })}
        </svg>
      </div>
    </div>
  );
}

function EvolutionMetricLine({ evolution }: { evolution: PatientEvolution }) {
  const { t } = useLanguage();
  const values = getEvolutionMetricDefs(t)
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

function PatientCheckInLine({ checkIn }: { checkIn: PatientCheckIn }) {
  return (
    <article className="py-3">
      <div className="font-medium">{formatLocalDateTime(checkIn.created_at)}</div>
      <CheckInMetricLine checkIn={checkIn} />
      <p className="text-sm text-gray-700 mt-2">{checkIn.notes}</p>
    </article>
  );
}

function CheckInMetricLine({ checkIn }: { checkIn: PatientCheckIn }) {
  const { t } = useLanguage();
  const values = [
    checkIn.pain_level == null ? null : `${t("portal.pain")}: ${checkIn.pain_level}/10`,
    checkIn.mobility_score == null ? null : `${t("portal.mobility")}: ${checkIn.mobility_score}/100`,
    checkIn.strength_score == null ? null : `${t("portal.strength")}: ${checkIn.strength_score}/100`,
    checkIn.functional_score == null ? null : `${t("portal.functional")}: ${checkIn.functional_score}/100`,
  ].filter((value): value is string => value != null);

  if (values.length === 0) return null;

  return (
    <div className="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-sm text-gray-600">
      {values.map((value) => (
        <span key={value}>{value}</span>
      ))}
    </div>
  );
}
