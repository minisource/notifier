import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { StatusBadge } from '@/components/shared/status-badge';

describe('StatusBadge', () => {
  it('renders sent status', () => {
    render(<StatusBadge status="sent" />);
    const badge = screen.getByText('sent');
    expect(badge).toBeDefined();
  });

  it('renders failed status', () => {
    render(<StatusBadge status="failed" />);
    const badge = screen.getByText('failed');
    expect(badge).toBeDefined();
  });

  it('renders pending status', () => {
    render(<StatusBadge status="pending" />);
    const badge = screen.getByText('pending');
    expect(badge).toBeDefined();
  });

  it('renders delivered status', () => {
    render(<StatusBadge status="delivered" />);
    const badge = screen.getByText('delivered');
    expect(badge).toBeDefined();
  });

  it('renders sm size class', () => {
    render(<StatusBadge status="sent" size="sm" />);
    const badge = screen.getByText('sent');
    expect(badge.className).toContain('text-xs');
  });
});
