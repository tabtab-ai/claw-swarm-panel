import { fetcher } from '../utils/fetch/fetcher';

export interface AuditLog {
  id: number;
  timestamp: string;
  operator_id: number;
  operator_username: string;
  action: string;
  target_user_id: string;
  instance_name: string;
  status: string;
  detail: string;
}

export interface AuditLogsResult {
  data: AuditLog[];
  total: number;
  page: number;
  page_size: number;
}

class AuditService {
  listLogs = async (params: { page?: number; page_size?: number; action?: string }): Promise<AuditLogsResult> => {
    const q = new URLSearchParams();
    if (params.page) q.set('page', String(params.page));
    if (params.page_size) q.set('page_size', String(params.page_size));
    if (params.action) q.set('action', params.action);
    const query = q.toString() ? `?${q.toString()}` : '';
    return fetcher<AuditLogsResult>(`/audit/logs${query}`);
  };
}

export const auditService = new AuditService();
