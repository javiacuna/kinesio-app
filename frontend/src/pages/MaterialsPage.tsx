import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useAuth } from "@/features/auth/AuthProvider";
import { listKinesiologists } from "@/features/kinesiologists/api";
import {
  createMaterial,
  listMaterialLoans,
  listMaterials,
  listPatientMaterialLoans,
  loanMaterial,
  returnMaterialLoan,
  updateMaterial,
} from "@/features/materials/api";
import type { Material, MaterialLoan } from "@/features/materials/api";
import { PatientSearch } from "@/features/patients/components/PatientSearch";
import { listPatients } from "@/features/patients/api";
import type { Patient } from "@/features/patients/types";
import { formatLocalDate, formatLocalDateTime } from "@/shared/time/format";
import { RequiredLabel } from "@/shared/ui/RequiredLabel";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

function isStockError(error: unknown) {
  return (error as any)?.status === 409 || (error as any)?.message === "insufficient_stock";
}

function isAlreadyReturned(error: unknown) {
  return (error as any)?.status === 409 || (error as any)?.message === "already_returned";
}

function isOverdue(loan: MaterialLoan): boolean {
  if (!loan.due_date || loan.returned_at) return false;
  return new Date(loan.due_date).getTime() < Date.now();
}

export default function MaterialsPage() {
  const { user } = useAuth();
  const { t } = useLanguage();
  const [loanPatient, setLoanPatient] = useState<Patient | null>(null);
  const [returnPatient, setReturnPatient] = useState<Patient | null>(null);
  const [materialId, setMaterialId] = useState("");
  const [kinesiologistId, setKinesiologistId] = useState("");
  const [qty, setQty] = useState(1);
  const [notes, setNotes] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [newMaterialName, setNewMaterialName] = useState("");
  const [newMaterialQty, setNewMaterialQty] = useState(1);
  const [newMaterialDescription, setNewMaterialDescription] = useState("");
  const [editingMaterialId, setEditingMaterialId] = useState<string | null>(null);
  const [editMaterialName, setEditMaterialName] = useState("");
  const [editMaterialQty, setEditMaterialQty] = useState(1);
  const [editMaterialDescription, setEditMaterialDescription] = useState("");
  const [showMaterialValidation, setShowMaterialValidation] = useState(false);
  const [showLoanValidation, setShowLoanValidation] = useState(false);

  const materialsQ = useQuery({
    queryKey: ["materials"],
    queryFn: () => listMaterials(),
  });

  const kinesiologistsQ = useQuery({
    queryKey: ["kinesiologists", "materials"],
    queryFn: () => listKinesiologists({ includeInactive: true }),
  });

  const patientsQ = useQuery({
    queryKey: ["patients", "materials", "all"],
    queryFn: () => listPatients(500, true),
  });

  const pendingLoansQ = useQuery({
    queryKey: ["materials", "loans", "pending"],
    queryFn: () => listMaterialLoans(true, 200),
  });

  const loanHistoryQ = useQuery({
    queryKey: ["materials", "loans", "history"],
    queryFn: () => listMaterialLoans(false, 200),
  });

  const returnLoansQ = useQuery({
    queryKey: ["materials", "loans", "active", returnPatient?.id],
    queryFn: () => listPatientMaterialLoans(returnPatient!.id, true),
    enabled: Boolean(returnPatient?.id),
  });

  const materials = materialsQ.data ?? [];
  const kinesiologists = kinesiologistsQ.data ?? [];
  const patients = patientsQ.data ?? [];
  const pendingLoans = pendingLoansQ.data ?? [];
  const loanHistory = loanHistoryQ.data ?? [];
  const activeLoans = returnLoansQ.data ?? [];
  const selectedMaterial = materials.find((material) => material.id === materialId);
  const materialById = useMemo(
    () => new Map(materials.map((material) => [material.id, material])),
    [materials],
  );
  const patientById = useMemo(
    () => new Map(patients.map((patient) => [patient.id, patient])),
    [patients],
  );
  const kinesiologistById = useMemo(
    () => new Map(kinesiologists.map((kinesiologist) => [kinesiologist.id, kinesiologist])),
    [kinesiologists],
  );

  useEffect(() => {
    if (!materialId && materials.length > 0) {
      setMaterialId(materials[0].id);
    }
  }, [materialId, materials]);

  useEffect(() => {
    if (kinesiologistId || kinesiologists.length === 0) return;

    const ownProfile =
      user?.email
        ? kinesiologists.find((kinesiologist) => kinesiologist.email.toLowerCase() === user.email.toLowerCase())
        : undefined;
    setKinesiologistId((ownProfile ?? kinesiologists[0]).id);
  }, [kinesiologistId, kinesiologists, user?.email]);

  const createMaterialM = useMutation({
    mutationFn: () =>
      createMaterial({
        name: newMaterialName.trim(),
        description: newMaterialDescription.trim() || null,
        total_qty: newMaterialQty,
      }),
    onSuccess: () => {
      setShowMaterialValidation(false);
      setNewMaterialName("");
      setNewMaterialQty(1);
      setNewMaterialDescription("");
      materialsQ.refetch();
    },
  });

  const updateMaterialM = useMutation({
    mutationFn: () =>
      updateMaterial({
        id: editingMaterialId!,
        name: editMaterialName.trim(),
        description: editMaterialDescription.trim() || null,
        total_qty: editMaterialQty,
      }),
    onSuccess: () => {
      setEditingMaterialId(null);
      materialsQ.refetch();
    },
  });

  const loanM = useMutation({
    mutationFn: () =>
      loanMaterial({
        material_id: materialId,
        patient_id: loanPatient!.id,
        kinesiologist_id: kinesiologistId,
        qty,
        notes: notes.trim() || null,
        due_date: dueDate.trim() || null,
      }),
    onSuccess: () => {
      setShowLoanValidation(false);
      setQty(1);
      setNotes("");
      setDueDate("");
      materialsQ.refetch();
      pendingLoansQ.refetch();
      loanHistoryQ.refetch();
      if (returnPatient?.id === loanPatient?.id) returnLoansQ.refetch();
    },
  });

  const returnM = useMutation({
    mutationFn: (loan: MaterialLoan) => returnMaterialLoan(loan.id),
    onSuccess: () => {
      materialsQ.refetch();
      returnLoansQ.refetch();
      pendingLoansQ.refetch();
      loanHistoryQ.refetch();
    },
  });

  const canLoan =
    Boolean(loanPatient?.id) &&
    Boolean(materialId) &&
    Boolean(kinesiologistId) &&
    qty > 0 &&
    qty <= (selectedMaterial?.available_qty ?? 0);

  function startEditingMaterial(material: Material) {
    setEditingMaterialId(material.id);
    setEditMaterialName(material.name);
    setEditMaterialQty(material.total_qty);
    setEditMaterialDescription(material.description ?? "");
  }

  function patientName(id: string) {
    const patient = patientById.get(id);
    return patient ? `${patient.last_name}, ${patient.first_name}` : t("materials.defaultPatientName");
  }

  function kinesiologistName(id: string) {
    const kinesiologist = kinesiologistById.get(id);
    return kinesiologist ? `${kinesiologist.last_name}, ${kinesiologist.first_name}` : t("materials.defaultKinesiologistName");
  }

  function auditText(email?: string | null, role?: string | null) {
    if (!email && !role) return t("materials.noAudit");
    if (email && role) return `${email} · ${role}`;
    return email || role || t("materials.noAudit");
  }

  function borrowedQty(material: Material) {
    return material.total_qty - material.available_qty;
  }

  return (
    <main className="max-w-5xl mx-auto p-6 space-y-6">
      <header>
        <h1 className="text-2xl font-semibold">{t("materials.title")}</h1>
        <p className="text-sm text-gray-600">{t("materials.subtitle")}</p>
      </header>

      <section className="bg-white rounded-xl shadow p-4 space-y-4">
        <h2 className="text-lg font-semibold">{t("materials.stock")}</h2>

        <div className="grid grid-cols-1 md:grid-cols-[1fr_120px_1fr_auto] gap-3">
          <div>
            <label className="text-sm font-medium"><RequiredLabel required>{t("materials.material")}</RequiredLabel></label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              required
              value={newMaterialName}
              onChange={(event) => setNewMaterialName(event.target.value)}
            />
            {showMaterialValidation && !newMaterialName.trim() && (
              <p className="text-xs text-red-600 mt-1">{t("materials.errorMaterialNameRequired")}</p>
            )}
          </div>
          <div>
            <label className="text-sm font-medium"><RequiredLabel required>{t("materials.quantity")}</RequiredLabel></label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              type="number"
              min={1}
              required
              value={newMaterialQty}
              onChange={(event) => setNewMaterialQty(Number(event.target.value))}
            />
            {showMaterialValidation && newMaterialQty <= 0 && (
              <p className="text-xs text-red-600 mt-1">{t("materials.errorQtyPositive")}</p>
            )}
          </div>
          <div>
            <label className="text-sm font-medium">{t("materials.description")}</label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              value={newMaterialDescription}
              onChange={(event) => setNewMaterialDescription(event.target.value)}
            />
          </div>
          <button
            type="button"
            className="px-4 py-2 rounded-lg bg-black text-white self-end disabled:opacity-50"
            disabled={createMaterialM.isPending}
            onClick={() => {
              setShowMaterialValidation(true);
              if (!newMaterialName.trim() || newMaterialQty <= 0) return;
              createMaterialM.mutate();
            }}
          >
            {createMaterialM.isPending ? t("common.saving") : t("materials.create")}
          </button>
        </div>

        {createMaterialM.isError && (
          <p className="text-sm text-red-600">{t("materials.errorCreateFailed")}</p>
        )}

        {materialsQ.isLoading && <p className="text-sm text-gray-600">{t("materials.loadingStock")}</p>}
        {materialsQ.isError && <p className="text-sm text-red-600">{t("materials.errorLoadStockFailed")}</p>}
        {materials.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {materials.map((material) => (
              <article key={material.id} className="rounded-lg border p-3">
                {editingMaterialId === material.id ? (
                  <div className="space-y-3">
                    <div>
                      <label className="text-sm font-medium">{t("materials.material")}</label>
                      <input
                        className="mt-1 w-full border rounded-lg p-2"
                        value={editMaterialName}
                        onChange={(event) => setEditMaterialName(event.target.value)}
                      />
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-[120px_1fr] gap-3">
                      <div>
                        <label className="text-sm font-medium">{t("materials.quantity")}</label>
                        <input
                          className="mt-1 w-full border rounded-lg p-2"
                          type="number"
                          min={borrowedQty(material)}
                          value={editMaterialQty}
                          onChange={(event) => setEditMaterialQty(Number(event.target.value))}
                        />
                      </div>
                      <div>
                        <label className="text-sm font-medium">{t("materials.description")}</label>
                        <input
                          className="mt-1 w-full border rounded-lg p-2"
                          value={editMaterialDescription}
                          onChange={(event) => setEditMaterialDescription(event.target.value)}
                        />
                      </div>
                    </div>
                    <div className="flex flex-wrap gap-2">
                      <button
                        type="button"
                        className="px-3 py-2 rounded-lg bg-black text-white text-sm disabled:opacity-50"
                        disabled={
                          !editMaterialName.trim() ||
                          editMaterialQty < borrowedQty(material) ||
                          updateMaterialM.isPending
                        }
                        onClick={() => updateMaterialM.mutate()}
                      >
                        {updateMaterialM.isPending ? t("common.saving") : t("materials.save")}
                      </button>
                      <button
                        type="button"
                        className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50"
                        onClick={() => setEditingMaterialId(null)}
                      >
                        {t("common.cancel")}
                      </button>
                    </div>
                    {updateMaterialM.isError && (
                      <p className="text-sm text-red-600">{t("materials.errorEditFailed")}</p>
                    )}
                  </div>
                ) : (
                  <div className="space-y-2">
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <div className="font-medium">{material.name}</div>
                        <div className="text-sm text-gray-600">
                          {t("materials.available")}: {material.available_qty} / {material.total_qty}
                        </div>
                      </div>
                      <button
                        type="button"
                        className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50"
                        onClick={() => startEditingMaterial(material)}
                      >
                        {t("materials.edit")}
                      </button>
                    </div>
                    {material.description && <p className="text-sm text-gray-700">{material.description}</p>}
                    <p className="text-xs text-gray-500">
                      {t("materials.updatedBy")} {auditText(material.updated_by_email, material.updated_by_role)}
                    </p>
                  </div>
                )}
              </article>
            ))}
          </div>
        )}
      </section>

      <section className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <h2 className="text-lg font-semibold">{t("materials.loanTitle")}</h2>

          <PatientSearch
            valuePatientId={loanPatient?.id ?? ""}
            onSelect={setLoanPatient}
            placeholder={t("materials.searchActivePatient")}
          />

          <div>
            <label className="text-sm font-medium">{t("materials.responsibleKinesiologist")}</label>
            <select
              className="mt-1 w-full border rounded-lg p-2"
              value={kinesiologistId}
              onChange={(event) => setKinesiologistId(event.target.value)}
            >
              <option value="">{t("materials.select")}</option>
              {kinesiologists.map((kinesiologist) => (
                <option key={kinesiologist.id} value={kinesiologist.id}>
                  {kinesiologist.last_name}, {kinesiologist.first_name}
                </option>
              ))}
            </select>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-[1fr_120px] gap-3">
            <div>
              <label className="text-sm font-medium"><RequiredLabel required>{t("materials.material")}</RequiredLabel></label>
              <select
                className="mt-1 w-full border rounded-lg p-2"
                value={materialId}
                onChange={(event) => setMaterialId(event.target.value)}
              >
                <option value="">{t("materials.select")}</option>
                {materials.map((material) => (
                  <option key={material.id} value={material.id} disabled={material.available_qty <= 0}>
                    {material.name} ({material.available_qty} {t("materials.availableAbbrev")})
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-sm font-medium"><RequiredLabel required>{t("materials.quantity")}</RequiredLabel></label>
              <input
                className="mt-1 w-full border rounded-lg p-2"
                type="number"
                min={1}
                max={selectedMaterial?.available_qty ?? 1}
                value={qty}
                onChange={(event) => setQty(Number(event.target.value))}
              />
            </div>
          </div>

          <div>
            <label className="text-sm font-medium">{t("materials.dueDate")}</label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              type="date"
              value={dueDate}
              onChange={(event) => setDueDate(event.target.value)}
            />
          </div>

          <div>
            <label className="text-sm font-medium">{t("materials.notes")}</label>
            <textarea
              className="mt-1 w-full border rounded-lg p-2 min-h-20"
              value={notes}
              onChange={(event) => setNotes(event.target.value)}
            />
          </div>

          <button
            type="button"
            className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
            disabled={loanM.isPending}
            onClick={() => {
              setShowLoanValidation(true);
              if (!canLoan) return;
              loanM.mutate();
            }}
          >
            {loanM.isPending ? t("materials.registering") : t("materials.registerLoan")}
          </button>

          {showLoanValidation && !canLoan && (
            <p className="text-sm text-red-600">
              {!loanPatient?.id
                ? t("materials.errorSelectPatient")
                : !materialId
                  ? t("materials.errorSelectMaterial")
                  : !kinesiologistId
                    ? t("materials.errorSelectKinesiologist")
                    : qty <= 0
                      ? t("materials.errorQtyPositive")
                      : t("materials.errorQtyExceedsStock")}
            </p>
          )}

          {loanM.isError && (
            <p className="text-sm text-red-600">
              {isStockError(loanM.error) ? t("materials.errorInsufficientStock") : t("materials.errorLoanFailed")}
            </p>
          )}
          {loanM.isSuccess && <p className="text-sm text-green-700">{t("materials.loanRegistered")}</p>}
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <h2 className="text-lg font-semibold">{t("materials.returnTitle")}</h2>

          <PatientSearch
            valuePatientId={returnPatient?.id ?? ""}
            onSelect={setReturnPatient}
            placeholder={t("materials.searchPatientWithLoan")}
          />

          {returnLoansQ.isLoading && <p className="text-sm text-gray-600">{t("materials.loadingActiveLoans")}</p>}
          {returnLoansQ.isError && <p className="text-sm text-red-600">{t("materials.errorLoadLoansFailed")}</p>}
          {returnPatient && !returnLoansQ.isLoading && activeLoans.length === 0 && (
            <p className="text-sm text-gray-600">{t("materials.noPendingReturns")}</p>
          )}

          {activeLoans.length > 0 && (
            <div className="divide-y">
              {activeLoans.map((loan) => {
                const material = materialById.get(loan.material_id);
                return (
                  <article key={loan.id} className="py-3 flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
                    <div>
                      <div className="font-medium">{material?.name ?? t("materials.defaultMaterialName")}</div>
                      <div className="text-sm text-gray-600">
                        {t("materials.quantity")}: {loan.qty} · {t("materials.loaned")}: {formatLocalDateTime(loan.loaned_at)}
                      </div>
                      {loan.due_date && (
                        <div className={`text-sm ${isOverdue(loan) ? "text-red-600 font-medium" : "text-gray-600"}`}>
                          {t("materials.dueDateLabel")}: {formatLocalDate(loan.due_date)}
                          {isOverdue(loan) && ` · ${t("materials.overdue")}`}
                        </div>
                      )}
                      {loan.notes && <p className="text-sm text-gray-700 mt-1">{loan.notes}</p>}
                    </div>
                    <button
                      type="button"
                      className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50 disabled:opacity-50"
                      disabled={returnM.isPending}
                      onClick={() => returnM.mutate(loan)}
                    >
                      {returnM.isPending ? t("materials.registering") : t("materials.registerReturn")}
                    </button>
                  </article>
                );
              })}
            </div>
          )}

          {returnM.isError && (
            <p className="text-sm text-red-600">
              {isAlreadyReturned(returnM.error) ? t("materials.errorAlreadyReturned") : t("materials.errorReturnFailed")}
            </p>
          )}
          {returnM.isSuccess && <p className="text-sm text-green-700">{t("materials.returnRegistered")}</p>}
        </section>
      </section>

      <section className="bg-white rounded-xl shadow p-4 space-y-4">
        <h2 className="text-lg font-semibold">{t("materials.pendingTitle")}</h2>

        {pendingLoansQ.isLoading && <p className="text-sm text-gray-600">{t("materials.loadingPending")}</p>}
        {pendingLoansQ.isError && <p className="text-sm text-red-600">{t("materials.errorLoadPendingFailed")}</p>}
        {!pendingLoansQ.isLoading && pendingLoans.length === 0 && (
          <p className="text-sm text-gray-600">{t("materials.noPending")}</p>
        )}
        {pendingLoans.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {pendingLoans.map((loan) => (
              <LoanRow
                key={loan.id}
                loan={loan}
                material={materialById.get(loan.material_id)}
                patientName={patientName(loan.patient_id)}
                kinesiologistName={kinesiologistName(loan.kinesiologist_id)}
                auditText={auditText}
                onReturn={() => returnM.mutate(loan)}
                isReturning={returnM.isPending}
              />
            ))}
          </div>
        )}
      </section>

      <section className="bg-white rounded-xl shadow p-4 space-y-4">
        <h2 className="text-lg font-semibold">{t("materials.historyTitle")}</h2>

        {loanHistoryQ.isLoading && <p className="text-sm text-gray-600">{t("materials.loadingHistory")}</p>}
        {loanHistoryQ.isError && <p className="text-sm text-red-600">{t("materials.errorLoadHistoryFailed")}</p>}
        {!loanHistoryQ.isLoading && loanHistory.length === 0 && (
          <p className="text-sm text-gray-600">{t("materials.noHistory")}</p>
        )}
        {loanHistory.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {loanHistory.map((loan) => (
              <LoanRow
                key={loan.id}
                loan={loan}
                material={materialById.get(loan.material_id)}
                patientName={patientName(loan.patient_id)}
                kinesiologistName={kinesiologistName(loan.kinesiologist_id)}
                auditText={auditText}
              />
            ))}
          </div>
        )}
      </section>
    </main>
  );
}

type LoanRowProps = {
  loan: MaterialLoan;
  material?: Material;
  patientName: string;
  kinesiologistName: string;
  auditText: (email?: string | null, role?: string | null) => string;
  onReturn?: () => void;
  isReturning?: boolean;
};

function LoanRow({ loan, material, patientName, kinesiologistName, auditText, onReturn, isReturning }: LoanRowProps) {
  const { t } = useLanguage();
  return (
    <article className="rounded-lg border p-3 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
      <div className="space-y-1.5">
        <div className="font-medium">{material?.name ?? t("materials.defaultMaterialName")}</div>
        <div className="text-sm text-gray-600">
          {patientName} · {t("materials.quantity")}: {loan.qty} · {t("materials.loaned")}: {formatLocalDateTime(loan.loaned_at)}
        </div>
        <div className="text-sm text-gray-600">{t("materials.responsible")}: {kinesiologistName}</div>
        {loan.returned_at ? (
          <div className="text-sm text-gray-600">{t("materials.returned")}: {formatLocalDateTime(loan.returned_at)}</div>
        ) : (
          <div className="text-sm font-medium text-amber-700">{t("materials.pendingReturn")}</div>
        )}
        {loan.due_date && (
          <div className={`text-sm ${isOverdue(loan) ? "text-red-600 font-medium" : "text-gray-600"}`}>
            {t("materials.dueDateLabel")}: {formatLocalDate(loan.due_date)}
            {isOverdue(loan) && ` · ${t("materials.overdue")}`}
          </div>
        )}
        {loan.notes && <p className="text-sm text-gray-700">{loan.notes}</p>}
        <p className="text-xs text-gray-500">{t("materials.lentBy")}: {auditText(loan.loaned_by_email, loan.loaned_by_role)}</p>
        {loan.returned_at && (
          <p className="text-xs text-gray-500">
            {t("materials.receivedBy")}: {auditText(loan.returned_by_email, loan.returned_by_role)}
          </p>
        )}
      </div>
      {onReturn && !loan.returned_at && (
        <button
          type="button"
          className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50 disabled:opacity-50 shrink-0"
          disabled={isReturning}
          onClick={onReturn}
        >
          {isReturning ? t("materials.registering") : t("materials.registerReturn")}
        </button>
      )}
    </article>
  );
}
