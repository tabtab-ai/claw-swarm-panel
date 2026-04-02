import { Link, useLocation, useNavigate } from "react-router";
import { Activity, Grid3x3, Plus, Users, LogOut, KeyRound, ChevronDown, Copy, RefreshCw, Eye, EyeOff, BookOpen, Shield } from "lucide-react";
import { Button } from "./ui/button";
import { Badge } from "./ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "./ui/dropdown-menu";
import { useAuthStore } from "../../store/auth/store";
import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "./ui/dialog";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { authService } from "../../services/auth";

function ChangePasswordDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const changePassword = useAuthStore((s) => s.changePassword);
  const [oldPwd, setOldPwd] = useState("");
  const [newPwd, setNewPwd] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    if (newPwd.length < 6) {
      setError("New password must be at least 6 characters");
      return;
    }
    setLoading(true);
    try {
      await changePassword(oldPwd, newPwd);
      setOldPwd("");
      setNewPwd("");
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to change password");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="bg-slate-900 border-slate-800 text-white">
        <DialogHeader>
          <DialogTitle>Change Password</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <Label className="text-slate-300">Current Password</Label>
            <Input
              type="password"
              className="bg-slate-800 border-slate-700 text-white"
              value={oldPwd}
              onChange={(e) => setOldPwd(e.target.value)}
              required
            />
          </div>
          <div className="space-y-1.5">
            <Label className="text-slate-300">New Password</Label>
            <Input
              type="password"
              className="bg-slate-800 border-slate-700 text-white"
              value={newPwd}
              onChange={(e) => setNewPwd(e.target.value)}
              required
            />
          </div>
          {error && <p className="text-red-400 text-sm">{error}</p>}
          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              className="border-slate-700 text-slate-300"
              onClick={onClose}
            >
              Cancel
            </Button>
            <Button type="submit" className="bg-cyan-600 hover:bg-cyan-700" disabled={loading}>
              {loading ? "Updating…" : "Update Password"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function APISecretDialog({ open, onClose }: { open: boolean; onClose: () => void }) {
  const [secret, setSecret] = useState("");
  const [visible, setVisible] = useState(false);
  const [loading, setLoading] = useState(false);
  const [regenerating, setRegenerating] = useState(false);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) return;
    setError("");
    setVisible(false);
    setLoading(true);
    authService.getAPISecret()
      .then((res) => setSecret(res.api_secret))
      .catch(() => setError("Failed to load API secret"))
      .finally(() => setLoading(false));
  }, [open]);

  const handleRegenerate = async () => {
    setRegenerating(true);
    setError("");
    try {
      const res = await authService.regenerateAPISecret();
      setSecret(res.api_secret);
      setVisible(true);
    } catch {
      setError("Failed to regenerate API secret");
    } finally {
      setRegenerating(false);
    }
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(secret);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="bg-slate-900 border-slate-800 text-white">
        <DialogHeader>
          <DialogTitle>API Secret</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <p className="text-sm text-slate-400">
            Use this secret to authenticate API requests via the{" "}
            <code className="bg-slate-800 px-1 rounded text-cyan-400 text-xs">X-API-Key</code>{" "}
            header.
          </p>
          <div className="space-y-1.5">
            <Label className="text-slate-300">Your API Secret</Label>
            <div className="flex gap-2">
              <div className="relative flex-1">
                <Input
                  type={visible ? "text" : "password"}
                  readOnly
                  className="bg-slate-800 border-slate-700 text-white font-mono text-sm pr-10"
                  value={loading ? "Loading…" : secret}
                />
                <button
                  type="button"
                  className="absolute right-2 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white"
                  onClick={() => setVisible((v) => !v)}
                  tabIndex={-1}
                >
                  {visible ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
              <Button
                type="button"
                variant="outline"
                size="icon"
                className="border-slate-700 text-slate-300 hover:bg-slate-800 shrink-0"
                onClick={handleCopy}
                disabled={loading || !secret}
                title="Copy to clipboard"
              >
                <Copy className="h-4 w-4" />
              </Button>
            </div>
            {copied && <p className="text-cyan-400 text-xs">Copied to clipboard!</p>}
          </div>
          {error && <p className="text-red-400 text-sm">{error}</p>}
          <div className="flex justify-between items-center pt-1">
            <Button
              type="button"
              variant="outline"
              className="gap-2 border-slate-700 text-slate-300 hover:bg-slate-800"
              onClick={handleRegenerate}
              disabled={regenerating || loading}
            >
              <RefreshCw className={`h-4 w-4 ${regenerating ? "animate-spin" : ""}`} />
              {regenerating ? "Regenerating…" : "Regenerate"}
            </Button>
            <Button
              type="button"
              variant="outline"
              className="border-slate-700 text-slate-300"
              onClick={onClose}
            >
              Close
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

export function Navigation() {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();
  const [showChangePwd, setShowChangePwd] = useState(false);
  const [showAPISecret, setShowAPISecret] = useState(false);

  const isActive = (path: string) => {
    if (path === "/") return location.pathname === "/";
    return location.pathname.startsWith(path);
  };

  const handleLogout = () => {
    logout();
    navigate("/login", { replace: true });
  };

  return (
    <>
      <nav className="border-b border-slate-800 bg-slate-900/50 backdrop-blur-sm">
        <div className="container mx-auto px-4">
          <div className="flex h-16 items-center justify-between">
            <div className="flex items-center gap-8">
              <Link to="/" className="flex items-center gap-2">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-cyan-500 to-blue-600">
                  <Activity className="h-6 w-6 text-white" />
                </div>
                <div>
                  <h1 className="font-semibold text-white">Claw Swarm</h1>
                  <p className="text-xs text-slate-400">Operation Console</p>
                </div>
              </Link>

              <div className="flex gap-1">
                <Link to="/">
                  <Button
                    variant={isActive("/") ? "secondary" : "ghost"}
                    size="sm"
                    className={`gap-2 ${isActive("/") ? "" : "text-slate-300 hover:text-white"}`}
                  >
                    <Activity className="h-4 w-4" />
                    Dashboard
                  </Button>
                </Link>
                <Link to="/instances">
                  <Button
                    variant={isActive("/instances") ? "secondary" : "ghost"}
                    size="sm"
                    className={`gap-2 ${isActive("/instances") ? "" : "text-slate-300 hover:text-white"}`}
                  >
                    <Grid3x3 className="h-4 w-4" />
                    Instances
                  </Button>
                </Link>
                {user?.role === "admin" && (
                  <Link to="/users">
                    <Button
                      variant={isActive("/users") ? "secondary" : "ghost"}
                      size="sm"
                      className={`gap-2 ${isActive("/users") ? "" : "text-slate-300 hover:text-white"}`}
                    >
                      <Users className="h-4 w-4" />
                      Users
                    </Button>
                  </Link>
                )}
                {user?.role === "admin" && (
                  <Link to="/audit">
                    <Button
                      variant={isActive("/audit") ? "secondary" : "ghost"}
                      size="sm"
                      className={`gap-2 ${isActive("/audit") ? "" : "text-slate-300 hover:text-white"}`}
                    >
                      <Shield className="h-4 w-4" />
                      审计日志
                    </Button>
                  </Link>
                )}
                <Link to="/guide">
                  <Button
                    variant={isActive("/guide") ? "secondary" : "ghost"}
                    size="sm"
                    className={`gap-2 ${isActive("/guide") ? "" : "text-slate-300 hover:text-white"}`}
                  >
                    <BookOpen className="h-4 w-4" />
                    使用说明
                  </Button>
                </Link>
              </div>
            </div>

            <div className="flex items-center gap-3">
              <Link to="/create">
                <Button className="gap-2 bg-cyan-600 hover:bg-cyan-700">
                  <Plus className="h-4 w-4" />
                  Allocate Instance
                </Button>
              </Link>

              {user && (
                <>
                  <DropdownMenu>
                    <DropdownMenuTrigger className="flex items-center gap-2 rounded-md px-3 h-8 text-slate-300 hover:text-white hover:bg-accent/50 outline-none">
                      <span className="text-sm">{user.username}</span>
                      <Badge
                        className={
                          user.role === "admin"
                            ? "bg-cyan-600/20 text-cyan-400 border-cyan-600/30 text-xs"
                            : "bg-slate-700/50 text-slate-400 border-slate-600 text-xs"
                        }
                      >
                        {user.role}
                      </Badge>
                      <ChevronDown className="h-3 w-3" />
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="bg-slate-900 border-slate-800 text-slate-200">
                      <DropdownMenuItem
                        className="gap-2 cursor-pointer hover:bg-slate-800"
                        onClick={() => setShowChangePwd(true)}
                      >
                        <KeyRound className="h-4 w-4" />
                        Change Password
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        className="gap-2 cursor-pointer hover:bg-slate-800"
                        onClick={() => setShowAPISecret(true)}
                      >
                        <Copy className="h-4 w-4" />
                        API Secret
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="gap-2 text-red-400 hover:text-red-300 hover:bg-red-900/20"
                    onClick={handleLogout}
                  >
                    <LogOut className="h-4 w-4" />
                    Logout
                  </Button>
                </>
              )}
            </div>
          </div>
        </div>
      </nav>
      <ChangePasswordDialog open={showChangePwd} onClose={() => setShowChangePwd(false)} />
      <APISecretDialog open={showAPISecret} onClose={() => setShowAPISecret(false)} />
    </>
  );
}
