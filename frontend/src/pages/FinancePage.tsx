import { useMemo, useState, type FormEvent, type ReactNode } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { listKinesiologists } from "@/features/kinesiologists/api";
import { listPractices } from "@/features/kinesiologists/api";
import { listPatients } from "@/features/patients/api";
import {
  listFinanciers,
  listFinancialMovements,
  listPracticeTariffs,
  listProfessionalFeeRules,
  saveFinancier,
  savePracticeTariff,
  saveProfessionalFeeRule,
  updateFinancialMovementStatus,
  type PracticeTariff,
  type ProfessionalFeeRule,
} from "@/features/finance/api";
import { formatLocalDateTime } from "@/shared/time/format";
import { RequiredLabel } from "@/shared/ui/RequiredLabel";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

type Translate = (key: string) => string;

type FinancierForm = {
  id?: string;
  name: string;
  kind: string;
  active: boolean;
};

type TariffForm = {
  id?: string;
  practice_id: string;
  financier_id: string;
  billing_value: string;
  copay: string;
  currency: string;
  valid_from: string;
  valid_to: string;
  active: boolean;
};

type FeeRuleForm = {
  id?: string;
  kinesiologist_id: string;
  practice_id: string;
  rule_type: "fixed" | "percentage";
  fixed_value: string;
  percentage: string;
  active: boolean;
};

const emptyFinancierForm: FinancierForm = {
  name: "",
  kind: "particular",
  active: true,
};

const emptyTariffForm: TariffForm = {
  practice_id: "",
  financier_id: "",
  billing_value: "",
  copay: "0",
  currency: "ARS",
  valid_from: todayISO(),
  valid_to: "",
  active: true,
};

const emptyFeeRuleForm: FeeRuleForm = {
  kinesiologist_id: "",
  practice_id: "",
  rule_type: "percentage",
  fixed_value: "",
  percentage: "60",
  active: true,
};

export default function FinancePage() {
  const { t } = useLanguage();
  const [financierForm, setFinancierForm] = useState<FinancierForm>(emptyFinancierForm);
  const [tariffForm, setTariffForm] = useState<TariffForm>(emptyTariffForm);
  const [feeRuleForm, setFeeRuleForm] = useState<FeeRuleForm>(emptyFeeRuleForm);
  const [movementFrom, setMovementFrom] = useState(monthStartISO());
  const [movementTo, setMovementTo] = useState(todayISO());
  const [movementKinesiologistId, setMovementKinesiologistId] = useState("");
  const [movementPracticeId, setMovementPracticeId] = useState("");
  const [movementFinancierId, setMovementFinancierId] = useState("");
  const [movementCollectionStatus, setMovementCollectionStatus] = useState("");
  const [movementProfessionalPaymentStatus, setMovementProfessionalPaymentStatus] = useState("");
  const [cancellingMovementId, setCancellingMovementId] = useState("");
  const [cancellationReason, setCancellationReason] = useState("");
  const [isPayingSettlement, setIsPayingSettlement] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const financiersQ = useQuery({
    queryKey: ["financiers", "finance", "all"],
    queryFn: () => listFinanciers({ includeInactive: true }),
  });
  const practicesQ = useQuery({
    queryKey: ["practices", "finance", "all"],
    queryFn: () => listPractices({ includeInactive: true }),
  });
  const kinesiologistsQ = useQuery({
    queryKey: ["kinesiologists", "finance", "all"],
    queryFn: () => listKinesiologists({ includeInactive: true }),
  });
  const tariffsQ = useQuery({
    queryKey: ["financial", "tariffs", "all"],
    queryFn: () => listPracticeTariffs({ includeInactive: true }),
  });
  const feeRulesQ = useQuery({
    queryKey: ["financial", "fee-rules", "all"],
    queryFn: () => listProfessionalFeeRules({ includeInactive: true }),
  });
  const movementsQ = useQuery({
    queryKey: [
      "financial",
      "movements",
      movementFrom,
      movementTo,
      movementKinesiologistId,
      movementPracticeId,
      movementFinancierId,
      movementCollectionStatus,
      movementProfessionalPaymentStatus,
    ],
    queryFn: () =>
      listFinancialMovements({
        from: movementFrom ? localDateStartISO(movementFrom) : undefined,
        to: movementTo ? nextLocalDateStartISO(movementTo) : undefined,
        kinesiologist_id: movementKinesiologistId || undefined,
        practice_id: movementPracticeId || undefined,
        financier_id: movementFinancierId || undefined,
        collection_status: movementCollectionStatus || undefined,
        professional_payment_status: movementProfessionalPaymentStatus || undefined,
      }),
  });
  const patientsQ = useQuery({
    queryKey: ["patients", "finance", "all"],
    queryFn: () => listPatients(500, true),
  });

  const financiers = financiersQ.data ?? [];
  const practices = practicesQ.data ?? [];
  const activePractices = practices.filter((practice) => practice.active);
  const kinesiologists = kinesiologistsQ.data ?? [];
  const tariffs = tariffsQ.data ?? [];
  const feeRules = feeRulesQ.data ?? [];
  const movements = movementsQ.data ?? [];

  const practiceById = useMemo(() => new Map(practices.map((item) => [item.id, item])), [practices]);
  const financierById = useMemo(() => new Map(financiers.map((item) => [item.id, item])), [financiers]);
  const kinesiologistById = useMemo(() => new Map(kinesiologists.map((item) => [item.id, item])), [kinesiologists]);
  const patientById = useMemo(() => new Map((patientsQ.data ?? []).map((item) => [item.id, item])), [patientsQ.data]);

  const saveFinancierM = useMutation({
    mutationFn: saveFinancier,
    onSuccess: async () => {
      setMessage(financierForm.id ? t("finance.financierUpdated") : t("finance.financierCreated"));
      setError("");
      setFinancierForm(emptyFinancierForm);
      await financiersQ.refetch();
    },
    onError: (err) => setError((err as Error).message),
  });

  const saveTariffM = useMutation({
    mutationFn: savePracticeTariff,
    onSuccess: async () => {
      setMessage(tariffForm.id ? t("finance.tariffUpdated") : t("finance.tariffCreated"));
      setError("");
      setTariffForm(emptyTariffForm);
      await tariffsQ.refetch();
    },
    onError: (err) => setError((err as Error).message),
  });

  const saveFeeRuleM = useMutation({
    mutationFn: saveProfessionalFeeRule,
    onSuccess: async () => {
      setMessage(feeRuleForm.id ? t("finance.feeUpdated") : t("finance.feeCreated"));
      setError("");
      setFeeRuleForm(emptyFeeRuleForm);
      await feeRulesQ.refetch();
    },
    onError: (err) => setError((err as Error).message),
  });

  const updateMovementStatusM = useMutation({
    mutationFn: updateFinancialMovementStatus,
    onSuccess: async () => {
      setMessage(t("finance.movementUpdated"));
      setError("");
      setCancellingMovementId("");
      setCancellationReason("");
      await movementsQ.refetch();
    },
    onError: (err) => setError((err as Error).message),
  });

  function submitFinancier(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setMessage("");
    if (!financierForm.name.trim()) {
      setError(t("finance.errorFinancierNameRequired"));
      return;
    }
    saveFinancierM.mutate({
      id: financierForm.id,
      name: financierForm.name.trim(),
      kind: financierForm.kind.trim() || "particular",
      active: financierForm.active,
    });
  }

  function submitTariff(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setMessage("");
    const billing = amountToCents(tariffForm.billing_value);
    const copay = amountToCents(tariffForm.copay);
    if (!tariffForm.practice_id || !tariffForm.financier_id || billing == null || copay == null || !tariffForm.valid_from) {
      setError(t("finance.errorTariffRequired"));
      return;
    }
    saveTariffM.mutate({
      id: tariffForm.id,
      practice_id: tariffForm.practice_id,
      financier_id: tariffForm.financier_id,
      billing_value_cents: billing,
      copay_cents: copay,
      currency: tariffForm.currency || "ARS",
      valid_from: tariffForm.valid_from,
      valid_to: tariffForm.valid_to || null,
      active: tariffForm.active,
    });
  }

  function submitFeeRule(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setMessage("");
    const fixed = amountToCents(feeRuleForm.fixed_value);
    const percentage = parseNumber(feeRuleForm.percentage);
    if (!feeRuleForm.kinesiologist_id || !feeRuleForm.practice_id) {
      setError(t("finance.errorFeeKinesiologistPractice"));
      return;
    }
    if (feeRuleForm.rule_type === "fixed" && fixed == null) {
      setError(t("finance.errorFeeFixedRequired"));
      return;
    }
    if (feeRuleForm.rule_type === "percentage" && (percentage == null || percentage < 0 || percentage > 100)) {
      setError(t("finance.errorFeePercentageRequired"));
      return;
    }
    saveFeeRuleM.mutate({
      id: feeRuleForm.id,
      kinesiologist_id: feeRuleForm.kinesiologist_id,
      practice_id: feeRuleForm.practice_id,
      rule_type: feeRuleForm.rule_type,
      fixed_value_cents: feeRuleForm.rule_type === "fixed" ? fixed : null,
      percentage: feeRuleForm.rule_type === "percentage" ? percentage : null,
      active: feeRuleForm.active,
    });
  }

  async function markSettlementPaid() {
    const pending = settlementMovements.filter((movement) => movement.professional_payment_status !== "paid");
    if (pending.length === 0) return;
    setMessage("");
    setError("");
    setIsPayingSettlement(true);
    try {
      await Promise.all(
        pending.map((movement) =>
          updateFinancialMovementStatus({
            movement_id: movement.id,
            professional_payment_status: "paid",
          }),
        ),
      );
      setMessage(t("finance.settlementPaid"));
      await movementsQ.refetch();
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setIsPayingSettlement(false);
    }
  }

  const totalBilling = movements.reduce((sum, item) => sum + item.billing_value_cents, 0);
  const totalCenter = movements.reduce((sum, item) => sum + item.center_amount_cents, 0);
  const totalCollected = movements
    .filter((item) => item.collection_status === "collected")
    .reduce((sum, item) => sum + item.billing_value_cents, 0);
  const totalPendingProfessional = movements
    .filter((item) => item.professional_payment_status !== "paid")
    .reduce((sum, item) => sum + item.professional_fee_cents, 0);
  const settlementMovements = movementKinesiologistId
    ? movements.filter((item) => item.professional_payment_status !== "cancelled" && item.collection_status !== "cancelled")
    : [];
  const settlementTotal = settlementMovements.reduce((sum, item) => sum + item.professional_fee_cents, 0);
  const settlementPendingTotal = settlementMovements
    .filter((item) => item.professional_payment_status !== "paid")
    .reduce((sum, item) => sum + item.professional_fee_cents, 0);

  return (
    <main>
      <div className="max-w-6xl mx-auto p-6 space-y-6">
        <header>
          <h1 className="text-2xl font-semibold">{t("finance.title")}</h1>
          <p className="text-sm text-gray-600">{t("finance.subtitle")}</p>
        </header>

        {(message || error) && (
          <div className={`rounded-lg border p-3 text-sm ${error ? "border-red-200 bg-red-50 text-red-700" : "border-green-200 bg-green-50 text-green-700"}`}>
            {error || message}
          </div>
        )}

        <section className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <SummaryBox label={t("finance.billing")} value={formatMoney(totalBilling)} />
          <SummaryBox label={t("finance.collected")} value={formatMoney(totalCollected)} />
          <SummaryBox label={t("finance.pendingFees")} value={formatMoney(totalPendingProfessional)} />
          <SummaryBox label={t("finance.center")} value={formatMoney(totalCenter)} />
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div>
            <h2 className="text-lg font-semibold">{t("finance.financiersTitle")}</h2>
            <p className="text-sm text-gray-600">{t("finance.financiersSubtitle")}</p>
          </div>

          <form className="grid grid-cols-1 md:grid-cols-[1fr_180px_120px_auto] gap-3 items-end" onSubmit={submitFinancier}>
            <div>
              <label className="text-sm font-medium"><RequiredLabel required>{t("finance.name")}</RequiredLabel></label>
              <input className="mt-1 w-full border rounded-lg p-2" value={financierForm.name} onChange={(event) => setFinancierForm({ ...financierForm, name: event.target.value })} />
            </div>
            <div>
              <label className="text-sm font-medium">{t("finance.type")}</label>
              <select className="mt-1 w-full border rounded-lg p-2" value={financierForm.kind} onChange={(event) => setFinancierForm({ ...financierForm, kind: event.target.value })}>
                <option value="particular">{t("finance.particular")}</option>
                <option value="obra_social">{t("finance.obraSocial")}</option>
                <option value="prepaga">{t("finance.prepaga")}</option>
                <option value="convenio">{t("finance.convenio")}</option>
              </select>
            </div>
            <label className="flex items-center gap-2 text-sm pb-3">
              <input type="checkbox" checked={financierForm.active} onChange={(event) => setFinancierForm({ ...financierForm, active: event.target.checked })} />
              {t("finance.active")}
            </label>
            <div className="flex gap-2">
              <button className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50" disabled={saveFinancierM.isPending}>
                {saveFinancierM.isPending ? t("common.saving") : financierForm.id ? t("finance.save") : t("finance.create")}
              </button>
              {financierForm.id && <button type="button" className="px-3 py-2 rounded-lg border" onClick={() => setFinancierForm(emptyFinancierForm)}>{t("common.cancel")}</button>}
            </div>
          </form>

          <div className="divide-y">
            {financiers.map((financier) => (
              <div key={financier.id} className="py-2 flex items-center justify-between gap-3">
                <div>
                  <div className="font-medium">{financier.name}</div>
                  <div className="text-sm text-gray-600">{financier.kind} · {financier.active ? t("finance.active") : t("finance.inactive")}</div>
                </div>
                <button type="button" className="px-3 py-1 rounded-lg border text-sm" onClick={() => setFinancierForm(financier)}>
                  {t("finance.edit")}
                </button>
              </div>
            ))}
          </div>
        </section>

        <section id="tariffs-section" className="bg-white rounded-xl shadow p-4 space-y-4">
          <div>
            <h2 className="text-lg font-semibold">{t("finance.tariffsTitle")}</h2>
            <p className="text-sm text-gray-600">{t("finance.tariffsSubtitle")}</p>
          </div>

          <form className="grid grid-cols-1 md:grid-cols-3 gap-3" onSubmit={submitTariff}>
            <SelectField label={t("finance.practice")} value={tariffForm.practice_id} onChange={(value) => setTariffForm({ ...tariffForm, practice_id: value })}>
              <option value="">{t("finance.select")}</option>
              {activePractices.map((practice) => <option key={practice.id} value={practice.id}>{practice.name}</option>)}
            </SelectField>
            <SelectField label={t("finance.financier")} value={tariffForm.financier_id} onChange={(value) => setTariffForm({ ...tariffForm, financier_id: value })}>
              <option value="">{t("finance.select")}</option>
              {financiers.filter((item) => item.active).map((financier) => <option key={financier.id} value={financier.id}>{financier.name}</option>)}
            </SelectField>
            <InputField label={t("finance.billingValue")} value={tariffForm.billing_value} onChange={(value) => setTariffForm({ ...tariffForm, billing_value: value })} />
            <InputField label={t("finance.copay")} value={tariffForm.copay} onChange={(value) => setTariffForm({ ...tariffForm, copay: value })} />
            <InputField label={t("finance.currency")} value={tariffForm.currency} onChange={(value) => setTariffForm({ ...tariffForm, currency: value.toUpperCase() })} />
            <InputField label={t("finance.validFrom")} type="date" value={tariffForm.valid_from} onChange={(value) => setTariffForm({ ...tariffForm, valid_from: value })} />
            <InputField label={t("finance.validTo")} type="date" value={tariffForm.valid_to} onChange={(value) => setTariffForm({ ...tariffForm, valid_to: value })} required={false} />
            <label className="flex items-center gap-2 text-sm md:self-end md:pb-3">
              <input type="checkbox" checked={tariffForm.active} onChange={(event) => setTariffForm({ ...tariffForm, active: event.target.checked })} />
              {t("finance.activeFem")}
            </label>
            <div className="flex gap-2 md:self-end">
              <button className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50" disabled={saveTariffM.isPending}>
                {saveTariffM.isPending ? t("common.saving") : tariffForm.id ? t("finance.saveTariff") : t("finance.createTariff")}
              </button>
              {tariffForm.id && <button type="button" className="px-3 py-2 rounded-lg border" onClick={() => setTariffForm(emptyTariffForm)}>{t("common.cancel")}</button>}
            </div>
          </form>

          <div className="divide-y">
            {tariffs.map((tariff) => (
              <div key={tariff.id} className="py-2 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <div className="font-medium">{practiceById.get(tariff.practice_id)?.name ?? tariff.practice_id}</div>
                    <span className={`text-xs px-2 py-0.5 rounded-md ${tariff.active ? "bg-green-100 text-green-700" : "bg-red-100 text-red-700"}`}>
                      {tariff.active ? t("finance.activeFem") : t("finance.inactiveFem")}
                    </span>
                  </div>
                  <div className="text-sm text-gray-600">
                    {financierById.get(tariff.financier_id)?.name ?? tariff.financier_id} · {formatMoney(tariff.billing_value_cents)} · {t("finance.copayInline")} {formatMoney(tariff.copay_cents)} · {t("finance.since")} {tariff.valid_from}
                    {tariff.valid_to && ` · ${t("finance.until")} ${tariff.valid_to}`}
                  </div>
                </div>
                <button type="button" className="px-3 py-1 rounded-lg border text-sm" onClick={() => editTariff(tariff, setTariffForm)}>
                  {t("finance.edit")}
                </button>
              </div>
            ))}
          </div>
        </section>

        <section id="fee-rules-section" className="bg-white rounded-xl shadow p-4 space-y-4">
          <div>
            <h2 className="text-lg font-semibold">{t("finance.feeRulesTitle")}</h2>
            <p className="text-sm text-gray-600">{t("finance.feeRulesSubtitle")}</p>
          </div>

          <form className="grid grid-cols-1 md:grid-cols-3 gap-3" onSubmit={submitFeeRule}>
            <SelectField label={t("finance.kinesiologist")} value={feeRuleForm.kinesiologist_id} onChange={(value) => setFeeRuleForm({ ...feeRuleForm, kinesiologist_id: value })}>
              <option value="">{t("finance.select")}</option>
              {kinesiologists.filter((item) => item.active).map((kinesiologist) => (
                <option key={kinesiologist.id} value={kinesiologist.id}>{kinesiologist.last_name}, {kinesiologist.first_name}</option>
              ))}
            </SelectField>
            <SelectField label={t("finance.practice")} value={feeRuleForm.practice_id} onChange={(value) => setFeeRuleForm({ ...feeRuleForm, practice_id: value })}>
              <option value="">{t("finance.select")}</option>
              {activePractices.map((practice) => <option key={practice.id} value={practice.id}>{practice.name}</option>)}
            </SelectField>
            <SelectField label={t("finance.ruleType")} value={feeRuleForm.rule_type} onChange={(value) => setFeeRuleForm({ ...feeRuleForm, rule_type: value as "fixed" | "percentage" })}>
              <option value="percentage">{t("finance.percentage")}</option>
              <option value="fixed">{t("finance.fixedAmount")}</option>
            </SelectField>
            {feeRuleForm.rule_type === "percentage" ? (
              <InputField label={t("finance.percentage")} value={feeRuleForm.percentage} onChange={(value) => setFeeRuleForm({ ...feeRuleForm, percentage: value })} />
            ) : (
              <InputField label={t("finance.fixedAmount")} value={feeRuleForm.fixed_value} onChange={(value) => setFeeRuleForm({ ...feeRuleForm, fixed_value: value })} />
            )}
            <label className="flex items-center gap-2 text-sm md:self-end md:pb-3">
              <input type="checkbox" checked={feeRuleForm.active} onChange={(event) => setFeeRuleForm({ ...feeRuleForm, active: event.target.checked })} />
              {t("finance.activeFem")}
            </label>
            <div className="flex gap-2 md:self-end">
              <button className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50" disabled={saveFeeRuleM.isPending}>
                {saveFeeRuleM.isPending ? t("common.saving") : feeRuleForm.id ? t("finance.saveFee") : t("finance.createFee")}
              </button>
              {feeRuleForm.id && <button type="button" className="px-3 py-2 rounded-lg border" onClick={() => setFeeRuleForm(emptyFeeRuleForm)}>{t("common.cancel")}</button>}
            </div>
          </form>

          <div className="divide-y">
            {feeRules.map((rule) => (
              <div key={rule.id} className="py-2 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <div className="font-medium">{kinesiologistName(kinesiologistById.get(rule.kinesiologist_id), t)}</div>
                    <span className={`text-xs px-2 py-0.5 rounded-md ${rule.active ? "bg-green-100 text-green-700" : "bg-red-100 text-red-700"}`}>
                      {rule.active ? t("finance.activeFem") : t("finance.inactiveFem")}
                    </span>
                  </div>
                  <div className="text-sm text-gray-600">
                    {practiceById.get(rule.practice_id)?.name ?? rule.practice_id} · {rule.rule_type === "fixed" ? formatMoney(rule.fixed_value_cents ?? 0) : `${rule.percentage ?? 0}%`}
                  </div>
                </div>
                <button type="button" className="px-3 py-1 rounded-lg border text-sm" onClick={() => editFeeRule(rule, setFeeRuleForm)}>
                  {t("finance.edit")}
                </button>
              </div>
            ))}
          </div>
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div>
            <h2 className="text-lg font-semibold">{t("finance.movementsTitle")}</h2>
            <p className="text-sm text-gray-600">{t("finance.movementsSubtitle")}</p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
            <InputField label={t("finance.from")} type="date" value={movementFrom} onChange={setMovementFrom} required={false} />
            <InputField label={t("finance.to")} type="date" value={movementTo} onChange={setMovementTo} required={false} />
            <SelectField label={t("finance.kinesiologist")} value={movementKinesiologistId} onChange={setMovementKinesiologistId} required={false}>
              <option value="">{t("finance.all")}</option>
              {kinesiologists.map((kinesiologist) => (
                <option key={kinesiologist.id} value={kinesiologist.id}>
                  {kinesiologist.last_name}, {kinesiologist.first_name}
                </option>
              ))}
            </SelectField>
            <SelectField label={t("finance.practice")} value={movementPracticeId} onChange={setMovementPracticeId} required={false}>
              <option value="">{t("finance.allFem")}</option>
              {practices.map((practice) => (
                <option key={practice.id} value={practice.id}>
                  {practice.name}
                </option>
              ))}
            </SelectField>
            <SelectField label={t("finance.financier")} value={movementFinancierId} onChange={setMovementFinancierId} required={false}>
              <option value="">{t("finance.all")}</option>
              {financiers.map((financier) => (
                <option key={financier.id} value={financier.id}>
                  {financier.name}
                </option>
              ))}
            </SelectField>
            <SelectField label={t("finance.collectionStatus")} value={movementCollectionStatus} onChange={setMovementCollectionStatus} required={false}>
              <option value="">{t("finance.all")}</option>
              <option value="pending">{t("finance.pending")}</option>
              <option value="collected">{t("finance.collected")}</option>
              <option value="cancelled">{t("finance.cancelled")}</option>
            </SelectField>
            <SelectField label={t("finance.professionalPayment")} value={movementProfessionalPaymentStatus} onChange={setMovementProfessionalPaymentStatus} required={false}>
              <option value="">{t("finance.all")}</option>
              <option value="pending">{t("finance.pending")}</option>
              <option value="paid">{t("finance.paid")}</option>
              <option value="cancelled">{t("finance.cancelled")}</option>
            </SelectField>
            <div className="flex items-end">
              <button
                type="button"
                className="px-3 py-2 rounded-lg border text-sm hover:bg-gray-50"
                onClick={() => {
                  setMovementFrom(monthStartISO());
                  setMovementTo(todayISO());
                  setMovementKinesiologistId("");
                  setMovementPracticeId("");
                  setMovementFinancierId("");
                  setMovementCollectionStatus("");
                  setMovementProfessionalPaymentStatus("");
                }}
              >
                {t("finance.clearFilters")}
              </button>
            </div>
          </div>

          <div className="border rounded-lg p-3 space-y-3">
            <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
              <div>
                <h3 className="font-medium">{t("finance.settlementTitle")}</h3>
                <p className="text-sm text-gray-600">
                  {movementKinesiologistId
                    ? `${kinesiologistName(kinesiologistById.get(movementKinesiologistId), t)} · ${movementFrom || t("finance.settlementNoStart")} ${t("portal.errorOutsideScheduleMiddle")} ${movementTo || t("finance.settlementNoEnd")}`
                    : t("finance.settlementSelectKinesiologist")}
                </p>
              </div>
              <button
                type="button"
                className="px-3 py-2 rounded-lg bg-black text-white text-sm disabled:opacity-50"
                disabled={!movementKinesiologistId || settlementPendingTotal <= 0 || isPayingSettlement}
                onClick={markSettlementPaid}
              >
                {isPayingSettlement ? t("finance.settling") : t("finance.markSettlementPaid")}
              </button>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              <SummaryBox label={t("finance.sessions")} value={String(settlementMovements.length)} />
              <SummaryBox label={t("finance.periodFees")} value={formatMoney(settlementTotal)} />
              <SummaryBox label={t("finance.pendingPayment")} value={formatMoney(settlementPendingTotal)} />
            </div>
          </div>

          {movementsQ.isLoading && <p className="text-sm text-gray-600">{t("finance.loadingMovements")}</p>}
          {movementsQ.isError && <p className="text-sm text-red-600">{t("finance.errorLoadMovements")}</p>}
          {!movementsQ.isLoading && movements.length === 0 && <p className="text-sm text-gray-600">{t("finance.noMovements")}</p>}
          {movements.length > 0 && (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left border-b">
                    <th className="py-2 pr-3">{t("finance.colDate")}</th>
                    <th className="py-2 pr-3">{t("finance.colPatient")}</th>
                    <th className="py-2 pr-3">{t("finance.colKinesiologist")}</th>
                    <th className="py-2 pr-3">{t("finance.colPractice")}</th>
                    <th className="py-2 pr-3">{t("finance.colFinancier")}</th>
                    <th className="py-2 pr-3">{t("finance.colBilled")}</th>
                    <th className="py-2 pr-3">{t("finance.colCollection")}</th>
                    <th className="py-2 pr-3">{t("finance.colFee")}</th>
                    <th className="py-2 pr-3">{t("finance.colProfPayment")}</th>
                    <th className="py-2 pr-3">{t("finance.colCenter")}</th>
                    <th className="py-2 pr-3">{t("finance.colActions")}</th>
                  </tr>
                </thead>
                <tbody>
                  {movements.map((movement) => (
                    <tr key={movement.id} className="border-b last:border-b-0">
                      <td className="py-2 pr-3 whitespace-nowrap">{formatLocalDateTime(movement.created_at)}</td>
                      <td className="py-2 pr-3">{patientName(patientById.get(movement.patient_id), t)}</td>
                      <td className="py-2 pr-3">{kinesiologistName(kinesiologistById.get(movement.kinesiologist_id), t)}</td>
                      <td className="py-2 pr-3">{practiceById.get(movement.practice_id)?.name ?? movement.practice_id}</td>
                      <td className="py-2 pr-3">{financierById.get(movement.financier_id)?.name ?? movement.financier_id}</td>
                      <td className="py-2 pr-3">{formatMoney(movement.billing_value_cents)}</td>
                      <td className="py-2 pr-3">
                        {collectionStatusLabel(movement.collection_status, t)}
                        {movement.cancellation_reason && (
                          <div className="text-xs text-gray-500 mt-1">{t("finance.reasonLabel")}: {movement.cancellation_reason}</div>
                        )}
                      </td>
                      <td className="py-2 pr-3">{formatMoney(movement.professional_fee_cents)}</td>
                      <td className="py-2 pr-3">{professionalPaymentStatusLabel(movement.professional_payment_status, t)}</td>
                      <td className="py-2 pr-3">{formatMoney(movement.center_amount_cents)}</td>
                      <td className="py-2 pr-3">
                        <div className="flex flex-wrap gap-2">
                          {movement.collection_status !== "collected" ? (
                            <button
                              type="button"
                              className="px-2 py-1 rounded-lg border text-xs hover:bg-green-50 disabled:opacity-50"
                              disabled={updateMovementStatusM.isPending || movement.collection_status === "cancelled"}
                              onClick={() => updateMovementStatusM.mutate({ movement_id: movement.id, collection_status: "collected" })}
                            >
                              {t("finance.markCollected")}
                            </button>
                          ) : (
                            <button
                              type="button"
                              className="px-2 py-1 rounded-lg border text-xs hover:bg-gray-50 disabled:opacity-50"
                              disabled={updateMovementStatusM.isPending}
                              onClick={() => updateMovementStatusM.mutate({ movement_id: movement.id, collection_status: "pending" })}
                            >
                              {t("finance.reopenCollection")}
                            </button>
                          )}
                          {movement.professional_payment_status !== "paid" ? (
                            <button
                              type="button"
                              className="px-2 py-1 rounded-lg border text-xs hover:bg-green-50 disabled:opacity-50"
                              disabled={updateMovementStatusM.isPending || movement.professional_payment_status === "cancelled"}
                              onClick={() => updateMovementStatusM.mutate({ movement_id: movement.id, professional_payment_status: "paid" })}
                            >
                              {t("finance.payProfessional")}
                            </button>
                          ) : (
                            <button
                              type="button"
                              className="px-2 py-1 rounded-lg border text-xs hover:bg-gray-50 disabled:opacity-50"
                              disabled={updateMovementStatusM.isPending}
                              onClick={() => updateMovementStatusM.mutate({ movement_id: movement.id, professional_payment_status: "pending" })}
                            >
                              {t("finance.reopenPayment")}
                            </button>
                          )}
                          {movement.collection_status === "cancelled" || movement.professional_payment_status === "cancelled" ? (
                            <button
                              type="button"
                              className="px-2 py-1 rounded-lg border text-xs hover:bg-gray-50 disabled:opacity-50"
                              disabled={updateMovementStatusM.isPending}
                              onClick={() =>
                                updateMovementStatusM.mutate({
                                  movement_id: movement.id,
                                  collection_status: "pending",
                                  professional_payment_status: "pending",
                                })
                              }
                            >
                              {t("finance.reopenMovement")}
                            </button>
                          ) : (
                            <button
                              type="button"
                              className="px-2 py-1 rounded-lg border text-xs hover:bg-red-50 disabled:opacity-50"
                              disabled={updateMovementStatusM.isPending}
                              onClick={() => {
                                setCancellingMovementId(movement.id);
                                setCancellationReason("");
                              }}
                            >
                              {t("finance.cancelMovement")}
                            </button>
                          )}
                          {cancellingMovementId === movement.id && (
                            <div className="w-full min-w-56 rounded-lg border bg-gray-50 p-2 space-y-2">
                              <label className="text-xs font-medium">{t("finance.cancellationReason")}</label>
                              <textarea
                                className="w-full border rounded-lg p-2 text-xs min-h-20"
                                value={cancellationReason}
                                onChange={(event) => setCancellationReason(event.target.value)}
                              />
                              <div className="flex gap-2">
                                <button
                                  type="button"
                                  className="px-2 py-1 rounded-lg bg-black text-white text-xs disabled:opacity-50"
                                  disabled={!cancellationReason.trim() || updateMovementStatusM.isPending}
                                  onClick={() =>
                                    updateMovementStatusM.mutate({
                                      movement_id: movement.id,
                                      collection_status: "cancelled",
                                      professional_payment_status: "cancelled",
                                      cancellation_reason: cancellationReason.trim(),
                                    })
                                  }
                                >
                                  {t("finance.confirm")}
                                </button>
                                <button
                                  type="button"
                                  className="px-2 py-1 rounded-lg border text-xs"
                                  onClick={() => {
                                    setCancellingMovementId("");
                                    setCancellationReason("");
                                  }}
                                >
                                  {t("common.cancel")}
                                </button>
                              </div>
                            </div>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      </div>
    </main>
  );
}

function SummaryBox({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-white rounded-xl shadow p-4">
      <div className="text-sm text-gray-600">{label}</div>
      <div className="text-2xl font-semibold mt-1">{value}</div>
    </div>
  );
}

function InputField({
  label,
  value,
  onChange,
  type = "text",
  required = true,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
  required?: boolean;
}) {
  return (
    <div>
      <label className="text-sm font-medium"><RequiredLabel required={required}>{label}</RequiredLabel></label>
      <input className="mt-1 w-full border rounded-lg p-2" type={type} value={value} onChange={(event) => onChange(event.target.value)} />
    </div>
  );
}

function SelectField({
  label,
  value,
  onChange,
  children,
  required = true,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  children: ReactNode;
  required?: boolean;
}) {
  return (
    <div>
      <label className="text-sm font-medium"><RequiredLabel required={required}>{label}</RequiredLabel></label>
      <select className="mt-1 w-full border rounded-lg p-2" value={value} onChange={(event) => onChange(event.target.value)}>
        {children}
      </select>
    </div>
  );
}

function editTariff(tariff: PracticeTariff, setForm: (form: TariffForm) => void) {
  setForm({
    id: tariff.id,
    practice_id: tariff.practice_id,
    financier_id: tariff.financier_id,
    billing_value: centsToAmount(tariff.billing_value_cents),
    copay: centsToAmount(tariff.copay_cents),
    currency: tariff.currency,
    valid_from: tariff.valid_from,
    valid_to: tariff.valid_to ?? "",
    active: tariff.active,
  });
  document.getElementById("tariffs-section")?.scrollIntoView({ behavior: "smooth", block: "start" });
}

function editFeeRule(rule: ProfessionalFeeRule, setForm: (form: FeeRuleForm) => void) {
  setForm({
    id: rule.id,
    kinesiologist_id: rule.kinesiologist_id,
    practice_id: rule.practice_id,
    rule_type: rule.rule_type,
    fixed_value: centsToAmount(rule.fixed_value_cents ?? 0),
    percentage: String(rule.percentage ?? 0),
    active: rule.active,
  });
  document.getElementById("fee-rules-section")?.scrollIntoView({ behavior: "smooth", block: "start" });
}

function todayISO() {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
}

function monthStartISO() {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-01`;
}

function localDateStartISO(dateISO: string) {
  const [year, month, day] = dateISO.split("-").map(Number);
  return new Date(year, month - 1, day, 0, 0, 0, 0).toISOString();
}

function nextLocalDateStartISO(dateISO: string) {
  const [year, month, day] = dateISO.split("-").map(Number);
  const date = new Date(year, month - 1, day, 0, 0, 0, 0);
  date.setDate(date.getDate() + 1);
  return date.toISOString();
}

function parseNumber(value: string) {
  const normalized = value.replace(",", ".").trim();
  if (!normalized) return null;
  const parsed = Number(normalized);
  return Number.isFinite(parsed) ? parsed : null;
}

function amountToCents(value: string) {
  const parsed = parseNumber(value);
  if (parsed == null || parsed < 0) return null;
  return Math.round(parsed * 100);
}

function centsToAmount(value: number) {
  return (value / 100).toFixed(2);
}

function formatMoney(value: number) {
  return new Intl.NumberFormat("es-AR", {
    style: "currency",
    currency: "ARS",
    maximumFractionDigits: 2,
  }).format(value / 100);
}

function kinesiologistName(kinesiologist: { first_name: string; last_name: string } | undefined, t: Translate) {
  return kinesiologist ? `${kinesiologist.last_name}, ${kinesiologist.first_name}` : t("finance.kinesiologist");
}

function patientName(patient: { first_name: string; last_name: string } | undefined, t: Translate) {
  return patient ? `${patient.last_name}, ${patient.first_name}` : t("finance.colPatient");
}

function collectionStatusLabel(status: string, t: Translate) {
  if (status === "collected") return t("finance.collected");
  if (status === "cancelled") return t("finance.cancelled");
  return t("finance.pending");
}

function professionalPaymentStatusLabel(status: string, t: Translate) {
  if (status === "paid") return t("finance.paid");
  if (status === "cancelled") return t("finance.cancelled");
  return t("finance.pending");
}
