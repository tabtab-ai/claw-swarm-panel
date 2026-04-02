import { ClawInstance } from '../../app/types/claw';

export interface ClawState {
  instances: ClawInstance[];
  loading: boolean;
  error: string | null;
  occupiedFilter: boolean | undefined;  // undefined = all, true = allocated, false = free
}

export const initialState: ClawState = {
  instances: [],
  loading: false,
  error: null,
  occupiedFilter: undefined,
};
