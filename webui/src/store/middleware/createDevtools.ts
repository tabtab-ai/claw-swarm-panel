import { devtools } from 'zustand/middleware';

export const createDevtools =
  (name: string) =>
  <T>(fn: any) =>
    devtools(fn, { name, enabled: import.meta.env.DEV });
