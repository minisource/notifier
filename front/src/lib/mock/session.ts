export type MockRole = 'admin' | 'operator' | 'viewer';

export interface MockSession {
  user: {
    id: string;
    email: string;
    name: string;
    role: MockRole;
  };
  tenant: {
    id: string;
    name: string;
    slug: string;
  };
}

export const mockSession: MockSession = {
  user: {
    id: 'user-mock-001',
    email: 'admin@notifier.local',
    name: 'Admin User',
    role: 'admin',
  },
  tenant: {
    id: 'tenant-default',
    name: 'Default Project',
    slug: 'default',
  },
};

export function getMockSession(): MockSession {
  return mockSession;
}

export function isMockMode(): boolean {
  return process.env.NEXT_PUBLIC_API_MODE === 'mock';
}
