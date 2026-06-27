export type UserRole = 'user' | 'admin' | 'operator' | 'service' | 'super_admin';
export type SessionSource = 'mock' | 'real' | 'none';

export interface AuthSession {
  /** Source of the session — mock, real auth, or none */
  source: SessionSource;
  accessToken: string | null;
  userId: string | null;
  tenantId: string | null;
  projectId: string | null;
  roles: UserRole[];
  permissions: string[];
  isAuthenticated: boolean;
}

export interface AuthAdapter {
  getSession(): AuthSession;
  getAccessToken(): string | null;
  getUserId(): string | null;
  getTenantId(): string | null;
  getProjectId(): string | null;
  getRoles(): UserRole[];
  isAuthenticated(): boolean;
  hasRole(role: UserRole): boolean;
  hasAnyRole(roles: UserRole[]): boolean;
  isAdminLike(): boolean;
  hasPermission(permission: string): boolean;
}
