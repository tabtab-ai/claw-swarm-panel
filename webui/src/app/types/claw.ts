export type ClawStatus = "active" | "idle" | "error" | "stopped" | "occupied";
export type ClawType = "worker" | "sentinel" | "harvester" | "scout";

export interface ClawResources {
  cpuRequest: string;
  cpuLimit: string;
  memoryRequest: string;
  memoryLimit: string;
}

export interface ClawInstance {
  id: string;
  name: string;
  namespace: string;
  conversation: string;  // conversation_id assigned to this instance
  deployed: boolean;
  allocStatus: 'allocating' | 'allocated' | 'idle';
  port: number;
  webUIUrl: string;
  status: ClawStatus;
  resources: ClawResources;
  // UI display fields (not from backend, kept for design compatibility)
  type: ClawType;
  cpu: number;
  memory: number;
  uptime: number;
  region: string;
  createdAt: Date;
  lastActivity: Date;
  isShared: boolean;
  occupiedBy?: string;
}
