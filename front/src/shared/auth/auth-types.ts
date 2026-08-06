export type UserRole = 'user' | 'admin' | 'operator' | 'service' | 'super_admin';
export type SessionSource = 'local' | 'real' | 'none';

export interface AuthSession {
  /** Source of the session — local (auth disabled), real auth, or none */
  source: SessionSource;
  accessToken: string | null;
  refreshToken?: string | null;
  userId: string | null;
  email: string | null;
  /** Display name (username / first+last name) — never the raw userId */
  name: string | null;
  tenantId: string | null;
  projectId: string | null;
  roles: UserRole[];
  permissions: string[];
  isAuthenticated: boolean;
  isSuperAdmin?: boolean;
}

export interface AuthAdapter {
  getSession(): AuthSession;
  getAccessToken(): string | null;
  getUserId(): string | null;
  getEmail(): string | null;
  getName(): string | null;
  getTenantId(): string | null;
  getProjectId(): string | null;
  getRoles(): UserRole[];
  getPermissions(): string[];
  isAuthenticated(): boolean;
  isSuperAdmin(): boolean;
  hasRole(role: UserRole): boolean;
  hasAnyRole(roles: UserRole[]): boolean;
  isAdminLike(): boolean;
  hasPermission(permission: string): boolean;
}
