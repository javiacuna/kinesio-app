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
import { formatLocalDateTime } from "@/shared/time/format";

function isStockError(error: unknown) {
  return (error as any)?.status === 409 || (error as any)?.message === "insufficient_stock";
}

function isAlreadyReturned(error: unknown) {
  return (error as any)?.status === 409 || (error as any)?.message === "already_returned";
}

export default function MaterialsPage() {
  const { user } = useAuth();
  const [loanPatient, setLoanPatient] = useState<Patient | null>(null);
  const [returnPatient, setReturnPatient] = useState<Patient | null>(null);
  const [materialId, setMaterialId] = useState("");
  const [kinesiologistId, setKinesiologistId] = useState("");
  const [qty, setQty] = useState(1);
  const [notes, setNotes] = useState("");
  const [newMaterialName, setNewMaterialName] = useState("");
  const [newMaterialQty, setNewMaterialQty] = useState(1);
  const [newMaterialDescription, setNewMaterialDescription] = useState("");
  const [editingMaterialId, setEditingMaterialId] = useState<string | null>(null);
  const [editMaterialName, setEditMaterialName] = useState("");
  const [editMaterialQty, setEditMaterialQty] = useState(1);
  const [editMaterialDescription, setEditMaterialDescription] = useState("");

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
      }),
    onSuccess: () => {
      setQty(1);
      setNotes("");
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
    return patient ? `${patient.last_name}, ${patient.first_name}` : "Paciente";
  }

  function kinesiologistName(id: string) {
    const kinesiologist = kinesiologistById.get(id);
    return kinesiologist ? `${kinesiologist.last_name}, ${kinesiologist.first_name}` : "Kinesiólogo";
  }

  function auditText(email?: string | null, role?: string | null) {
    if (!email && !role) return "Sin auditoría";
    if (email && role) return `${email} · ${role}`;
    return email || role || "Sin auditoría";
  }

  function borrowedQty(material: Material) {
    return material.total_qty - material.available_qty;
  }

  return (
    <main className="max-w-5xl mx-auto p-6 space-y-6">
      <header>
        <h1 className="text-2xl font-semibold">Materiales</h1>
        <p className="text-sm text-gray-600">Préstamos, devoluciones y stock disponible.</p>
      </header>

      <section className="bg-white rounded-xl shadow p-4 space-y-4">
        <h2 className="text-lg font-semibold">Stock</h2>

        <div className="grid grid-cols-1 md:grid-cols-[1fr_120px_1fr_auto] gap-3">
          <div>
            <label className="text-sm font-medium">Material</label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              value={newMaterialName}
              onChange={(event) => setNewMaterialName(event.target.value)}
            />
          </div>
          <div>
            <label className="text-sm font-medium">Cantidad</label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              type="number"
              min={1}
              value={newMaterialQty}
              onChange={(event) => setNewMaterialQty(Number(event.target.value))}
            />
          </div>
          <div>
            <label className="text-sm font-medium">Descripción</label>
            <input
              className="mt-1 w-full border rounded-lg p-2"
              value={newMaterialDescription}
              onChange={(event) => setNewMaterialDescription(event.target.value)}
            />
          </div>
          <button
            type="button"
            className="px-4 py-2 rounded-lg bg-black text-white self-end disabled:opacity-50"
            disabled={!newMaterialName.trim() || newMaterialQty <= 0 || createMaterialM.isPending}
            onClick={() => createMaterialM.mutate()}
          >
            {createMaterialM.isPending ? "Guardando..." : "Crear"}
          </button>
        </div>

        {createMaterialM.isError && (
          <p className="text-sm text-red-600">No se pudo crear el material.</p>
        )}

        {materialsQ.isLoading && <p className="text-sm text-gray-600">Cargando stock...</p>}
        {materialsQ.isError && <p className="text-sm text-red-600">No se pudo cargar el stock.</p>}
        {materials.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {materials.map((material) => (
              <article key={material.id} className="rounded-lg border p-3">
                {editingMaterialId === material.id ? (
                  <div className="space-y-3">
                    <div>
                      <label className="text-sm font-medium">Material</label>
                      <input
                        className="mt-1 w-full border rounded-lg p-2"
                        value={editMaterialName}
                        onChange={(event) => setEditMaterialName(event.target.value)}
                      />
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-[120px_1fr] gap-3">
                      <div>
                        <label className="text-sm font-medium">Cantidad</label>
                        <input
                          className="mt-1 w-full border rounded-lg p-2"
                          type="number"
                          min={borrowedQty(material)}
                          value={editMaterialQty}
                          onChange={(event) => setEditMaterialQty(Number(event.target.value))}
                        />
                      </div>
                      <div>
                        <label className="text-sm font-medium">Descripción</label>
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
                        {updateMaterialM.isPending ? "Guardando..." : "Guardar"}
                      </button>
                      <button
                        type="button"
                        className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50"
                        onClick={() => setEditingMaterialId(null)}
                      >
                        Cancelar
                      </button>
                    </div>
                    {updateMaterialM.isError && (
                      <p className="text-sm text-red-600">No se pudo editar el material.</p>
                    )}
                  </div>
                ) : (
                  <div className="space-y-2">
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <div className="font-medium">{material.name}</div>
                        <div className="text-sm text-gray-600">
                          Disponible: {material.available_qty} / {material.total_qty}
                        </div>
                      </div>
                      <button
                        type="button"
                        className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50"
                        onClick={() => startEditingMaterial(material)}
                      >
                        Editar
                      </button>
                    </div>
                    {material.description && <p className="text-sm text-gray-700">{material.description}</p>}
                    <p className="text-xs text-gray-500">
                      Actualizado por {auditText(material.updated_by_email, material.updated_by_role)}
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
          <h2 className="text-lg font-semibold">Préstamo de material</h2>

          <PatientSearch
            valuePatientId={loanPatient?.id ?? ""}
            onSelect={setLoanPatient}
            placeholder="Buscar paciente activo..."
          />

          <div>
            <label className="text-sm font-medium">Kinesiólogo responsable</label>
            <select
              className="mt-1 w-full border rounded-lg p-2"
              value={kinesiologistId}
              onChange={(event) => setKinesiologistId(event.target.value)}
            >
              <option value="">Seleccionar...</option>
              {kinesiologists.map((kinesiologist) => (
                <option key={kinesiologist.id} value={kinesiologist.id}>
                  {kinesiologist.last_name}, {kinesiologist.first_name}
                </option>
              ))}
            </select>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-[1fr_120px] gap-3">
            <div>
              <label className="text-sm font-medium">Material</label>
              <select
                className="mt-1 w-full border rounded-lg p-2"
                value={materialId}
                onChange={(event) => setMaterialId(event.target.value)}
              >
                <option value="">Seleccionar...</option>
                {materials.map((material) => (
                  <option key={material.id} value={material.id} disabled={material.available_qty <= 0}>
                    {material.name} ({material.available_qty} disp.)
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-sm font-medium">Cantidad</label>
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
            <label className="text-sm font-medium">Notas</label>
            <textarea
              className="mt-1 w-full border rounded-lg p-2 min-h-20"
              value={notes}
              onChange={(event) => setNotes(event.target.value)}
            />
          </div>

          <button
            type="button"
            className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50"
            disabled={!canLoan || loanM.isPending}
            onClick={() => loanM.mutate()}
          >
            {loanM.isPending ? "Registrando..." : "Registrar préstamo"}
          </button>

          {loanM.isError && (
            <p className="text-sm text-red-600">
              {isStockError(loanM.error) ? "No hay stock suficiente." : "No se pudo registrar el préstamo."}
            </p>
          )}
          {loanM.isSuccess && <p className="text-sm text-green-700">Préstamo registrado.</p>}
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <h2 className="text-lg font-semibold">Devolución de material</h2>

          <PatientSearch
            valuePatientId={returnPatient?.id ?? ""}
            onSelect={setReturnPatient}
            placeholder="Buscar paciente con préstamo activo..."
          />

          {returnLoansQ.isLoading && <p className="text-sm text-gray-600">Cargando préstamos activos...</p>}
          {returnLoansQ.isError && <p className="text-sm text-red-600">No se pudieron cargar los préstamos.</p>}
          {returnPatient && !returnLoansQ.isLoading && activeLoans.length === 0 && (
            <p className="text-sm text-gray-600">El paciente no tiene materiales pendientes de devolución.</p>
          )}

          {activeLoans.length > 0 && (
            <div className="divide-y">
              {activeLoans.map((loan) => {
                const material = materialById.get(loan.material_id);
                return (
                  <article key={loan.id} className="py-3 flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
                    <div>
                      <div className="font-medium">{material?.name ?? "Material"}</div>
                      <div className="text-sm text-gray-600">
                        Cantidad: {loan.qty} · Prestado: {formatLocalDateTime(loan.loaned_at)}
                      </div>
                      {loan.notes && <p className="text-sm text-gray-700 mt-1">{loan.notes}</p>}
                    </div>
                    <button
                      type="button"
                      className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50 disabled:opacity-50"
                      disabled={returnM.isPending}
                      onClick={() => returnM.mutate(loan)}
                    >
                      {returnM.isPending ? "Registrando..." : "Registrar devolución"}
                    </button>
                  </article>
                );
              })}
            </div>
          )}

          {returnM.isError && (
            <p className="text-sm text-red-600">
              {isAlreadyReturned(returnM.error) ? "Ese préstamo ya fue devuelto." : "No se pudo registrar la devolución."}
            </p>
          )}
          {returnM.isSuccess && <p className="text-sm text-green-700">Devolución registrada.</p>}
        </section>
      </section>

      <section className="bg-white rounded-xl shadow p-4 space-y-4">
        <h2 className="text-lg font-semibold">Pendientes de devolución</h2>

        {pendingLoansQ.isLoading && <p className="text-sm text-gray-600">Cargando materiales pendientes...</p>}
        {pendingLoansQ.isError && <p className="text-sm text-red-600">No se pudieron cargar los pendientes.</p>}
        {!pendingLoansQ.isLoading && pendingLoans.length === 0 && (
          <p className="text-sm text-gray-600">No hay materiales pendientes de devolución.</p>
        )}
        {pendingLoans.length > 0 && (
          <div className="divide-y">
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
        <h2 className="text-lg font-semibold">Historial de préstamos</h2>

        {loanHistoryQ.isLoading && <p className="text-sm text-gray-600">Cargando historial...</p>}
        {loanHistoryQ.isError && <p className="text-sm text-red-600">No se pudo cargar el historial.</p>}
        {!loanHistoryQ.isLoading && loanHistory.length === 0 && (
          <p className="text-sm text-gray-600">Todavía no hay préstamos registrados.</p>
        )}
        {loanHistory.length > 0 && (
          <div className="divide-y">
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
  return (
    <article className="py-3 flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
      <div className="space-y-1">
        <div className="font-medium">{material?.name ?? "Material"}</div>
        <div className="text-sm text-gray-600">
          {patientName} · Cantidad: {loan.qty} · Prestado: {formatLocalDateTime(loan.loaned_at)}
        </div>
        <div className="text-sm text-gray-600">Responsable: {kinesiologistName}</div>
        {loan.returned_at ? (
          <div className="text-sm text-gray-600">Devuelto: {formatLocalDateTime(loan.returned_at)}</div>
        ) : (
          <div className="text-sm text-amber-700">Pendiente de devolución</div>
        )}
        {loan.notes && <p className="text-sm text-gray-700">{loan.notes}</p>}
        <p className="text-xs text-gray-500">Prestó: {auditText(loan.loaned_by_email, loan.loaned_by_role)}</p>
        {loan.returned_at && (
          <p className="text-xs text-gray-500">
            Recibió: {auditText(loan.returned_by_email, loan.returned_by_role)}
          </p>
        )}
      </div>
      {onReturn && !loan.returned_at && (
        <button
          type="button"
          className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50 disabled:opacity-50"
          disabled={isReturning}
          onClick={onReturn}
        >
          {isReturning ? "Registrando..." : "Registrar devolución"}
        </button>
      )}
    </article>
  );
}
