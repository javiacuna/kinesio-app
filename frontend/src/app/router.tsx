import { createBrowserRouter } from "react-router-dom";
import { AppShell } from "./AppShell";
import { ProtectedRoute } from "./ProtectedRoute";
import { RoleRedirect } from "./RoleRedirect";
import AgendaPage from "../pages/AgendaPage";
import LoginPage from "../pages/LoginPage";
import PatientPortalPage from "../pages/PatientPortalPage";
import PatientDetailPage from "../pages/PatientDetailPage";
import PatientsPage from "../pages/PatientsPage";
import StaffPage from "../pages/StaffPage";
import UnauthorizedPage from "../pages/UnauthorizedPage";

export const router = createBrowserRouter([
  { path: "/", element: <RoleRedirect /> },
  { path: "/login", element: <LoginPage /> },
  { path: "/unauthorized", element: <UnauthorizedPage /> },
  {
    element: <ProtectedRoute roles={["recepcionista", "kinesiologo"]} />,
    children: [
      {
        element: <AppShell />,
        children: [{ path: "/agenda", element: <AgendaPage /> }],
      },
    ],
  },
  {
    element: <ProtectedRoute roles={["paciente"]} />,
    children: [
      {
        element: <AppShell />,
        children: [{ path: "/portal", element: <PatientPortalPage /> }],
      },
    ],
  },
  {
    element: <ProtectedRoute roles={["recepcionista"]} />,
    children: [
      {
        element: <AppShell />,
        children: [{ path: "/patients", element: <PatientsPage /> }],
      },
    ],
  },
  {
    element: <ProtectedRoute roles={["recepcionista", "kinesiologo"]} />,
    children: [
      {
        element: <AppShell />,
        children: [{ path: "/patients/:patientId", element: <PatientDetailPage /> }],
      },
    ],
  },
  {
    element: <ProtectedRoute roles={["admin"]} />,
    children: [
      {
        element: <AppShell />,
        children: [{ path: "/staff", element: <StaffPage /> }],
      },
    ],
  },
]);
