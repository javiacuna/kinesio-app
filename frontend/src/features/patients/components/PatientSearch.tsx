import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { searchPatients } from "../api";
import type { Patient } from "../types";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

type Props = {
  valuePatientId: string;
  onSelect: (p: Patient) => void;
  label?: string | null;
  placeholder?: string;
  selectedText?: string;
  showSelected?: boolean;
  resetKey?: number;
};

export function PatientSearch({
  valuePatientId,
  onSelect,
  label,
  placeholder,
  selectedText,
  showSelected = true,
  resetKey,
}: Props) {
  const { t } = useLanguage();
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);

  // Debounce simple (evita pegarle al backend en cada tecla)
  const [debounced, setDebounced] = useState("");
  useEffect(() => {
    const t = setTimeout(() => setDebounced(query.trim()), 250);
    return () => clearTimeout(t);
  }, [query]);

  useEffect(() => {
    setQuery("");
    setDebounced("");
    setOpen(false);
  }, [resetKey]);

  useEffect(() => {
    if (!valuePatientId) {
      setQuery("");
      setDebounced("");
      setOpen(false);
    }
  }, [valuePatientId]);

  const enabled = debounced.length >= 3;

  const q = useQuery({
    queryKey: ["patients", "search", debounced],
    queryFn: () => searchPatients(debounced, 20),
    enabled,
  });

  const items = useMemo(() => (q.data ?? []).filter((patient) => patient.active), [q.data]);

  return (
    <div className="relative">
      {label !== null && (
        <label className="text-sm font-medium">{label ?? t("agenda.patient")}</label>
      )}
      <input
        className={`${label === null ? "" : "mt-1"} w-full border rounded-lg p-2`}
        value={query}
        placeholder={placeholder ?? t("agenda.patientSearchDefault")}
        onChange={(e) => {
          setQuery(e.target.value);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
      />

      {showSelected && (
        <div className="mt-2 text-xs text-gray-500">
          {t("agenda.selected")}: <span>{selectedText || valuePatientId || "—"}</span>
        </div>
      )}

      {open && enabled && (
        <div className="absolute z-10 mt-2 w-full rounded-lg border bg-white shadow">
          {q.isLoading && <div className="p-3 text-sm text-gray-600">{t("agenda.searching")}</div>}

          {q.isError && (
            <div className="p-3 text-sm text-red-600">
              Error: {String((q.error as any)?.message)}
            </div>
          )}

          {!q.isLoading && !q.isError && items.length === 0 && (
            <div className="p-3 text-sm text-gray-600">{t("agenda.searchNoResults")}</div>
          )}

          {!q.isLoading && !q.isError && items.length > 0 && (
            <div className="max-h-56 overflow-auto divide-y">
              {items.map((p) => (
                <button
                  key={p.id}
                  type="button"
                  className="w-full text-left px-3 py-2 hover:bg-gray-50"
                  onClick={() => {
                    onSelect(p);
                    setOpen(false);
                    setQuery(`${p.last_name}, ${p.first_name} — DNI ${p.dni}`);
                  }}
                >
                  <div className="text-sm font-medium">
                    {p.last_name}, {p.first_name}
                  </div>
                  <div className="text-xs text-gray-600">
                    {p.email} • DNI {p.dni}
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
