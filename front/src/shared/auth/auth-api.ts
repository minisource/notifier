/**
 * Auth API client for Notifier frontend.
 *
 * Provides login, logout, refresh, and userinfo calls against the Auth service.
 * This is the real Auth service API integration used by the auth adapter.
 */

export interface AuthLoginRequest {
  email: string;
  password: string;
}

export interface AuthLoginResponse {
  accessToken: string;
  refreshToken: string;
  expiresAt: string;
  tokenType: string;
  user: AuthUserInfo;
}

export interface AuthUserInfo {
  id: string;
  email: string;
  username: string;
  firstName: string;
  lastName: string;
  phone?: string;
  avatar?: string;
  emailVerified: boolean;
  phoneVerified: boolean;
  roles: string[];
}

export interface AuthUserinfoResponse {
  sub: string;
  email: string;
  emailVerified: boolean;
  phone?: string;
  phoneVerified: boolean;
  name: string;
  givenName: string;
  familyName: string;
  picture?: string;
  roles: string[];
  permissions: string[];
  tenantId?: string;
  isSuperAdmin: boolean;
}

export interface AuthRefreshRequest {
  refreshToken: string;
}

export interface AuthErrorResponse {
  success: boolean;
  error: {
    code: string;
    message: string;
  };
}

function getAuthBaseUrl(): string {
  if (typeof process !== 'undefined' && process.env && process.env.NEXT_PUBLIC_AUTH_API_URL) {
    return process.env.NEXT_PUBLIC_AUTH_API_URL;
  }
  if (typeof window !== 'undefined') {
    return '/v1';
  }
  return 'http://minisource-auth-backend:9001/v1';
}

function getApiUrl(path: string): string {
  const base = getAuthBaseUrl();
  const cleanPath = path.startsWith('/') ? path : `/${path}`;
  return `${base}${cleanPath}`;
}

async function handleAuthResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    let body: AuthErrorResponse | undefined;
    try {
      body = (await response.json()) as AuthErrorResponse;
    } catch {
      // Body not JSON
    }
    const message = body?.error?.message || response.statusText || 'Authentication failed';
    const code = body?.error?.code || 'AUTH_ERROR';
    const err = new Error(message) as Error & { code: string; status: number };
    err.code = code;
    err.status = response.status;
    throw err;
  }

  const body = (await response.json()) as { success: boolean; data: T };

  // Unwrap standard response envelope
  if (body && typeof body === 'object' && 'success' in body && body.success === true && 'data' in body) {
    return body.data as T;
  }

  return body as unknown as T;
}

export const authApi = {
  /**
   * Login with email and password via Auth service
   */
  login: async (email: string, password: string): Promise<AuthLoginResponse> => {
    const response = await fetch(getApiUrl('/auth/login'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password } satisfies AuthLoginRequest),
    });
    return handleAuthResponse<AuthLoginResponse>(response);
  },

  /**
   * Refresh access token using refresh token
   */
  refresh: async (refreshToken: string): Promise<AuthLoginResponse> => {
    const response = await fetch(getApiUrl('/auth/refresh'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken } satisfies AuthRefreshRequest),
    });
    return handleAuthResponse<AuthLoginResponse>(response);
  },

  /**
   * Logout and revoke session
   */
  logout: async (accessToken: string): Promise<void> => {
    const response = await fetch(getApiUrl('/auth/logout'), {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${accessToken}`,
      },
      body: JSON.stringify({ revokeAll: false }),
    });
    if (!response.ok) {
      // Logout failure is non-critical
      console.warn('Logout API call failed:', response.status);
    }
  },

  /**
   * Get user info from Auth service
   */
  userinfo: async (accessToken: string): Promise<AuthUserinfoResponse> => {
    const response = await fetch(getApiUrl('/auth/userinfo'), {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    });
    return handleAuthResponse<AuthUserinfoResponse>(response);
  },
};
