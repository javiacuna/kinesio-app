import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { searchPatients } from "../api";
import type { Patient } from "../types";

type Props = {
  valuePatientId: string;
  onSelect: (p: Patient) => void;
  placeholder?: string;
};

export function PatientSearch({ valuePatientId, onSelect, placeholder }: Props) {
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);

  // Debounce simple (evita pegarle al backend en cada tecla)
  const [debounced, setDebounced] = useState("");
  useEffect(() => {
    const t = setTimeout(() => setDebounced(query.trim()), 250);
    return () => clearTimeout(t);
  }, [query]);

  const enabled = debounced.length >= 3;

  const q = useQuery({
    queryKey: ["patients", "search", debounced],
    queryFn: () => searchPatients(debounced, 20),
    enabled,
  });

  const items = useMemo(() => q.data ?? [], [q.data]);

  return (
    <div className="relative">
      <label className="text-sm font-medium">Paciente</label>
      <input
        className="mt-1 w-full border rounded-lg p-2"
        value={query}
        placeholder={placeholder ?? "Buscar por DNI, email o apellido (mín. 3 caracteres)…"}
        onChange={(e) => {
          setQuery(e.target.value);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
      />

      <div className="mt-2 text-xs text-gray-500">
        Seleccionado: <span className="font-mono">{valuePatientId || "—"}</span>
      </div>

      {open && enabled && (
        <div className="absolute z-10 mt-2 w-full rounded-lg border bg-white shadow">
          {q.isLoading && <div className="p-3 text-sm text-gray-600">Buscando…</div>}

          {q.isError && (
            <div className="p-3 text-sm text-red-600">
              Error: {String((q.error as any)?.message)}
            </div>
          )}

          {!q.isLoading && !q.isError && items.length === 0 && (
            <div className="p-3 text-sm text-gray-600">Sin resultados.</div>
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
