import { createBrowserRouter } from "react-router";
import { Dashboard } from "./components/Dashboard";
import { InstanceList } from "./components/InstanceList";
import { InstanceDetail } from "./components/InstanceDetail";
import { PodTerminal } from "./components/PodTerminal";
import { CreateInstance } from "./components/CreateInstance";
import { UserManagement } from "./components/UserManagement";
import { Guide } from "./components/Guide";
import { AuditLogs } from "./components/AuditLogs";
import { Root } from "./components/Root";
import { Login } from "./components/Login";

export const router = createBrowserRouter([
  {
    path: "/login",
    Component: Login,
  },
  {
    path: "/",
    Component: Root,
    children: [
      { index: true, Component: Dashboard },
      { path: "instances", Component: InstanceList },
      { path: "instances/:id", Component: InstanceDetail },
      { path: "instances/:id/terminal", Component: PodTerminal },
      { path: "create", Component: CreateInstance },
      { path: "users", Component: UserManagement },
      { path: "audit", Component: AuditLogs },
      { path: "guide", Component: Guide },
    ],
  },
]);
