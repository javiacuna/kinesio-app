import { createBrowserRouter } from "react-router-dom";
import { AppShell } from "./AppShell";
import { ProtectedRoute } from "./ProtectedRoute";
import { RoleRedirect } from "./RoleRedirect";
import AgendaPage from "../pages/AgendaPage";
import DashboardPage from "../pages/DashboardPage";
import ForgotPasswordPage from "../pages/ForgotPasswordPage";
import FinancePage from "../pages/FinancePage";
import LoginPage from "../pages/LoginPage";
import PatientPortalPage from "../pages/PatientPortalPage";
import PatientDetailPage from "../pages/PatientDetailPage";
import PatientsPage from "../pages/PatientsPage";
import ProfilePage from "../pages/ProfilePage";
import ReportsPage from "../pages/ReportsPage";
import StaffPage from "../pages/StaffPage";
import MaterialsPage from "../pages/MaterialsPage";
import UnauthorizedPage from "../pages/UnauthorizedPage";
import { ErrorPage } from "./ErrorPage";

const errorElement = <ErrorPage />;

export const router = createBrowserRouter([
  { path: "/", element: <RoleRedirect />, errorElement },
  { path: "/login", element: <LoginPage />, errorElement },
  { path: "/forgot-password", element: <ForgotPasswordPage />, errorElement },
  { path: "/unauthorized", element: <UnauthorizedPage />, errorElement },
  {
    element: <ProtectedRoute />,
    errorElement,
    children: [
      {
        element: <AppShell />,
        children: [{ path: "/profile", element: <ProfilePage /> }],
      },
    ],
  },
  {
    element: <ProtectedRoute roles={["admin", "recepcionista"]} />,
    errorElement,
    children: [
      {
        element: <AppShell />,
        children: [
          { path: "/finance", element: <FinancePage /> },
          { path: "/reports", element: <ReportsPage /> },
        ],
      },
    ],
  },
  {
    element: <ProtectedRoute roles={["recepcionista", "kinesiologo"]} />,
    errorElement,
    children: [
      {
        element: <AppShell />,
        children: [
          { path: "/dashboard", element: <DashboardPage /> },
          { path: "/agenda", element: <AgendaPage /> },
        ],
      },
    ],
  },
  {
    element: <ProtectedRoute roles={["paciente"]} />,
    errorElement,
    children: [
      {
        element: <AppShell />,
        children: [{ path: "/portal", element: <PatientPortalPage /> }],
      },
    ],
  },
  {
    element: <ProtectedRoute roles={["recepcionista"]} />,
    errorElement,
    children: [
      {
        element: <AppShell />,
        children: [{ path: "/patients", element: <PatientsPage /> }],
      },
    ],
  },
  {
    element: <ProtectedRoute roles={["recepcionista", "kinesiologo"]} />,
    errorElement,
    children: [
      {
        element: <AppShell />,
        children: [{ path: "/patients/:patientId", element: <PatientDetailPage /> }],
      },
    ],
  },
  {
    element: <ProtectedRoute roles={["recepcionista", "kinesiologo"]} />,
    errorElement,
    children: [
      {
        element: <AppShell />,
        children: [{ path: "/materials", element: <MaterialsPage /> }],
      },
    ],
  },
  {
    element: <ProtectedRoute roles={["admin"]} />,
    errorElement,
    children: [
      {
        element: <AppShell />,
        children: [{ path: "/staff", element: <StaffPage /> }],
      },
    ],
  },
]);
