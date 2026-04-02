import { useEffect, useState } from 'react'
import { Trash2, UserPlus } from 'lucide-react'
import { AuthUser } from '../types/user'
import { authService } from '../../services/auth'
import { useAuthStore } from '../../store/auth/store'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from './ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './ui/select'
import { Badge } from './ui/badge'

export function UserManagement() {
  const [users, setUsers] = useState<AuthUser[]>([])
  const [loading, setLoading] = useState(false)
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState({ username: '', password: '', role: 'user' as 'admin' | 'user' })
  const [formError, setFormError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const currentUser = useAuthStore((s) => s.user)

  const load = async () => {
    setLoading(true)
    try {
      const data = await authService.listUsers()
      setUsers(data)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormError('')
    if (form.password.length < 6) {
      setFormError('Password must be at least 6 characters')
      return
    }
    setSubmitting(true)
    try {
      await authService.createUser(form)
      setOpen(false)
      setForm({ username: '', password: '', role: 'user' })
      await load()
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to create user')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async (user: AuthUser) => {
    if (!confirm(`Delete user "${user.username}"?`)) return
    try {
      await authService.deleteUser(user.id)
      await load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete user')
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-white">User Management</h1>
        <Button className="gap-2 bg-cyan-600 hover:bg-cyan-700" onClick={() => setOpen(true)}>
          <UserPlus className="h-4 w-4" />
          New User
        </Button>
      </div>

      <div className="rounded-lg border border-slate-800 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-900 text-slate-400">
            <tr>
              <th className="px-4 py-3 text-left">ID</th>
              <th className="px-4 py-3 text-left">Username</th>
              <th className="px-4 py-3 text-left">Role</th>
              <th className="px-4 py-3 text-left">Force Change Pwd</th>
              <th className="px-4 py-3 text-left">Created At</th>
              <th className="px-4 py-3 text-left">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            {loading && (
              <tr>
                <td colSpan={6} className="px-4 py-6 text-center text-slate-500">Loading…</td>
              </tr>
            )}
            {!loading && users.map((u) => (
              <tr key={u.id} className="bg-slate-950 hover:bg-slate-900/50">
                <td className="px-4 py-3 text-slate-400">{u.id}</td>
                <td className="px-4 py-3 text-white font-medium">{u.username}</td>
                <td className="px-4 py-3">
                  {u.role === 'admin'
                    ? <Badge className="bg-cyan-600/20 text-cyan-400 border-cyan-600/30">admin</Badge>
                    : <Badge className="bg-slate-700/50 text-slate-300 border-slate-600">user</Badge>
                  }
                </td>
                <td className="px-4 py-3">
                  {u.force_change_password
                    ? <span className="text-yellow-400 text-xs">Yes</span>
                    : <span className="text-slate-500 text-xs">No</span>
                  }
                </td>
                <td className="px-4 py-3 text-slate-400 text-xs">
                  {u.created_at ? new Date(u.created_at).toLocaleString() : '—'}
                </td>
                <td className="px-4 py-3">
                  <Button
                    size="sm"
                    variant="ghost"
                    className="text-red-400 hover:text-red-300 hover:bg-red-900/20"
                    disabled={u.id === currentUser?.id}
                    onClick={() => handleDelete(u)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="bg-slate-900 border-slate-800 text-white">
          <DialogHeader>
            <DialogTitle>Create New User</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="space-y-1.5">
              <Label className="text-slate-300">Username</Label>
              <Input
                className="bg-slate-800 border-slate-700 text-white"
                value={form.username}
                onChange={(e) => setForm((f) => ({ ...f, username: e.target.value }))}
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label className="text-slate-300">Password</Label>
              <Input
                type="password"
                className="bg-slate-800 border-slate-700 text-white"
                value={form.password}
                onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
                required
              />
            </div>
            <div className="space-y-1.5">
              <Label className="text-slate-300">Role</Label>
              <Select value={form.role} onValueChange={(v) => setForm((f) => ({ ...f, role: v as 'admin' | 'user' }))}>
                <SelectTrigger className="bg-slate-800 border-slate-700 text-white">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="bg-slate-800 border-slate-700">
                  <SelectItem value="user">user</SelectItem>
                  <SelectItem value="admin">admin</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {formError && <p className="text-red-400 text-sm">{formError}</p>}
            <DialogFooter>
              <Button type="button" variant="outline" className="border-slate-700 text-slate-300" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" className="bg-cyan-600 hover:bg-cyan-700" disabled={submitting}>
                {submitting ? 'Creating…' : 'Create'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
