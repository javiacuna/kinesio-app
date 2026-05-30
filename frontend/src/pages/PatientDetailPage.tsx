import { useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { inviteUserAccess } from "@/features/auth/adminApi";
import { useAuth } from "@/features/auth/AuthProvider";
import { listPatientAppointments } from "@/features/appointments/api";
import { listKinesiologists } from "@/features/kinesiologists/api";
import { formatLocalDateTime, formatLocalTime } from "@/shared/time/format";
import { archivePatient, getPatient, updatePatient } from "@/features/patients/api";
import {
  type ExercisePlan,
  type PatientDiagnosis,
  type PatientAttachment,
  type PatientEvolution,
  createPatientDiagnosis,
  createPatientPlan,
  createPatientEvolution,
  deletePatientDiagnosis,
  deletePatientAttachment,
  downloadPatientAttachment,
  listPatientDiagnoses,
  listPatientAttachments,
  listPatientEvolutions,
  listPatientMaterialLoans,
  listPatientPlans,
  searchCIE10Codes,
  updatePatientDiagnosis,
  updatePatientAttachment,
  updatePatientPlan,
  uploadPatientAttachment,
} from "@/features/patients/detailApi";

type PlanItemForm = {
  name: string;
  estimated_minutes: number;
  sets: string;
  reps: string;
  description: string;
};

type AttachmentEditForm = {
  id: string;
  file_name: string;
  notes: string;
  category: string;
  patient_visible: boolean;
  confirm_delete: boolean;
};

const attachmentCategories = [
  { value: "otro", label: "Otro" },
  { value: "radiografia", label: "Radiografía" },
  { value: "resonancia", label: "Resonancia" },
  { value: "ecografia", label: "Ecografía" },
  { value: "laboratorio", label: "Laboratorio" },
  { value: "foto", label: "Foto" },
  { value: "video", label: "Video" },
  { value: "documento", label: "Documento" },
];

function historyRange() {
  const from = new Date();
  from.setFullYear(from.getFullYear() - 1);
  from.setHours(0, 0, 0, 0);

  const to = new Date();
  to.setFullYear(to.getFullYear() + 1);
  to.setHours(23, 59, 59, 999);

  return { from: from.toISOString(), to: to.toISOString() };
}

function todayISO() {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
}

export default function PatientDetailPage() {
  const { patientId = "" } = useParams();
  const { user } = useAuth();
  const canManagePatient = user?.role === "admin" || user?.role === "recepcionista";
  const canEditClinical = user?.role === "admin" || user?.role === "recepcionista" || user?.role === "kinesiologo";
  const canCreateEvolution = user?.role === "admin" || user?.role === "kinesiologo";
  const backPath = user?.role === "kinesiologo" ? "/agenda" : "/patients";
  const range = useMemo(() => historyRange(), []);
  const [phone, setPhone] = useState("");
  const [birthDate, setBirthDate] = useState("");
  const [clinicalNotes, setClinicalNotes] = useState("");
  const [cie10Query, setCie10Query] = useState("");
  const [diagnosisCode, setDiagnosisCode] = useState("");
  const [diagnosisKind, setDiagnosisKind] = useState<"primary" | "secondary">("primary");
  const [diagnosisStatus, setDiagnosisStatus] = useState<"suspected" | "confirmed" | "resolved">("suspected");
  const [diagnosisDate, setDiagnosisDate] = useState(todayISO());
  const [diagnosisNotes, setDiagnosisNotes] = useState("");
  const [editingDiagnosisId, setEditingDiagnosisId] = useState("");
  const [selectedClinicalDiagnosisId, setSelectedClinicalDiagnosisId] = useState("");
  const [evolutionKinesiologistId, setEvolutionKinesiologistId] = useState("");
  const [evolutionAppointmentId, setEvolutionAppointmentId] = useState("");
  const [evolutionDiagnosisId, setEvolutionDiagnosisId] = useState("");
  const [painLevel, setPainLevel] = useState("");
  const [mobilityScore, setMobilityScore] = useState("");
  const [strengthScore, setStrengthScore] = useState("");
  const [functionalScore, setFunctionalScore] = useState("");
  const [evolutionNotes, setEvolutionNotes] = useState("");
  const [planKinesiologistId, setPlanKinesiologistId] = useState("");
  const [planDiagnosisId, setPlanDiagnosisId] = useState("");
  const [planFrequency, setPlanFrequency] = useState<"daily" | "weekly">("weekly");
  const [planStatus, setPlanStatus] = useState<"active" | "closed">("active");
  const [planDurationWeeks, setPlanDurationWeeks] = useState(4);
  const [planObservations, setPlanObservations] = useState("");
  const [editingPlanId, setEditingPlanId] = useState("");
  const [attachmentFile, setAttachmentFile] = useState<File | null>(null);
  const [attachmentNotes, setAttachmentNotes] = useState("");
  const [attachmentCategory, setAttachmentCategory] = useState("otro");
  const [attachmentPatientVisible, setAttachmentPatientVisible] = useState(false);
  const [editingAttachment, setEditingAttachment] = useState<AttachmentEditForm | null>(null);
  const [planItems, setPlanItems] = useState<PlanItemForm[]>([
    { name: "", estimated_minutes: 10, sets: "", reps: "", description: "" },
  ]);

  const patientQ = useQuery({
    queryKey: ["patients", "detail", patientId],
    queryFn: () => getPatient(patientId),
    enabled: Boolean(patientId),
  });

  const appointmentsQ = useQuery({
    queryKey: ["appointments", "patient-detail", patientId, range.from, range.to],
    queryFn: () => listPatientAppointments({ patient_id: patientId, from: range.from, to: range.to }),
    enabled: Boolean(patientId),
  });

  const plansQ = useQuery({
    queryKey: ["patients", "plans", patientId],
    queryFn: () => listPatientPlans(patientId),
    enabled: Boolean(patientId),
  });

  const evolutionsQ = useQuery({
    queryKey: ["patients", "evolutions", patientId],
    queryFn: () => listPatientEvolutions(patientId),
    enabled: Boolean(patientId),
  });

  const loansQ = useQuery({
    queryKey: ["patients", "material-loans", patientId],
    queryFn: () => listPatientMaterialLoans(patientId),
    enabled: Boolean(patientId),
  });

  const attachmentsQ = useQuery({
    queryKey: ["patients", "attachments", patientId],
    queryFn: () => listPatientAttachments(patientId),
    enabled: Boolean(patientId) && canEditClinical,
  });

  const diagnosesQ = useQuery({
    queryKey: ["patients", "diagnoses", patientId],
    queryFn: () => listPatientDiagnoses(patientId),
    enabled: Boolean(patientId) && canEditClinical,
  });

  const cie10Q = useQuery({
    queryKey: ["cie10", cie10Query],
    queryFn: () => searchCIE10Codes(cie10Query, 30),
    enabled: canEditClinical,
  });

  const kinesiologistsQ = useQuery({
    queryKey: ["kinesiologists", "patient-detail"],
    queryFn: () => listKinesiologists(),
    enabled: canCreateEvolution,
  });

  const archiveM = useMutation({
    mutationFn: archivePatient,
    onSuccess: () => patientQ.refetch(),
  });

  const reactivateM = useMutation({
    mutationFn: async () => {
      const patient = patientQ.data;
      if (!patient) return;
      await updatePatient({
        id: patient.id,
        dni: patient.dni,
        first_name: patient.first_name,
        last_name: patient.last_name,
        email: patient.email,
        phone: patient.phone ?? null,
        birth_date: patient.birth_date ?? null,
        clinical_notes: patient.clinical_notes ?? null,
        active: true,
      });
    },
    onSuccess: () => patientQ.refetch(),
  });

  const inviteM = useMutation({
    mutationFn: (email: string) => inviteUserAccess({ email, role: "paciente" }),
  });

  const updateClinicalM = useMutation({
    mutationFn: async () => {
      const patient = patientQ.data;
      if (!patient) return;
      await updatePatient({
        id: patient.id,
        dni: patient.dni,
        first_name: patient.first_name,
        last_name: patient.last_name,
        email: patient.email,
        phone: phone.trim() || null,
        birth_date: birthDate || null,
        clinical_notes: clinicalNotes.trim() || null,
        active: patient.active,
      });
    },
    onSuccess: () => patientQ.refetch(),
  });

  const saveDiagnosisM = useMutation({
    mutationFn: () => {
      const input = {
        cie10_code: diagnosisCode,
        kind: diagnosisKind,
        status: diagnosisStatus,
        diagnosed_at: diagnosisDate,
        notes: diagnosisNotes.trim() || null,
      };
      return editingDiagnosisId
        ? updatePatientDiagnosis(patientId, editingDiagnosisId, input)
        : createPatientDiagnosis(patientId, input);
    },
    onSuccess: () => {
      resetDiagnosisForm();
      diagnosesQ.refetch();
    },
  });

  const deleteDiagnosisM = useMutation({
    mutationFn: (diagnosisId: string) => deletePatientDiagnosis(patientId, diagnosisId),
    onSuccess: () => {
      resetDiagnosisForm();
      diagnosesQ.refetch();
    },
  });

  const createEvolutionM = useMutation({
    mutationFn: () =>
      createPatientEvolution(patientId, {
        kinesiologist_id: evolutionKinesiologistId,
        appointment_id: evolutionAppointmentId || null,
        patient_diagnosis_id: evolutionDiagnosisId || null,
        pain_level: painLevel === "" ? null : Number(painLevel),
        mobility_score: mobilityScore === "" ? null : Number(mobilityScore),
        strength_score: strengthScore === "" ? null : Number(strengthScore),
        functional_score: functionalScore === "" ? null : Number(functionalScore),
        notes: evolutionNotes,
      }),
    onSuccess: () => {
      setEvolutionAppointmentId("");
      setEvolutionDiagnosisId("");
      setPainLevel("");
      setMobilityScore("");
      setStrengthScore("");
      setFunctionalScore("");
      setEvolutionNotes("");
      evolutionsQ.refetch();
    },
  });

  const savePlanM = useMutation({
    mutationFn: () =>
      editingPlanId
        ? updatePatientPlan(editingPlanId, {
            kinesiologist_id: planKinesiologistId,
            patient_diagnosis_id: planDiagnosisId || null,
            frequency: planFrequency,
            duration_weeks: planDurationWeeks,
            observations: planObservations.trim() || null,
            status: planStatus,
            items: planItems.map((item) => ({
              name: item.name.trim(),
              estimated_minutes: Number(item.estimated_minutes),
              sets: item.sets === "" ? null : Number(item.sets),
              reps: item.reps === "" ? null : Number(item.reps),
              description: item.description.trim() || null,
            })),
          })
        : createPatientPlan(patientId, {
        kinesiologist_id: planKinesiologistId,
        patient_diagnosis_id: planDiagnosisId || null,
        frequency: planFrequency,
        duration_weeks: planDurationWeeks,
        observations: planObservations.trim() || null,
        items: planItems.map((item) => ({
          name: item.name.trim(),
          estimated_minutes: Number(item.estimated_minutes),
          sets: item.sets === "" ? null : Number(item.sets),
          reps: item.reps === "" ? null : Number(item.reps),
          description: item.description.trim() || null,
        })),
      }),
    onSuccess: () => {
      resetPlanForm();
      plansQ.refetch();
    },
  });

  const uploadAttachmentM = useMutation({
    mutationFn: () => {
      if (!attachmentFile) throw new Error("file_required");
      return uploadPatientAttachment(patientId, attachmentFile, {
        notes: attachmentNotes,
        category: attachmentCategory,
        patient_visible: attachmentPatientVisible,
      });
    },
    onSuccess: () => {
      setAttachmentFile(null);
      setAttachmentNotes("");
      setAttachmentCategory("otro");
      setAttachmentPatientVisible(false);
      attachmentsQ.refetch();
    },
  });

  const updateAttachmentM = useMutation({
    mutationFn: (attachment: AttachmentEditForm) =>
      updatePatientAttachment(attachment.id, {
        file_name: attachment.file_name,
        notes: attachment.notes.trim() || null,
        category: attachment.category || "otro",
        patient_visible: attachment.patient_visible,
      }),
    onSuccess: () => {
      setEditingAttachment(null);
      attachmentsQ.refetch();
    },
  });

  const deleteAttachmentM = useMutation({
    mutationFn: (attachmentId: string) => deletePatientAttachment(attachmentId),
    onSuccess: () => attachmentsQ.refetch(),
  });

  const patient = patientQ.data;
  const appointments = appointmentsQ.data ?? [];
  const plans = plansQ.data ?? [];
  const evolutions = evolutionsQ.data ?? [];
  const loans = loansQ.data ?? [];
  const attachments = attachmentsQ.data ?? [];
  const diagnoses = diagnosesQ.data ?? [];
  const cie10Results = cie10Q.data ?? [];
  const kinesiologists = kinesiologistsQ.data ?? [];
  const diagnosisById = useMemo(() => {
    const index = new Map<string, PatientDiagnosis>();
    for (const diagnosis of diagnoses) {
      index.set(diagnosis.id, diagnosis);
    }
    return index;
  }, [diagnoses]);
  const selectedClinicalDiagnosis = useMemo(
    () => (selectedClinicalDiagnosisId ? diagnosisById.get(selectedClinicalDiagnosisId) : undefined),
    [diagnosisById, selectedClinicalDiagnosisId],
  );
  const filteredDiagnoses = useMemo(
    () => (selectedClinicalDiagnosis ? [selectedClinicalDiagnosis] : diagnoses),
    [diagnoses, selectedClinicalDiagnosis],
  );
  const filteredEvolutions = useMemo(
    () =>
      selectedClinicalDiagnosisId
        ? evolutions.filter((evolution) => evolution.patient_diagnosis_id === selectedClinicalDiagnosisId)
        : evolutions,
    [evolutions, selectedClinicalDiagnosisId],
  );
  const filteredPlans = useMemo(
    () =>
      selectedClinicalDiagnosisId
        ? plans.filter((plan) => plan.patient_diagnosis_id === selectedClinicalDiagnosisId)
        : plans,
    [plans, selectedClinicalDiagnosisId],
  );
  const upcomingScheduledAppointments = appointments.filter(
    (appointment) => appointment.status !== "cancelled",
  );
  const timeline = useMemo(
    () =>
      [
        ...(selectedClinicalDiagnosisId
          ? []
          : appointments.map((appointment) => ({
              id: `appointment:${appointment.id}`,
              at: appointment.start_at,
              kind: "Turno",
              title:
                appointment.status === "cancelled"
                  ? "Turno cancelado"
                  : "Turno programado",
              detail: `${formatLocalTime(appointment.start_at)} a ${formatLocalTime(appointment.end_at)}${appointment.notes ? ` · ${appointment.notes}` : ""}`,
            }))),
        ...filteredEvolutions.map((evolution) => ({
          id: `evolution:${evolution.id}`,
          at: evolution.created_at,
          kind: "Evolución",
          title: evolutionMetricSummary(evolution) || "Evolución clínica",
          detail: `${diagnosisSummary(diagnosisById.get(evolution.patient_diagnosis_id ?? ""))}${evolution.notes}`,
        })),
        ...filteredPlans.map((plan) => ({
          id: `plan:${plan.id}`,
          at: plan.created_at,
          kind: "Plan",
          title: `${plan.frequency} · ${plan.duration_weeks} semanas`,
          detail: `${diagnosisSummary(diagnosisById.get(plan.patient_diagnosis_id ?? ""))}${plan.items.length} ejercicios · estado ${plan.status}`,
        })),
        ...filteredDiagnoses.map((diagnosis) => ({
          id: `diagnosis:${diagnosis.id}`,
          at: diagnosis.created_at,
          kind: "Diagnóstico",
          title: `${diagnosis.cie10_code} · ${diagnosis.cie10_description}`,
          detail: `${diagnosisKindLabel(diagnosis.kind)} · ${diagnosisStatusLabel(diagnosis.status)}${diagnosis.notes ? ` · ${diagnosis.notes}` : ""}`,
        })),
        ...(selectedClinicalDiagnosisId
          ? []
          : loans.map((loan) => ({
              id: `loan:${loan.id}`,
              at: loan.loaned_at,
              kind: "Material",
              title: loan.returned_at ? "Material devuelto" : "Material prestado",
              detail: `Cantidad ${loan.qty}${loan.notes ? ` · ${loan.notes}` : ""}`,
            }))),
        ...(selectedClinicalDiagnosisId
          ? []
          : attachments.map((attachment) => ({
              id: `attachment:${attachment.id}`,
              at: attachment.created_at,
              kind: "Archivo",
              title: attachment.file_name,
              detail: `${attachment.kind}${attachment.uploaded_by_email ? ` · subido por ${attachment.uploaded_by_email}` : ""}`,
            }))),
      ].sort((a, b) => new Date(b.at).getTime() - new Date(a.at).getTime()),
    [
      appointments,
      attachments,
      diagnosisById,
      filteredDiagnoses,
      filteredEvolutions,
      filteredPlans,
      loans,
      selectedClinicalDiagnosisId,
    ],
  );

  useEffect(() => {
    if (evolutionKinesiologistId || kinesiologists.length === 0) return;

    const ownProfile =
      user?.email
        ? kinesiologists.find((kinesiologist) => kinesiologist.email.toLowerCase() === user.email.toLowerCase())
        : undefined;
    setEvolutionKinesiologistId((ownProfile ?? kinesiologists[0]).id);
  }, [evolutionKinesiologistId, kinesiologists, user?.email]);

  useEffect(() => {
    if (planKinesiologistId || kinesiologists.length === 0) return;

    const ownProfile =
      user?.email
        ? kinesiologists.find((kinesiologist) => kinesiologist.email.toLowerCase() === user.email.toLowerCase())
        : undefined;
    setPlanKinesiologistId((ownProfile ?? kinesiologists[0]).id);
  }, [kinesiologists, planKinesiologistId, user?.email]);

  useEffect(() => {
    if (!patient) return;
    setPhone(patient.phone ?? "");
    setBirthDate(patient.birth_date ?? "");
    setClinicalNotes(patient.clinical_notes ?? "");
  }, [patient]);

  useEffect(() => {
    if (selectedClinicalDiagnosisId && !diagnosisById.has(selectedClinicalDiagnosisId)) {
      setSelectedClinicalDiagnosisId("");
    }
  }, [diagnosisById, selectedClinicalDiagnosisId]);

  const hasValidPlanItems = planItems.every((item) => item.name.trim() && Number(item.estimated_minutes) > 0);

  function resetPlanForm() {
    setEditingPlanId("");
    setPlanFrequency("weekly");
    setPlanStatus("active");
    setPlanDiagnosisId("");
    setPlanDurationWeeks(4);
    setPlanObservations("");
    setPlanItems([{ name: "", estimated_minutes: 10, sets: "", reps: "", description: "" }]);
  }

  function loadPlanForEdit(plan: ExercisePlan) {
    setEditingPlanId(plan.id);
    setPlanKinesiologistId(plan.kinesiologist_id);
    setPlanDiagnosisId(plan.patient_diagnosis_id ?? "");
    setPlanFrequency(plan.frequency === "daily" ? "daily" : "weekly");
    setPlanStatus(plan.status === "closed" ? "closed" : "active");
    setPlanDurationWeeks(plan.duration_weeks);
    setPlanObservations(plan.observations ?? "");
    setPlanItems(
      plan.items.length > 0
        ? plan.items.map((item) => ({
            name: item.name,
            estimated_minutes: item.estimated_minutes,
            sets: item.sets == null ? "" : String(item.sets),
            reps: item.reps == null ? "" : String(item.reps),
            description: item.description ?? "",
          }))
        : [{ name: "", estimated_minutes: 10, sets: "", reps: "", description: "" }],
    );
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function loadDiagnosisForEdit(diagnosis: PatientDiagnosis) {
    setEditingDiagnosisId(diagnosis.id);
    setDiagnosisCode(diagnosis.cie10_code);
    setDiagnosisKind(diagnosis.kind === "secondary" ? "secondary" : "primary");
    setDiagnosisStatus(
      diagnosis.status === "confirmed" || diagnosis.status === "resolved" ? diagnosis.status : "suspected",
    );
    setDiagnosisDate(diagnosis.diagnosed_at || todayISO());
    setDiagnosisNotes(diagnosis.notes ?? "");
    setCie10Query(`${diagnosis.cie10_code} ${diagnosis.cie10_description}`);
  }

  function resetDiagnosisForm() {
    setEditingDiagnosisId("");
    setDiagnosisCode("");
    setDiagnosisKind("primary");
    setDiagnosisStatus("suspected");
    setDiagnosisDate(todayISO());
    setDiagnosisNotes("");
    setCie10Query("");
  }

  async function openAttachment(attachment: PatientAttachment) {
    const blob = await downloadPatientAttachment(attachment);
    const url = URL.createObjectURL(blob);
    window.open(url, "_blank", "noopener,noreferrer");
    window.setTimeout(() => URL.revokeObjectURL(url), 60_000);
  }

  function startEditingAttachment(attachment: PatientAttachment) {
    setEditingAttachment({
      id: attachment.id,
      file_name: attachment.file_name,
      notes: attachment.notes ?? "",
      category: attachment.category || "otro",
      patient_visible: attachment.patient_visible,
      confirm_delete: false,
    });
  }

  function patchEditingAttachment(patch: Partial<AttachmentEditForm>) {
    setEditingAttachment((current) => (current ? { ...current, ...patch } : current));
  }

  return (
    <main>
      <div className="max-w-5xl mx-auto p-6 space-y-6">
        <header className="flex items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold">Detalle paciente</h1>
            <p className="text-sm text-gray-600">
              {patient ? `${patient.last_name}, ${patient.first_name}` : "Cargando ficha..."}
            </p>
          </div>
          <Link className="text-sm underline" to={backPath}>
            Volver
          </Link>
        </header>

        {patientQ.isError && (
          <div className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            No se pudo cargar el paciente.
          </div>
        )}

        {patient && (
          <section className="bg-white rounded-xl shadow p-4 space-y-4">
            <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
              <div>
                <h2 className="text-lg font-semibold">Ficha</h2>
                <div className="mt-2 grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-1 text-sm text-gray-700">
                  <div>DNI: <span className="font-medium">{patient.dni}</span></div>
                  <div>Email: <span className="font-medium">{patient.email}</span></div>
                  <div>Teléfono: <span className="font-medium">{patient.phone || "-"}</span></div>
                  <div>Estado: <span className="font-medium">{patient.active ? "Activo" : "Archivado"}</span></div>
                  {patient.birth_date && <div>Fecha nacimiento: <span className="font-medium">{patient.birth_date}</span></div>}
                </div>
                {patient.clinical_notes && (
                  <p className="text-sm text-gray-700 mt-3">{patient.clinical_notes}</p>
                )}
              </div>

              {canManagePatient && (
                <div className="flex flex-wrap gap-2">
                  <Link
                    className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50"
                    to="/agenda"
                    onClick={() => localStorage.setItem("last_patient_id", patient.id)}
                  >
                    Usar en Agenda
                  </Link>
                  {user?.role === "admin" && (
                    <button
                      type="button"
                      className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50 disabled:opacity-50"
                      disabled={inviteM.isPending}
                      onClick={() => inviteM.mutate(patient.email)}
                    >
                      {inviteM.isPending ? "Enviando..." : "Enviar acceso"}
                    </button>
                  )}
                  <button
                    type="button"
                    className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50 disabled:opacity-50"
                    disabled={archiveM.isPending || reactivateM.isPending}
                    onClick={() => {
                      if (patient.active) archiveM.mutate(patient.id);
                      else reactivateM.mutate();
                    }}
                  >
                    {patient.active ? "Archivar" : "Reactivar"}
                  </button>
                </div>
              )}
            </div>

            {inviteM.isSuccess && <p className="text-sm text-green-700">Invitación enviada.</p>}
            {(archiveM.isError || reactivateM.isError || inviteM.isError) && (
              <p className="text-sm text-red-600">No se pudo completar la acción.</p>
            )}
          </section>
        )}

        {patient && canEditClinical && (
          <section className="bg-white rounded-xl shadow p-4 space-y-4">
            <div>
              <h2 className="text-lg font-semibold">Historia clínica</h2>
              <p className="text-sm text-gray-600">Resumen clínico y datos útiles para el seguimiento.</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div>
                <label className="text-sm font-medium">Teléfono</label>
                <input
                  className="mt-1 w-full border rounded-lg p-2"
                  value={phone}
                  onChange={(event) => setPhone(event.target.value)}
                />
              </div>

              <div>
                <label className="text-sm font-medium">Fecha de nacimiento</label>
                <input
                  className="mt-1 w-full border rounded-lg p-2"
                  type="date"
                  value={birthDate}
                  onChange={(event) => setBirthDate(event.target.value)}
                />
              </div>

              <div className="md:col-span-2">
                <label className="text-sm font-medium">Resumen clínico</label>
                <textarea
                  className="mt-1 w-full border rounded-lg p-2 min-h-36"
                  value={clinicalNotes}
                  onChange={(event) => setClinicalNotes(event.target.value)}
                  placeholder="Motivo de consulta, diagnóstico, antecedentes, objetivos, alertas o contraindicaciones."
                />
              </div>
            </div>

            <button
              type="button"
              className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
              disabled={updateClinicalM.isPending}
              onClick={() => updateClinicalM.mutate()}
            >
              {updateClinicalM.isPending ? "Guardando..." : "Guardar historia clínica"}
            </button>

            {updateClinicalM.isError && (
              <p className="text-sm text-red-600">No se pudo guardar la historia clínica.</p>
            )}
            {updateClinicalM.isSuccess && (
              <p className="text-sm text-green-700">Historia clínica actualizada.</p>
            )}
            {patient.clinical_notes_updated_at && (
              <p className="text-sm text-gray-600">
                Última edición clínica: {formatLocalDateTime(patient.clinical_notes_updated_at)}
                {patient.clinical_notes_updated_by_email ? ` · ${patient.clinical_notes_updated_by_email}` : ""}
                {patient.clinical_notes_updated_by_role ? ` (${patient.clinical_notes_updated_by_role})` : ""}
              </p>
            )}
          </section>
        )}

        {patient && canEditClinical && (
          <section className="bg-white rounded-xl shadow p-4 space-y-4">
            <div>
              <h2 className="text-lg font-semibold">Diagnósticos CIE-10</h2>
              <p className="text-sm text-gray-600">Registro de diagnóstico principal y diagnósticos asociados.</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
              <div className="md:col-span-2">
                <label className="text-sm font-medium">Buscar diagnóstico</label>
                <input
                  className="mt-1 w-full border rounded-lg p-2"
                  value={cie10Query}
                  onChange={(event) => setCie10Query(event.target.value)}
                  placeholder="Ej: M54, dorsalgia, respiratoria..."
                />
                {cie10Q.isFetching && <p className="text-xs text-gray-500 mt-1">Buscando CIE-10...</p>}
              </div>

              <div>
                <label className="text-sm font-medium">Tipo</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={diagnosisKind}
                  onChange={(event) => setDiagnosisKind(event.target.value as "primary" | "secondary")}
                >
                  <option value="primary">Principal</option>
                  <option value="secondary">Secundario</option>
                </select>
              </div>

              <div>
                <label className="text-sm font-medium">Estado</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={diagnosisStatus}
                  onChange={(event) => setDiagnosisStatus(event.target.value as "suspected" | "confirmed" | "resolved")}
                >
                  <option value="suspected">Sospechado</option>
                  <option value="confirmed">Confirmado</option>
                  <option value="resolved">Resuelto</option>
                </select>
              </div>

              <div>
                <label className="text-sm font-medium">Fecha</label>
                <input
                  className="mt-1 w-full border rounded-lg p-2"
                  type="date"
                  value={diagnosisDate}
                  onChange={(event) => setDiagnosisDate(event.target.value)}
                />
              </div>

              <div className="md:col-span-3">
                <label className="text-sm font-medium">Observaciones</label>
                <input
                  className="mt-1 w-full border rounded-lg p-2"
                  value={diagnosisNotes}
                  onChange={(event) => setDiagnosisNotes(event.target.value)}
                  placeholder="Detalle clínico opcional."
                />
              </div>
            </div>

            <div className="rounded-lg border divide-y">
              {cie10Results.length === 0 ? (
                <p className="text-sm text-gray-600 p-3">Buscá por código o texto para seleccionar un diagnóstico.</p>
              ) : (
                cie10Results.map((item) => (
                  <button
                    type="button"
                    key={item.code}
                    className={`w-full text-left p-3 hover:bg-gray-50 ${diagnosisCode === item.code ? "bg-gray-100" : ""}`}
                    onClick={() => {
                      setDiagnosisCode(item.code);
                      setCie10Query(`${item.code} ${item.description}`);
                    }}
                  >
                    <div className="font-medium">{item.code} · {item.description}</div>
                    {item.chapter && <div className="text-xs text-gray-500">{item.chapter}</div>}
                  </button>
                ))
              )}
            </div>

            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
                disabled={!diagnosisCode || !diagnosisDate || saveDiagnosisM.isPending}
                onClick={() => saveDiagnosisM.mutate()}
              >
                {saveDiagnosisM.isPending ? "Guardando..." : editingDiagnosisId ? "Guardar diagnóstico" : "Agregar diagnóstico"}
              </button>
              {editingDiagnosisId && (
                <button type="button" className="px-4 py-2 rounded-lg border hover:bg-gray-50" onClick={resetDiagnosisForm}>
                  Cancelar edición
                </button>
              )}
            </div>

            {saveDiagnosisM.isError && (
              <p className="text-sm text-red-600">No se pudo guardar el diagnóstico.</p>
            )}
            {deleteDiagnosisM.isError && (
              <p className="text-sm text-red-600">No se pudo quitar el diagnóstico.</p>
            )}
            {saveDiagnosisM.isSuccess && (
              <p className="text-sm text-green-700">Diagnóstico guardado.</p>
            )}

            {diagnosesQ.isLoading && <p className="text-sm text-gray-600">Cargando diagnósticos...</p>}
            {diagnosesQ.isError && <p className="text-sm text-red-600">No se pudieron cargar los diagnósticos.</p>}
            {!diagnosesQ.isLoading && !diagnosesQ.isError && diagnoses.length === 0 && (
              <p className="text-sm text-gray-600">Sin diagnósticos registrados.</p>
            )}
            {diagnoses.length > 0 && (
              <div className="divide-y">
                {diagnoses.map((diagnosis) => (
                  <div key={diagnosis.id} className="py-3 flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                    <div>
                      <div className="font-medium">{diagnosis.cie10_code} · {diagnosis.cie10_description}</div>
                      <div className="text-sm text-gray-600">
                        {diagnosisKindLabel(diagnosis.kind)} · {diagnosisStatusLabel(diagnosis.status)} · {diagnosis.diagnosed_at}
                      </div>
                      {diagnosis.cie10_chapter && <div className="text-xs text-gray-500">{diagnosis.cie10_chapter}</div>}
                      {diagnosis.notes && <p className="text-sm text-gray-700 mt-1">{diagnosis.notes}</p>}
                      <div className="text-xs text-gray-500 mt-1">
                        Cargado: {formatLocalDateTime(diagnosis.created_at)}
                        {diagnosis.created_by_email ? ` · ${diagnosis.created_by_email}` : ""}
                      </div>
                      <div className="text-xs text-gray-500">
                        Última edición: {formatLocalDateTime(diagnosis.updated_at)}
                        {diagnosis.updated_by_email ? ` · ${diagnosis.updated_by_email}` : ""}
                      </div>
                    </div>
                    <div className="flex flex-wrap gap-2">
                      <button
                        type="button"
                        className={`px-3 py-1 rounded-lg border text-sm ${
                          selectedClinicalDiagnosisId === diagnosis.id ? "bg-black text-white" : "hover:bg-gray-50"
                        }`}
                        onClick={() => setSelectedClinicalDiagnosisId(diagnosis.id)}
                      >
                        Ver historia
                      </button>
                      <button
                        type="button"
                        className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                        onClick={() => loadDiagnosisForEdit(diagnosis)}
                      >
                        Editar
                      </button>
                      <button
                        type="button"
                        className="px-3 py-1 rounded-lg border text-sm hover:bg-red-50 disabled:opacity-50"
                        disabled={deleteDiagnosisM.isPending}
                        onClick={() => deleteDiagnosisM.mutate(diagnosis.id)}
                      >
                        Quitar
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>
        )}

        {patient && canEditClinical && (
          <section className="bg-white rounded-xl shadow p-4 space-y-4">
            <div>
              <h2 className="text-lg font-semibold">Estudios y archivos</h2>
              <p className="text-sm text-gray-600">Imágenes, videos y PDFs asociados al paciente.</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
              <div>
                <label className="text-sm font-medium">Archivo</label>
                <input
                  className="mt-1 w-full border rounded-lg p-2"
                  type="file"
                  accept="image/*,video/*,application/pdf"
                  onChange={(event) => setAttachmentFile(event.target.files?.[0] ?? null)}
                />
              </div>
              <div>
                <label className="text-sm font-medium">Categoría</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={attachmentCategory}
                  onChange={(event) => setAttachmentCategory(event.target.value)}
                >
                  {attachmentCategories.map((category) => (
                    <option key={category.value} value={category.value}>
                      {category.label}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="text-sm font-medium">Nota</label>
                <input
                  className="mt-1 w-full border rounded-lg p-2"
                  value={attachmentNotes}
                  onChange={(event) => setAttachmentNotes(event.target.value)}
                  placeholder="Ej: Resonancia lumbar, control postural, video de marcha."
                />
              </div>
              <label className="flex items-center gap-2 text-sm font-medium md:self-end md:pb-3">
                <input
                  type="checkbox"
                  checked={attachmentPatientVisible}
                  onChange={(event) => setAttachmentPatientVisible(event.target.checked)}
                />
                Visible para paciente
              </label>
            </div>

            <button
              type="button"
              className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
              disabled={!attachmentFile || uploadAttachmentM.isPending}
              onClick={() => uploadAttachmentM.mutate()}
            >
              {uploadAttachmentM.isPending ? "Subiendo..." : "Adjuntar archivo"}
            </button>

            {uploadAttachmentM.isError && (
              <p className="text-sm text-red-600">No se pudo subir el archivo.</p>
            )}
            {uploadAttachmentM.isSuccess && (
              <p className="text-sm text-green-700">Archivo adjuntado.</p>
            )}

            {attachmentsQ.isLoading && <p className="text-sm text-gray-600">Cargando archivos...</p>}
            {attachmentsQ.isError && <p className="text-sm text-red-600">No se pudieron cargar los archivos.</p>}
            {!attachmentsQ.isLoading && !attachmentsQ.isError && attachments.length === 0 && (
              <p className="text-sm text-gray-600">Sin archivos adjuntos.</p>
            )}
            {attachments.length > 0 && (
              <div className="divide-y">
                {attachments.map((attachment) => {
                  const isEditing = editingAttachment?.id === attachment.id;
                  return (
                    <div key={attachment.id} className="py-3">
                      <div className="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
                        <div>
                          <div className="font-medium">{attachment.file_name}</div>
                          <div className="text-sm text-gray-600">
                            {attachment.category || "otro"} · {attachment.kind} · {Math.max(1, Math.round(attachment.size_bytes / 1024))} KB ·{" "}
                            {formatLocalDateTime(attachment.created_at)}
                          </div>
                          <div className="text-sm text-gray-600">
                            {attachment.patient_visible ? "Visible en portal paciente" : "Solo equipo"}
                          </div>
                          {attachment.uploaded_by_email && (
                            <div className="text-sm text-gray-600">
                              Subido por {attachment.uploaded_by_email}
                              {attachment.uploaded_by_role ? ` (${attachment.uploaded_by_role})` : ""}
                            </div>
                          )}
                          {attachment.updated_at && (
                            <div className="text-sm text-gray-600">
                              Editado: {formatLocalDateTime(attachment.updated_at)}
                              {attachment.updated_by_email ? ` · ${attachment.updated_by_email}` : ""}
                            </div>
                          )}
                          {attachment.notes && <p className="text-sm text-gray-700 mt-1">{attachment.notes}</p>}
                        </div>
                        <div className="flex flex-wrap gap-2">
                          <button
                            type="button"
                            className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50"
                            onClick={() => openAttachment(attachment)}
                          >
                            Ver archivo
                          </button>
                          <button
                            type="button"
                            className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50 disabled:opacity-50"
                            disabled={updateAttachmentM.isPending}
                            onClick={() => startEditingAttachment(attachment)}
                          >
                            Editar
                          </button>
                        </div>
                      </div>

                      {isEditing && editingAttachment && (
                        <div className="mt-3 rounded-lg border bg-gray-50 p-3 space-y-3">
                          <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                            <div>
                              <label className="text-sm font-medium">Nombre</label>
                              <input
                                className="mt-1 w-full border rounded-lg p-2"
                                value={editingAttachment.file_name}
                                onChange={(event) => patchEditingAttachment({ file_name: event.target.value })}
                              />
                            </div>
                            <div>
                              <label className="text-sm font-medium">Categoría</label>
                              <select
                                className="mt-1 w-full border rounded-lg p-2"
                                value={editingAttachment.category}
                                onChange={(event) => patchEditingAttachment({ category: event.target.value })}
                              >
                                {attachmentCategories.map((category) => (
                                  <option key={category.value} value={category.value}>
                                    {category.label}
                                  </option>
                                ))}
                              </select>
                            </div>
                            <label className="flex items-center gap-2 text-sm font-medium md:self-end md:pb-3">
                              <input
                                type="checkbox"
                                checked={editingAttachment.patient_visible}
                                onChange={(event) => patchEditingAttachment({ patient_visible: event.target.checked })}
                              />
                              Visible para paciente
                            </label>
                            <div className="md:col-span-3">
                              <label className="text-sm font-medium">Nota</label>
                              <textarea
                                className="mt-1 w-full border rounded-lg p-2 min-h-20"
                                value={editingAttachment.notes}
                                onChange={(event) => patchEditingAttachment({ notes: event.target.value })}
                              />
                            </div>
                          </div>

                          <div className="flex flex-wrap items-center gap-2">
                            <button
                              type="button"
                              className="px-3 py-2 rounded-lg bg-black text-white text-sm disabled:opacity-50"
                              disabled={!editingAttachment.file_name.trim() || updateAttachmentM.isPending}
                              onClick={() => updateAttachmentM.mutate(editingAttachment)}
                            >
                              {updateAttachmentM.isPending ? "Guardando..." : "Guardar cambios"}
                            </button>
                            <button
                              type="button"
                              className="px-3 py-2 rounded-lg border text-sm hover:bg-white"
                              onClick={() => setEditingAttachment(null)}
                            >
                              Cancelar
                            </button>
                            <button
                              type="button"
                              className="px-3 py-2 rounded-lg border text-sm hover:bg-red-50 disabled:opacity-50"
                              disabled={deleteAttachmentM.isPending}
                              onClick={() => {
                                if (!editingAttachment.confirm_delete) {
                                  patchEditingAttachment({ confirm_delete: true });
                                  return;
                                }
                                deleteAttachmentM.mutate(editingAttachment.id, {
                                  onSuccess: () => setEditingAttachment(null),
                                });
                              }}
                            >
                              {editingAttachment.confirm_delete ? "Confirmar borrado" : "Borrar archivo"}
                            </button>
                          </div>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
            {(updateAttachmentM.isError || deleteAttachmentM.isError) && (
              <p className="text-sm text-red-600">No se pudo actualizar el archivo.</p>
            )}
          </section>
        )}

        {canCreateEvolution && (
          <section className="bg-white rounded-xl shadow p-4 space-y-4">
            <div>
              <h2 className="text-lg font-semibold">Registrar evolución</h2>
              <p className="text-sm text-gray-600">Nueva nota clínica asociada a este paciente.</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
              <div className="md:col-span-2">
                <label className="text-sm font-medium">Kinesiólogo</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={evolutionKinesiologistId}
                  onChange={(event) => setEvolutionKinesiologistId(event.target.value)}
                >
                  <option value="">Seleccionar...</option>
                  {kinesiologists.map((kinesiologist) => (
                    <option key={kinesiologist.id} value={kinesiologist.id}>
                      {kinesiologist.last_name}, {kinesiologist.first_name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-sm font-medium">Dolor</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={painLevel}
                  onChange={(event) => setPainLevel(event.target.value)}
                >
                  <option value="">Sin registrar</option>
                  {Array.from({ length: 11 }, (_, value) => (
                    <option key={value} value={value}>
                      {value}/10
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-sm font-medium">Movilidad</label>
                <input
                  type="number"
                  min={0}
                  max={100}
                  className="mt-1 w-full border rounded-lg p-2"
                  placeholder="0 a 100"
                  value={mobilityScore}
                  onChange={(event) => setMobilityScore(event.target.value)}
                />
              </div>

              <div>
                <label className="text-sm font-medium">Fuerza</label>
                <input
                  type="number"
                  min={0}
                  max={100}
                  className="mt-1 w-full border rounded-lg p-2"
                  placeholder="0 a 100"
                  value={strengthScore}
                  onChange={(event) => setStrengthScore(event.target.value)}
                />
              </div>

              <div>
                <label className="text-sm font-medium">Funcionalidad</label>
                <input
                  type="number"
                  min={0}
                  max={100}
                  className="mt-1 w-full border rounded-lg p-2"
                  placeholder="0 a 100"
                  value={functionalScore}
                  onChange={(event) => setFunctionalScore(event.target.value)}
                />
              </div>

              <div className="md:col-span-2">
                <label className="text-sm font-medium">Turno asociado</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={evolutionAppointmentId}
                  onChange={(event) => setEvolutionAppointmentId(event.target.value)}
                >
                  <option value="">Sin asociar</option>
                  {upcomingScheduledAppointments.map((appointment) => (
                    <option key={appointment.id} value={appointment.id}>
                      {formatLocalDateTime(appointment.start_at)}
                    </option>
                  ))}
                </select>
              </div>

              <div className="md:col-span-2">
                <label className="text-sm font-medium">Diagnóstico asociado</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={evolutionDiagnosisId}
                  onChange={(event) => setEvolutionDiagnosisId(event.target.value)}
                >
                  <option value="">Sin asociar</option>
                  {diagnoses.map((diagnosis) => (
                    <option key={diagnosis.id} value={diagnosis.id}>
                      {diagnosis.cie10_code} · {diagnosis.cie10_description} · {diagnosisKindLabel(diagnosis.kind)}
                    </option>
                  ))}
                </select>
              </div>

              <div className="md:col-span-4">
                <label className="text-sm font-medium">Notas</label>
                <textarea
                  className="mt-1 w-full border rounded-lg p-2 min-h-28"
                  value={evolutionNotes}
                  onChange={(event) => setEvolutionNotes(event.target.value)}
                />
              </div>
            </div>

            <button
              type="button"
              className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
              disabled={!evolutionKinesiologistId || !evolutionNotes.trim() || createEvolutionM.isPending}
              onClick={() => createEvolutionM.mutate()}
            >
              {createEvolutionM.isPending ? "Guardando..." : "Guardar evolución"}
            </button>

            {createEvolutionM.isError && (
              <p className="text-sm text-red-600">No se pudo guardar la evolución.</p>
            )}
            {createEvolutionM.isSuccess && (
              <p className="text-sm text-green-700">Evolución registrada.</p>
            )}
          </section>
        )}

        {canCreateEvolution && (
          <section className="bg-white rounded-xl shadow p-4 space-y-4">
            <div>
              <h2 className="text-lg font-semibold">
                {editingPlanId ? "Editar plan de ejercicios" : "Crear plan de ejercicios"}
              </h2>
              <p className="text-sm text-gray-600">Plan con ejercicios para seguimiento del paciente.</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div className="md:col-span-3">
                <label className="text-sm font-medium">Kinesiólogo</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={planKinesiologistId}
                  onChange={(event) => setPlanKinesiologistId(event.target.value)}
                >
                  <option value="">Seleccionar...</option>
                  {kinesiologists.map((kinesiologist) => (
                    <option key={kinesiologist.id} value={kinesiologist.id}>
                      {kinesiologist.last_name}, {kinesiologist.first_name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="md:col-span-3">
                <label className="text-sm font-medium">Diagnóstico asociado</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={planDiagnosisId}
                  onChange={(event) => setPlanDiagnosisId(event.target.value)}
                >
                  <option value="">Sin asociar</option>
                  {diagnoses.map((diagnosis) => (
                    <option key={diagnosis.id} value={diagnosis.id}>
                      {diagnosis.cie10_code} · {diagnosis.cie10_description} · {diagnosisKindLabel(diagnosis.kind)}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-sm font-medium">Frecuencia</label>
                <select
                  className="mt-1 w-full border rounded-lg p-2"
                  value={planFrequency}
                  onChange={(event) => setPlanFrequency(event.target.value as "daily" | "weekly")}
                >
                  <option value="weekly">Semanal</option>
                  <option value="daily">Diaria</option>
                </select>
              </div>

              <div>
                <label className="text-sm font-medium">Duración</label>
                <input
                  className="mt-1 w-full border rounded-lg p-2"
                  type="number"
                  min={1}
                  value={planDurationWeeks}
                  onChange={(event) => setPlanDurationWeeks(Number(event.target.value))}
                />
              </div>

              <div>
                <label className="text-sm font-medium">Unidad</label>
                <div className="mt-1 w-full border rounded-lg p-2 bg-gray-50 text-gray-700">semanas</div>
              </div>

              {editingPlanId && (
                <div>
                  <label className="text-sm font-medium">Estado</label>
                  <select
                    className="mt-1 w-full border rounded-lg p-2"
                    value={planStatus}
                    onChange={(event) => setPlanStatus(event.target.value as "active" | "closed")}
                  >
                    <option value="active">Activo</option>
                    <option value="closed">Cerrado</option>
                  </select>
                </div>
              )}

              <div className="md:col-span-3">
                <label className="text-sm font-medium">Observaciones</label>
                <textarea
                  className="mt-1 w-full border rounded-lg p-2 min-h-20"
                  value={planObservations}
                  onChange={(event) => setPlanObservations(event.target.value)}
                />
              </div>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="font-medium">Ejercicios</h3>
                <button
                  type="button"
                  className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                  onClick={() =>
                    setPlanItems((current) => [
                      ...current,
                      { name: "", estimated_minutes: 10, sets: "", reps: "", description: "" },
                    ])
                  }
                >
                  Agregar ejercicio
                </button>
              </div>

              {planItems.map((item, index) => (
                <div key={index} className="border rounded-lg p-3 grid grid-cols-1 md:grid-cols-4 gap-3">
                  <div className="md:col-span-2">
                    <label className="text-sm font-medium">Nombre</label>
                    <input
                      className="mt-1 w-full border rounded-lg p-2"
                      value={item.name}
                      onChange={(event) =>
                        setPlanItems((current) =>
                          current.map((it, i) => (i === index ? { ...it, name: event.target.value } : it)),
                        )
                      }
                    />
                  </div>

                  <div>
                    <label className="text-sm font-medium">Minutos</label>
                    <input
                      className="mt-1 w-full border rounded-lg p-2"
                      type="number"
                      min={1}
                      value={item.estimated_minutes}
                      onChange={(event) =>
                        setPlanItems((current) =>
                          current.map((it, i) =>
                            i === index ? { ...it, estimated_minutes: Number(event.target.value) } : it,
                          ),
                        )
                      }
                    />
                  </div>

                  <div className="flex items-end">
                    <button
                      type="button"
                      className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50 disabled:opacity-50"
                      disabled={planItems.length === 1}
                      onClick={() => setPlanItems((current) => current.filter((_, i) => i !== index))}
                    >
                      Quitar
                    </button>
                  </div>

                  <div>
                    <label className="text-sm font-medium">Series</label>
                    <input
                      className="mt-1 w-full border rounded-lg p-2"
                      type="number"
                      min={1}
                      value={item.sets}
                      onChange={(event) =>
                        setPlanItems((current) =>
                          current.map((it, i) => (i === index ? { ...it, sets: event.target.value } : it)),
                        )
                      }
                    />
                  </div>

                  <div>
                    <label className="text-sm font-medium">Repeticiones</label>
                    <input
                      className="mt-1 w-full border rounded-lg p-2"
                      type="number"
                      min={1}
                      value={item.reps}
                      onChange={(event) =>
                        setPlanItems((current) =>
                          current.map((it, i) => (i === index ? { ...it, reps: event.target.value } : it)),
                        )
                      }
                    />
                  </div>

                  <div className="md:col-span-2">
                    <label className="text-sm font-medium">Descripción</label>
                    <input
                      className="mt-1 w-full border rounded-lg p-2"
                      value={item.description}
                      onChange={(event) =>
                        setPlanItems((current) =>
                          current.map((it, i) => (i === index ? { ...it, description: event.target.value } : it)),
                        )
                      }
                    />
                  </div>
                </div>
              ))}
            </div>

            <button
              type="button"
              className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
              disabled={!planKinesiologistId || !hasValidPlanItems || planDurationWeeks <= 0 || savePlanM.isPending}
              onClick={() => savePlanM.mutate()}
            >
              {savePlanM.isPending ? "Guardando..." : editingPlanId ? "Guardar cambios" : "Crear plan"}
            </button>
            {editingPlanId && (
              <button
                type="button"
                className="ml-2 px-4 py-2 rounded-lg border hover:bg-gray-50"
                onClick={resetPlanForm}
              >
                Cancelar edición
              </button>
            )}

            {savePlanM.isError && (
              <p className="text-sm text-red-600">No se pudo guardar el plan.</p>
            )}
            {savePlanM.isSuccess && (
              <p className="text-sm text-green-700">Plan guardado.</p>
            )}
          </section>
        )}

        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
            <div>
              <h2 className="text-lg font-semibold">Timeline clínico</h2>
              <p className="text-sm text-gray-600">
                {selectedClinicalDiagnosis
                  ? `Filtrado por ${diagnosisLabel(selectedClinicalDiagnosis)}.`
                  : "Toda la actividad clínica del paciente."}
              </p>
              {selectedClinicalDiagnosis && (
                <p className="text-sm text-gray-600">
                  {filteredEvolutions.length} evoluciones · {filteredPlans.length} planes
                </p>
              )}
            </div>

            <div className="flex flex-col gap-2 sm:flex-row sm:items-end">
              <div>
                <label className="text-sm font-medium">Filtrar por diagnóstico</label>
                <select
                  className="mt-1 w-full min-w-72 border rounded-lg p-2"
                  value={selectedClinicalDiagnosisId}
                  onChange={(event) => setSelectedClinicalDiagnosisId(event.target.value)}
                >
                  <option value="">Todos</option>
                  {diagnoses.map((diagnosis) => (
                    <option key={diagnosis.id} value={diagnosis.id}>
                      {diagnosis.cie10_code} · {diagnosis.cie10_description}
                    </option>
                  ))}
                </select>
              </div>
              {selectedClinicalDiagnosisId && (
                <button
                  type="button"
                  className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50"
                  onClick={() => setSelectedClinicalDiagnosisId("")}
                >
                  Limpiar
                </button>
              )}
            </div>
          </div>
          {timeline.length === 0 ? (
            <p className="text-sm text-gray-600">Sin actividad registrada.</p>
          ) : (
            <div className="divide-y">
              {timeline.map((item) => (
                <div key={item.id} className="py-3 grid grid-cols-1 md:grid-cols-[160px_120px_1fr] gap-2">
                  <div className="text-sm text-gray-600">{formatLocalDateTime(item.at)}</div>
                  <div>
                    <span className="text-xs px-2 py-1 rounded-md bg-gray-100 text-gray-700">
                      {item.kind}
                    </span>
                  </div>
                  <div>
                    <div className="font-medium">{item.title}</div>
                    <div className="text-sm text-gray-600">{item.detail}</div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <h2 className="text-lg font-semibold">Turnos</h2>
          {appointmentsQ.isLoading && <p className="text-sm text-gray-600">Cargando turnos...</p>}
          {appointmentsQ.isError && <p className="text-sm text-red-600">No se pudieron cargar los turnos.</p>}
          {!appointmentsQ.isLoading && appointments.length === 0 ? (
            <p className="text-sm text-gray-600">No hay turnos en el rango consultado.</p>
          ) : (
            <div className="divide-y">
              {appointments.map((appointment) => (
                <div key={appointment.id} className="py-3 flex items-center justify-between gap-3">
                  <div>
                    <div className="font-medium">{formatLocalDateTime(appointment.start_at)}</div>
                    <div className="text-sm text-gray-600">
                      {formatLocalTime(appointment.start_at)} a {formatLocalTime(appointment.end_at)}
                    </div>
                    {appointment.notes && <div className="text-sm text-gray-600">Notas: {appointment.notes}</div>}
                  </div>
                  <span className={`text-xs px-2 py-1 rounded-md ${appointment.status === "cancelled" ? "bg-gray-100 text-gray-700" : "bg-blue-100 text-blue-700"}`}>
                    {appointment.status === "cancelled" ? "Cancelado" : "Programado"}
                  </span>
                </div>
              ))}
            </div>
          )}
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div>
            <h2 className="text-lg font-semibold">
              {selectedClinicalDiagnosis ? "Evolución clínica del diagnóstico" : "Evolución clínica"}
            </h2>
            <p className="text-sm text-gray-600">Dolor, movilidad, fuerza y funcionalidad registrados en cada evolución.</p>
          </div>
          <ClinicalEvolutionChart evolutions={filteredEvolutions} />
        </section>

        <section className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Panel
            title={selectedClinicalDiagnosis ? "Evoluciones del diagnóstico" : "Evoluciones"}
            isLoading={evolutionsQ.isLoading}
            isError={evolutionsQ.isError}
            empty={filteredEvolutions.length === 0}
          >
            {filteredEvolutions.map((evolution) => (
              <div key={evolution.id} className="py-3 border-b last:border-b-0">
                <div className="font-medium">{formatLocalDateTime(evolution.created_at)}</div>
                <EvolutionMetrics evolution={evolution} />
                {evolution.patient_diagnosis_id && diagnosisById.has(evolution.patient_diagnosis_id) && (
                  <div className="text-sm text-gray-600">
                    Diagnóstico: {diagnosisLabel(diagnosisById.get(evolution.patient_diagnosis_id))}
                  </div>
                )}
                <p className="text-sm text-gray-700 mt-1">{evolution.notes}</p>
              </div>
            ))}
          </Panel>

          <Panel
            title={selectedClinicalDiagnosis ? "Planes del diagnóstico" : "Planes"}
            isLoading={plansQ.isLoading}
            isError={plansQ.isError}
            empty={filteredPlans.length === 0}
          >
            {filteredPlans.map((plan) => (
              <div key={plan.id} className="py-3 border-b last:border-b-0">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <div className="font-medium">
                      {plan.frequency === "daily" ? "Diario" : "Semanal"} · {plan.duration_weeks} semanas
                    </div>
                    <div className="text-sm text-gray-600">Estado: {plan.status === "closed" ? "Cerrado" : "Activo"}</div>
                    {plan.patient_diagnosis_id && diagnosisById.has(plan.patient_diagnosis_id) && (
                      <div className="text-sm text-gray-600">
                        Diagnóstico: {diagnosisLabel(diagnosisById.get(plan.patient_diagnosis_id))}
                      </div>
                    )}
                    <div className="text-sm text-gray-600">Ejercicios: {plan.items.length}</div>
                    <div className="text-sm text-gray-600">
                      Creado: {formatLocalDateTime(plan.created_at)}
                      {plan.created_by_email ? ` · ${plan.created_by_email}` : ""}
                    </div>
                    <div className="text-sm text-gray-600">
                      Última edición: {formatLocalDateTime(plan.updated_at)}
                      {plan.updated_by_email ? ` · ${plan.updated_by_email}` : ""}
                    </div>
                  </div>
                  {canCreateEvolution && (
                    <button
                      type="button"
                      className="px-3 py-1 rounded-lg border text-sm hover:bg-gray-50"
                      onClick={() => loadPlanForEdit(plan)}
                    >
                      Editar
                    </button>
                  )}
                </div>
                {plan.observations && <p className="text-sm text-gray-700 mt-1">{plan.observations}</p>}
                {plan.items.length > 0 && (
                  <div className="mt-2 space-y-2">
                    {plan.items.map((item) => (
                      <div key={item.id} className="rounded-lg border p-2">
                        <div className="text-sm font-medium">{item.name}</div>
                        <div className="text-sm text-gray-600">
                          {item.estimated_minutes} min
                          {item.sets ? ` · ${item.sets} series` : ""}
                          {item.reps ? ` · ${item.reps} repeticiones` : ""}
                        </div>
                        {item.description && <p className="text-sm text-gray-700 mt-1">{item.description}</p>}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </Panel>

          <Panel title="Materiales" isLoading={loansQ.isLoading} isError={loansQ.isError} empty={loans.length === 0}>
            {loans.map((loan) => (
              <div key={loan.id} className="py-3 border-b last:border-b-0">
                <div className="font-medium">Cantidad: {loan.qty}</div>
                <div className="text-sm text-gray-600">Prestado: {formatLocalDateTime(loan.loaned_at)}</div>
                <div className="text-sm text-gray-600">
                  {loan.returned_at ? `Devuelto: ${formatLocalDateTime(loan.returned_at)}` : "Pendiente de devolución"}
                </div>
                {loan.notes && <p className="text-sm text-gray-700 mt-1">{loan.notes}</p>}
              </div>
            ))}
          </Panel>
        </section>
      </div>
    </main>
  );
}

function Panel({
  title,
  isLoading,
  isError,
  empty,
  children,
}: {
  title: string;
  isLoading: boolean;
  isError: boolean;
  empty: boolean;
  children: ReactNode;
}) {
  return (
    <section className="bg-white rounded-xl shadow p-4">
      <h2 className="text-lg font-semibold">{title}</h2>
      {isLoading && <p className="text-sm text-gray-600 mt-2">Cargando...</p>}
      {isError && <p className="text-sm text-red-600 mt-2">No se pudo cargar la información.</p>}
      {!isLoading && !isError && empty ? (
        <p className="text-sm text-gray-600 mt-2">Sin registros.</p>
      ) : (
        <div className="mt-2">{children}</div>
      )}
    </section>
  );
}

type EvolutionMetricKey = "pain_level" | "mobility_score" | "strength_score" | "functional_score";

const evolutionMetricDefs: Array<{
  key: EvolutionMetricKey;
  label: string;
  color: string;
  max: number;
}> = [
  { key: "pain_level", label: "Dolor", color: "#dc2626", max: 10 },
  { key: "mobility_score", label: "Movilidad", color: "#2563eb", max: 100 },
  { key: "strength_score", label: "Fuerza", color: "#16a34a", max: 100 },
  { key: "functional_score", label: "Funcionalidad", color: "#9333ea", max: 100 },
];

function ClinicalEvolutionChart({ evolutions }: { evolutions: PatientEvolution[] }) {
  const points = [...evolutions]
    .filter((evolution) => evolutionMetricDefs.some((metric) => evolution[metric.key] != null))
    .sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());

  if (points.length === 0) {
    return <p className="text-sm text-gray-600">Todavía no hay métricas clínicas registradas.</p>;
  }

  const width = 760;
  const height = 280;
  const padding = { top: 22, right: 28, bottom: 54, left: 46 };
  const chartWidth = width - padding.left - padding.right;
  const chartHeight = height - padding.top - padding.bottom;
  const xFor = (index: number) => padding.left + (points.length === 1 ? chartWidth / 2 : (chartWidth * index) / (points.length - 1));
  const yFor = (normalizedValue: number) => padding.top + chartHeight - (chartHeight * normalizedValue) / 100;
  const labelIndexes = new Set([0, Math.floor((points.length - 1) / 2), points.length - 1]);

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
        <svg className="min-w-[680px] w-full" viewBox={`0 0 ${width} ${height}`} role="img" aria-label="Gráfico de evolución clínica">
          {[0, 25, 50, 75, 100].map((value) => (
            <g key={value}>
              <line
                x1={padding.left}
                x2={width - padding.right}
                y1={yFor(value)}
                y2={yFor(value)}
                stroke="#e5e7eb"
              />
              <text x={padding.left - 10} y={yFor(value) + 4} textAnchor="end" className="fill-gray-500 text-[11px]">
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

          {points.map((evolution, index) => (
            <g key={evolution.id}>
              <line
                x1={xFor(index)}
                x2={xFor(index)}
                y1={padding.top}
                y2={height - padding.bottom}
                stroke="#f3f4f6"
              />
              {labelIndexes.has(index) && (
                <text x={xFor(index)} y={height - 18} textAnchor="middle" className="fill-gray-600 text-[11px]">
                  {formatLocalDateTime(evolution.created_at)}
                </text>
              )}
            </g>
          ))}
        </svg>
      </div>
    </div>
  );
}

function EvolutionMetrics({ evolution }: { evolution: PatientEvolution }) {
  const values = evolutionMetricDefs
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

function evolutionMetricSummary(evolution: PatientEvolution) {
  return evolutionMetricDefs
    .map((metric) => {
      const value = evolution[metric.key];
      return value == null ? null : `${metric.label} ${value}/${metric.max}`;
    })
    .filter((value): value is string => value != null)
    .join(" · ");
}

function diagnosisKindLabel(kind: string) {
  return kind === "secondary" ? "Secundario" : "Principal";
}

function diagnosisStatusLabel(status: string) {
  if (status === "confirmed") return "Confirmado";
  if (status === "resolved") return "Resuelto";
  return "Sospechado";
}

function diagnosisLabel(diagnosis: PatientDiagnosis | undefined) {
  if (!diagnosis) return "";
  return `${diagnosis.cie10_code} · ${diagnosis.cie10_description}`;
}

function diagnosisSummary(diagnosis: PatientDiagnosis | undefined) {
  const label = diagnosisLabel(diagnosis);
  return label ? `${label} · ` : "";
}
