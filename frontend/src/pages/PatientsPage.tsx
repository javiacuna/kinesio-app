import { useEffect, useState } from "react";
import { apiFetch } from "../shared/api/http";

type Patient = {
  id: string;
  dni: string;
  first_name: string;
  last_name: string;
  email: string;
};

export default function PatientsPage() {
  const [dni, setDni] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");

  const [created, setCreated] = useState<Patient | null>(null);
  const [error, setError] = useState<string>("");

  useEffect(() => {
    const last = localStorage.getItem("last_patient_id");
    if (last) {
      // no hacemos GET por id acá para no agregar endpoints; solo mostramos que existe
    }
  }, []);

  async function submit() {
    setError("");
    setCreated(null);

    try {
      const res = await apiFetch<Patient>("/api/v1/patients", {
        method: "POST",
        body: JSON.stringify({
          dni,
          first_name: firstName,
          last_name: lastName,
          email,
        }),
      });

      setCreated(res);
      localStorage.setItem("last_patient_id", res.id);
    } catch (e: any) {
      setError(e?.message ?? "Error");
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-3xl mx-auto p-6 space-y-6">
        <header className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold">Pacientes</h1>
            <p className="text-sm text-gray-600">Alta de paciente (por ahora sin búsqueda/listado).</p>
          </div>
          <a className="text-sm underline" href="/">Ir a Agenda</a>
        </header>

        <div className="bg-white rounded-xl shadow p-4 space-y-3">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label className="text-sm font-medium">DNI</label>
              <input className="mt-1 w-full border rounded-lg p-2" value={dni} onChange={(e) => setDni(e.target.value)} />
            </div>
            <div>
              <label className="text-sm font-medium">Email</label>
              <input className="mt-1 w-full border rounded-lg p-2" value={email} onChange={(e) => setEmail(e.target.value)} />
            </div>
            <div>
              <label className="text-sm font-medium">Nombre</label>
              <input className="mt-1 w-full border rounded-lg p-2" value={firstName} onChange={(e) => setFirstName(e.target.value)} />
            </div>
            <div>
              <label className="text-sm font-medium">Apellido</label>
              <input className="mt-1 w-full border rounded-lg p-2" value={lastName} onChange={(e) => setLastName(e.target.value)} />
            </div>
          </div>

          <button className="px-4 py-2 rounded-lg bg-black text-white" onClick={submit}>
            Crear paciente
          </button>

          {error && <p className="text-sm text-red-600">Error: {error}</p>}

          {created && (
            <div className="mt-2 border rounded-lg p-3 bg-gray-50">
              <div className="font-medium">Paciente creado</div>
              <div className="text-sm text-gray-700">ID: <span className="font-mono">{created.id}</span></div>
              <div className="text-sm text-gray-700">{created.last_name}, {created.first_name} — {created.dni}</div>
              <p className="text-xs text-gray-500 mt-1">Se guardó como last_patient_id para usar en Agenda.</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
