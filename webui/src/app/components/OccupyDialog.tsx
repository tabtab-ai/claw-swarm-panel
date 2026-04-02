import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "./ui/dialog";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { ClawInstance } from "../types/claw";
import { Hash } from "lucide-react";

interface OccupyDialogProps {
  instance: ClawInstance | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: (userId: string) => void;
}

export function OccupyDialog({
  instance,
  open,
  onOpenChange,
  onConfirm,
}: OccupyDialogProps) {
  const [userId, setUserId] = useState("");

  const handleConfirm = () => {
    if (userId.trim()) {
      onConfirm(userId.trim());
      setUserId("");
    }
  };

  const handleClose = () => {
    setUserId("");
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>占用实例</DialogTitle>
          <DialogDescription>
            为实例{" "}
            <span className="font-medium text-white">{instance?.name}</span>{" "}
            指定 User ID
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="user-id-input" className="flex items-center gap-2">
              <Hash className="h-4 w-4" />
              User ID
            </Label>
            <Input
              id="user-id-input"
              placeholder="e.g., user-abc123..."
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleConfirm()}
              autoFocus
            />
          </div>

          {instance && (
            <div className="rounded-lg border border-slate-800 bg-slate-900/50 p-4">
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-slate-400">实例 ID</span>
                  <span className="font-medium text-white">{instance.id}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">Namespace</span>
                  <span className="font-medium text-white">
                    {instance.namespace}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-slate-400">Web UI</span>
                  <a
                    href={instance.webUIUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-cyan-400 hover:text-cyan-300"
                  >
                    {instance.webUIUrl}
                  </a>
                </div>
              </div>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
            取消
          </Button>
          <Button
            onClick={handleConfirm}
            disabled={!userId.trim()}
            className="bg-blue-600 hover:bg-blue-700"
          >
            确认占用
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
