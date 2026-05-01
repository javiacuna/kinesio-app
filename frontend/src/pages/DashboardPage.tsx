import { useMemo } from "react";
import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { listAppointmentsDay } from "@/features/appointments/api";
import type { Appointment } from "@/features/appointments/types";
import { useAuth } from "@/features/auth/AuthProvider";
import { listKinesiologists } from "@/features/kinesiologists/api";
import type { Kinesiologist } from "@/features/kinesiologists/types";
import { listMaterialLoans, listMaterials } from "@/features/materials/api";
import { listPatients } from "@/features/patients/api";
import type { Patient } from "@/features/patients/types";
import { formatLocalDateTime, formatLocalTime } from "@/shared/time/format";

function todayISO() {
  const d = new Date();
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function patientName(patient?: Patient) {
  return patient ? `${patient.last_name}, ${patient.first_name}` : "Paciente";
}

function kinesiologistName(kinesiologist?: Kinesiologist) {
  return kinesiologist ? `${kinesiologist.last_name}, ${kinesiologist.first_name}` : "Kinesiólogo";
}

export default function DashboardPage() {
  const { user } = useAuth();
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
      <header className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Dashboard</h1>
          <p className="text-sm text-gray-600">Actividad operativa de hoy.</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Link className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-100" to="/agenda">
            Agenda
          </Link>
          <Link className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-100" to="/materials">
            Materiales
          </Link>
        </div>
      </header>

      <section className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
        <Metric title="Turnos hoy" value={scheduledToday.length} />
        <Metric title="Materiales pendientes" value={pendingLoans.length} />
        <Metric title="Pacientes activos" value={activePatients} />
        <Metric title="Bajo stock" value={lowStock.length} />
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Panel title="Próximos turnos">
          {appointmentsQ.isLoading && <p className="text-sm text-gray-600">Cargando turnos...</p>}
          {appointmentsQ.isError && <p className="text-sm text-red-600">No se pudieron cargar los turnos.</p>}
          {!appointmentsQ.isLoading && nextAppointments.length === 0 && (
            <p className="text-sm text-gray-600">No hay turnos programados para hoy.</p>
          )}
          <div className="divide-y">
            {nextAppointments.map((appointment) => (
              <AppointmentLine
                key={appointment.id}
                appointment={appointment}
                patient={patientById.get(appointment.patient_id)}
                kinesiologist={kinesiologistById.get(appointment.kinesiologist_id)}
              />
            ))}
          </div>
        </Panel>

        <Panel title="Pendientes de devolución">
          {pendingLoansQ.isLoading && <p className="text-sm text-gray-600">Cargando pendientes...</p>}
          {pendingLoansQ.isError && <p className="text-sm text-red-600">No se pudieron cargar los pendientes.</p>}
          {!pendingLoansQ.isLoading && pendingLoans.length === 0 && (
            <p className="text-sm text-gray-600">No hay materiales pendientes.</p>
          )}
          <div className="divide-y">
            {pendingLoans.slice(0, 6).map((loan) => (
              <div key={loan.id} className="py-3">
                <div className="font-medium">{materialById.get(loan.material_id)?.name ?? "Material"}</div>
                <div className="text-sm text-gray-600">
                  {patientName(patientById.get(loan.patient_id))} · Cantidad: {loan.qty}
                </div>
                <div className="text-xs text-gray-500">Prestado: {formatLocalDateTime(loan.loaned_at)}</div>
              </div>
            ))}
          </div>
        </Panel>
      </section>

      <Panel title="Alertas de stock">
        {materialsQ.isLoading && <p className="text-sm text-gray-600">Cargando stock...</p>}
        {materialsQ.isError && <p className="text-sm text-red-600">No se pudo cargar el stock.</p>}
        {!materialsQ.isLoading && lowStock.length === 0 && (
          <p className="text-sm text-gray-600">No hay materiales con bajo stock.</p>
        )}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {lowStock.map((material) => (
            <div key={material.id} className="border rounded-lg p-3">
              <div className="font-medium">{material.name}</div>
              <div className="text-sm text-gray-600">
                Disponible: {material.available_qty} / {material.total_qty}
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
}: {
  appointment: Appointment;
  patient?: Patient;
  kinesiologist?: Kinesiologist;
}) {
  return (
    <div className="py-3">
      <div className="font-medium">
        {formatLocalTime(appointment.start_at)} a {formatLocalTime(appointment.end_at)}
      </div>
      <div className="text-sm text-gray-600">{patientName(patient)}</div>
      <div className="text-xs text-gray-500">{kinesiologistName(kinesiologist)}</div>
    </div>
  );
}
