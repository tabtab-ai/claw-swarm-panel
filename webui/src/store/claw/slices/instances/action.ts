import { StateCreator } from 'zustand';

import { ClawInstance, ClawResources, ClawStatus } from '../../../../app/types/claw';
import { AllocResult, BackendInstance, clawService } from '../../../../services/claw';
import { ClawStore } from '../../store';

function mapState(state: string): ClawStatus {
  switch (state) {
    case 'running':
      return 'active';
    case 'paused':
      return 'stopped';
    default:
      return 'idle';
  }
}

function mapResources(b: BackendInstance): ClawResources {
  return {
    cpuRequest: b.resources?.cpu_request ?? '',
    cpuLimit: b.resources?.cpu_limit ?? '',
    memoryRequest: b.resources?.memory_request ?? '',
    memoryLimit: b.resources?.memory_limit ?? '',
  };
}

function mapInstance(b: BackendInstance): ClawInstance {
  return {
    id: b.name,
    name: b.name,
    namespace: '',
    conversation: b.user_id,
    deployed: b.occupied,
    allocStatus: b.alloc_status ?? '',
    port: 0,
    webUIUrl: b.claw_webui_url,
    status: mapState(b.state),
    resources: mapResources(b),
    type: 'worker',
    cpu: 0,
    memory: 0,
    uptime: 0,
    region: '',
    createdAt: b.created_at ? new Date(b.created_at) : new Date(),
    lastActivity: new Date(),
    isShared: false,
  };
}

export interface InstanceAction {
  refresh: () => Promise<void>;
  setOccupiedFilter: (filter: boolean | undefined) => Promise<void>;
  createInstance: (userId: string) => Promise<AllocResult>;
  deleteInstance: (id: string) => Promise<void>;
  updateInstanceStatus: (id: string, status: ClawStatus) => Promise<void>;
  occupyInstance: (id: string, userId: string) => Promise<void>;
  releaseInstance: (id: string) => Promise<void>;
}

export const instanceActions: StateCreator<
  ClawStore,
  [['zustand/devtools', never]],
  [],
  InstanceAction
> = (set, get) => ({
  refresh: async () => {
    set({ loading: true, error: null });
    try {
      const data = await clawService.listInstances();
      set({ instances: data.map(mapInstance), loading: false });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to fetch instances',
        loading: false,
      });
    }
  },

  setOccupiedFilter: async (filter: boolean | undefined) => {
    set({ occupiedFilter: filter });
  },

  createInstance: async (userId: string) => {
    const result = await clawService.allocInstance(userId);
    await get().refresh();
    return result;
  },

  deleteInstance: async (id: string) => {
    const instance = get().instances.find((i) => i.id === id);
    if (!instance) throw new Error('Instance not found');
    await clawService.freeInstance(instance.name, instance.conversation || undefined);
    await get().refresh();
  },

  updateInstanceStatus: async (id: string, status: ClawStatus) => {
    const instance = get().instances.find((i) => i.id === id);
    if (!instance) throw new Error('Instance not found');
    if (status === 'stopped') {
      await clawService.pauseInstance(instance.name, instance.conversation || undefined);
    } else if (status === 'active') {
      await clawService.resumeInstance(instance.name, instance.conversation || undefined);
    }
    await get().refresh();
  },

  occupyInstance: async (_id: string, userId: string) => {
    await clawService.allocInstance(userId);
    await get().refresh();
  },

  releaseInstance: async (id: string) => {
    const instance = get().instances.find((i) => i.id === id);
    if (!instance) throw new Error('Instance not found');
    await clawService.freeInstance(instance.name, instance.conversation || undefined);
    await get().refresh();
  },
});
