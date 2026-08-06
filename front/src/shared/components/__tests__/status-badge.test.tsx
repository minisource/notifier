import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { StatusBadge } from '@/components/shared/status-badge';
import { WithIntl } from './test-utils';

describe('StatusBadge', () => {
  it('renders sent status with localized label', () => {
    render(
      <WithIntl>
        <StatusBadge status="sent" />
      </WithIntl>
    );
    // en.json statuses.sent = "Sent"
    const badge = screen.getByText('Sent');
    expect(badge).toBeDefined();
  });

  it('renders failed status with localized label', () => {
    render(
      <WithIntl>
        <StatusBadge status="failed" />
      </WithIntl>
    );
    const badge = screen.getByText('Failed');
    expect(badge).toBeDefined();
  });

  it('renders pending status with localized label', () => {
    render(
      <WithIntl>
        <StatusBadge status="pending" />
      </WithIntl>
    );
    const badge = screen.getByText('Pending');
    expect(badge).toBeDefined();
  });

  it('renders delivered status with localized label', () => {
    render(
      <WithIntl>
        <StatusBadge status="delivered" />
      </WithIntl>
    );
    const badge = screen.getByText('Delivered');
    expect(badge).toBeDefined();
  });

  it('renders retrying status with localized label', () => {
    render(
      <WithIntl>
        <StatusBadge status="retrying" />
      </WithIntl>
    );
    const badge = screen.getByText('Retrying');
    expect(badge).toBeDefined();
  });

  it('renders dead status with localized label', () => {
    render(
      <WithIntl>
        <StatusBadge status="dead" />
      </WithIntl>
    );
    const badge = screen.getByText('Dead Letter');
    expect(badge).toBeDefined();
  });

  it('renders sm size class', () => {
    render(
      <WithIntl>
        <StatusBadge status="sent" size="sm" />
      </WithIntl>
    );
    const badge = screen.getByText('Sent');
    expect(badge.className).toContain('text-xs');
  });
});
