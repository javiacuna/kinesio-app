import { useAuth } from "@/features/auth/AuthProvider";

export default function PatientPortalPage() {
  const { user } = useAuth();

  return (
    <main className="max-w-4xl mx-auto p-6 space-y-6">
      <header>
        <h1 className="text-2xl font-semibold">Portal paciente</h1>
        <p className="text-sm text-gray-600">{user?.email}</p>
      </header>

      <section className="bg-white border rounded-lg p-4">
        <h2 className="text-lg font-semibold">Seguimiento</h2>
        <p className="text-sm text-gray-600 mt-2">
          Tus turnos, planes y evolución van a estar disponibles acá.
        </p>
      </section>
    </main>
  );
}
