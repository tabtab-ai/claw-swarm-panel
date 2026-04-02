import { fetcher } from '../utils/fetch/fetcher'
import { AuthUser } from '../app/types/user'

export interface LoginResponse {
  token: string
  username: string
  role: string
  force_change_password: boolean
}

export interface ChangePasswordResponse {
  msg: string
  token: string
}

export interface CreateUserRequest {
  username: string
  password: string
  role: 'admin' | 'user'
}

export const authService = {
  login: (username: string, password: string) =>
    fetcher<LoginResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),

  me: () => fetcher<AuthUser>('/auth/me'),

  changePassword: (old_password: string, new_password: string) =>
    fetcher<ChangePasswordResponse>('/auth/change-password', {
      method: 'POST',
      body: JSON.stringify({ old_password, new_password }),
    }),

  listUsers: () =>
    fetcher<{ data: AuthUser[] }>('/users').then((res) => res.data ?? []),

  createUser: (req: CreateUserRequest) =>
    fetcher<AuthUser>('/users', {
      method: 'POST',
      body: JSON.stringify(req),
    }),

  deleteUser: (id: number) =>
    fetcher<void>(`/users/${id}`, { method: 'DELETE' }),

  getAPISecret: () =>
    fetcher<{ api_secret: string }>('/auth/api-secret'),

  regenerateAPISecret: () =>
    fetcher<{ api_secret: string }>('/auth/api-secret/regenerate', { method: 'POST' }),
}
