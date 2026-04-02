import { create } from 'zustand'
import { AuthUser } from '../../app/types/user'
import { authService } from '../../services/auth'
import { getToken, setToken, removeToken } from '../../utils/auth/token'

interface AuthStore {
  user: AuthUser | null
  isAuthenticated: boolean
  init: () => Promise<void>
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  changePassword: (oldPwd: string, newPwd: string) => Promise<void>
}

export const useAuthStore = create<AuthStore>((set) => ({
  user: null,
  isAuthenticated: false,

  init: async () => {
    const token = getToken()
    if (!token) return
    try {
      const user = await authService.me()
      set({ user, isAuthenticated: true })
    } catch {
      removeToken()
      set({ user: null, isAuthenticated: false })
    }
  },

  login: async (username, password) => {
    const res = await authService.login(username, password)
    setToken(res.token)
    set({
      user: {
        id: 0, // will be filled by me() but login response has enough info
        username: res.username,
        role: res.role as 'admin' | 'user',
        force_change_password: res.force_change_password,
      },
      isAuthenticated: true,
    })
  },

  logout: () => {
    removeToken()
    set({ user: null, isAuthenticated: false })
  },

  changePassword: async (oldPwd, newPwd) => {
    const res = await authService.changePassword(oldPwd, newPwd)
    setToken(res.token)
    set((state) => ({
      user: state.user ? { ...state.user, force_change_password: false } : null,
    }))
  },
}))
