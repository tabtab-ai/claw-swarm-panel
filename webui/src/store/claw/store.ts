import { StateCreator } from 'zustand';
import { subscribeWithSelector } from 'zustand/middleware';
import { shallow } from 'zustand/shallow';
import { createWithEqualityFn } from 'zustand/traditional';

import { createDevtools } from '../middleware/createDevtools';
import { ClawState, initialState } from './initialState';
import { InstanceAction, instanceActions } from './slices/instances/action';

export interface ClawStoreAction extends InstanceAction {}

export type ClawStore = ClawStoreAction & ClawState;

const createStore: StateCreator<ClawStore, [['zustand/devtools', never]]> = (...params) => ({
  ...initialState,
  ...instanceActions(...params),
});

const devtools = createDevtools('claw');

export const useClawStore = createWithEqualityFn<ClawStore>()(
  subscribeWithSelector(devtools(createStore)),
  shallow,
);
