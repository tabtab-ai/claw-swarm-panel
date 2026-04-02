import { useState } from "react";
import {
  ArrowLeft,
  ArrowRight,
  CheckCircle,
  Copy,
  ExternalLink,
  Hash,
  KeyRound,
} from "lucide-react";
import { useNavigate } from "react-router";
import { toast } from "sonner";

import { useClawStore } from "../../store/claw/store";
import { AllocResult } from "../../services/claw";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Card } from "./ui/card";
import { Input } from "./ui/input";
import { Label } from "./ui/label";

export function CreateInstance() {
  const navigate = useNavigate();
  const createInstance = useClawStore((s) => s.createInstance);
  const [userId, setUserId] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<AllocResult | null>(null);

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard");
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!userId.trim()) {
      toast.error("Please enter a user ID");
      return;
    }
    try {
      setSubmitting(true);
      const alloc = await createInstance(userId.trim());
      setResult(alloc);
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : "Failed to allocate instance"
      );
    } finally {
      setSubmitting(false);
    }
  };

  /* ── Result view ── */
  if (result) {
    return (
      <div className="mx-auto max-w-3xl space-y-6">
        <div>
          <Button
            variant="ghost"
            size="sm"
            className="mb-4 gap-2"
            onClick={() => navigate("/")}
          >
            <ArrowLeft className="h-4 w-4" />
            Back to Dashboard
          </Button>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-green-500/10">
              <CheckCircle className="h-5 w-5 text-green-500" />
            </div>
            <div>
              <h2 className="text-2xl font-semibold text-white">
                Instance Allocated
              </h2>
              <p className="text-slate-400">
                Successfully assigned to user{" "}
                <span className="text-slate-200">{result.user_id}</span>
              </p>
            </div>
          </div>
        </div>

        {/* Instance details */}
        <Card className="border-slate-800 bg-slate-900/50 p-6 space-y-5">
          <h3 className="font-semibold text-white">Instance Details</h3>

          <div className="space-y-4 text-sm">
            {/* Name */}
            <div className="flex items-center justify-between">
              <span className="text-slate-400">Name</span>
              <div className="flex items-center gap-2">
                <span className="font-mono text-white">{result.name}</span>
                <button
                  onClick={() => copyToClipboard(result.name)}
                  className="text-slate-500 hover:text-slate-300"
                >
                  <Copy className="h-3 w-3" />
                </button>
              </div>
            </div>

            {/* State */}
            <div className="flex items-center justify-between">
              <span className="text-slate-400">State</span>
              <Badge
                className={
                  result.state === "running"
                    ? "bg-green-500/10 text-green-400"
                    : "bg-yellow-500/10 text-yellow-400"
                }
              >
                {result.state}
              </Badge>
            </div>

            {/* User ID */}
            <div className="flex items-center justify-between">
              <span className="text-slate-400">User ID</span>
              <div className="flex items-center gap-2">
                <Hash className="h-3 w-3 text-slate-500" />
                <span className="font-mono text-slate-200 truncate max-w-[240px]">
                  {result.user_id}
                </span>
                <button
                  onClick={() => copyToClipboard(result.user_id)}
                  className="text-slate-500 hover:text-slate-300"
                >
                  <Copy className="h-3 w-3" />
                </button>
              </div>
            </div>
          </div>

          {/* Gateway Token */}
          {result.token && (
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <KeyRound className="h-4 w-4 text-amber-400" />
                <span className="text-sm text-slate-400">Gateway Token</span>
              </div>
              <div className="flex items-center gap-3 rounded-lg border border-slate-800 bg-slate-900/80 px-4 py-3">
                <p className="flex-1 break-all font-mono text-sm text-slate-300">
                  {result.token}
                </p>
                <button
                  onClick={() => copyToClipboard(result.token!)}
                  className="shrink-0 text-slate-500 hover:text-slate-300"
                >
                  <Copy className="h-4 w-4" />
                </button>
              </div>
            </div>
          )}

          {/* Claw Web UI URL */}
          <div className="space-y-2">
            <span className="text-sm text-slate-400">Claw Web UI</span>
            <div className="flex items-center gap-3 rounded-lg border border-slate-800 bg-slate-900/80 px-4 py-3">
              <p className="flex-1 truncate font-mono text-sm text-slate-300">
                {result.claw_webui_url}
              </p>
              <button
                onClick={() => copyToClipboard(result.claw_webui_url)}
                className="text-slate-500 hover:text-slate-300"
              >
                <Copy className="h-4 w-4" />
              </button>
              <a
                href={result.claw_webui_url}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-1 text-sm text-cyan-400 hover:text-cyan-300"
              >
                <ExternalLink className="h-4 w-4" />
                Open
              </a>
            </div>
          </div>
        </Card>

        {/* Actions */}
        <div className="flex gap-3">
          <Button
            variant="outline"
            className="flex-1 border-slate-700"
            onClick={() => navigate("/")}
          >
            Back to Dashboard
          </Button>
          <Button
            className="flex-1 bg-cyan-600 hover:bg-cyan-700 gap-2"
            onClick={() => navigate(`/instances/${result.name}`)}
          >
            View Instance Details
            <ArrowRight className="h-4 w-4" />
          </Button>
        </div>
      </div>
    );
  }

  /* ── Form view ── */
  return (
    <div className="mx-auto max-w-3xl space-y-6">
      {/* Header */}
      <div>
        <Button
          variant="ghost"
          size="sm"
          className="mb-4 gap-2"
          onClick={() => navigate(-1)}
        >
          <ArrowLeft className="h-4 w-4" />
          Back
        </Button>
        <h2 className="text-2xl font-semibold text-white">
          Allocate Instance
        </h2>
        <p className="text-slate-400">
          Assign an available OpenClaw instance to a user
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* User ID input */}
        <Card className="border-slate-800 bg-slate-900/50 p-6">
          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-cyan-500/10">
                <Hash className="h-5 w-5 text-cyan-400" />
              </div>
              <div>
                <Label htmlFor="user-id" className="text-white">
                  User ID
                </Label>
                <p className="text-sm text-slate-400">
                  A unique identifier for the user that will own this instance
                </p>
              </div>
            </div>
            <Input
              id="user-id"
              placeholder="e.g., user-abc123, alice..."
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              className="border-slate-800 bg-slate-900 text-white"
              autoFocus
            />
          </div>
        </Card>

        {/* Info */}
        <Card className="border-slate-800 bg-slate-900/50 p-6">
          <div className="space-y-2 text-sm text-slate-400">
            <p className="font-medium text-slate-300">How allocation works</p>
            <ul className="list-inside list-disc space-y-1">
              <li>
                The system finds the earliest created available instance in the
                pool
              </li>
              <li>The instance is assigned exclusively to the given user ID</li>
              <li>
                If the user already has an instance, the same one is returned
              </li>
              <li>If no instances are available, allocation will fail</li>
            </ul>
          </div>
        </Card>

        {/* Actions */}
        <div className="flex gap-3">
          <Button
            type="button"
            variant="outline"
            className="flex-1"
            onClick={() => navigate(-1)}
            disabled={submitting}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            className="flex-1 bg-cyan-600 hover:bg-cyan-700"
            disabled={submitting}
          >
            {submitting ? "Allocating..." : "Allocate Instance"}
          </Button>
        </div>
      </form>
    </div>
  );
}
