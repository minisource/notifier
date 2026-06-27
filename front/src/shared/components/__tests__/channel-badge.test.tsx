import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ChannelBadge } from '@/components/shared/channel-badge';

describe('ChannelBadge', () => {
  it('renders sms channel', () => {
    render(<ChannelBadge channel="sms" />);
    const badge = screen.getByText('sms');
    expect(badge).toBeDefined();
  });

  it('renders email channel', () => {
    render(<ChannelBadge channel="email" />);
    const badge = screen.getByText('email');
    expect(badge).toBeDefined();
  });

  it('renders push channel', () => {
    render(<ChannelBadge channel="push" />);
    const badge = screen.getByText('push');
    expect(badge).toBeDefined();
  });

  it('renders in_app channel', () => {
    render(<ChannelBadge channel="in_app" />);
    const badge = screen.getByText('in_app');
    expect(badge).toBeDefined();
  });

  it('renders sm size', () => {
    render(<ChannelBadge channel="email" size="sm" />);
    const badge = screen.getByText('email');
    expect(badge.className).toContain('text-xs');
  });
});
