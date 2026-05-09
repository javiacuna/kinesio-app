import { useEffect, useMemo, useState, type FormEvent } from "react";
import {
  createKinesiologist,
  listKinesiologists,
  updateKinesiologist,
  uploadKinesiologistAttachment,
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

type TeamForm = SaveStaffMemberInput & {
  license_number: string;
  work_start_time: string;
  work_end_time: string;
  work_days: number[];
};

type RoleFilter = "all" | StaffRole;

type TeamListItem = {
  key: string;
  staff?: StaffMember;
  profile?: Kinesiologist;
  first_name: string;
  last_name: string;
  email: string;
  role: StaffRole;
  phone?: string | null;
  active: boolean;
};

const defaultWorkDays = [1, 2, 3, 4, 5];

const workDayOptions = [
  { value: 1, label: "Lun" },
  { value: 2, label: "Mar" },
  { value: 3, label: "Mié" },
  { value: 4, label: "Jue" },
  { value: 5, label: "Vie" },
  { value: 6, label: "Sáb" },
  { value: 7, label: "Dom" },
];

const emptyTeamForm: TeamForm = {
  first_name: "",
  last_name: "",
  email: "",
  role: "recepcionista",
  phone: "",
  license_number: "",
  work_start_time: "08:00",
  work_end_time: "20:00",
  work_days: defaultWorkDays,
  active: true,
};

export default function StaffPage() {
  const [staffMembers, setStaffMembers] = useState<StaffMember[]>([]);
  const [kinesiologists, setKinesiologists] = useState<Kinesiologist[]>([]);
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [form, setForm] = useState<TeamForm>(emptyTeamForm);
  const [editingStaffId, setEditingStaffId] = useState<string | null>(null);
  const [editingStaffEmail, setEditingStaffEmail] = useState<string | null>(null);
  const [sendAccess, setSendAccess] = useState(true);
  const [professionalFiles, setProfessionalFiles] = useState<File[]>([]);
  const [roleFilter, setRoleFilter] = useState<RoleFilter>("all");
  const [roleEmail, setRoleEmail] = useState("");
  const [role, setRole] = useState<AuthRole>("recepcionista");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [isSaving, setIsSaving] = useState(false);
  const [isAssigningRole, setIsAssigningRole] = useState(false);

  const kinesiologistByEmail = useMemo(() => {
    const index = new Map<string, Kinesiologist>();
    for (const item of kinesiologists) {
      index.set(item.email.toLowerCase(), item);
    }
    return index;
  }, [kinesiologists]);

  const sortedUsers = useMemo(
    () => [...users].sort((a, b) => a.email.localeCompare(b.email)),
    [users],
  );

  const visibleTeamMembers = useMemo(() => {
    const byEmail = new Set(staffMembers.map((member) => member.email.toLowerCase()));
    const unified: TeamListItem[] = [
      ...staffMembers.map((member) => ({
        key: `staff:${member.id}`,
        staff: member,
        profile: kinesiologistByEmail.get(member.email.toLowerCase()),
        first_name: member.first_name,
        last_name: member.last_name,
        email: member.email,
        role: member.role,
        phone: member.phone,
        active: member.active,
      })),
      ...kinesiologists
        .filter((profile) => !byEmail.has(profile.email.toLowerCase()))
        .map((profile) => ({
          key: `kinesiologist:${profile.id}`,
          profile,
          first_name: profile.first_name,
          last_name: profile.last_name,
          email: profile.email,
          role: "kinesiologo" as StaffRole,
          phone: null,
          active: profile.active,
        })),
    ];

    const items =
      roleFilter === "all"
        ? unified
        : unified.filter((member) => member.role === roleFilter);

    return items.sort((a, b) =>
      `${a.last_name} ${a.first_name}`.localeCompare(`${b.last_name} ${b.first_name}`),
    );
  }, [kinesiologistByEmail, kinesiologists, roleFilter, staffMembers]);

  const isKinesiologistRole = form.role === "kinesiologo";

  useEffect(() => {
    refreshTeam();
    refreshUsers();
  }, []);

  async function refreshTeam() {
    setError("");
    try {
      const [staff, kinesioProfiles] = await Promise.all([
        listStaffMembers({ includeInactive: true }),
        listKinesiologists({ includeInactive: true }),
      ]);
      setStaffMembers(staff);
      setKinesiologists(kinesioProfiles);
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

  async function submitTeamMember(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setMessage("");
    setIsSaving(true);

    try {
      const payload: SaveStaffMemberInput = {
        first_name: form.first_name.trim(),
        last_name: form.last_name.trim(),
        email: form.email.trim().toLowerCase(),
        role: form.role,
        phone: form.phone?.trim() || null,
        active: form.active,
        firebase_uid: form.firebase_uid ?? null,
      };

      const saved = editingStaffId
        ? await updateStaffMember({ ...payload, id: editingStaffId })
        : await createStaffMember(payload);

      const professionalProfile = await syncKinesiologistProfile(saved);

      if (professionalProfile && professionalFiles.length > 0) {
        await Promise.all(
          professionalFiles.map((file) =>
            uploadKinesiologistAttachment(professionalProfile.id, file, {
              category: professionalFileCategory(file),
            }),
          ),
        );
        setProfessionalFiles([]);
      }

      if (sendAccess) {
        await inviteUserAccess({ email: saved.email, role: saved.role });
        await refreshUsers();
      }

      await refreshTeam();
      setMessage(teamMessage(saved));
      resetForm();
    } catch (err) {
      setError((err as Error)?.message ?? "No se pudo guardar el miembro del equipo.");
    } finally {
      setIsSaving(false);
    }
  }

  async function syncKinesiologistProfile(saved: StaffMember): Promise<Kinesiologist | undefined> {
    const currentEmail = saved.email.toLowerCase();
    const previousEmail = editingStaffEmail?.toLowerCase();
    const existing =
      kinesiologistByEmail.get(currentEmail) ??
      (previousEmail ? kinesiologistByEmail.get(previousEmail) : undefined);

    if (saved.role !== "kinesiologo") {
      if (existing?.active) {
        await updateKinesiologist({
          ...toKinesiologistPayload(saved, existing, false),
          id: existing.id,
        });
      }
      return undefined;
    }

    const payload = toKinesiologistPayload(saved, existing, saved.active);
    if (existing) {
      return updateKinesiologist({ ...payload, id: existing.id });
    }
    return createKinesiologist(payload);
  }

  function toKinesiologistPayload(
    member: StaffMember,
    existing: Kinesiologist | undefined,
    active: boolean,
  ): SaveKinesiologistInput {
    return {
      first_name: member.first_name,
      last_name: member.last_name,
      email: member.email,
      license_number: form.license_number.trim() || existing?.license_number || null,
      work_start_time: form.work_start_time || existing?.work_start_time || "08:00",
      work_end_time: form.work_end_time || existing?.work_end_time || "20:00",
      work_days: form.work_days.length > 0 ? form.work_days : existing?.work_days ?? defaultWorkDays,
      active,
    };
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

  function edit(member: TeamListItem) {
    const profile = member.profile;
    setEditingStaffId(member.staff?.id ?? null);
    setEditingStaffEmail(member.email);
    setForm({
      first_name: member.first_name,
      last_name: member.last_name,
      email: member.email,
      role: member.role,
      phone: member.phone ?? "",
      active: member.active,
      firebase_uid: member.staff?.firebase_uid ?? null,
      license_number: profile?.license_number ?? "",
      work_start_time: profile?.work_start_time ?? "08:00",
      work_end_time: profile?.work_end_time ?? "20:00",
      work_days: profile?.work_days?.length ? profile.work_days : defaultWorkDays,
    });
    setSendAccess(false);
    setProfessionalFiles([]);
    setMessage("");
    setError("");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function resetForm() {
    setEditingStaffId(null);
    setEditingStaffEmail(null);
    setForm(emptyTeamForm);
    setProfessionalFiles([]);
    setSendAccess(true);
  }

  function teamMessage(saved: StaffMember) {
    const access = sendAccess ? " e invitación enviada" : "";
    return `${saved.first_name} ${saved.last_name} guardado${access}.`;
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
            <p className="text-sm text-gray-600">Los datos profesionales aparecen cuando el rol es kinesiólogo.</p>
          </div>

          <form className="grid grid-cols-1 md:grid-cols-2 gap-3" onSubmit={submitTeamMember}>
            <Field label="Nombre" value={form.first_name} onChange={(value) => setForm({ ...form, first_name: value })} />
            <Field label="Apellido" value={form.last_name} onChange={(value) => setForm({ ...form, last_name: value })} />
            <Field label="Email" type="email" value={form.email} onChange={(value) => setForm({ ...form, email: value })} />
            <Field label="Teléfono" value={form.phone ?? ""} onChange={(value) => setForm({ ...form, phone: value })} required={false} />

            <div>
              <label className="text-sm font-medium">Rol</label>
              <select
                className="mt-1 w-full border rounded-lg p-2"
                value={form.role}
                onChange={(event) => setForm({ ...form, role: event.target.value as StaffRole })}
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
                  checked={form.active}
                  onChange={(event) => setForm({ ...form, active: event.target.checked })}
                />
                Activo
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={sendAccess}
                  onChange={(event) => setSendAccess(event.target.checked)}
                />
                Enviar acceso por email
              </label>
            </div>

            {isKinesiologistRole && (
              <>
                <Field label="Matrícula" value={form.license_number} onChange={(value) => setForm({ ...form, license_number: value })} required={false} />
                <Field label="Horario inicio" type="time" value={form.work_start_time} onChange={(value) => setForm({ ...form, work_start_time: value })} />
                <Field label="Horario fin" type="time" value={form.work_end_time} onChange={(value) => setForm({ ...form, work_end_time: value })} />
                <div className="md:col-span-2">
                  <label className="text-sm font-medium">Días laborales</label>
                  <div className="mt-2 flex flex-wrap gap-2">
                    {workDayOptions.map((day) => {
                      const checked = form.work_days.includes(day.value);
                      return (
                        <label
                          key={day.value}
                          className={`px-3 py-2 rounded-lg border text-sm cursor-pointer ${checked ? "bg-black text-white" : "bg-white hover:bg-gray-50"}`}
                        >
                          <input
                            className="sr-only"
                            type="checkbox"
                            checked={checked}
                            onChange={(event) => {
                              const next = event.target.checked
                                ? [...form.work_days, day.value]
                                : form.work_days.filter((value) => value !== day.value);
                              setForm({ ...form, work_days: next.sort((a, b) => a - b) });
                            }}
                          />
                          {day.label}
                        </label>
                      );
                    })}
                  </div>
                </div>
                <div className="md:col-span-2">
                  <label className="text-sm font-medium">Documentación profesional</label>
                  <input
                    className="mt-1 w-full border rounded-lg p-2"
                    type="file"
                    multiple
                    accept="image/*,video/*,application/pdf"
                    onChange={(event) => setProfessionalFiles(Array.from(event.target.files ?? []))}
                  />
                  {professionalFiles.length > 0 && (
                    <p className="text-xs text-gray-500 mt-1">
                      {professionalFiles.length} archivo(s) se van a subir al guardar el kinesiólogo.
                    </p>
                  )}
                </div>
              </>
            )}

            <div className="flex gap-2 md:col-span-2">
              <button className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50" disabled={isSaving}>
                {isSaving ? "Guardando..." : editingStaffId ? "Guardar cambios" : "Crear miembro"}
              </button>
              {editingStaffId && (
                <button type="button" className="px-4 py-2 rounded-lg border" onClick={resetForm}>
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
          <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <h2 className="text-lg font-semibold">Staff</h2>
              <p className="text-sm text-gray-600">Listado unificado del equipo.</p>
            </div>
            <div className="flex items-end gap-2">
              <div>
                <label className="text-sm font-medium">Filtrar por rol</label>
                <select
                  className="mt-1 w-full min-w-48 border rounded-lg p-2"
                  value={roleFilter}
                  onChange={(event) => setRoleFilter(event.target.value as RoleFilter)}
                >
                  <option value="all">Todos</option>
                  <option value="recepcionista">Recepcionista</option>
                  <option value="admin">Admin</option>
                  <option value="kinesiologo">Kinesiólogo</option>
                </select>
              </div>
              <button type="button" className="px-3 py-2 rounded-lg border text-sm" onClick={refreshTeam}>
                Actualizar
              </button>
            </div>
          </div>

          <div className="divide-y">
            {visibleTeamMembers.length === 0 ? (
              <p className="text-sm text-gray-600 py-3">No hay miembros del equipo para mostrar.</p>
            ) : (
              visibleTeamMembers.map((member) => {
                const profile = member.profile;

                return (
                  <div key={member.key} className="py-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <div>
                      <div className="font-medium">{member.last_name}, {member.first_name}</div>
                      <div className="text-sm text-gray-600">{member.email}</div>
                      {member.phone && <div className="text-sm text-gray-600">Tel: {member.phone}</div>}
                      {member.role === "kinesiologo" && profile && (
                        <div className="text-sm text-gray-600">
                          {profile.license_number ? `Matrícula: ${profile.license_number} · ` : ""}
                          Horario: {profile.work_start_time} a {profile.work_end_time} · Días: {workDaysLabel(profile.work_days)}
                        </div>
                      )}
                      {member.role === "kinesiologo" && !profile && (
                        <div className="text-sm text-red-600">Perfil profesional pendiente.</div>
                      )}
                      {!member.staff && (
                        <div className="text-sm text-amber-700">Solo perfil profesional, falta crear miembro de staff.</div>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-xs px-2 py-1 rounded-md bg-blue-100 text-blue-700">
                        {member.role}
                      </span>
                      <span className={`text-xs px-2 py-1 rounded-md ${member.active ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-700"}`}>
                        {member.active ? "Activo" : "Inactivo"}
                      </span>
                      <button type="button" className="px-3 py-1 rounded-lg border text-sm" onClick={() => edit(member)}>
                        Editar
                      </button>
                    </div>
                  </div>
                );
              })
            )}
          </div>
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

function workDaysLabel(days: number[] | undefined) {
  const selected = days?.length ? days : defaultWorkDays;
  return selected
    .map((value) => workDayOptions.find((day) => day.value === value)?.label)
    .filter(Boolean)
    .join(", ");
}

function professionalFileCategory(file: File) {
  if (file.type.startsWith("image/")) return "foto";
  if (file.type.startsWith("video/")) return "video";
  return "documento";
}
