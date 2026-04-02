import { getToken, removeToken } from '../auth/token';

const BASE_URL = (import.meta.env.VITE_API_BASE_URL as string) || '';

export const fetcher = async <T = any>(
  url: string,
  options: RequestInit = {},
): Promise<T> => {
  const token = getToken();
  const authHeaders: Record<string, string> = token
    ? { Authorization: `Bearer ${token}` }
    : {};

  const response = await fetch(`${BASE_URL}${url}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders,
      ...options.headers,
    },
  });

  if (!response.ok) {
    if (response.status === 401) {
      removeToken();
      window.location.href = '/login';
      return undefined as unknown as T;
    }
    const data = await response.json().catch(() => ({}));
    const error: Error & { response?: Response; data?: any } = new Error(
      (data as { msg?: string }).msg || `HTTP ${response.status}`,
    );
    error.response = response;
    error.data = data;
    throw error;
  }

  const contentType = response.headers.get('content-type') ?? '';
  const contentLength = response.headers.get('content-length');
  if (
    response.status === 204 ||
    contentLength === '0' ||
    !contentType.includes('application/json')
  ) {
    return undefined as unknown as T;
  }
  return response.json();
};
