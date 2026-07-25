import { useEffect, useMemo, useState, type FormEvent } from "react";
import {
  createPractice,
  createKinesiologist,
  createSpecialty,
  listPractices,
  listKinesiologists,
  listSpecialties,
  updatePractice,
  updateKinesiologist,
  updateSpecialty,
  uploadKinesiologistAttachment,
  type SavePracticeInput,
  type SaveKinesiologistInput,
} from "@/features/kinesiologists/api";
import type { Kinesiologist, Practice, Specialty } from "@/features/kinesiologists/types";
import { inviteUserAccess, listAdminUsers, type AdminUser } from "@/features/auth/adminApi";
import type { AuthRole } from "@/features/auth/types";
import {
  createStaffMember,
  listStaffMembers,
  updateStaffMember,
  type SaveStaffMemberInput,
} from "@/features/staff/api";
import type { StaffMember, StaffRole } from "@/features/staff/types";
import { RequiredLabel } from "@/shared/ui/RequiredLabel";
import { Tabs } from "@/shared/ui/Tabs";
import { useLanguage } from "@/shared/i18n/LanguageProvider";

type Translate = (key: string) => string;

function getStaffTabs(t: Translate) {
  return [
    { key: "equipo", label: t("staff.tabTeam") },
    { key: "especialidades", label: t("staff.tabSpecialties") },
    { key: "accesos", label: t("staff.tabAccess") },
  ];
}

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

type TeamForm = SaveStaffMemberInput & {
  license_number: string;
  work_start_time: string;
  work_end_time: string;
  work_days: number[];
  practice_ids: string[];
};

type PracticeForm = {
  specialty_id: string;
  name: string;
  description: string;
  active: boolean;
};

type SpecialtyForm = {
  name: string;
  active: boolean;
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

function getWorkDayOptions(t: Translate) {
  return [
    { value: 1, label: t("day.mon") },
    { value: 2, label: t("day.tue") },
    { value: 3, label: t("day.wed") },
    { value: 4, label: t("day.thu") },
    { value: 5, label: t("day.fri") },
    { value: 6, label: t("day.sat") },
    { value: 7, label: t("day.sun") },
  ];
}

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
  practice_ids: [],
  active: true,
};

const emptyPracticeForm: PracticeForm = {
  specialty_id: "",
  name: "",
  description: "",
  active: true,
};

function getEmptySpecialtyForm(t: Translate): SpecialtyForm {
  return {
    name: t("staff.defaultSpecialtyName"),
    active: true,
  };
}

export default function StaffPage() {
  const { t } = useLanguage();
  const staffTabs = getStaffTabs(t);
  const workDayOptions = getWorkDayOptions(t);
  const [staffMembers, setStaffMembers] = useState<StaffMember[]>([]);
  const [kinesiologists, setKinesiologists] = useState<Kinesiologist[]>([]);
  const [specialties, setSpecialties] = useState<Specialty[]>([]);
  const [practices, setPractices] = useState<Practice[]>([]);
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [form, setForm] = useState<TeamForm>(emptyTeamForm);
  const [specialtyForm, setSpecialtyForm] = useState<SpecialtyForm>(() => getEmptySpecialtyForm(t));
  const [practiceForm, setPracticeForm] = useState<PracticeForm>(emptyPracticeForm);
  const [editingStaffId, setEditingStaffId] = useState<string | null>(null);
  const [editingStaffEmail, setEditingStaffEmail] = useState<string | null>(null);
  const [editingSpecialtyId, setEditingSpecialtyId] = useState<string | null>(null);
  const [editingPracticeId, setEditingPracticeId] = useState<string | null>(null);
  const [sendAccess, setSendAccess] = useState(true);
  const [professionalFiles, setProfessionalFiles] = useState<File[]>([]);
  const [roleFilter, setRoleFilter] = useState<RoleFilter>("all");
  const [roleEmail, setRoleEmail] = useState("");
  const [role, setRole] = useState<AuthRole>("recepcionista");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [teamFormErrors, setTeamFormErrors] = useState<Record<string, string>>({});
  const [specialtyFormErrors, setSpecialtyFormErrors] = useState<Record<string, string>>({});
  const [practiceFormErrors, setPracticeFormErrors] = useState<Record<string, string>>({});
  const [roleFormErrors, setRoleFormErrors] = useState<Record<string, string>>({});
  const [isSaving, setIsSaving] = useState(false);
  const [isSavingSpecialty, setIsSavingSpecialty] = useState(false);
  const [isSavingPractice, setIsSavingPractice] = useState(false);
  const [isAssigningRole, setIsAssigningRole] = useState(false);
  const [activeTab, setActiveTab] = useState("equipo");

  const kinesiologistByEmail = useMemo(() => {
    const index = new Map<string, Kinesiologist>();
    for (const item of kinesiologists) {
      index.set(item.email.toLowerCase(), item);
    }
    return index;
  }, [kinesiologists]);

  const primarySpecialty = useMemo(
    () => specialties.find((item) => item.active) ?? specialties[0],
    [specialties],
  );

  const activePractices = useMemo(
    () => practices.filter((practice) => practice.active),
    [practices],
  );

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
      const [staff, kinesioProfiles, specialtyItems, practiceItems] = await Promise.all([
        listStaffMembers({ includeInactive: true }),
        listKinesiologists({ includeInactive: true }),
        listSpecialties({ includeInactive: true }),
        listPractices({ includeInactive: true }),
      ]);
      setStaffMembers(staff);
      setKinesiologists(kinesioProfiles);
      setSpecialties(specialtyItems);
      setPractices(practiceItems);
      if (!practiceForm.specialty_id && specialtyItems.length > 0) {
        setPracticeForm((current) => ({
          ...current,
          specialty_id: current.specialty_id || specialtyItems.find((item) => item.active)?.id || specialtyItems[0].id,
        }));
      }
    } catch (err) {
      setError((err as Error)?.message ?? t("staff.errorLoadTeam"));
    }
  }

  async function refreshUsers() {
    setError("");
    try {
      setUsers(await listAdminUsers());
    } catch (err) {
      setError((err as Error)?.message ?? t("staff.errorLoadUsers"));
    }
  }

  async function submitTeamMember(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setMessage("");
    const validation = validateTeamForm();
    setTeamFormErrors(validation);
    if (Object.keys(validation).length > 0) return;
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
      setError((err as Error)?.message ?? t("staff.errorSaveMember"));
    } finally {
      setIsSaving(false);
    }
  }

  function validateTeamForm() {
    const validation: Record<string, string> = {};
    if (!form.first_name.trim()) validation.first_name = t("staff.errorFirstName");
    if (!form.last_name.trim()) validation.last_name = t("staff.errorLastName");
    if (!form.email.trim()) {
      validation.email = t("staff.errorEmailRequired");
    } else if (!emailPattern.test(form.email.trim())) {
      validation.email = t("staff.errorEmailInvalid");
    }
    if (form.role === "kinesiologo") {
      if (!form.work_start_time) validation.work_start_time = t("staff.errorWorkStart");
      if (!form.work_end_time) validation.work_end_time = t("staff.errorWorkEnd");
      if (form.work_start_time && form.work_end_time && form.work_start_time >= form.work_end_time) {
        validation.work_end_time = t("staff.errorWorkEndAfterStart");
      }
      if (form.work_days.length === 0) validation.work_days = t("staff.errorWorkDays");
    }
    return validation;
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
      practice_ids: form.practice_ids,
      active,
    };
  }

  async function submitSpecialty(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setMessage("");
    const validation: Record<string, string> = {};
    if (!specialtyForm.name.trim()) validation.name = t("staff.errorSpecialtyName");
    setSpecialtyFormErrors(validation);
    if (Object.keys(validation).length > 0) return;
    setIsSavingSpecialty(true);

    try {
      const payload = {
        name: specialtyForm.name.trim(),
        active: specialtyForm.active,
      };
      if (editingSpecialtyId) {
        await updateSpecialty({ ...payload, id: editingSpecialtyId });
      } else {
        await createSpecialty(payload);
      }
      setMessage(editingSpecialtyId ? t("staff.specialtyUpdated") : t("staff.specialtyCreated"));
      resetSpecialtyForm();
      await refreshTeam();
    } catch (err) {
      const apiError = err as Error;
      setError(apiError.message === "nombre_duplicado" ? t("staff.errorDuplicateSpecialty") : apiError.message);
    } finally {
      setIsSavingSpecialty(false);
    }
  }

  async function toggleSpecialty(specialty: Specialty) {
    setError("");
    setMessage("");
    try {
      await updateSpecialty({
        id: specialty.id,
        name: specialty.name,
        active: !specialty.active,
      });
      setMessage(!specialty.active ? t("staff.specialtyActivated") : t("staff.specialtyDeactivated"));
      await refreshTeam();
    } catch (err) {
      setError((err as Error)?.message ?? t("staff.errorUpdateSpecialty"));
    }
  }

  function editSpecialty(specialty: Specialty) {
    setEditingSpecialtyId(specialty.id);
    setSpecialtyForm({
      name: specialty.name,
      active: specialty.active,
    });
    setSpecialtyFormErrors({});
    setMessage("");
    setError("");
  }

  function resetSpecialtyForm() {
    setEditingSpecialtyId(null);
    setSpecialtyForm(getEmptySpecialtyForm(t));
    setSpecialtyFormErrors({});
  }

  async function submitPractice(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setMessage("");
    const validation: Record<string, string> = {};
    const specialtyId = practiceForm.specialty_id || primarySpecialty?.id || "";
    if (!specialtyId) validation.specialty_id = t("staff.errorSpecialtyFirst");
    if (!practiceForm.name.trim()) validation.name = t("staff.errorPracticeName");
    setPracticeFormErrors(validation);
    if (Object.keys(validation).length > 0) return;
    setIsSavingPractice(true);

    try {
      const payload: SavePracticeInput = {
        specialty_id: specialtyId,
        name: practiceForm.name.trim(),
        description: practiceForm.description.trim() || null,
        active: practiceForm.active,
      };
      if (editingPracticeId) {
        await updatePractice({ ...payload, id: editingPracticeId });
      } else {
        await createPractice(payload);
      }
      setMessage(editingPracticeId ? t("staff.practiceUpdated") : t("staff.practiceCreated"));
      resetPracticeForm();
      await refreshTeam();
    } catch (err) {
      const apiError = err as Error;
      setError(apiError.message === "nombre_duplicado" ? t("staff.errorDuplicatePractice") : apiError.message);
    } finally {
      setIsSavingPractice(false);
    }
  }

  async function togglePractice(practice: Practice) {
    setError("");
    setMessage("");
    try {
      await updatePractice({
        id: practice.id,
        specialty_id: practice.specialty_id,
        name: practice.name,
        description: practice.description ?? null,
        active: !practice.active,
      });
      setMessage(!practice.active ? t("staff.practiceActivated") : t("staff.practiceDeactivated"));
      await refreshTeam();
    } catch (err) {
      setError((err as Error)?.message ?? t("staff.errorUpdatePractice"));
    }
  }

  function editPractice(practice: Practice) {
    setEditingPracticeId(practice.id);
    setPracticeForm({
      specialty_id: practice.specialty_id,
      name: practice.name,
      description: practice.description ?? "",
      active: practice.active,
    });
    setPracticeFormErrors({});
    setMessage("");
    setError("");
  }

  function resetPracticeForm() {
    setEditingPracticeId(null);
    setPracticeForm({
      ...emptyPracticeForm,
      specialty_id: primarySpecialty?.id ?? specialties[0]?.id ?? "",
    });
    setPracticeFormErrors({});
  }

  async function submitRole(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setMessage("");
    const validation: Record<string, string> = {};
    if (!roleEmail.trim()) {
      validation.roleEmail = t("staff.errorEmailRequired");
    } else if (!emailPattern.test(roleEmail.trim())) {
      validation.roleEmail = t("staff.errorEmailInvalid");
    }
    setRoleFormErrors(validation);
    if (Object.keys(validation).length > 0) return;
    setIsAssigningRole(true);

    try {
      const res = await inviteUserAccess({ email: roleEmail.trim().toLowerCase(), role });
      setMessage(
        res.created
          ? `${t("staff.userCreatedInvitedPrefix")} ${res.email}.`
          : `${t("staff.invitationSentPrefix")} ${res.email} ${t("staff.roleConfirmedMiddle")} ${res.role} ${t("staff.roleConfirmedSuffix")}`,
      );
      setRoleEmail("");
      await refreshUsers();
    } catch (err) {
      setError((err as Error)?.message ?? t("staff.errorAssignRole"));
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
      practice_ids: profile?.practice_ids?.length
        ? profile.practice_ids
        : profile?.practices?.map((practice) => practice.id) ?? [],
    });
    setSendAccess(false);
    setProfessionalFiles([]);
    setMessage("");
    setError("");
    setActiveTab("equipo");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function resetForm() {
    setEditingStaffId(null);
    setEditingStaffEmail(null);
    setForm({ ...emptyTeamForm });
    setProfessionalFiles([]);
    setSendAccess(true);
  }

  function teamMessage(saved: StaffMember) {
    const access = sendAccess ? ` ${t("staff.andInvitationSent")}` : "";
    return `${saved.first_name} ${saved.last_name} ${t("staff.savedSuffix")}${access}.`;
  }

  return (
    <main>
      <div className="max-w-5xl mx-auto p-6 space-y-6">
        <header>
          <h1 className="text-2xl font-semibold">{t("staff.title")}</h1>
          <p className="text-sm text-gray-600">{t("staff.subtitle")}</p>
        </header>

        {(message || error) && (
          <div className={`rounded-lg border p-3 text-sm ${error ? "border-red-200 bg-red-50 text-red-700" : "border-green-200 bg-green-50 text-green-700"}`}>
            {error || message}
          </div>
        )}

        <Tabs tabs={staffTabs} active={activeTab} onChange={setActiveTab} />

        {activeTab === "equipo" && (
        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div>
            <h2 className="text-lg font-semibold">{editingStaffId ? t("staff.editMember") : t("staff.createMember")}</h2>
            <p className="text-sm text-gray-600">{t("staff.professionalDataNote")}</p>
          </div>

          <form className="grid grid-cols-1 md:grid-cols-2 gap-3" onSubmit={submitTeamMember} noValidate>
            <p className="text-xs text-gray-500 md:col-span-2">{t("staff.requiredFieldsNote")}</p>
            <Field label={t("staff.firstName")} value={form.first_name} onChange={(value) => setForm({ ...form, first_name: value })} error={teamFormErrors.first_name} />
            <Field label={t("staff.lastName")} value={form.last_name} onChange={(value) => setForm({ ...form, last_name: value })} error={teamFormErrors.last_name} />
            <Field label={t("staff.email")} type="email" value={form.email} onChange={(value) => setForm({ ...form, email: value })} error={teamFormErrors.email} />
            <Field label={t("staff.phone")} value={form.phone ?? ""} onChange={(value) => setForm({ ...form, phone: value })} required={false} />

            <div>
              <label className="text-sm font-medium">{t("staff.role")}</label>
              <select
                className="mt-1 w-full border rounded-lg p-2"
                value={form.role}
                onChange={(event) => setForm({ ...form, role: event.target.value as StaffRole })}
              >
                <option value="recepcionista">{t("staff.roleReceptionist")}</option>
                <option value="admin">{t("staff.roleAdmin")}</option>
                <option value="kinesiologo">{t("staff.roleKinesiologist")}</option>
              </select>
            </div>

            <div className="flex flex-col justify-end gap-2">
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={form.active}
                  onChange={(event) => setForm({ ...form, active: event.target.checked })}
                />
                {t("staff.active")}
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={sendAccess}
                  onChange={(event) => setSendAccess(event.target.checked)}
                />
                {t("staff.sendAccessEmail")}
              </label>
            </div>

            {isKinesiologistRole && (
              <>
                <Field label={t("staff.licenseNumber")} value={form.license_number} onChange={(value) => setForm({ ...form, license_number: value })} required={false} />
                <Field label={t("staff.workStart")} type="time" value={form.work_start_time} onChange={(value) => setForm({ ...form, work_start_time: value })} error={teamFormErrors.work_start_time} />
                <Field label={t("staff.workEnd")} type="time" value={form.work_end_time} onChange={(value) => setForm({ ...form, work_end_time: value })} error={teamFormErrors.work_end_time} />
                <div className="md:col-span-2">
                  <label className="text-sm font-medium"><RequiredLabel required>{t("staff.workDays")}</RequiredLabel></label>
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
                  {teamFormErrors.work_days && <p className="text-xs text-red-600 mt-1">{teamFormErrors.work_days}</p>}
                </div>
                <div className="md:col-span-2">
                  <label className="text-sm font-medium">{t("staff.practices")}</label>
                  <p className="text-xs text-gray-500 mt-1">{t("staff.specialtyLabel")} {primarySpecialty?.name ?? t("staff.defaultSpecialtyName")}</p>
                  {activePractices.length === 0 ? (
                    <p className="text-sm text-amber-700 mt-2">{t("staff.noActivePractices")}</p>
                  ) : (
                    <div className="mt-2 flex flex-wrap gap-2">
                      {activePractices.map((practice) => {
                        const checked = form.practice_ids.includes(practice.id);
                        return (
                          <label
                            key={practice.id}
                            className={`px-3 py-2 rounded-lg border text-sm cursor-pointer ${checked ? "bg-black text-white" : "bg-white hover:bg-gray-50"}`}
                          >
                            <input
                              className="sr-only"
                              type="checkbox"
                              checked={checked}
                              onChange={(event) => {
                                const next = event.target.checked
                                  ? [...form.practice_ids, practice.id]
                                  : form.practice_ids.filter((value) => value !== practice.id);
                                setForm({ ...form, practice_ids: next });
                              }}
                            />
                            {practice.name}
                          </label>
                        );
                      })}
                    </div>
                  )}
                </div>
                <div className="md:col-span-2">
                  <label className="text-sm font-medium">{t("staff.professionalDocs")}</label>
                  <input
                    className="mt-1 w-full border rounded-lg p-2"
                    type="file"
                    multiple
                    accept="image/*,video/*,application/pdf"
                    onChange={(event) => setProfessionalFiles(Array.from(event.target.files ?? []))}
                  />
                  {professionalFiles.length > 0 && (
                    <p className="text-xs text-gray-500 mt-1">
                      {professionalFiles.length} {t("staff.filesWillUploadOnSave")}
                    </p>
                  )}
                </div>
              </>
            )}

            <div className="flex gap-2 md:col-span-2">
              <button className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50" disabled={isSaving}>
                {isSaving ? t("common.saving") : editingStaffId ? t("common.saveChanges") : t("staff.createMemberBtn")}
              </button>
              {editingStaffId && (
                <button type="button" className="px-4 py-2 rounded-lg border" onClick={resetForm}>
                  {t("common.cancel")}
                </button>
              )}
            </div>
          </form>
        </section>
        )}

        {activeTab === "especialidades" && (
        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div>
            <h2 className="text-lg font-semibold">{t("staff.specialtiesTitle")}</h2>
            <p className="text-sm text-gray-600">{t("staff.specialtiesSubtitle")}</p>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <div className="border rounded-lg p-3 space-y-3">
              <div>
                <h3 className="font-medium">{t("staff.specialtiesBoxTitle")}</h3>
                <p className="text-xs text-gray-500">{t("staff.specialtiesBoxNote")}</p>
              </div>
              <form className="space-y-3" onSubmit={submitSpecialty} noValidate>
                <Field
                  label={t("staff.specialtyNameLabel")}
                  value={specialtyForm.name}
                  onChange={(value) => setSpecialtyForm({ ...specialtyForm, name: value })}
                  error={specialtyFormErrors.name}
                />
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={specialtyForm.active}
                    onChange={(event) => setSpecialtyForm({ ...specialtyForm, active: event.target.checked })}
                  />
                  {t("staff.activeFem")}
                </label>
                <div className="flex gap-2">
                  <button className="px-3 py-2 rounded-lg bg-black text-white text-sm disabled:opacity-50" disabled={isSavingSpecialty}>
                    {isSavingSpecialty ? t("common.saving") : editingSpecialtyId ? t("staff.saveSpecialty") : t("staff.createSpecialty")}
                  </button>
                  {editingSpecialtyId && (
                    <button type="button" className="px-3 py-2 rounded-lg border text-sm" onClick={resetSpecialtyForm}>
                      {t("common.cancel")}
                    </button>
                  )}
                </div>
              </form>

              <div className="divide-y">
                {specialties.length === 0 ? (
                  <p className="text-sm text-gray-600 py-2">{t("staff.noSpecialties")}</p>
                ) : (
                  specialties.map((specialty) => (
                    <div key={specialty.id} className="py-2 flex items-center justify-between gap-2">
                      <div>
                        <div className="font-medium">{specialty.name}</div>
                        <div className="text-xs text-gray-500">{specialty.active ? t("staff.activeFem") : t("staff.inactiveFem")}</div>
                      </div>
                      <div className="flex gap-2">
                        <button type="button" className="px-3 py-1 rounded-lg border text-sm" onClick={() => editSpecialty(specialty)}>
                          {t("staff.edit")}
                        </button>
                        <button type="button" className="px-3 py-1 rounded-lg border text-sm" onClick={() => toggleSpecialty(specialty)}>
                          {specialty.active ? t("staff.deactivate") : t("staff.activateBtn")}
                        </button>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>

            <div className="border rounded-lg p-3 space-y-3">
              <div>
                <h3 className="font-medium">{t("staff.practicesBoxTitle")}</h3>
                <p className="text-xs text-gray-500">{t("staff.practicesBoxNote")}</p>
              </div>
              <form className="space-y-3" onSubmit={submitPractice} noValidate>
                <div>
                  <label className="text-sm font-medium"><RequiredLabel required>{t("staff.specialty")}</RequiredLabel></label>
                  <select
                    className="mt-1 w-full border rounded-lg p-2"
                    value={practiceForm.specialty_id || primarySpecialty?.id || ""}
                    onChange={(event) => setPracticeForm({ ...practiceForm, specialty_id: event.target.value })}
                  >
                    <option value="">{t("staff.select")}</option>
                    {specialties.map((specialty) => (
                      <option key={specialty.id} value={specialty.id}>
                        {specialty.name}{specialty.active ? "" : ` ${t("staff.inactiveSuffix")}`}
                      </option>
                    ))}
                  </select>
                  {practiceFormErrors.specialty_id && <p className="text-xs text-red-600 mt-1">{practiceFormErrors.specialty_id}</p>}
                </div>
                <Field
                  label={t("staff.practiceNameLabel")}
                  value={practiceForm.name}
                  onChange={(value) => setPracticeForm({ ...practiceForm, name: value })}
                  error={practiceFormErrors.name}
                />
                <div>
                  <label className="text-sm font-medium">{t("staff.description")}</label>
                  <textarea
                    className="mt-1 w-full border rounded-lg p-2"
                    rows={3}
                    value={practiceForm.description}
                    onChange={(event) => setPracticeForm({ ...practiceForm, description: event.target.value })}
                  />
                </div>
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={practiceForm.active}
                    onChange={(event) => setPracticeForm({ ...practiceForm, active: event.target.checked })}
                  />
                  {t("staff.activeFem")}
                </label>
                <div className="flex gap-2">
                  <button className="px-3 py-2 rounded-lg bg-black text-white text-sm disabled:opacity-50" disabled={isSavingPractice}>
                    {isSavingPractice ? t("common.saving") : editingPracticeId ? t("staff.savePractice") : t("staff.createPractice")}
                  </button>
                  {editingPracticeId && (
                    <button type="button" className="px-3 py-2 rounded-lg border text-sm" onClick={resetPracticeForm}>
                      {t("common.cancel")}
                    </button>
                  )}
                </div>
              </form>

              <div className="divide-y">
                {practices.length === 0 ? (
                  <p className="text-sm text-gray-600 py-2">{t("staff.noPractices")}</p>
                ) : (
                  practices.map((practice) => (
                    <div key={practice.id} className="py-2 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                      <div>
                        <div className="font-medium">{practice.name}</div>
                        <div className="text-xs text-gray-500">
                          {specialties.find((item) => item.id === practice.specialty_id)?.name ?? t("staff.defaultSpecialtyName")} · {practice.active ? t("staff.activeFem") : t("staff.inactiveFem")}
                        </div>
                        {practice.description && <div className="text-sm text-gray-600">{practice.description}</div>}
                      </div>
                      <div className="flex gap-2">
                        <button type="button" className="px-3 py-1 rounded-lg border text-sm" onClick={() => editPractice(practice)}>
                          {t("staff.edit")}
                        </button>
                        <button type="button" className="px-3 py-1 rounded-lg border text-sm" onClick={() => togglePractice(practice)}>
                          {practice.active ? t("staff.deactivate") : t("staff.activateBtn")}
                        </button>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        </section>
        )}

        {activeTab === "accesos" && (
        <section className="bg-white rounded-xl shadow p-4 space-y-4">
          <div>
            <h2 className="text-lg font-semibold">{t("staff.resendAccessTitle")}</h2>
            <p className="text-sm text-gray-600">
              {t("staff.resendAccessNote")}
            </p>
          </div>

          <form className="grid grid-cols-1 md:grid-cols-[1fr_220px_auto] gap-3" onSubmit={submitRole} noValidate>
            <Field label={t("staff.email")} type="email" value={roleEmail} onChange={setRoleEmail} error={roleFormErrors.roleEmail} />
            <div>
              <label className="text-sm font-medium">{t("staff.role")}</label>
              <select
                className="mt-1 w-full border rounded-lg p-2"
                value={role}
                onChange={(event) => setRole(event.target.value as AuthRole)}
              >
                <option value="recepcionista">{t("staff.roleReceptionist")}</option>
                <option value="admin">{t("staff.roleAdmin")}</option>
                <option value="kinesiologo">{t("staff.roleKinesiologist")}</option>
              </select>
            </div>
            <div className="flex items-end">
              <button className="px-4 py-2 rounded-lg bg-black text-white disabled:opacity-50" disabled={isAssigningRole}>
                {isAssigningRole ? t("staff.sending") : t("staff.sendInvitation")}
              </button>
            </div>
          </form>
        </section>
        )}

        {activeTab === "equipo" && (
        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <h2 className="text-lg font-semibold">{t("staff.staffTitle")}</h2>
              <p className="text-sm text-gray-600">{t("staff.staffSubtitle")}</p>
            </div>
            <div className="flex items-end gap-2">
              <div>
                <label className="text-sm font-medium">{t("staff.filterByRole")}</label>
                <select
                  className="mt-1 w-full min-w-48 border rounded-lg p-2"
                  value={roleFilter}
                  onChange={(event) => setRoleFilter(event.target.value as RoleFilter)}
                >
                  <option value="all">{t("staff.all")}</option>
                  <option value="recepcionista">{t("staff.roleReceptionist")}</option>
                  <option value="admin">{t("staff.roleAdmin")}</option>
                  <option value="kinesiologo">{t("staff.roleKinesiologist")}</option>
                </select>
              </div>
              <button type="button" className="px-3 py-2 rounded-lg border text-sm" onClick={refreshTeam}>
                {t("staff.refresh")}
              </button>
            </div>
          </div>

          <div className="divide-y">
            {visibleTeamMembers.length === 0 ? (
              <p className="text-sm text-gray-600 py-3">{t("staff.noMembers")}</p>
            ) : (
              visibleTeamMembers.map((member) => {
                const profile = member.profile;

                return (
                  <div key={member.key} className="py-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <div>
                      <div className="font-medium">{member.last_name}, {member.first_name}</div>
                      <div className="text-sm text-gray-600">{member.email}</div>
                      {member.phone && <div className="text-sm text-gray-600">{t("staff.tel")}: {member.phone}</div>}
                      {member.role === "kinesiologo" && profile && (
                        <>
                          <div className="text-sm text-gray-600">
                            {profile.license_number ? `${t("staff.licenseNumber")}: ${profile.license_number} · ` : ""}
                            {t("staff.scheduleLabel")}: {profile.work_start_time} {t("portal.errorOutsideScheduleMiddle")} {profile.work_end_time} · {t("staff.daysLabel")}: {workDaysLabel(profile.work_days, t)}
                          </div>
                          <div className="text-sm text-gray-600">
                            {t("staff.practicesLabel")}: {profile.practices?.length ? profile.practices.map((practice) => practice.name).join(", ") : t("staff.unassigned")}
                          </div>
                        </>
                      )}
                      {member.role === "kinesiologo" && !profile && (
                        <div className="text-sm text-red-600">{t("staff.pendingProfile")}</div>
                      )}
                      {!member.staff && (
                        <div className="text-sm text-amber-700">{t("staff.onlyProfessionalProfile")}</div>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-xs px-2 py-1 rounded-md bg-blue-100 text-blue-700">
                        {member.role}
                      </span>
                      <span className={`text-xs px-2 py-1 rounded-md ${member.active ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-700"}`}>
                        {member.active ? t("staff.active") : t("staff.inactive")}
                      </span>
                      <button type="button" className="px-3 py-1 rounded-lg border text-sm" onClick={() => edit(member)}>
                        {t("staff.edit")}
                      </button>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </section>
        )}

        {activeTab === "accesos" && (
        <section className="bg-white rounded-xl shadow p-4 space-y-3">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold">{t("staff.firebaseUsersTitle")}</h2>
              <p className="text-sm text-gray-600">{t("staff.firebaseUsersNote")}</p>
            </div>
            <button type="button" className="px-3 py-2 rounded-lg border text-sm" onClick={refreshUsers}>
              {t("staff.refresh")}
            </button>
          </div>

          <div className="divide-y">
            {sortedUsers.length === 0 ? (
              <p className="text-sm text-gray-600 py-3">{t("staff.noUsers")}</p>
            ) : (
              sortedUsers.map((user) => (
                <div key={user.uid} className="py-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <div className="font-medium">{user.email}</div>
                    <div className="text-xs text-gray-500 font-mono">{user.uid}</div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`text-xs px-2 py-1 rounded-md ${user.role ? "bg-blue-100 text-blue-700" : "bg-gray-100 text-gray-700"}`}>
                      {user.role || t("staff.noRole")}
                    </span>
                    {user.disabled && (
                      <span className="text-xs px-2 py-1 rounded-md bg-red-100 text-red-700">
                        {t("staff.disabled")}
                      </span>
                    )}
                    <button
                      type="button"
                      className="px-3 py-1 rounded-lg border text-sm"
                      onClick={() => {
                        setRoleEmail(user.email);
                        setRole((user.role || "kinesiologo") as AuthRole);
                        setActiveTab("accesos");
                        window.scrollTo({ top: 0, behavior: "smooth" });
                      }}
                    >
                      {t("staff.inviteReset")}
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </section>
        )}
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
  error,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
  required?: boolean;
  error?: string;
}) {
  return (
    <div>
      <label className="text-sm font-medium"><RequiredLabel required={required}>{label}</RequiredLabel></label>
      <input
        className="mt-1 w-full border rounded-lg p-2"
        type={type}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        required={required}
      />
      {error && <p className="text-xs text-red-600 mt-1">{error}</p>}
    </div>
  );
}

function workDaysLabel(days: number[] | undefined, t: Translate) {
  const selected = days?.length ? days : defaultWorkDays;
  const options = getWorkDayOptions(t);
  return selected
    .map((value) => options.find((day) => day.value === value)?.label)
    .filter(Boolean)
    .join(", ");
}

function professionalFileCategory(file: File) {
  if (file.type.startsWith("image/")) return "foto";
  if (file.type.startsWith("video/")) return "video";
  return "documento";
}
