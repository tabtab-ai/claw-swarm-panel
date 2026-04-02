import { useState } from "react";
import {
  AlertCircle,
  Copy,
  ExternalLink,
  Hash,
  MoreVertical,
  Pause,
  Play,
  RefreshCw,
  Search,
  Trash2,
} from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { toast } from "sonner";
import { useNavigate } from "react-router";

import { useClawStore } from "../../store/claw/store";
import { ClawInstance, ClawStatus } from "../types/claw";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "./ui/alert-dialog";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Card } from "./ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "./ui/dropdown-menu";
import { Input } from "./ui/input";
import { OccupyDialog } from "./OccupyDialog";
import { StatusIndicator } from "./StatusIndicator";

export function InstanceList() {
  const navigate = useNavigate();
  const instances = useClawStore((s) => s.instances);
  const loading = useClawStore((s) => s.loading);
  const error = useClawStore((s) => s.error);
  const refresh = useClawStore((s) => s.refresh);
  const occupiedFilter = useClawStore((s) => s.occupiedFilter);
  const setOccupiedFilter = useClawStore((s) => s.setOccupiedFilter);
  const deleteInstance = useClawStore((s) => s.deleteInstance);
  const updateInstanceStatus = useClawStore((s) => s.updateInstanceStatus);
  const occupyInstance = useClawStore((s) => s.occupyInstance);
  const releaseInstance = useClawStore((s) => s.releaseInstance);

  const [searchQuery, setSearchQuery] = useState("");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [instanceToDelete, setInstanceToDelete] = useState<string | null>(null);
  const [occupyDialogOpen, setOccupyDialogOpen] = useState(false);
  const [instanceToOccupy, setInstanceToOccupy] =
    useState<ClawInstance | null>(null);

  const allCount = instances.length;
  const allocatedCount = instances.filter((i) => i.deployed).length;
  const freeCount = instances.filter((i) => !i.deployed).length;

  const filteredInstances = instances.filter((instance) => {
    if (occupiedFilter !== undefined && instance.deployed !== occupiedFilter) return false;
    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      return instance.name.toLowerCase().includes(q) || instance.conversation.toLowerCase().includes(q);
    }
    return true;
  });

  const handleDelete = (id: string) => {
    setInstanceToDelete(id);
    setDeleteDialogOpen(true);
  };

  const confirmDelete = async () => {
    if (instanceToDelete) {
      try {
        await deleteInstance(instanceToDelete);
        toast.success("Instance freed successfully");
      } catch (err) {
        toast.error(err instanceof Error ? err.message : "Failed to free instance");
      }
      setDeleteDialogOpen(false);
      setInstanceToDelete(null);
    }
  };

  const handleStatusChange = async (id: string, status: ClawStatus) => {
    try {
      await updateInstanceStatus(id, status);
      toast.success(
        status === "active" ? "Instance resumed" : "Instance paused"
      );
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to update instance");
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard");
  };

  const handleOccupy = (instance: ClawInstance) => {
    setInstanceToOccupy(instance);
    setOccupyDialogOpen(true);
  };

  const confirmOccupy = async (userId: string) => {
    if (instanceToOccupy) {
      try {
        await occupyInstance(instanceToOccupy.id, userId);
        toast.success(`Instance allocated to user ${userId}`);
      } catch (err) {
        toast.error(err instanceof Error ? err.message : "Failed to allocate instance");
      }
      setOccupyDialogOpen(false);
      setInstanceToOccupy(null);
    }
  };

  const handleRelease = async (id: string) => {
    try {
      await releaseInstance(id);
      toast.success("实例已释放");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to release instance");
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold text-white">
            {occupiedFilter === true ? 'Allocated Instances' : occupiedFilter === false ? 'Free Instances' : 'All Instances'}
          </h2>
          <p className="text-slate-400">
            {filteredInstances.length} of {allCount} OpenClaw instances
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

      {/* Filter Tabs */}
      <div className="flex items-center gap-1 rounded-lg border border-slate-800 bg-slate-900/50 p-1 w-fit">
        {(
          [
            { label: 'All', value: undefined, count: allCount },
            { label: 'Allocated', value: true, count: allocatedCount },
            { label: 'Free', value: false, count: freeCount },
          ] as { label: string; value: boolean | undefined; count: number }[]
        ).map(({ label, value, count }) => (
          <button
            key={label}
            onClick={() => setOccupiedFilter(value)}
            className={`flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
              occupiedFilter === value
                ? 'bg-slate-700 text-white shadow-sm'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            {label}
            <span className={`rounded-full px-1.5 py-0.5 text-xs ${
              occupiedFilter === value
                ? 'bg-slate-600 text-slate-200'
                : 'bg-slate-800 text-slate-500'
            }`}>
              {count}
            </span>
          </button>
        ))}
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
        <Input
          placeholder="Search by name or user ID..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="border-slate-800 bg-slate-900/50 pl-10 text-white placeholder:text-slate-500"
        />
      </div>

      {/* Instances Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {filteredInstances.map((instance) => (
          <Card
            key={instance.id}
            className="border-slate-800 bg-slate-900/50 p-6"
          >
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <StatusIndicator status={instance.status} />
                  <h3
                    className="font-semibold text-white cursor-pointer hover:text-cyan-400 transition-colors"
                    onClick={() => navigate(`/instances/${instance.id}`)}
                  >
                    {instance.name}
                  </h3>
                </div>
              </div>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                    <MoreVertical className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  {instance.status === "active" || instance.status === "idle" ? (
                    <DropdownMenuItem
                      onClick={() => handleStatusChange(instance.id, "stopped")}
                    >
                      <Pause className="mr-2 h-4 w-4" />
                      Pause Instance
                    </DropdownMenuItem>
                  ) : (
                    <DropdownMenuItem
                      onClick={() => handleStatusChange(instance.id, "active")}
                    >
                      <Play className="mr-2 h-4 w-4" />
                      Resume Instance
                    </DropdownMenuItem>
                  )}
                  <DropdownMenuItem onClick={() => handleOccupy(instance)}>
                    <Hash className="mr-2 h-4 w-4" />
                    Allocate to User
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => handleRelease(instance.id)}>
                    <Trash2 className="mr-2 h-4 w-4" />
                    Free Instance
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    onClick={() => handleDelete(instance.id)}
                    className="text-red-400"
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    Delete Instance
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            <div className="mt-4 flex flex-wrap gap-2">
              <Badge variant="secondary" className="text-xs">
                {instance.namespace}
              </Badge>
              {instance.deployed ? (
                <Badge className="bg-cyan-500/10 text-xs text-cyan-400 hover:bg-cyan-500/20">
                  Allocated
                </Badge>
              ) : (
                <Badge className="bg-slate-500/10 text-xs text-slate-400 hover:bg-slate-500/20">
                  Free
                </Badge>
              )}
              {instance.status === "error" && (
                <Badge className="bg-red-500/10 text-xs text-red-400 hover:bg-red-500/20">
                  <AlertCircle className="mr-1 h-3 w-3" />
                  Error
                </Badge>
              )}
            </div>

            {/* User ID */}
            {instance.conversation && (
              <div className="mt-3 flex items-center gap-2 rounded-lg border border-slate-800 bg-slate-900/80 px-3 py-2">
                <Hash className="h-4 w-4 text-slate-400" />
                <p className="flex-1 truncate text-sm text-slate-300">
                  {instance.conversation}
                </p>
                <button
                  onClick={() => copyToClipboard(instance.conversation)}
                  className="text-slate-500 hover:text-slate-300"
                >
                  <Copy className="h-3 w-3" />
                </button>
              </div>
            )}

            <div className="mt-4 space-y-3 border-t border-slate-800 pt-4">
              <div className="flex items-center justify-between text-sm">
                <span className="text-slate-400">Claw Web UI</span>
                <div className="flex items-center gap-2">
                  <a
                    href={instance.webUIUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-1 text-cyan-400 hover:text-cyan-300"
                  >
                    <ExternalLink className="h-3 w-3" />
                    Open
                  </a>
                  <button
                    onClick={() => copyToClipboard(instance.webUIUrl)}
                    className="text-slate-400 hover:text-slate-300"
                  >
                    <Copy className="h-3 w-3" />
                  </button>
                </div>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-slate-400">Last Active</span>
                <span className="font-medium text-white">
                  {formatDistanceToNow(instance.lastActivity, {
                    addSuffix: true,
                  })}
                </span>
              </div>
            </div>
          </Card>
        ))}
      </div>

      {filteredInstances.length === 0 && !loading && (
        <Card className="border-slate-800 bg-slate-900/50 p-12 text-center">
          <p className="text-slate-400">
            {searchQuery
              ? "No instances found matching your search."
              : occupiedFilter === true
              ? "No allocated instances. Use 'Allocate Instance' to assign one."
              : occupiedFilter === false
              ? "No free instances available in the pool."
              : "No instances found. Use 'Allocate Instance' to assign an instance to a user."}
          </p>
        </Card>
      )}

      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you sure?</AlertDialogTitle>
            <AlertDialogDescription>
              This will free the instance and release its user assignment.
              The instance will be returned to the available pool.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              className="bg-red-600 hover:bg-red-700"
            >
              Free Instance
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <OccupyDialog
        instance={instanceToOccupy}
        open={occupyDialogOpen}
        onOpenChange={setOccupyDialogOpen}
        onConfirm={confirmOccupy}
      />
    </div>
  );
}
