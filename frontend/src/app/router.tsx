import { createBrowserRouter } from "react-router-dom";
import AgendaPage from "@/pages/AgendaPage";

export const router = createBrowserRouter([
  { path: "/", element: <AgendaPage /> },
]);
