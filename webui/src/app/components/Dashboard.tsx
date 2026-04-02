import {
  Activity,
  AlertCircle,
  Clock,
  ExternalLink,
  Hash,
  RefreshCw,
  Server,
} from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { useState } from "react";
import { useNavigate } from "react-router";

import { useClawStore } from "../../store/claw/store";
import { ClawInstance } from "../types/claw";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Card } from "./ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./ui/select";
import { StatusIndicator } from "./StatusIndicator";

export function Dashboard() {
  const instances = useClawStore((s) => s.instances);
  const loading = useClawStore((s) => s.loading);
  const error = useClawStore((s) => s.error);
  const refresh = useClawStore((s) => s.refresh);
  const [statusFilter, setStatusFilter] = useState<"all" | "active" | "stopped" | "idle">("all");
  const [allocFilter, setAllocFilter] = useState<"all" | "allocated" | "free">("all");

  const stats = {
    total: instances.length,
    active: instances.filter((i) => i.status === "active").length,
    idle: instances.filter((i) => !i.deployed).length,
    error: instances.filter((i) => i.status === "error").length,
    stopped: instances.filter((i) => i.status === "stopped").length,
    deployed: instances.filter((i) => i.deployed).length,
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold text-white">Swarm Overview</h2>
          <p className="text-slate-400">
            Monitor and manage your OpenClaw deployment
          </p>
        </div>
        <Button
          variant="ghost"
          size="sm"
          className="gap-2"
          onClick={refresh}
          disabled={loading}
        >
          <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
          Refresh
        </Button>
      </div>

      {/* Error Banner */}
      {error && (
        <Card className="border-red-900/50 bg-red-950/20 p-4">
          <div className="flex items-center gap-3">
            <AlertCircle className="h-5 w-5 text-red-500" />
            <p className="text-sm text-red-400">{error}</p>
          </div>
        </Card>
      )}

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card className="border-slate-800 bg-slate-900/50 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-slate-400">Total Instances</p>
              <p className="mt-2 text-3xl font-semibold text-white">
                {stats.total}
              </p>
            </div>
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-blue-500/10">
              <Activity className="h-6 w-6 text-blue-500" />
            </div>
          </div>
          <div className="mt-4 flex gap-2 text-xs">
            <span className="text-green-400">{stats.active} active</span>
            <span className="text-slate-500">•</span>
            <span className="text-yellow-400">{stats.idle} free</span>
            <span className="text-slate-500">•</span>
            <span className="text-slate-400">{stats.stopped} stopped</span>
          </div>
        </Card>

        <Card className="border-slate-800 bg-slate-900/50 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-slate-400">Allocated Instances</p>
              <p className="mt-2 text-3xl font-semibold text-white">
                {stats.deployed}
              </p>
            </div>
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-green-500/10">
              <Server className="h-6 w-6 text-green-500" />
            </div>
          </div>
          <div className="mt-4 text-xs text-slate-400">
            {stats.active} actively running
          </div>
        </Card>
      </div>

      {/* Instances */}
      <Card className="border-slate-800 bg-slate-900/50">
        <div className="border-b border-slate-800 p-6">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h3 className="font-semibold text-white">Instances</h3>
              <p className="text-sm text-slate-400">
                All OpenClaw instances in the pool
              </p>
            </div>
            <div className="flex gap-2">
              <Select value={statusFilter} onValueChange={(v) => setStatusFilter(v as typeof statusFilter)}>
                <SelectTrigger className="w-36 border-slate-700 bg-slate-800 text-sm text-slate-200">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="border-slate-700 bg-slate-800 text-slate-200">
                  <SelectItem value="all">All Status</SelectItem>
                  <SelectItem value="active">Running</SelectItem>
                  <SelectItem value="idle">Pending</SelectItem>
                  <SelectItem value="stopped">Paused</SelectItem>
                </SelectContent>
              </Select>
              <Select value={allocFilter} onValueChange={(v) => setAllocFilter(v as typeof allocFilter)}>
                <SelectTrigger className="w-36 border-slate-700 bg-slate-800 text-sm text-slate-200">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="border-slate-700 bg-slate-800 text-slate-200">
                  <SelectItem value="all">All Types</SelectItem>
                  <SelectItem value="allocated">Allocated</SelectItem>
                  <SelectItem value="free">Free</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>
        <div className="divide-y divide-slate-800">
          {instances
            .filter((i) => {
              if (statusFilter !== "all" && i.status !== statusFilter) return false;
              if (allocFilter === "allocated") return i.deployed;
              if (allocFilter === "free") return !i.deployed;
              return true;
            })
            .sort((a, b) => a.createdAt.getTime() - b.createdAt.getTime())
            .slice(0, 5)
            .map((instance) => (
              <InstanceRow key={instance.id} instance={instance} />
            ))}
          {instances.filter((i) => {
            if (statusFilter !== "all" && i.status !== statusFilter) return false;
            if (allocFilter === "allocated") return i.deployed;
            if (allocFilter === "free") return !i.deployed;
            return true;
          }).length === 0 && (
            <div className="p-6 text-center text-sm text-slate-400">
              No instances
            </div>
          )}
        </div>
      </Card>

      {/* Alerts */}
      {stats.error > 0 && (
        <Card className="border-red-900/50 bg-red-950/20">
          <div className="flex items-start gap-4 p-6">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-red-500/10">
              <AlertCircle className="h-5 w-5 text-red-500" />
            </div>
            <div className="flex-1">
              <h3 className="font-semibold text-white">
                Instances Require Attention
              </h3>
              <p className="mt-1 text-sm text-slate-400">
                {stats.error} instance{stats.error > 1 ? "s are" : " is"}{" "}
                experiencing errors and may need intervention.
              </p>
            </div>
          </div>
        </Card>
      )}
    </div>
  );
}

function InstanceRow({ instance }: { instance: ClawInstance }) {
  const navigate = useNavigate();
  return (
    <div className="flex items-center gap-4 p-6">
      <StatusIndicator status={instance.status} />
      <div className="flex-1">
        <div className="flex items-center gap-2">
          <p
            className="font-medium text-white cursor-pointer hover:text-cyan-400 transition-colors"
            onClick={() => navigate(`/instances/${instance.id}`)}
          >
            {instance.name}
          </p>
          {instance.deployed && (
            <Badge className="bg-cyan-500/10 text-xs text-cyan-400 hover:bg-cyan-500/20">
              Allocated
            </Badge>
          )}
        </div>
        <div className="mt-1 flex items-center gap-2">
          <p className="text-sm text-slate-400">{instance.namespace}</p>
          <a
            href={instance.webUIUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1 text-sm text-cyan-400 hover:text-cyan-300"
          >
            <ExternalLink className="h-3 w-3" />
            Claw Web UI
          </a>
        </div>
        {instance.conversation && (
          <div className="mt-1 flex items-center gap-1 text-xs text-slate-500">
            <Hash className="h-3 w-3" />
            <span className="truncate max-w-[200px]">
              {instance.conversation}
            </span>
          </div>
        )}
      </div>
      <div className="flex items-center gap-2 text-sm text-slate-400">
        <Clock className="h-4 w-4" />
        {formatDistanceToNow(instance.lastActivity, { addSuffix: true })}
      </div>
    </div>
  );
}
