import { useEffect, useMemo, useState, type FormEvent } from "react";
import {
  createKinesiologist,
  listKinesiologists,
  updateKinesiologist,
  type SaveKinesiologistInput,
} from "@/features/kinesiologists/api";
import type { Kinesiologist } from "@/features/kinesiologists/types";
import { assignUserRole, listAdminUsers, type AdminUser } from "@/features/auth/adminApi";
import type { AuthRole } from "@/features/auth/types";

const emptyForm: SaveKinesiologistInput = {
  first_name: "",
  last_name: "",
  email: "",
  license_number: "",
  active: true,
};

export default function StaffPage() {
  const [items, setItems] = useState<Kinesiologist[]>([]);
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [form, setForm] = useState<SaveKinesiologistInput>(emptyForm);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [roleEmail, setRoleEmail] = useState("");
  const [role, setRole] = useState<AuthRole>("kinesiologo");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [isSaving, setIsSaving] = useState(false);
  const [isAssigningRole, setIsAssigningRole] = useState(false);

  const sortedItems = useMemo(
    () => [...items].sort((a, b) => `${a.last_name} ${a.first_name}`.localeCompare(`${b.last_name} ${b.first_name}`)),
    [items],
  );
  const sortedUsers = useMemo(
    () => [...users].sort((a, b) => a.email.localeCompare(b.email)),
    [users],
  );

  useEffect(() => {
    refresh();
    refreshUsers();
  }, []);

  async function refresh() {
    setError("");
    try {
      setItems(await listKinesiologists({ includeInactive: true }));
    } catch (err) {
      setError((err as Error)?.message ?? "No se pudo cargar el equipo.");
    }
  }

  async function refreshUsers() {
    setError("");
    try {
      setUsers(await listAdminUsers());
    } catch (err) {
      setError((err as Error)?.message ?? "No se pudieron cargar los usuarios.");
    }
  }

  async function submitKinesiologist(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setMessage("");
    setIsSaving(true);

    try {
      const payload = {
        ...form,
        email: form.email.trim().toLowerCase(),
        license_number: form.license_number?.trim() || null,
      };

      const saved = editingId
        ? await updateKinesiologist({ ...payload, id: editingId })
        : await createKinesiologist(payload);

      setItems((current) => {
        const exists = current.some((item) => item.id === saved.id);
        return exists ? current.map((item) => (item.id === saved.id ? saved : item)) : [...current, saved];
      });
      setMessage(editingId ? "Kinesiólogo actualizado." : "Kinesiólogo creado.");
      resetForm();
    } catch (err) {
      setError((err as Error)?.message ?? "No se pudo guardar el kinesiólogo.");
    } finally {
      setIsSaving(false);
    }
  }

  async function submitRole(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setMessage("");
    setIsAssigningRole(true);

    try {
      const res = await assignUserRole({ email: roleEmail.trim().toLowerCase(), role });
      setMessage(`Rol ${res.role} asignado a ${res.email}. Debe cerrar sesión y volver a entrar.`);
      setRoleEmail("");
      await refreshUsers();
    } catch (err) {
      setError((err as Error)?.message ?? "No se pudo asignar el rol.");
    } finally {
      setIsAssigningRole(false);
    }
  }

  function edit(item: Kinesiologist) {
    setEditingId(item.id);
    setForm({
      first_name: item.first_name,
      last_name: item.last_name,
      email: item.email,
      license_number: item.license_number ?? "",
      active: item.active,
    });
    setMessage("");
    setError("");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function resetForm() {
    setEditingId(null);
    setForm(emptyForm);
  }

  return (
    <main>
      <div className="max-w-5xl mx-auto p-6 space-y-6">
        <header>
          <h1 className="text-2xl font-semibold">Equipo</h1>
          <p className="text-sm text-gray-600">Administración de profesionales y roles de acceso.</p>
        </header>

        {(message || error) && (
          <div className={`rounded-lg border p-3 text-sm ${error ? "border-red-200 bg-red-50 text-red-700" : "border-green-200 bg-green-50 text-green-700"}`}>
            {error || message}
          </div>
        )}

        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div>
            <h2 className="text-lg font-semibold">{editingId ? "Editar kinesiólogo" : "Crear kinesiólogo"}</h2>
            <p className="text-sm text-gray-600">El email debe coincidir con el usuario de Firebase para que vea sus turnos.</p>
          </div>

          <form className="grid grid-cols-1 md:grid-cols-2 gap-3" onSubmit={submitKinesiologist}>
            <Field label="Nombre" value={form.first_name} onChange={(value) => setForm({ ...form, first_name: value })} />
            <Field label="Apellido" value={form.last_name} onChange={(value) => setForm({ ...form, last_name: value })} />
            <Field label="Email" type="email" value={form.email} onChange={(value) => setForm({ ...form, email: value })} />
            <Field label="Matrícula" value={form.license_number ?? ""} onChange={(value) => setForm({ ...form, license_number: value })} />

            <label className="flex items-center gap-2 text-sm md:col-span-2">
              <input
                type="checkbox"
                checked={form.active}
                onChange={(event) => setForm({ ...form, active: event.target.checked })}
              />
              Activo
            </label>

            <div className="flex gap-2 md:col-span-2">
              <button className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50" disabled={isSaving}>
                {isSaving ? "Guardando..." : editingId ? "Guardar cambios" : "Crear kinesiólogo"}
              </button>
              {editingId && (
                <button type="button" className="px-4 py-2 rounded-lg border" onClick={resetForm}>
                  Cancelar
                </button>
              )}
            </div>
          </form>
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div>
            <h2 className="text-lg font-semibold">Asignar rol</h2>
            <p className="text-sm text-gray-600">El usuario debe existir en Firebase Authentication.</p>
          </div>

          <form className="grid grid-cols-1 md:grid-cols-[1fr_220px_auto] gap-3" onSubmit={submitRole}>
            <Field label="Email" type="email" value={roleEmail} onChange={setRoleEmail} />
            <div>
              <label className="text-sm font-medium">Rol</label>
              <select
                className="mt-1 w-full border rounded-lg p-2"
                value={role}
                onChange={(event) => setRole(event.target.value as AuthRole)}
              >
                <option value="admin">Admin</option>
                <option value="recepcionista">Recepcionista</option>
                <option value="kinesiologo">Kinesiólogo</option>
                <option value="paciente">Paciente</option>
              </select>
            </div>
            <div className="flex items-end">
              <button className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50" disabled={isAssigningRole}>
                {isAssigningRole ? "Asignando..." : "Asignar"}
              </button>
            </div>
          </form>
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold">Usuarios Firebase</h2>
              <p className="text-sm text-gray-600">Usuarios existentes y rol configurado en custom claims.</p>
            </div>
            <button type="button" className="px-3 py-2 rounded-lg border text-sm" onClick={refreshUsers}>
              Actualizar
            </button>
          </div>

          <div className="divide-y">
            {sortedUsers.length === 0 ? (
              <p className="text-sm text-gray-600 py-3">No hay usuarios para mostrar.</p>
            ) : (
              sortedUsers.map((user) => (
                <div key={user.uid} className="py-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <div className="font-medium">{user.email}</div>
                    <div className="text-xs text-gray-500 font-mono">{user.uid}</div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`text-xs px-2 py-1 rounded-md ${user.role ? "bg-blue-100 text-blue-700" : "bg-gray-100 text-gray-700"}`}>
                      {user.role || "sin rol"}
                    </span>
                    {user.disabled && (
                      <span className="text-xs px-2 py-1 rounded-md bg-red-100 text-red-700">
                        Deshabilitado
                      </span>
                    )}
                    <button
                      type="button"
                      className="px-3 py-1 rounded-lg border text-sm"
                      onClick={() => {
                        setRoleEmail(user.email);
                        setRole((user.role || "kinesiologo") as AuthRole);
                        window.scrollTo({ top: 0, behavior: "smooth" });
                      }}
                    >
                      Cambiar rol
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Kinesiólogos</h2>
            <button type="button" className="px-3 py-2 rounded-lg border text-sm" onClick={refresh}>
              Actualizar
            </button>
          </div>

          <div className="divide-y">
            {sortedItems.length === 0 ? (
              <p className="text-sm text-gray-600 py-3">No hay kinesiólogos cargados.</p>
            ) : (
              sortedItems.map((item) => (
                <div key={item.id} className="py-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <div className="font-medium">{item.last_name}, {item.first_name}</div>
                    <div className="text-sm text-gray-600">{item.email}</div>
                    {item.license_number && <div className="text-sm text-gray-600">Matrícula: {item.license_number}</div>}
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`text-xs px-2 py-1 rounded-md ${item.active ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-700"}`}>
                      {item.active ? "Activo" : "Inactivo"}
                    </span>
                    <button type="button" className="px-3 py-1 rounded-lg border text-sm" onClick={() => edit(item)}>
                      Editar
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </section>
      </div>
    </main>
  );
}

function Field({
  label,
  value,
  onChange,
  type = "text",
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
}) {
  return (
    <div>
      <label className="text-sm font-medium">{label}</label>
      <input
        className="mt-1 w-full border rounded-lg p-2"
        type={type}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        required={label !== "Matrícula"}
      />
    </div>
  );
}
