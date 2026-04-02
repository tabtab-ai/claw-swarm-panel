import { ClawStatus } from "../types/claw";

interface StatusIndicatorProps {
  status: ClawStatus;
  showLabel?: boolean;
}

export function StatusIndicator({ status, showLabel = false }: StatusIndicatorProps) {
  const colors = {
    active: "bg-green-500",
    idle: "bg-yellow-500",
    error: "bg-red-500",
    stopped: "bg-slate-500",
    occupied: "bg-blue-500",
  };

  const labels = {
    active: "Active",
    idle: "Idle",
    error: "Error",
    stopped: "Stopped",
    occupied: "Occupied",
  };

  return (
    <div className="flex items-center gap-2">
      <div className={`h-2 w-2 rounded-full ${colors[status]} animate-pulse`} />
      {showLabel && (
        <span className="text-sm text-slate-400">{labels[status]}</span>
      )}
    </div>
  );
}