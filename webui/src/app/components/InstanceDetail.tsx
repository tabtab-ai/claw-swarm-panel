import {
  ArrowLeft,
  Copy,
  Cpu,
  Eye,
  EyeOff,
  ExternalLink,
  HardDrive,
  Hash,
  KeyRound,
  Loader2,
  Pause,
  Play,
  Terminal,
  Trash2,
} from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router";
import { toast } from "sonner";

import { clawService, BackendInstance } from "../../services/claw";
import { useClawStore } from "../../store/claw/store";
import { ClawStatus } from "../types/claw";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Card } from "./ui/card";
import { StatusIndicator } from "./StatusIndicator";

function stateToStatus(state: BackendInstance["state"]): ClawStatus {
  if (state === "running") return "active";
  if (state === "paused") return "stopped";
  return "idle";
}

export function InstanceDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const deleteInstance = useClawStore((s) => s.deleteInstance);
  const updateInstanceStatus = useClawStore((s) => s.updateInstanceStatus);

  const [detail, setDetail] = useState<BackendInstance | null>(null);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

  // Token: null = not yet revealed, "" = revealed but empty, string = revealed token
  const [gatewayToken, setGatewayToken] = useState<string | null>(null);
  const [tokenLoading, setTokenLoading] = useState(false);
  const [tokenVisible, setTokenVisible] = useState(false);

  const fetchDetail = () => {
    if (!id) return;
    clawService
      .getInstance(id)
      .then((d) => {
        setDetail(d);
        setNotFound(false);
      })
      .catch(() => setNotFound(true))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchDetail();
  }, [id]);

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard");
  };

  const handleRevealToken = async () => {
    if (!id) return;
    setTokenLoading(true);
    try {
      const token = await clawService.getInstanceWithToken(id);
      setGatewayToken(token);
      setTokenVisible(true);
    } catch {
      toast.error("Failed to fetch token");
      setGatewayToken("");
    } finally {
      setTokenLoading(false);
    }
  };

  const handlePause = async () => {
    try {
      await updateInstanceStatus(id!, "stopped");
      toast.success("Instance paused");
      fetchDetail();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to pause instance");
    }
  };

  const handleResume = async () => {
    try {
      await updateInstanceStatus(id!, "active");
      toast.success("Instance resumed");
      fetchDetail();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to resume instance");
    }
  };

  const handleDelete = async () => {
    try {
      await deleteInstance(id!);
      toast.success("Instance freed");
      navigate("/instances");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to free instance");
    }
  };

  if (loading) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" size="sm" className="gap-2" onClick={() => navigate(-1)}>
          <ArrowLeft className="h-4 w-4" />
          Back
        </Button>
        <Card className="border-slate-800 bg-slate-900/50 p-12 flex items-center justify-center gap-2 text-slate-400">
          <Loader2 className="h-4 w-4 animate-spin" />
          <span className="text-sm">Loading instance...</span>
        </Card>
      </div>
    );
  }

  if (notFound || !detail) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" size="sm" className="gap-2" onClick={() => navigate(-1)}>
          <ArrowLeft className="h-4 w-4" />
          Back
        </Button>
        <Card className="border-slate-800 bg-slate-900/50 p-12 text-center">
          <p className="text-slate-400">Instance not found.</p>
        </Card>
      </div>
    );
  }

  const status = stateToStatus(detail.state);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <Button variant="ghost" size="sm" className="mb-4 gap-2" onClick={() => navigate(-1)}>
          <ArrowLeft className="h-4 w-4" />
          Back
        </Button>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <StatusIndicator status={status} />
            <h2 className="text-2xl font-semibold text-white">{detail.name}</h2>
            {detail.occupied ? (
              <Badge className="bg-cyan-500/10 text-xs text-cyan-400 hover:bg-cyan-500/20">
                Allocated
              </Badge>
            ) : (
              <Badge className="bg-slate-500/10 text-xs text-slate-400 hover:bg-slate-500/20">
                Free
              </Badge>
            )}
          </div>
          <div className="flex gap-2">
            <Button
              size="sm"
              variant="outline"
              className="gap-2 border-slate-700 text-cyan-400 hover:text-cyan-300"
              disabled={status !== "active"}
              onClick={() => navigate(`/instances/${id}/terminal`)}
            >
              <Terminal className="h-4 w-4" />
              Terminal
            </Button>
            {status === "stopped" ? (
              <Button
                size="sm"
                variant="outline"
                className="gap-2 border-slate-700"
                onClick={handleResume}
              >
                <Play className="h-4 w-4" />
                Resume
              </Button>
            ) : (
              <Button
                size="sm"
                variant="outline"
                className="gap-2 border-slate-700"
                onClick={handlePause}
              >
                <Pause className="h-4 w-4" />
                Pause
              </Button>
            )}
            <Button
              size="sm"
              variant="outline"
              className="gap-2 border-red-900 text-red-400 hover:bg-red-950"
              onClick={handleDelete}
            >
              <Trash2 className="h-4 w-4" />
              Free Instance
            </Button>
          </div>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {/* Basic Info */}
        <Card className="border-slate-800 bg-slate-900/50 p-6 space-y-4">
          <h3 className="font-semibold text-white">Basic Info</h3>
          <div className="space-y-3 text-sm">
            <div className="flex justify-between">
              <span className="text-slate-400">Name</span>
              <span className="font-mono text-white">{detail.name}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-400">Status</span>
              <StatusIndicator status={status} showLabel />
            </div>
            <div className="flex justify-between">
              <span className="text-slate-400">Allocation</span>
              {detail.alloc_status === "allocating" ? (
                <Badge className="bg-yellow-500/10 text-xs text-yellow-400 hover:bg-yellow-500/20">
                  Allocating
                </Badge>
              ) : detail.alloc_status === "allocated" ? (
                <Badge className="bg-cyan-500/10 text-xs text-cyan-400 hover:bg-cyan-500/20">
                  Allocated
                </Badge>
              ) : (
                <Badge className="bg-slate-500/10 text-xs text-slate-400 hover:bg-slate-500/20">
                  Idle
                </Badge>
              )}
            </div>
          </div>
        </Card>

        {/* User Assignment */}
        <Card className="border-slate-800 bg-slate-900/50 p-6 space-y-4">
          <h3 className="font-semibold text-white">User Assignment</h3>
          {detail.user_id ? (
            <div className="flex items-center gap-2 rounded-lg border border-slate-800 bg-slate-900/80 px-3 py-2">
              <Hash className="h-4 w-4 shrink-0 text-slate-400" />
              <p className="flex-1 truncate font-mono text-sm text-slate-300">
                {detail.user_id}
              </p>
              <button
                onClick={() => copyToClipboard(detail.user_id)}
                className="text-slate-500 hover:text-slate-300"
              >
                <Copy className="h-3 w-3" />
              </button>
            </div>
          ) : (
            <p className="text-sm text-slate-500">Not assigned to any user.</p>
          )}
        </Card>

        {/* Resources */}
        <Card className="border-slate-800 bg-slate-900/50 p-6 space-y-4">
          <h3 className="font-semibold text-white">Resources</h3>
          <div className="space-y-3 text-sm">
            <div className="flex items-center gap-2 text-slate-400 mb-1">
              <Cpu className="h-4 w-4" />
              <span>CPU</span>
            </div>
            <div className="flex justify-between pl-6">
              <span className="text-slate-400">Request</span>
              <span className="font-mono text-white">{detail.resources?.cpu_request || "—"}</span>
            </div>
            <div className="flex justify-between pl-6">
              <span className="text-slate-400">Limit</span>
              <span className="font-mono text-white">{detail.resources?.cpu_limit || "—"}</span>
            </div>
            <div className="flex items-center gap-2 text-slate-400 mt-2 mb-1">
              <HardDrive className="h-4 w-4" />
              <span>Memory</span>
            </div>
            <div className="flex justify-between pl-6">
              <span className="text-slate-400">Request</span>
              <span className="font-mono text-white">{detail.resources?.memory_request || "—"}</span>
            </div>
            <div className="flex justify-between pl-6">
              <span className="text-slate-400">Limit</span>
              <span className="font-mono text-white">{detail.resources?.memory_limit || "—"}</span>
            </div>
          </div>
        </Card>

        {/* Claw Web UI */}
        <Card className="border-slate-800 bg-slate-900/50 p-6 space-y-4">
          <h3 className="font-semibold text-white">Claw Web UI</h3>
          <div className="flex items-center gap-3 rounded-lg border border-slate-800 bg-slate-900/80 px-4 py-3">
            <p className="flex-1 truncate font-mono text-sm text-slate-300">
              {detail.claw_webui_url}
            </p>
            <button
              onClick={() => copyToClipboard(detail.claw_webui_url)}
              className="text-slate-500 hover:text-slate-300"
            >
              <Copy className="h-4 w-4" />
            </button>
            <a
              href={detail.claw_webui_url}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1 text-sm text-cyan-400 hover:text-cyan-300"
            >
              <ExternalLink className="h-4 w-4" />
              Open
            </a>
          </div>
        </Card>

        {/* Gateway Token */}
        <Card className="border-slate-800 bg-slate-900/50 p-6 space-y-4 md:col-span-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <KeyRound className="h-4 w-4 text-amber-400" />
              <h3 className="font-semibold text-white">Gateway Token</h3>
            </div>
            {gatewayToken === null ? (
              <Button
                size="sm"
                variant="outline"
                className="gap-2 border-slate-700 text-slate-300"
                onClick={handleRevealToken}
                disabled={tokenLoading}
              >
                {tokenLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
                Reveal Token
              </Button>
            ) : (
              <button
                className="flex items-center gap-1 text-xs text-slate-500 hover:text-slate-300"
                onClick={() => setTokenVisible((v) => !v)}
              >
                {tokenVisible ? <EyeOff className="h-3 w-3" /> : <Eye className="h-3 w-3" />}
                {tokenVisible ? "Hide" : "Show"}
              </button>
            )}
          </div>

          {gatewayToken === null ? (
            <p className="text-sm text-slate-500">Click "Reveal Token" to load the gateway token.</p>
          ) : gatewayToken === "" ? (
            <p className="text-sm text-slate-500">暂无 Gateway Token</p>
          ) : (
            <div className="flex items-center gap-3 rounded-lg border border-slate-800 bg-slate-900/80 px-4 py-3">
              <p className="flex-1 break-all font-mono text-sm text-slate-300">
                {tokenVisible ? gatewayToken : "•".repeat(Math.min(gatewayToken.length, 48))}
              </p>
              <button
                onClick={() => copyToClipboard(gatewayToken)}
                className="shrink-0 text-slate-500 hover:text-slate-300"
              >
                <Copy className="h-4 w-4" />
              </button>
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
