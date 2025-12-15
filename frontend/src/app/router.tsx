import { createBrowserRouter } from "react-router-dom";
import AgendaPage from "../pages/AgendaPage";
import PatientsPage from "../pages/PatientsPage";

export const router = createBrowserRouter([
  { path: "/", element: <AgendaPage /> },
  { path: "/patients", element: <PatientsPage /> },
]);
