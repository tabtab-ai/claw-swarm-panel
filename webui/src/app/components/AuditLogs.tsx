import { useEffect, useState } from 'react';
import { Shield, RefreshCw, ChevronLeft, ChevronRight } from 'lucide-react';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './ui/select';
import { auditService, AuditLog } from '../../services/audit';

const ACTION_LABELS: Record<string, string> = {
  alloc: '分配',
  free: '释放',
  pause: '暂停',
  resume: '恢复',
};

const ACTION_COLORS: Record<string, string> = {
  alloc: 'bg-cyan-600/20 text-cyan-400 border-cyan-600/30',
  free: 'bg-red-600/20 text-red-400 border-red-600/30',
  pause: 'bg-yellow-600/20 text-yellow-400 border-yellow-600/30',
  resume: 'bg-green-600/20 text-green-400 border-green-600/30',
};

const PAGE_SIZE = 20;

export function AuditLogs() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [actionFilter, setActionFilter] = useState('');
  const [loading, setLoading] = useState(false);

  const load = async (p: number, action: string) => {
    setLoading(true);
    try {
      const result = await auditService.listLogs({
        page: p,
        page_size: PAGE_SIZE,
        action: action || undefined,
      });
      setLogs(result.data ?? []);
      setTotal(result.total ?? 0);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(page, actionFilter);
  }, [page, actionFilter]);

  const handleActionChange = (val: string) => {
    setActionFilter(val === 'all' ? '' : val);
    setPage(1);
  };

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const formatTime = (ts: string) => {
    try {
      return new Date(ts).toLocaleString('zh-CN', { hour12: false });
    } catch {
      return ts;
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-cyan-500 to-blue-600">
            <Shield className="h-5 w-5 text-white" />
          </div>
          <div>
            <h1 className="text-xl font-semibold text-white">审计日志</h1>
            <p className="text-sm text-slate-400">共 {total} 条记录</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Select value={actionFilter || 'all'} onValueChange={handleActionChange}>
            <SelectTrigger className="w-32 bg-slate-800 border-slate-700 text-slate-200">
              <SelectValue placeholder="操作类型" />
            </SelectTrigger>
            <SelectContent className="bg-slate-900 border-slate-700 text-slate-200">
              <SelectItem value="all">全部操作</SelectItem>
              <SelectItem value="alloc">分配</SelectItem>
              <SelectItem value="free">释放</SelectItem>
              <SelectItem value="pause">暂停</SelectItem>
              <SelectItem value="resume">恢复</SelectItem>
            </SelectContent>
          </Select>
          <Button
            variant="outline"
            size="icon"
            className="border-slate-700 text-slate-300 hover:bg-slate-800"
            onClick={() => load(page, actionFilter)}
            disabled={loading}
          >
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          </Button>
        </div>
      </div>

      <div className="rounded-lg border border-slate-800 overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-slate-800 bg-slate-900/60">
              <th className="px-4 py-3 text-left text-slate-400 font-medium">时间</th>
              <th className="px-4 py-3 text-left text-slate-400 font-medium">操作者</th>
              <th className="px-4 py-3 text-left text-slate-400 font-medium">操作</th>
              <th className="px-4 py-3 text-left text-slate-400 font-medium">目标用户</th>
              <th className="px-4 py-3 text-left text-slate-400 font-medium">实例</th>
              <th className="px-4 py-3 text-left text-slate-400 font-medium">状态</th>
              <th className="px-4 py-3 text-left text-slate-400 font-medium">详情</th>
            </tr>
          </thead>
          <tbody>
            {loading && (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-slate-500">
                  加载中...
                </td>
              </tr>
            )}
            {!loading && logs.length === 0 && (
              <tr>
                <td colSpan={7} className="px-4 py-8 text-center text-slate-500">
                  暂无审计记录
                </td>
              </tr>
            )}
            {!loading && logs.map((log) => (
              <tr key={log.id} className="border-b border-slate-800/50 hover:bg-slate-800/30 transition-colors">
                <td className="px-4 py-3 text-slate-300 whitespace-nowrap font-mono text-xs">
                  {formatTime(log.timestamp)}
                </td>
                <td className="px-4 py-3 text-slate-200">{log.operator_username}</td>
                <td className="px-4 py-3">
                  <Badge className={ACTION_COLORS[log.action] ?? 'bg-slate-700 text-slate-300'}>
                    {ACTION_LABELS[log.action] ?? log.action}
                  </Badge>
                </td>
                <td className="px-4 py-3 text-slate-300 font-mono text-xs">{log.target_user_id || '—'}</td>
                <td className="px-4 py-3 text-slate-300 font-mono text-xs">{log.instance_name || '—'}</td>
                <td className="px-4 py-3">
                  <Badge
                    className={
                      log.status === 'success'
                        ? 'bg-green-600/20 text-green-400 border-green-600/30'
                        : 'bg-red-600/20 text-red-400 border-red-600/30'
                    }
                  >
                    {log.status === 'success' ? '成功' : '失败'}
                  </Badge>
                </td>
                <td className="px-4 py-3 text-slate-400 text-xs">{log.detail || '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between text-sm text-slate-400">
          <span>第 {page} / {totalPages} 页</span>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              className="border-slate-700 text-slate-300 hover:bg-slate-800"
              disabled={page <= 1}
              onClick={() => setPage((p) => p - 1)}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="border-slate-700 text-slate-300 hover:bg-slate-800"
              disabled={page >= totalPages}
              onClick={() => setPage((p) => p + 1)}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
