import { useMemo } from "react";
import type { ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { listAppointmentsDay } from "@/features/appointments/api";
import type { Appointment } from "@/features/appointments/types";
import { useAuth } from "@/features/auth/AuthProvider";
import { listKinesiologists } from "@/features/kinesiologists/api";
import type { Kinesiologist } from "@/features/kinesiologists/types";
import { listMaterialLoans, listMaterials } from "@/features/materials/api";
import { listPatients } from "@/features/patients/api";
import type { Patient } from "@/features/patients/types";
import { useLanguage } from "@/shared/i18n/LanguageProvider";
import { formatLocalDateTime, formatLocalTime } from "@/shared/time/format";

function todayISO() {
  const d = new Date();
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function patientName(patient: Patient | undefined, fallback: string) {
  return patient ? `${patient.last_name}, ${patient.first_name}` : fallback;
}

function kinesiologistName(kinesiologist: Kinesiologist | undefined, fallback: string) {
  return kinesiologist ? `${kinesiologist.last_name}, ${kinesiologist.first_name}` : fallback;
}

export default function DashboardPage() {
  const { user } = useAuth();
  const { t } = useLanguage();
  const date = todayISO();
  const isKinesiologist = user?.role === "kinesiologo";

  const kinesiosQ = useQuery({
    queryKey: ["kinesiologists", "dashboard"],
    queryFn: () => listKinesiologists(),
  });

  const patientsQ = useQuery({
    queryKey: ["patients", "dashboard"],
    queryFn: () => listPatients(500, true),
  });

  const materialsQ = useQuery({
    queryKey: ["materials", "dashboard"],
    queryFn: () => listMaterials(200),
  });

  const pendingLoansQ = useQuery({
    queryKey: ["materials", "loans", "dashboard", "pending"],
    queryFn: () => listMaterialLoans(true, 200),
  });

  const kinesios = kinesiosQ.data ?? [];
  const ownKinesiologist = useMemo(
    () =>
      user?.email
        ? kinesios.find((kinesiologist) => kinesiologist.email.toLowerCase() === user.email.toLowerCase())
        : undefined,
    [kinesios, user?.email],
  );
  const dashboardKinesios = isKinesiologist ? (ownKinesiologist ? [ownKinesiologist] : []) : kinesios;

  const appointmentsQ = useQuery({
    queryKey: ["appointments", "dashboard", date, dashboardKinesios.map((kinesiologist) => kinesiologist.id).join(",")],
    queryFn: async () => {
      const results = await Promise.all(
        dashboardKinesios.map((kinesiologist) =>
          listAppointmentsDay({ date, kinesiologist_id: kinesiologist.id }),
        ),
      );
      return results.flat();
    },
    enabled: dashboardKinesios.length > 0,
  });

  const patients = patientsQ.data ?? [];
  const materials = materialsQ.data ?? [];
  const pendingLoans = pendingLoansQ.data ?? [];
  const appointments = appointmentsQ.data ?? [];

  const patientById = useMemo(() => new Map(patients.map((patient) => [patient.id, patient])), [patients]);
  const kinesiologistById = useMemo(
    () => new Map(kinesios.map((kinesiologist) => [kinesiologist.id, kinesiologist])),
    [kinesios],
  );
  const materialById = useMemo(
    () => new Map(materials.map((material) => [material.id, material])),
    [materials],
  );

  const activePatients = patients.filter((patient) => patient.active).length;
  const scheduledToday = appointments.filter((appointment) => appointment.status === "scheduled");
  const lowStock = materials.filter((material) => material.total_qty > 0 && material.available_qty <= 1);
  const nextAppointments = scheduledToday
    .slice()
    .sort((a, b) => new Date(a.start_at).getTime() - new Date(b.start_at).getTime())
    .slice(0, 6);

  return (
    <main className="max-w-5xl mx-auto p-6 space-y-6">
      <header>
        <div>
          <h1 className="text-2xl font-semibold">{t("dashboard.title")}</h1>
          <p className="text-sm text-gray-600">{t("dashboard.activityToday")}</p>
        </div>
      </header>

      <section className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
        <Metric title={t("dashboard.todayAppointments")} value={scheduledToday.length} />
        <Metric title={t("dashboard.pendingMaterials")} value={pendingLoans.length} />
        <Metric title={t("dashboard.activePatients")} value={activePatients} />
        <Metric title={t("dashboard.lowStock")} value={lowStock.length} />
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Panel title={t("dashboard.upcomingAppointments")}>
          {appointmentsQ.isLoading && <p className="text-sm text-gray-600">{t("dashboard.loadingAppointments")}</p>}
          {appointmentsQ.isError && <p className="text-sm text-red-600">{t("dashboard.loadingAppointments")}</p>}
          {!appointmentsQ.isLoading && nextAppointments.length === 0 && (
            <p className="text-sm text-gray-600">{t("dashboard.noAppointmentsToday")}</p>
          )}
          <div className="divide-y">
            {nextAppointments.map((appointment) => (
              <AppointmentLine
                key={appointment.id}
                appointment={appointment}
                patient={patientById.get(appointment.patient_id)}
                kinesiologist={kinesiologistById.get(appointment.kinesiologist_id)}
                patientFallback={t("agenda.patient")}
                kinesiologistFallback={t("agenda.kinesiologist")}
                inPersonLabel={t("agenda.inPerson")}
                virtualLabel={t("agenda.virtual")}
                openVideoCallLabel={t("agenda.openVideoCall")}
              />
            ))}
          </div>
        </Panel>

        <Panel title={t("dashboard.pendingReturns")}>
          {pendingLoansQ.isLoading && <p className="text-sm text-gray-600">{t("dashboard.loadingPending")}</p>}
          {pendingLoansQ.isError && <p className="text-sm text-red-600">{t("dashboard.loadingPending")}</p>}
          {!pendingLoansQ.isLoading && pendingLoans.length === 0 && (
            <p className="text-sm text-gray-600">{t("dashboard.noPendingMaterials")}</p>
          )}
          <div className="divide-y">
            {pendingLoans.slice(0, 6).map((loan) => (
              <div key={loan.id} className="py-3">
                <div className="font-medium">{materialById.get(loan.material_id)?.name ?? t("dashboard.material")}</div>
                <div className="text-sm text-gray-600">
                  {patientName(patientById.get(loan.patient_id), t("agenda.patient"))} · {t("dashboard.quantity")}: {loan.qty}
                </div>
                <div className="text-xs text-gray-500">{t("dashboard.borrowed")}: {formatLocalDateTime(loan.loaned_at)}</div>
              </div>
            ))}
          </div>
        </Panel>
      </section>

      <Panel title={t("dashboard.stockAlerts")}>
        {materialsQ.isLoading && <p className="text-sm text-gray-600">{t("dashboard.loadingStock")}</p>}
        {materialsQ.isError && <p className="text-sm text-red-600">{t("dashboard.loadingStock")}</p>}
        {!materialsQ.isLoading && lowStock.length === 0 && (
          <p className="text-sm text-gray-600">{t("dashboard.noStockAlerts")}</p>
        )}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {lowStock.map((material) => (
            <div key={material.id} className="border rounded-lg p-3">
              <div className="font-medium">{material.name}</div>
              <div className="text-sm text-gray-600">
                {t("dashboard.available")}: {material.available_qty} / {material.total_qty}
              </div>
            </div>
          ))}
        </div>
      </Panel>
    </main>
  );
}

function Metric({ title, value }: { title: string; value: number }) {
  return (
    <section className="bg-white rounded-xl shadow p-4">
      <div className="text-sm text-gray-600">{title}</div>
      <div className="text-3xl font-semibold mt-1">{value}</div>
    </section>
  );
}

function Panel({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="bg-white rounded-xl shadow p-4 space-y-3">
      <h2 className="text-lg font-semibold">{title}</h2>
      {children}
    </section>
  );
}

function AppointmentLine({
  appointment,
  patient,
  kinesiologist,
  patientFallback,
  kinesiologistFallback,
  inPersonLabel,
  virtualLabel,
  openVideoCallLabel,
}: {
  appointment: Appointment;
  patient?: Patient;
  kinesiologist?: Kinesiologist;
  patientFallback: string;
  kinesiologistFallback: string;
  inPersonLabel: string;
  virtualLabel: string;
  openVideoCallLabel: string;
}) {
  return (
    <div className="py-3">
      <div className="font-medium">
        {formatLocalTime(appointment.start_at)} a {formatLocalTime(appointment.end_at)}
      </div>
      <div className="text-sm text-gray-600">{patientName(patient, patientFallback)}</div>
      <div className="text-xs text-gray-500">{kinesiologistName(kinesiologist, kinesiologistFallback)}</div>
      <div className="text-xs text-gray-500">
        {appointment.modality === "virtual" ? virtualLabel : inPersonLabel}
        {appointment.modality === "virtual" && appointment.video_call_url && (
          <>
            {" · "}
            <a className="underline text-blue-700" href={appointment.video_call_url} target="_blank" rel="noreferrer">
              {openVideoCallLabel}
            </a>
          </>
        )}
      </div>
    </div>
  );
}
