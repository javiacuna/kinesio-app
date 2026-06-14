import { useMemo, useState } from "react";
import type { ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { getReportsSummary, type AppointmentDay, type LabelValue, type LowStockMaterial, type RevenueDay } from "@/features/reports/api";
import { listKinesiologists } from "@/features/kinesiologists/api";
import { listFinanciers } from "@/features/finance/api";
import { PatientSearch } from "@/features/patients/components/PatientSearch";
import type { Patient } from "@/features/patients/types";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

function todayISO() {
  const d = new Date();
  return d.toISOString().slice(0, 10);
}

function daysAgoISO(days: number) {
  const d = new Date();
  d.setDate(d.getDate() - days);
  return d.toISOString().slice(0, 10);
}

export default function ReportsPage() {
  const { language, t } = useLanguage();
  const [from, setFrom] = useState(daysAgoISO(29));
  const [to, setTo] = useState(todayISO());
  const [kinesiologistId, setKinesiologistId] = useState("");
  const [financierId, setFinancierId] = useState("");
  const [patient, setPatient] = useState<Patient | null>(null);
  const [patientSearchResetKey, setPatientSearchResetKey] = useState(0);

  const kinesiologistsQ = useQuery({
    queryKey: ["kinesiologists", "list"],
    queryFn: () => listKinesiologists(),
  });

  const financiersQ = useQuery({
    queryKey: ["financiers", "list"],
    queryFn: () => listFinanciers(),
  });

  const summaryQ = useQuery({
    queryKey: ["reports", "summary", from, to, kinesiologistId, financierId, patient?.id],
    queryFn: () =>
      getReportsSummary({
        from,
        to,
        kinesiologist_id: kinesiologistId || undefined,
        financier_id: financierId || undefined,
        patient_id: patient?.id || undefined,
      }),
  });

  const hasActiveFilters = Boolean(kinesiologistId || financierId || patient);

  function clearFilters() {
    setKinesiologistId("");
    setFinancierId("");
    setPatient(null);
    setPatientSearchResetKey((key) => key + 1);
  }

  const summary = summaryQ.data;
  const appointmentsByDay = asArray(summary?.appointments_by_day);
  const revenueByDay = asArray(summary?.revenue_by_day);
  const appointmentsByStatus = asArray(summary?.appointments_by_status);
  const appointmentsByKinesiologist = asArray(summary?.appointments_by_kinesiologist);
  const pendingLoansByMaterial = asArray(summary?.pending_loans_by_material);
  const lowStockMaterials = asArray(summary?.low_stock_materials);
  const money = useMemo(
    () =>
      new Intl.NumberFormat(language === "en" ? "en-US" : "es-AR", {
        style: "currency",
        currency: "ARS",
        maximumFractionDigits: 0,
      }),
    [language],
  );

  return (
    <main className="max-w-6xl mx-auto p-6 space-y-6">
      <header className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h1 className="text-2xl font-semibold">{t("reports.title")}</h1>
          <p className="text-sm text-gray-600">{t("reports.subtitle")}</p>
        </div>

        <form className="grid grid-cols-1 sm:grid-cols-3 gap-3 bg-white border rounded-lg p-3">
          <label className="text-sm font-medium">
            {t("reports.from")}
            <input className="mt-1 w-full border rounded-lg p-2" type="date" value={from} onChange={(e) => setFrom(e.target.value)} />
          </label>
          <label className="text-sm font-medium">
            {t("reports.to")}
            <input className="mt-1 w-full border rounded-lg p-2" type="date" value={to} onChange={(e) => setTo(e.target.value)} />
          </label>
          <div className="flex items-end">
            <button type="button" className="w-full border rounded-lg px-3 py-2 hover:bg-gray-100" onClick={() => { setFrom(daysAgoISO(29)); setTo(todayISO()); }}>
              {t("reports.last30")}
            </button>
          </div>
        </form>
      </header>

      <form className="grid grid-cols-1 sm:grid-cols-3 lg:grid-cols-4 gap-3 bg-white border rounded-lg p-3">
        <label className="text-sm font-medium">
          {t("reports.filterKinesiologist")}
          <select
            className="mt-1 w-full border rounded-lg p-2"
            value={kinesiologistId}
            onChange={(e) => setKinesiologistId(e.target.value)}
          >
            <option value="">{t("reports.filterKinesiologistAll")}</option>
            {(kinesiologistsQ.data ?? []).map((k) => (
              <option key={k.id} value={k.id}>
                {k.last_name}, {k.first_name}
              </option>
            ))}
          </select>
        </label>

        <label className="text-sm font-medium">
          {t("reports.filterFinancier")}
          <select
            className="mt-1 w-full border rounded-lg p-2"
            value={financierId}
            onChange={(e) => setFinancierId(e.target.value)}
          >
            <option value="">{t("reports.filterFinancierAll")}</option>
            {(financiersQ.data ?? []).map((f) => (
              <option key={f.id} value={f.id}>
                {f.name}
              </option>
            ))}
          </select>
        </label>

        <div className="sm:col-span-1 lg:col-span-1">
          <PatientSearch
            valuePatientId={patient?.id ?? ""}
            onSelect={setPatient}
            label={t("reports.filterPatient")}
            showSelected={false}
            resetKey={patientSearchResetKey}
          />
        </div>

        <div className="flex items-end">
          <button
            type="button"
            className="w-full border rounded-lg px-3 py-2 hover:bg-gray-100 disabled:opacity-50"
            onClick={clearFilters}
            disabled={!hasActiveFilters}
          >
            {t("reports.clearFilters")}
          </button>
        </div>
      </form>

      {summaryQ.isLoading && <p className="text-sm text-gray-600">{t("reports.loading")}</p>}
      {summaryQ.isError && <p className="text-sm text-red-600">{t("reports.error")}</p>}

      {summary && (
        <>
          <section className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-3">
            <Metric title={t("reports.appointments")} value={summary.totals.appointments_total} />
            <Metric title={t("reports.cancelled")} value={summary.totals.cancelled_appointments} />
            <Metric title={t("reports.billing")} value={money.format(summary.totals.billing_value_cents / 100)} />
            <Metric title={t("reports.pendingCollection")} value={money.format(summary.totals.pending_collection_cents / 100)} />
            <Metric title={t("reports.activePatients")} value={summary.totals.active_patients} />
            <Metric title={t("reports.pendingLoans")} value={summary.totals.pending_material_loans} />
            <Metric title={t("reports.lowStock")} value={summary.totals.low_stock_materials} />
            <Metric title={t("reports.professionalFees")} value={money.format(summary.totals.professional_fee_cents / 100)} />
          </section>

          <section className="grid grid-cols-1 xl:grid-cols-2 gap-6">
            <Panel title={t("reports.appointmentsByDay")}>
              <StackedBars
                data={appointmentsByDay}
                getLabel={(item) => compactDate(item.date)}
                getFirst={(item) => item.scheduled}
                getSecond={(item) => item.cancelled}
                firstLabel={t("reports.scheduled")}
                secondLabel={t("reports.cancelled")}
                emptyText={t("reports.noData")}
              />
            </Panel>

            <Panel title={t("reports.revenueByDay")}>
              <MoneyBars
                data={revenueByDay}
                getLabel={(item) => compactDate(item.date)}
                getValue={(item) => item.billing_value_cents}
                format={(value) => money.format(value / 100)}
                emptyText={t("reports.noData")}
              />
            </Panel>

            <Panel title={t("reports.byKinesiologist")}>
              <HorizontalBars data={appointmentsByKinesiologist} emptyText={t("reports.noData")} />
            </Panel>

            <Panel title={t("reports.pendingLoansByMaterial")}>
              <HorizontalBars data={pendingLoansByMaterial} emptyText={t("reports.noData")} />
            </Panel>
          </section>

          <section className="grid grid-cols-1 xl:grid-cols-2 gap-6">
            <Panel title={t("reports.status")}>
              <HorizontalBars data={appointmentsByStatus.map((item) => ({ ...item, label: statusLabel(item.label, t) }))} emptyText={t("reports.noData")} />
            </Panel>

            <Panel title={t("reports.lowStockList")}>
              <LowStockList items={lowStockMaterials} emptyText={t("reports.noLowStock")} />
            </Panel>
          </section>
        </>
      )}
    </main>
  );
}

function asArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}

function Metric({ title, value }: { title: string; value: string | number }) {
  return (
    <section className="bg-white rounded-lg border p-4">
      <div className="text-sm text-gray-600">{title}</div>
      <div className="text-2xl font-semibold mt-1 break-words">{value}</div>
    </section>
  );
}

function Panel({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="bg-white rounded-lg border p-4 space-y-4">
      <h2 className="text-lg font-semibold">{title}</h2>
      {children}
    </section>
  );
}

function StackedBars({
  data,
  getLabel,
  getFirst,
  getSecond,
  firstLabel,
  secondLabel,
  emptyText,
}: {
  data: AppointmentDay[];
  getLabel: (item: AppointmentDay) => string;
  getFirst: (item: AppointmentDay) => number;
  getSecond: (item: AppointmentDay) => number;
  firstLabel: string;
  secondLabel: string;
  emptyText: string;
}) {
  const max = Math.max(1, ...data.map((item) => getFirst(item) + getSecond(item)));
  if (data.length === 0) return <p className="text-sm text-gray-600">{emptyText}</p>;

  return (
    <div className="space-y-3">
      <div className="flex gap-4 text-xs text-gray-600">
        <Legend className="bg-black" label={firstLabel} />
        <Legend className="bg-gray-400" label={secondLabel} />
      </div>
      <div className="space-y-2">
        {data.map((item) => {
          const first = getFirst(item);
          const second = getSecond(item);
          const total = first + second;
          return (
            <div key={getLabel(item)} className="grid grid-cols-[64px_1fr_40px] items-center gap-2 text-sm">
              <span className="text-gray-600">{getLabel(item)}</span>
              <div className="h-6 bg-gray-100 rounded overflow-hidden flex">
                <div className="bg-black" style={{ width: `${(first / max) * 100}%` }} />
                <div className="bg-gray-400" style={{ width: `${(second / max) * 100}%` }} />
              </div>
              <span className="text-right tabular-nums">{total}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function MoneyBars({ data, getLabel, getValue, format, emptyText }: { data: RevenueDay[]; getLabel: (item: RevenueDay) => string; getValue: (item: RevenueDay) => number; format: (value: number) => string; emptyText: string }) {
  const max = Math.max(1, ...data.map(getValue));
  if (data.length === 0) return <p className="text-sm text-gray-600">{emptyText}</p>;

  return (
    <div className="space-y-2">
      {data.map((item) => {
        const value = getValue(item);
        return (
          <div key={getLabel(item)} className="grid grid-cols-[64px_1fr_88px] items-center gap-2 text-sm">
            <span className="text-gray-600">{getLabel(item)}</span>
            <div className="h-6 bg-gray-100 rounded overflow-hidden">
              <div className="h-full bg-black" style={{ width: `${(value / max) * 100}%` }} />
            </div>
            <span className="text-right tabular-nums text-xs">{format(value)}</span>
          </div>
        );
      })}
    </div>
  );
}

function HorizontalBars({ data, emptyText }: { data: LabelValue[]; emptyText: string }) {
  const max = Math.max(1, ...data.map((item) => item.value));
  if (data.length === 0) return <p className="text-sm text-gray-600">{emptyText}</p>;

  return (
    <div className="space-y-3">
      {data.map((item) => (
        <div key={item.label} className="space-y-1">
          <div className="flex justify-between gap-3 text-sm">
            <span className="truncate">{item.label}</span>
            <span className="tabular-nums">{item.value}</span>
          </div>
          <div className="h-3 bg-gray-100 rounded overflow-hidden">
            <div className="h-full bg-black" style={{ width: `${(item.value / max) * 100}%` }} />
          </div>
        </div>
      ))}
    </div>
  );
}

function LowStockList({ items, emptyText }: { items: LowStockMaterial[]; emptyText: string }) {
  if (items.length === 0) return <p className="text-sm text-gray-600">{emptyText}</p>;

  return (
    <div className="divide-y">
      {items.map((item) => (
        <div key={item.id} className="py-3 flex items-center justify-between gap-3">
          <span className="font-medium">{item.name}</span>
          <span className="text-sm text-gray-600 tabular-nums">
            {item.available_qty} / {item.total_qty}
          </span>
        </div>
      ))}
    </div>
  );
}

function Legend({ className, label }: { className: string; label: string }) {
  return (
    <span className="inline-flex items-center gap-2">
      <span className={`h-3 w-3 rounded-sm ${className}`} />
      {label}
    </span>
  );
}

function compactDate(value: string) {
  const [, month, day] = value.split("-");
  return `${day}/${month}`;
}

function statusLabel(value: string, t: (key: string) => string) {
  if (value === "scheduled") return t("reports.scheduled");
  if (value === "cancelled") return t("reports.cancelled");
  return value;
}
