import { fetcher } from '../utils/fetch/fetcher';

export interface BackendResources {
  cpu_request: string;
  cpu_limit: string;
  memory_request: string;
  memory_limit: string;
}

export interface BackendInstance {
  name: string;
  user_id: string;  // user_id assigned to this instance (empty if unoccupied)
  claw_webui_url: string;
  occupied: boolean;
  state: 'running' | 'pending' | 'paused';
  alloc_status: 'allocating' | 'allocated' | 'idle';
  resources?: BackendResources;
  created_at: string;  // RFC3339 timestamp
  token?: string;      // only present when requested with ?token=true
}

export interface AllocResult {
  name: string;
  user_id: string;
  claw_webui_url: string;
  occupied: boolean;
  state: string;
  token?: string;
}

class ClawService {
  listInstances = async (occupied?: boolean): Promise<BackendInstance[]> => {
    const query = occupied === undefined ? '' : `?occupied=${occupied}`;
    const data = await fetcher<{ data: BackendInstance[] }>(`/claw/instances${query}`);
    return data.data ?? [];
  };

  allocInstance = async (conversationId: string): Promise<AllocResult> => {
    return fetcher('/claw/alloc', {
      method: 'POST',
      body: JSON.stringify({ user_id: conversationId }),
    });
  };

  freeInstance = async (name: string, conversationId?: string): Promise<void> => {
    await fetcher('/claw/free', {
      method: 'POST',
      body: JSON.stringify({ name, user_id: conversationId ?? '' }),
    });
  };

  pauseInstance = async (name: string, conversationId?: string): Promise<void> => {
    await fetcher('/claw/pause', {
      method: 'POST',
      body: JSON.stringify({ name, user_id: conversationId ?? '' }),
    });
  };

  resumeInstance = async (name: string, conversationId?: string): Promise<void> => {
    await fetcher('/claw/resume', {
      method: 'POST',
      body: JSON.stringify({ name, user_id: conversationId ?? '' }),
    });
  };

  getInstance = async (name: string): Promise<BackendInstance> => {
    const data = await fetcher<{ data: BackendInstance }>(`/claw/instances/${encodeURIComponent(name)}`);
    return data.data;
  };

  getInstanceWithToken = async (name: string): Promise<string> => {
    const data = await fetcher<{ data: BackendInstance }>(`/claw/instances/${encodeURIComponent(name)}?token=true`);
    return data.data?.token ?? '';
  };

  getInstanceToken = async (name: string): Promise<string> => {
    const data = await fetcher<{ token: string }>(`/claw/token?name=${encodeURIComponent(name)}`);
    return data.token ?? '';
  };
}

export const clawService = new ClawService();
