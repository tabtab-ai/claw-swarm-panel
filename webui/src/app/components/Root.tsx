import { useEffect, useState } from "react";
import { Outlet, useNavigate } from "react-router";

import { useClawStore } from "../../store/claw/store";
import { useAuthStore } from "../../store/auth/store";
import { Navigation } from "./Navigation";
import { Toaster } from "./ui/sonner";

function ClawInit() {
  const refresh = useClawStore((s) => s.refresh);

  useEffect(() => {
    refresh();
    const interval = setInterval(refresh, 10000);
    return () => clearInterval(interval);
  }, [refresh]);

  return null;
}

export function Root() {
  const { isAuthenticated, init } = useAuthStore();
  const navigate = useNavigate();
  const [initializing, setInitializing] = useState(true);

  useEffect(() => {
    init().finally(() => {
      setInitializing(false);
    });
  }, []);

  useEffect(() => {
    if (!initializing && !isAuthenticated) {
      navigate("/login", { replace: true });
    }
  }, [initializing, isAuthenticated, navigate]);

  if (initializing) {
    return (
      <div className="min-h-screen bg-slate-950 flex items-center justify-center">
        <div className="text-slate-400 text-sm">Loading…</div>
      </div>
    );
  }

  if (!isAuthenticated) return null;

  return (
    <div className="min-h-screen bg-slate-950">
      <ClawInit />
      <Navigation />
      <main className="container mx-auto px-4 py-6">
        <Outlet />
      </main>
      <Toaster />
    </div>
  );
}
