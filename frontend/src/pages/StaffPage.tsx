import { useEffect, useMemo, useState, type FormEvent } from "react";
import {
  createKinesiologist,
  listKinesiologists,
  updateKinesiologist,
  type SaveKinesiologistInput,
} from "@/features/kinesiologists/api";
import type { Kinesiologist } from "@/features/kinesiologists/types";
import { inviteUserAccess, listAdminUsers, type AdminUser } from "@/features/auth/adminApi";
import type { AuthRole } from "@/features/auth/types";
import {
  createStaffMember,
  listStaffMembers,
  updateStaffMember,
  type SaveStaffMemberInput,
} from "@/features/staff/api";
import type { StaffMember, StaffRole } from "@/features/staff/types";

const emptyForm: SaveKinesiologistInput = {
  first_name: "",
  last_name: "",
  email: "",
  license_number: "",
  active: true,
};

const emptyStaffForm: SaveStaffMemberInput = {
  first_name: "",
  last_name: "",
  email: "",
  role: "recepcionista",
  phone: "",
  active: true,
};

export default function StaffPage() {
  const [staffMembers, setStaffMembers] = useState<StaffMember[]>([]);
  const [items, setItems] = useState<Kinesiologist[]>([]);
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [staffForm, setStaffForm] = useState<SaveStaffMemberInput>(emptyStaffForm);
  const [editingStaffId, setEditingStaffId] = useState<string | null>(null);
  const [sendStaffAccess, setSendStaffAccess] = useState(true);
  const [form, setForm] = useState<SaveKinesiologistInput>(emptyForm);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [roleEmail, setRoleEmail] = useState("");
  const [role, setRole] = useState<AuthRole>("recepcionista");
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
  const sortedStaffMembers = useMemo(
    () => [...staffMembers].sort((a, b) => `${a.last_name} ${a.first_name}`.localeCompare(`${b.last_name} ${b.first_name}`)),
    [staffMembers],
  );

  useEffect(() => {
    refreshStaff();
    refresh();
    refreshUsers();
  }, []);

  async function refreshStaff() {
    setError("");
    try {
      setStaffMembers(await listStaffMembers({ includeInactive: true }));
    } catch (err) {
      setError((err as Error)?.message ?? "No se pudo cargar el staff.");
    }
  }

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

  async function submitStaffMember(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setMessage("");
    setIsSaving(true);

    try {
      const payload = {
        ...staffForm,
        email: staffForm.email.trim().toLowerCase(),
        phone: staffForm.phone?.trim() || null,
      };
      const saved = editingStaffId
        ? await updateStaffMember({ ...payload, id: editingStaffId })
        : await createStaffMember(payload);

      setStaffMembers((current) => {
        const exists = current.some((item) => item.id === saved.id);
        return exists ? current.map((item) => (item.id === saved.id ? saved : item)) : [...current, saved];
      });

      if (sendStaffAccess) {
        await inviteUserAccess({ email: saved.email, role: saved.role });
        await refreshUsers();
        setMessage(`${saved.first_name} ${saved.last_name} guardado e invitación enviada.`);
      } else {
        setMessage(`${saved.first_name} ${saved.last_name} guardado.`);
      }

      resetStaffForm();
    } catch (err) {
      setError((err as Error)?.message ?? "No se pudo guardar el miembro del equipo.");
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
      const res = await inviteUserAccess({ email: roleEmail.trim().toLowerCase(), role });
      setMessage(
        res.created
          ? `Usuario creado e invitación enviada a ${res.email}.`
          : `Invitación enviada a ${res.email} y rol ${res.role} confirmado.`,
      );
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

  function editStaff(member: StaffMember) {
    setEditingStaffId(member.id);
    setStaffForm({
      first_name: member.first_name,
      last_name: member.last_name,
      email: member.email,
      role: member.role,
      phone: member.phone ?? "",
      active: member.active,
      firebase_uid: member.firebase_uid ?? null,
    });
    setSendStaffAccess(false);
    setMessage("");
    setError("");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function resetStaffForm() {
    setEditingStaffId(null);
    setStaffForm(emptyStaffForm);
    setSendStaffAccess(true);
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
            <h2 className="text-lg font-semibold">{editingStaffId ? "Editar miembro del equipo" : "Crear miembro del equipo"}</h2>
            <p className="text-sm text-gray-600">Usá este formulario para admins, recepcionistas y kinesiólogos.</p>
          </div>

          <form className="grid grid-cols-1 md:grid-cols-2 gap-3" onSubmit={submitStaffMember}>
            <Field label="Nombre" value={staffForm.first_name} onChange={(value) => setStaffForm({ ...staffForm, first_name: value })} />
            <Field label="Apellido" value={staffForm.last_name} onChange={(value) => setStaffForm({ ...staffForm, last_name: value })} />
            <Field label="Email" type="email" value={staffForm.email} onChange={(value) => setStaffForm({ ...staffForm, email: value })} />
            <Field label="Teléfono" value={staffForm.phone ?? ""} onChange={(value) => setStaffForm({ ...staffForm, phone: value })} required={false} />

            <div>
              <label className="text-sm font-medium">Rol</label>
              <select
                className="mt-1 w-full border rounded-lg p-2"
                value={staffForm.role}
                onChange={(event) => setStaffForm({ ...staffForm, role: event.target.value as StaffRole })}
              >
                <option value="recepcionista">Recepcionista</option>
                <option value="admin">Admin</option>
                <option value="kinesiologo">Kinesiólogo</option>
              </select>
            </div>

            <div className="flex flex-col justify-end gap-2">
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={staffForm.active}
                  onChange={(event) => setStaffForm({ ...staffForm, active: event.target.checked })}
                />
                Activo
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={sendStaffAccess}
                  onChange={(event) => setSendStaffAccess(event.target.checked)}
                />
                Enviar acceso por email
              </label>
            </div>

            <div className="flex gap-2 md:col-span-2">
              <button className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50" disabled={isSaving}>
                {isSaving ? "Guardando..." : editingStaffId ? "Guardar cambios" : "Crear miembro"}
              </button>
              {editingStaffId && (
                <button type="button" className="px-4 py-2 rounded-lg border" onClick={resetStaffForm}>
                  Cancelar
                </button>
              )}
            </div>
          </form>
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div>
            <h2 className="text-lg font-semibold">Reenviar acceso</h2>
            <p className="text-sm text-gray-600">
              Para resetear contraseña o corregir un rol sin editar el perfil interno.
            </p>
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
                <option value="recepcionista">Recepcionista</option>
                <option value="admin">Admin</option>
                <option value="kinesiologo">Kinesiólogo</option>
              </select>
            </div>
            <div className="flex items-end">
              <button className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50" disabled={isAssigningRole}>
                {isAssigningRole ? "Enviando..." : "Enviar invitación"}
              </button>
            </div>
          </form>
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Staff</h2>
            <button type="button" className="px-3 py-2 rounded-lg border text-sm" onClick={refreshStaff}>
              Actualizar
            </button>
          </div>

          <div className="divide-y">
            {sortedStaffMembers.length === 0 ? (
              <p className="text-sm text-gray-600 py-3">No hay miembros del equipo cargados.</p>
            ) : (
              sortedStaffMembers.map((member) => (
                <div key={member.id} className="py-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <div className="font-medium">{member.last_name}, {member.first_name}</div>
                    <div className="text-sm text-gray-600">{member.email}</div>
                    {member.phone && <div className="text-sm text-gray-600">Tel: {member.phone}</div>}
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs px-2 py-1 rounded-md bg-blue-100 text-blue-700">
                      {member.role}
                    </span>
                    <span className={`text-xs px-2 py-1 rounded-md ${member.active ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-700"}`}>
                      {member.active ? "Activo" : "Inactivo"}
                    </span>
                    <button type="button" className="px-3 py-1 rounded-lg border text-sm" onClick={() => editStaff(member)}>
                      Editar
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </section>

        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div>
            <h2 className="text-lg font-semibold">{editingId ? "Editar perfil de kinesiólogo" : "Crear perfil de kinesiólogo"}</h2>
            <p className="text-sm text-gray-600">Solo los kinesiólogos necesitan perfil profesional para aparecer en agenda.</p>
          </div>

          <form className="grid grid-cols-1 md:grid-cols-2 gap-3" onSubmit={submitKinesiologist}>
            <Field label="Nombre" value={form.first_name} onChange={(value) => setForm({ ...form, first_name: value })} />
            <Field label="Apellido" value={form.last_name} onChange={(value) => setForm({ ...form, last_name: value })} />
            <Field label="Email" type="email" value={form.email} onChange={(value) => setForm({ ...form, email: value })} />
            <Field label="Matrícula" value={form.license_number ?? ""} onChange={(value) => setForm({ ...form, license_number: value })} required={false} />

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
                {isSaving ? "Guardando..." : editingId ? "Guardar cambios" : "Crear perfil"}
              </button>
              {editingId && (
                <button type="button" className="px-4 py-2 rounded-lg border" onClick={resetForm}>
                  Cancelar
                </button>
              )}
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
                      Invitar/resetear
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
      <label className="text-sm font-medium">{label}</label>
      <input
        className="mt-1 w-full border rounded-lg p-2"
        type={type}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        required={required}
      />
    </div>
  );
}
