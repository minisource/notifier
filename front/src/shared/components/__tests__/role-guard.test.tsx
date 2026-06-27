import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { RoleGuard } from '@/shared/components/role-guard';
import { WithIntl } from './test-utils';

// Mock the auth adapter
vi.mock('@/shared/auth/auth-adapter', () => ({
  authAdapter: {
    getSession: vi.fn(),
  },
}));

import { authAdapter } from '@/shared/auth/auth-adapter';

describe('RoleGuard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders children when user has admin role', async () => {
    (authAdapter.getSession as ReturnType<typeof vi.fn>).mockReturnValue({
      roles: ['admin'],
    });

    render(
      <WithIntl>
        <RoleGuard requiredRoles={['admin', 'super_admin']}>
          <div data-testid="protected-content">Protected</div>
        </RoleGuard>
      </WithIntl>
    );

    await waitFor(() => {
      expect(screen.getByTestId('protected-content')).toBeDefined();
    });
  });

  it('renders forbidden state when user lacks required role', async () => {
    (authAdapter.getSession as ReturnType<typeof vi.fn>).mockReturnValue({
      roles: ['user'],
    });

    render(
      <WithIntl>
        <RoleGuard requiredRoles={['admin', 'super_admin']}>
          <div data-testid="protected-content">Protected</div>
        </RoleGuard>
      </WithIntl>
    );

    await waitFor(() => {
      expect(screen.queryByTestId('protected-content')).toBeNull();
    });
  });

  it('renders fallback when provided and access denied', async () => {
    (authAdapter.getSession as ReturnType<typeof vi.fn>).mockReturnValue({
      roles: ['user'],
    });

    render(
      <WithIntl>
        <RoleGuard
          requiredRoles={['admin']}
          fallback={<div data-testid="custom-fallback">Custom</div>}
        >
          <div data-testid="protected-content">Protected</div>
        </RoleGuard>
      </WithIntl>
    );

    await waitFor(() => {
      expect(screen.getByTestId('custom-fallback')).toBeDefined();
    });
  });

  it('allows access with operator role', async () => {
    (authAdapter.getSession as ReturnType<typeof vi.fn>).mockReturnValue({
      roles: ['operator'],
    });

    render(
      <WithIntl>
        <RoleGuard requiredRoles={['admin', 'operator', 'super_admin']}>
          <div data-testid="protected-content">Protected</div>
        </RoleGuard>
      </WithIntl>
    );

    await waitFor(() => {
      expect(screen.getByTestId('protected-content')).toBeDefined();
    });
  });

  it('denies access with empty roles', async () => {
    (authAdapter.getSession as ReturnType<typeof vi.fn>).mockReturnValue({
      roles: [],
    });

    render(
      <WithIntl>
        <RoleGuard requiredRoles={['admin']}>
          <div data-testid="protected-content">Protected</div>
        </RoleGuard>
      </WithIntl>
    );

    await waitFor(() => {
      expect(screen.queryByTestId('protected-content')).toBeNull();
    });
  });
});
