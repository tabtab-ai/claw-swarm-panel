export interface AuthUser {
  id: number
  username: string
  role: 'admin' | 'user'
  force_change_password: boolean
  created_at?: string
}
